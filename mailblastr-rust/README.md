# mailblastr

Official Rust SDK for the [MailBlastr](https://www.mailblastr.com) email API — send transactional and marketing email from your own verified domain.

## Install

```bash
cargo add mailblastr
cargo add tokio -F macros,rt-multi-thread
```

## Usage

```rust
use mailblastr::{CreateEmailBaseOptions, Mailblastr, Result};

#[tokio::main]
async fn main() -> Result<()> {
    let mailblastr = Mailblastr::new("mb_xxxxxxxxx");

    let from = "Acme <hello@yourdomain.com>";
    let to = ["user@example.com"];
    let subject = "Hello from MailBlastr";

    let email = CreateEmailBaseOptions::new(from, to, subject)
        .with_html("<p>Your first email 🎉</p>");

    let sent = mailblastr.emails.send(email).await?;
    println!("sent {}", sent.id);
    Ok(())
}
```

Every method is `async` and returns `Result<T, mailblastr::Error>`. API failures surface as `Error::Api { status_code, name, message }`, parsed from the standard MailBlastr error body; transport and decode problems are `Error::Http` / `Error::Json`.

```rust
use mailblastr::Error;

match mailblastr.emails.get("em_missing").await {
    Ok(email) => println!("{}", email.status),
    Err(Error::Api { status_code, name, message }) => {
        eprintln!("{status_code} {name}: {message}");
    }
    Err(other) => eprintln!("transport error: {other}"),
}
```

### Attachments

Attach files by hosted URL (fetched at send time) or inline base64:

```rust
use mailblastr::{Attachment, CreateEmailBaseOptions};

let email = CreateEmailBaseOptions::new(from, to, "Your invoice")
    .with_html("<p>Invoice attached.</p>")
    .with_attachment(Attachment::from_path(
        "invoice.pdf",
        "https://yourdomain.com/invoices/invoice.pdf",
    ))
    .with_attachment(
        Attachment::from_content("report.csv", base64_content).with_content_type("text/csv"),
    );
```

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

One client field per resource, each following a consistent `create` / `get` / `list` / `update` / `remove` shape plus resource-specific verbs:

`emails` (with nested `emails.receiving`), `batch`, `domains`, `audiences`, `contacts`, `contact_properties`, `campaigns`, `segments`, `topics`, `templates`, `automations`, `webhooks`, `logs`, `events`, `api_keys`, `polls`.

```rust
use mailblastr::*;

// Emails
mailblastr.emails.send(email).await?;
mailblastr.emails.list(Some(PaginationParams::new().with_limit(20))).await?;
mailblastr.emails.get(id).await?;
mailblastr.emails.list_attachments(id).await?;
mailblastr.emails.update(id, "2026-08-01T09:00:00Z").await?; // reschedule
mailblastr.emails.cancel(id).await?;

// Inbound email
mailblastr.emails.receiving.list(None).await?;
mailblastr.emails.receiving.get(id).await?;
let pdf: Vec<u8> = mailblastr.emails.receiving.get_attachment(id, att_id).await?;
mailblastr.emails.receiving.forward(id, ForwardReceivedEmailOptions::new(from, ["team@you.com"])).await?;
mailblastr.emails.receiving.reply(id, ReplyReceivedEmailOptions::new(from).with_html("<p>Thanks!</p>")).await?;

// Batch send (alias: mailblastr.emails.batch(...))
mailblastr.batch.send(vec![email_a, email_b]).await?;

// Domains (incl. claiming a domain verified elsewhere + one-click DNS)
mailblastr.domains.create(CreateDomainOptions::new("example.com")).await?;
mailblastr.domains.verify(id).await?;
mailblastr.domains.claim(ClaimDomainOptions::new("example.com")).await?;
mailblastr.domains.verify_claim(id).await?;
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
mailblastr.contacts.batch(audience_id, contacts, Some(OnConflict::Skip)).await?;
mailblastr.contacts.import(audience_id, csv_text, ImportCsvOptions::new()).await?;

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
mailblastr.segments.create(
    CreateSegmentOptions::new("yourdomain.com", "VIP")
        .with_filter(SegmentFilterOptions::new().with_status(SegmentStatus::Subscribed)),
).await?;
mailblastr.segments.list(ListSegmentsParams::new("example.com")).await?;
mailblastr.segments.contacts(id).await?; // preview who matches

// Templates
mailblastr.templates.create(CreateTemplateOptions::new("Welcome").with_subject("Hi {{first_name}}").with_html(html)).await?;
mailblastr.templates.duplicate(id, None).await?;
mailblastr.templates.publish(id).await?;
mailblastr.emails.send(
    CreateEmailBaseOptions::new(from, to, subject)
        .with_template_id(template_id)
        .with_variable("first_name", "Ada"),
).await?;

// API keys
mailblastr.api_keys.create(
    CreateApiKeyOptions::new("CI").with_permission(ApiKeyPermission::SendingAccess),
).await?;
mailblastr.api_keys.list().await?;
mailblastr.api_keys.remove(id).await?;

// Polls (read-only in-email poll results)
mailblastr.polls.list(None).await?;
mailblastr.polls.get(email_id).await?;

// Logs
mailblastr.logs.list(Some(ListLogsParams::new().with_limit(100))).await?;
mailblastr.logs.get(log_id).await?;
```

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

// Fire a custom event — only yourdomain.com's automations are triggered
mailblastr.events.send(
    SendEventOptions::new("signup.completed", "yourdomain.com")
        .with_email("user@example.com")
        .with_payload_entry("plan", "pro"),
).await?;

// Inspect execution
let runs = mailblastr.automations.runs(&automation.id, None).await?;
mailblastr.automations.get_run(&automation.id, &runs.data[0].id).await?;
mailblastr.automations.stop(&automation.id).await?;
```

### Webhooks

```rust
let hook = mailblastr.webhooks.create(CreateWebhookOptions::new(
    "https://yourapp.com/hooks/mailblastr",
    ["email.delivered", "email.bounced", "contact.unsubscribed"],
)).await?;
// hook.signing_secret is revealed ONCE — store it now.

mailblastr.webhooks.list(None).await?;
mailblastr.webhooks.rotate(&hook.id).await?; // new secret, revealed once
mailblastr.webhooks.test(&hook.id).await?;
```

Verify deliveries locally (pure HMAC-SHA256 — no HTTP request). Pass the EXACT raw body string; re-serializing parsed JSON breaks the signature:

```rust
use mailblastr::{verify_webhook_signature, VerifyWebhookOptions, WebhookHeaders};

let headers = WebhookHeaders::new(svix_id, svix_timestamp, svix_signature);
let result = verify_webhook_signature(raw_body, &headers, &signing_secret, &VerifyWebhookOptions::default());
if result.valid {
    // process the event
}
// also available as mailblastr.webhooks.verify_signature(...)
```

### Pagination

`list` methods accept optional cursor pagination:

```rust
use mailblastr::PaginationParams;

mailblastr.campaigns.list(Some(PaginationParams::new().with_limit(25).with_after("cursor_abc"))).await?;
```

### Idempotency

Safely retry creates with an idempotency key (24h window):

```rust
mailblastr.emails.send_with_idempotency_key(email, "order-123").await?;
mailblastr.batch.send_with_idempotency_key(emails, "batch-42").await?;
mailblastr.events.send_with_idempotency_key(event, "evt-7").await?;
```

## Requirements

- Rust 1.75+
- An async runtime ([`tokio`](https://tokio.rs) shown above; any executor that can drive `reqwest` futures works)

## Documentation

Full docs: <https://www.mailblastr.com/docs>

## License

MIT
