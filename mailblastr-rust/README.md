# mailblastr

Official Rust SDK for the [MailBlastr](https://www.mailblastr.com) email API — send transactional and marketing email from your own verified domain.

## Install

```bash
cargo add mailblastr
cargo add tokio -F macros,rt-multi-thread
```

## Usage

```rust
use mailblastr::{SendEmailOptions, Mailblastr, Result};

#[tokio::main]
async fn main() -> Result<()> {
    let mailblastr = Mailblastr::new("mb_xxxxxxxxx");

    let from = "Acme <hello@yourdomain.com>";
    let to = ["delivered@mailblastr.dev"]; // the simulator address — always delivers
    let subject = "Hello from MailBlastr";

    let email = SendEmailOptions::new(from, to, subject)
        .with_html("<p>Your first email 🎉</p>");

    let sent = mailblastr.emails.send(email).await?;
    println!("sent {}", sent.id);
    Ok(())
}
```

Every method is `async` and returns `Result<T, mailblastr::Error>`. API failures surface as `Error::Api(Box<ApiError>)`, parsed from the standard MailBlastr `{ statusCode, name, message }` error body; transport and decode problems are `Error::Http` / `Error::Json`.

Branch on `name`, never on `message` — messages are scrubbed of provider identifiers and are not a stable contract. A handler may also override the status a `name` normally maps to — a missing `User-Agent` is the one `validation_error` that is 403 rather than 422 — so read `status_code` off the error rather than a hard-coded table. The SDK sends the required `User-Agent` on every request for you.

```rust
use mailblastr::Error;

match mailblastr.emails.get("em_missing").await {
    Ok(email) => println!("{}", email.status),
    Err(Error::Api(api)) => {
        eprintln!("{} {}: {}", api.status_code, api.name, api.message);
    }
    Err(other) => eprintln!("transport error: {other}"),
}
```

#### Errors that carry more than the envelope

Some errors are a superset of `{ statusCode, name, message }`. `ApiError` models each extra as an `Option` (or an empty `Vec`) that is absent on ordinary errors, and keeps the whole parsed body in `api.body` for anything not modelled yet:

| Field | Present on | Tells you |
| --- | --- | --- |
| `limit` | `plan_limit_reached`, `daily_quota_exceeded`, `monthly_quota_exceeded`, `contact_limit_reached`, `ai_credits_exceeded`, `automation_quota_exceeded` | WHICH allowance ran out (`kind`), `used` / `limit` / `remaining`, the rolling `period`, the current `plan`, the cheapest `next_plan` that would fit (`None` when only Enterprise does), and prepaid `credits` for the email-quota kinds |
| `reputation` | `reputation_paused`, `reputation_limit_exceeded`, `sending_service_unavailable` | whether it is `retryable`, the `scope` (`tenant` / `domain` / `platform`), hourly and daily counters, and `retry_at` |
| `sent` / `sent_count` | a `POST /emails/batch` that failed part way through, sent with an `Idempotency-Key` | the emails that DID go out, so a retry does not send them twice. `sent_count` falls back to `sent.len()` when the server omits it |

```rust
match mailblastr.emails.send(email).await {
    Ok(sent) => println!("sent {}", sent.id),
    Err(err) => match err.api() {
        Some(api) => {
            if let Some(limit) = &api.limit {
                // e.g. "emails_daily cap hit: 100/100, upgrade to Pro"
                eprintln!("{} cap hit: {}/{}", limit.kind, limit.used, limit.limit);
            }
            if let Some(rep) = &api.reputation {
                eprintln!("{} sending gated, retryable={}", rep.scope, rep.retryable);
            }
            // On a partial batch failure, skip these on retry.
            for already in &api.sent {
                eprintln!("already delivered: {}", already.id);
            }
        }
        None => eprintln!("transport error: {err}"),
    },
}
```

### Attachments

Attach files by hosted URL (fetched at send time) or inline base64:

```rust
use mailblastr::{Attachment, SendEmailOptions};

let email = SendEmailOptions::new(from, to, "Your invoice")
    .with_html("<p>Invoice attached.</p>")
    .with_attachment(Attachment::from_path(
        "invoice.pdf",
        "https://yourdomain.com/invoices/invoice.pdf",
    ))
    .with_attachment(
        Attachment::from_content("report.csv", base64_content).with_content_type("text/csv"),
    );
```

Each attachment decodes to at most 25 MB, and all attachments on one message to at most 40 MB. Attachments are single-send only — the batch endpoint rejects them, along with `scheduled_at`.

Other caps worth knowing when you build a payload: `to` takes 1–50 recipients (`cc` and `bcc` up to 50 each), `from` is at most 320 characters, `preview_text` at most 150, `scheduled_at` at most 30 days ahead (ISO 8601 or a phrase like `"in 1 min"`), and a batch at most 100 emails. `subject` is uncapped. There is no `tags` field — sending one is a `422 validation_error`, not a silent drop.

### Options

```rust
// Override the API host (proxy, staging, tests):
let mailblastr = Mailblastr::with_base_url("mb_xxxxxxxxx", "https://www.mailblastr.com/api");

// Bring your own reqwest client (timeouts, proxies, pools):
let client = reqwest::Client::builder()
    .timeout(std::time::Duration::from_secs(30))
    .build()
    .unwrap();
let mailblastr = Mailblastr::with_client("mb_xxxxxxxxx", mailblastr::DEFAULT_BASE_URL, client);
```

## Domain-first parameters

MailBlastr is **domain-first**: each verified sending domain owns its own contact pool — the same address on two domains is two records with separate consent. That shows up in the SDK as required `domain` parameters:

| Surface | `domain` |
|---|---|
| `contacts` (flat `/contacts` API) | REQUIRED on create/list (email lookups take it to pick the pool); the nested audience API derives the pool from the path instead |
| `segments` | REQUIRED on create + list (names unique per domain; every domain has an auto-created "General" segment) |
| `topics` | REQUIRED on create + list |
| `campaigns.create` | REQUIRED (picks the contact pool; orthogonal to `from`) |
| `automations.create` | REQUIRED (only same-domain events trigger it) |
| `events.send` | REQUIRED (routes the event to that domain's automations and pool) |

## Resources

One client field per resource, each following a consistent `create` / `get` / `list` / `update` / `remove` shape plus resource-specific verbs. `api_keys` is the deliberate exception — it is read-only, see [API keys are dashboard-only](#api-keys-are-dashboard-only):

`emails` (with nested `emails.receiving`), `batch`, `domains`, `audiences`, `contacts`, `contact_properties`, `campaigns`, `segments`, `topics`, `templates`, `automations`, `webhooks`, `logs`, `events`, `api_keys`, `polls`.

```rust
use mailblastr::*;

// Emails
mailblastr.emails.send(email).await?;
mailblastr.emails.list(Some(PaginationParams::new().with_limit(20))).await?;
mailblastr.emails.list_filtered(Some(                         // server-side filters
    ListEmailsParams::new().with_status("bounced").with_search("acme"),
)).await?;
mailblastr.emails.sources().await?;                           // per-campaign/automation totals
mailblastr.emails.get(id).await?;
mailblastr.emails.list_attachments(id).await?;                // metadata only, no bytes
mailblastr.emails.update(id, "2026-08-01T09:00:00Z").await?;  // reschedule
mailblastr.emails.cancel(id).await?;

// Inbound email
mailblastr.emails.receiving.list(None).await?;
mailblastr.emails.receiving.list_addresses().await?;
mailblastr.emails.receiving.get(id).await?;
mailblastr.emails.receiving.list_attachments(id, None).await?;
let pdf: Vec<u8> = mailblastr.emails.receiving.get_attachment(id, att_id).await?;
mailblastr.emails.receiving.forward(id, ForwardReceivedEmailOptions::new(from, ["team@you.com"])).await?;
mailblastr.emails.receiving.reply(id, ReplyReceivedEmailOptions::new(from).with_html("<p>Thanks!</p>")).await?;

// Batch send — up to 100 per call. Items are `BatchEmailOptions`, which simply
// has no `attachments` / `scheduled_at` because the batch route rejects both.
// Above 40 emails the batch is QUEUED for the worker instead of sent inline:
// the response's `queued` flag is then true and the ids are emails that have
// not gone out yet.
let batch = mailblastr.batch.send_emails(vec![
    BatchEmailOptions::new(from, to, subject).with_html(html),
]).await?;
if batch.queued { /* not transmitted yet — poll emails.get(id) */ }

// Domains (incl. claiming a domain verified elsewhere + one-click DNS)
mailblastr.domains.create(CreateDomainOptions::new("example.com")).await?;
mailblastr.domains.verify(id).await?;
mailblastr.domains.claim(ClaimDomainOptions::new("example.com")).await?;
mailblastr.domains.verify_claim(id).await?;
mailblastr.domains.mx_check("example.com").await?;
let csv: Vec<u8> = mailblastr.domains.records_csv(id).await?;
mailblastr.domains.detect_dns(id).await?;
mailblastr.domains.apply_cloudflare_dns(id, "cf_token").await?;

// Contacts are DOMAIN-FIRST
mailblastr.contacts.create(
    CreateContactOptions::new("user@example.com")
        .with_domain("example.com")
        .with_first_name("Ada"),
).await?;
mailblastr.contacts.list(ListContactsParams::for_domain("example.com")).await?;
mailblastr.contacts.get(ContactLookup::new(contact_id)).await?;              // by id (exact) …
mailblastr.contacts.get(ContactLookup::new("user@example.com").with_domain("example.com")).await?; // … or by email + domain
mailblastr.contacts.update(UpdateContactOptions::new(contact_id).with_unsubscribed(true)).await?;
mailblastr.contacts.remove(ContactLookup::new(contact_id)).await?;
mailblastr.contacts.add_to_segment(contact_id, segment_id).await?;
mailblastr.contacts.list_segments(contact_id, None).await?;
mailblastr.contacts.batch(audience_id, contacts, Some(OnConflict::Skip)).await?;
mailblastr.contacts.import(audience_id, csv_text, ImportCsvOptions::new()).await?; // inline: ≤5 MB / 10k rows
// Bigger files: presign a direct upload, PUT the bytes yourself, then import by key.
let slot = mailblastr.contacts.create_import_upload(audience_id, "contacts.csv", size).await?;
mailblastr.contacts.import_from_storage_key(audience_id, &slot.storage_key, ImportCsvOptions::new()).await?;

// Contact properties (custom fields / merge tags)
mailblastr.contact_properties.create(
    CreateContactPropertyOptions::new("plan", ContactPropertyType::String),
).await?;

// Campaigns & Segments — also domain-first
mailblastr.campaigns.create(
    CreateCampaignOptions::new("example.com", from, subject)
        .with_html(html)
        .with_segment_id(segment_id),
).await?;
mailblastr.campaigns.send(id, Some("2026-08-01T09:00:00Z")).await?;
mailblastr.campaigns.stats(id).await?;
mailblastr.campaigns.engagement(id).await?; // who opened / clicked / replied (≤500 each)
mailblastr.segments.create(
    CreateSegmentOptions::new("yourdomain.com", "VIP")
        .with_filter(
            SegmentFilterOptions::new()
                .with_status(SegmentStatus::Subscribed)
                .with_engagement(SegmentEngagement::new(EngagementEvent::Clicked, campaign_id)),
        ),
).await?;
mailblastr.segments.list(ListSegmentsParams::new("example.com")).await?;
mailblastr.segments.contacts(id, None).await?; // preview who matches

// Templates
mailblastr.templates.create(CreateTemplateOptions::new("Welcome").with_subject("Hi {{first_name}}").with_html(html)).await?;
mailblastr.templates.duplicate(id, None).await?;
mailblastr.templates.publish(id).await?;
mailblastr.emails.send(
    SendEmailOptions::new(from, to, subject)
        .with_template_id(template_id)
        .with_variable("first_name", "Ada"),
).await?;

// API keys — read-only, see "API keys are dashboard-only" below. `token` is
// the 8-character display prefix, never the secret.
mailblastr.api_keys.list(None).await?;

// Polls (read-only in-email poll results)
mailblastr.polls.list(None).await?;
mailblastr.polls.get(email_id).await?;

// Logs
mailblastr.logs.list(Some(ListLogsParams::new().with_limit(100))).await?;
mailblastr.logs.get(log_id).await?;
```

### API keys are dashboard-only

`api_keys` is the one read-only service on the client: it has `list` and nothing else.

Keys are created, re-scoped and revoked **only from a signed-in session in the MailBlastr dashboard**. The API enforces it — `POST /api-keys`, `PATCH /api-keys/:id` and `DELETE /api-keys/:id` answer `403 dashboard_only` to any caller authenticating with an API key, whatever its permission. Every call this crate makes authenticates with a key, so there is deliberately no Rust method for those routes; `mailblastr.api_keys.create(..)` is a compile error, not a runtime 403.

That boundary is worth having: a key that leaks cannot mint itself a replacement, promote itself to `full_access`, add a domain to its own scope, or revoke the keys you would have used to shut it down. Containing a leaked key is a dashboard action taken by a human, and it stays that way.

`api_keys.list` still works with a key and is the SDK-side tool for it — each row carries the non-secret 8-character display prefix, the permission, the domain scoping and `last_used_at`, which is enough to audit what is live and spot a key being used when it should not be. Revoke it in the dashboard.

### Topics

Topics let contacts manage granular subscriptions (e.g. "Product updates"):

```rust
let topic = mailblastr.topics.create(
    CreateTopicOptions::new("example.com", "Product updates", SubscriptionState::OptIn)
        .with_description("New features and releases"),
).await?;

mailblastr.topics.list(ListTopicsParams::new("example.com").with_limit(50)).await?;

// Subscribe/unsubscribe a contact per-topic
mailblastr.contacts.update_topics(
    contact_id,
    UpdateContactTopicsOptions::new().with_topic(&topic.id, SubscriptionState::OptIn),
).await?;
```

### Automations & Events

Every automation belongs to one of your sending domains — `domain` is required on create, and `events.send` names the domain it targets, so the same event name across several products can never trigger the wrong automation.

```rust
let automation = mailblastr.automations.create(
    CreateAutomationOptions::new("Welcome series", "yourdomain.com")
        .with_trigger("contact.created"),
).await?;

mailblastr.automations.add_step(
    &automation.id,
    AddAutomationStepOptions::new("send_email")
        .with_config(serde_json::json!({ "template_id": "tmpl_welcome" })),
).await?;
mailblastr.automations.update(&automation.id, UpdateAutomationOptions::new().with_status("enabled")).await?;

// Or let the model build the graph (the automation must be disabled and empty)
mailblastr.automations.create_with_ai(
    &automation.id,
    AutomationAiOptions::new("Welcome new signups, then nudge non-openers after 3 days"),
).await?;

// Fire a custom event — only yourdomain.com's automations are triggered
mailblastr.events.send(
    SendEventOptions::new("signup.completed", "yourdomain.com")
        .with_email("user@example.com")
        .with_payload_entry("plan", "pro"),
).await?;

// Inspect execution
let runs = mailblastr.automations.runs_filtered(
    &automation.id,
    Some(ListAutomationRunsParams::new().with_status(["failed"])),
).await?;
mailblastr.automations.get_run(&automation.id, &runs.data[0].id).await?;
mailblastr.automations.stop(&automation.id).await?;
```

Structural edits — `domain`, `trigger`, `trigger_config`, `connections`, and any step add/edit/delete — require the automation to be **disabled** first; the API answers `422 validation_error` otherwise.

### Webhooks

```rust
let hook = mailblastr.webhooks.create(CreateWebhookOptions::new(
    "https://yourapp.com/hooks/mailblastr",
    ["email.delivered", "email.bounced", "email.unsubscribed"],
)).await?;
// hook.signing_secret is revealed ONCE — store it now.

mailblastr.webhooks.list(None).await?;
mailblastr.webhooks.rotate(&hook.id).await?; // new secret, revealed once
mailblastr.webhooks.test(&hook.id).await?;   // 200 even on failure — check `ok`
```

The endpoint must be `https://` and must not resolve to a private address. Event names are
`email.sent`, `email.delivered`, `email.delivery_delayed`, `email.bounced`, `email.complained`,
`email.opened`, `email.clicked`, `email.failed`, `email.scheduled`, `email.suppressed`,
`email.received`, `email.replied`, `email.unsubscribed`, `contact.created`, `contact.updated`,
`contact.deleted`, `domain.created`, `domain.updated`, `domain.deleted` — short aliases such as
`opened` or `bounce` are accepted on write and stored in canonical form.

Verify deliveries locally (pure HMAC-SHA256 — no HTTP request). Pass the EXACT raw body string; re-serializing parsed JSON breaks the signature:

```rust
use mailblastr::{verify_webhook_signature, VerifyWebhookOptions, WebhookHeaders};

let headers = WebhookHeaders::new(svix_id, svix_timestamp, svix_signature);
let result = verify_webhook_signature(raw_body, &headers, &signing_secret, &VerifyWebhookOptions::default());
if result.valid {
    // process the event
}
// also available as mailblastr.webhooks.verify(...)
```

### Pagination

`list` methods accept optional cursor pagination:

```rust
use mailblastr::PaginationParams;

mailblastr.campaigns.list(Some(PaginationParams::new().with_limit(25).with_after("cursor_abc"))).await?;
```

`limit` must be an integer 1–100 (default 20); `after` and `before` are item ids and are mutually exclusive — violating either is a `422 validation_error`. An unknown cursor is not an error: you get an empty page with `has_more: false`. Responses are `{ object: "list", has_more, data }` — there is no total and no next-cursor field, so page forward with the last `data[].id`.

Watch the unpaged default: `domains`, `api_keys`, `topics`, `campaigns`, `contacts`, `contact_properties`, `segments`, `segments.contacts`, `contacts.list_segments`, `contacts.get_topics` and `emails.receiving.list_attachments` skip the 20-row default when you pass no pagination params — but they are still capped, at **1,000** rows per page, and `has_more` reports the truncation. They do **not** return the whole collection: keep paging with `with_after(last_id)` while `has_more` is true. `templates`, `webhooks`, `audiences`, `automations`, `automations.runs` and `events` always cap at 20.

### Idempotency

Safely retry a send with an idempotency key. The key must be **1–255 characters**, measured after the server trims it — 255, not 256 (`mailblastr::IDEMPOTENCY_KEY_MAX_LEN`). The crate sends the key verbatim and lets the **server** be the authority: outside that range the API answers `400 invalid_idempotency_key`.

```rust
mailblastr.emails.send_with_idempotency_key(email, "order-123").await?;
mailblastr.batch.send_emails_with_idempotency_key(emails, "batch-42").await?;
```

Only `POST /emails` and `POST /emails/batch` implement it. Every other endpoint — including `events.send` — ignores the header, so a retry there creates a second resource; de-duplicate on your side instead. `events.send_with_idempotency_key` still exists and still sends the header, but is deprecated for exactly that reason.

Reusing a key with a different payload is `409 invalid_idempotent_request`; reusing it while the first request is still in flight is `409 concurrent_idempotent_requests`; reusing it after the first request finished replays the stored status and body.

### Rate limits

Only the `/emails` **send** surface is rate-limited: 30 requests / 60s per client IP. Reads (`GET /emails`, `GET /emails/:id`, the whole `emails.receiving` subtree and attachment listings) are NOT subject to that cap, so paging a large list no longer risks a 429. Capped responses carry `RateLimit-Limit` / `RateLimit-Remaining` / `RateLimit-Reset` on success as well as on a 429, which is the supported way to pace yourself. The client already retries `429` and `503` automatically (honoring `Retry-After`, up to `DEFAULT_MAX_RETRIES` times) — tune it with `Mailblastr::builder(key).max_retries(n)`.

## Requirements

- Rust 1.75+
- An async runtime ([`tokio`](https://tokio.rs) shown above; any executor that can drive `reqwest` futures works)

## Documentation

Full docs: <https://www.mailblastr.com/docs>

## License

MIT
