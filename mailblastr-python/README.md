# mailblastr

Official Python SDK for the [MailBlastr](https://www.mailblastr.com) email API — send transactional and marketing email from your own verified domain.

Zero dependencies (Python standard library only). Python 3.8+.

## Install

```bash
pip install mailblastr
```

## Setup

Grab your API key from the [MailBlastr dashboard](https://www.mailblastr.com).

```python
import mailblastr

mailblastr.api_key = "mb_xxxxxxxxx"
```

## Send your first email

```python
import mailblastr

mailblastr.api_key = "mb_xxxxxxxxx"

params: mailblastr.Emails.SendParams = {
    "from": "Acme <hello@yourdomain.com>",
    "to": ["delivered@mailblastr.dev"],
    "subject": "Hello from MailBlastr",
    "html": "<p>Your first email 🎉</p>",
}

email = mailblastr.Emails.send(params)
print(email["id"])
```

Every method returns the parsed JSON response — except the three binary
download helpers (`Emails.Receiving.get_attachment`, `Emails.Receiving.get_raw`
and `Domains.records_csv`), which return `bytes`. On any non-2xx status the SDK
raises `mailblastr.MailblastrError` carrying the API error body:

```python
try:
    mailblastr.Emails.send(params)
except mailblastr.MailblastrError as e:
    print(e.status_code, e.name, e.message)   # e.g. 422 validation_error "..."
```

### Attachments

Attach files by hosted URL (`path`, fetched at send time) or inline base64 (`content`):

```python
mailblastr.Emails.send({
    "from": "Acme <hello@yourdomain.com>",
    "to": ["delivered@mailblastr.dev"],
    "subject": "Your invoice",
    "html": "<p>Invoice attached.</p>",
    "attachments": [
        {"filename": "invoice.pdf", "path": "https://yourdomain.com/invoices/invoice.pdf"},
        {"filename": "report.csv", "content": base64_content, "content_type": "text/csv"},
    ],
})
```

### Batch send

```python
mailblastr.Batch.send([
    {"from": "hello@yourdomain.com", "to": ["delivered@mailblastr.dev"], "subject": "Hi A", "html": "<p>A</p>"},
    {"from": "hello@yourdomain.com", "to": ["delivered@mailblastr.dev"], "subject": "Hi B", "html": "<p>B</p>"},
])  # up to 100 emails per request
```

A batch succeeds in one of two ways, chosen by its **size** alone — branch on
`queued`, not on the absence of an error:

| Batch size | Status | Response | What happened |
|---|---|---|---|
| 1–40 | `200` | no `queued` key at all — never `queued: False` | every id in `data` is already handed to the mail service |
| 41–100 | `202` | `queued: True`, `queued_count == len(result["data"])` | ids are real, but the emails are still `scheduled` — **nothing has been transmitted yet** |

```python
result = mailblastr.Batch.send(payloads)
if result.get("queued"):
    # Accepted, not sent. Poll mailblastr.Emails.get(id) for the outcome.
    ...
```

Queuing is the only way the documented 100-email maximum can be accepted at all:
100 inline sends run past the platform's request ceiling. A batch carrying an
`@mailblastr.dev` simulator recipient in `to`, `cc` or `bcc` stays inline at any
size. An inline batch near the 40-email boundary can take ~100s server-side, far
past the 30s default `mailblastr.timeout` — raise it for batches that large, and
always pass an `idempotency_key`, since a client that gives up mid-request cannot
tell what was already sent.

### Options

```python
mailblastr.base_url = "https://www.mailblastr.com/api"   # override your API host
mailblastr.timeout = 30            # per-request timeout in seconds (default 30)
mailblastr.max_retries = 2         # auto-retry 429/503 responses (default 2; 0 disables)
```

Requests time out after 30 seconds by default. Automatic retries consider only HTTP 429 and 503 and honor `Retry-After`
(with capped exponential backoff otherwise). Partial or uncertain results stop
immediately. Generic 503 responses retry only reads or a send protected by an
operation key; other writes need a documented rejection before processing.
Network errors, timeouts, 409, 422, and other 5xx responses are never retried.
See **Recovery and tracking contracts** below before retrying in application code.

### Errors

`MailblastrError` carries the `{statusCode, name, message}` envelope. Match on
`name` and read `status_code` — messages are scrubbed server-side and a few
handlers override the status a name usually maps to, so neither is safe to
hard-code. Extra fields ride along on the exception:

```python
try:
    mailblastr.Emails.send(params)
except mailblastr.MailblastrError as e:
    if e.name == "daily_quota_exceeded":
        print(e.limit["used"], e.limit["limit"], e.limit["next_plan"])
    if e.retry_after:
        time.sleep(e.retry_after)
    print(e.body)          # the full parsed error body

try:
    mailblastr.Batch.send(payloads, options={"idempotency_key": "batch-1"})
except mailblastr.MailblastrError as e:
    already_sent = e.sent          # [{"id": ...}, ...] — do NOT resend these
    print(e.sent_count)
```

## The domain-first model

MailBlastr is DOMAIN-FIRST: each of your verified sending domains has its own
contact pool — the same address on two domains is two records with separate
consent. That means:

- `Contacts` take a `domain` (required to create/list on the flat `/contacts`
  API; disambiguates an email id on get/update/remove).
- `Segments` and `Topics` belong to a domain (`domain` required on create and list).
- `Campaigns.create` REQUIRES `domain` — it picks the contact pool the campaign
  targets (the `from` address may be a different verified domain).
- `Automations.create` REQUIRES `domain`, and `Events.send` REQUIRES `domain` —
  only automations belonging to that domain are triggered, so the same event
  name (e.g. `user.created`) across several products can never double-fire.

## Resources

Each resource is a class with methods following a consistent
`create` / `get` / `list` / `update` / `remove` shape (plus resource-specific verbs):
`Emails` (with nested `Emails.Receiving`), `Batch`,
`Domains`, `Audiences`, `Contacts`, `ContactProperties`, `Campaigns`,
`Segments`, `Topics`, `Templates`, `Automations`, `Webhooks`, `Events`,
`ApiKeys` (list only — see below), `Logs`, `Polls`.

```python
# Emails
mailblastr.Emails.send(params)
mailblastr.Emails.list({"limit": 20, "after": cursor})   # cursor pagination
mailblastr.Emails.list({"status": "bounced", "search": "acme.com"})  # filters
mailblastr.Emails.list({"folder": "sent"})  # outbox / sent / scheduled / failed — any other value 422s
mailblastr.Emails.get(email_id)
mailblastr.Emails.update(email_id, {"scheduled_at": "2026-08-01T09:00:00Z"})  # reschedule
mailblastr.Emails.cancel(email_id)
mailblastr.Emails.sources()                # per-campaign/automation send metrics
mailblastr.Emails.list_attachments(email_id)
mailblastr.Emails.get_attachment(email_id, attachment_id)

# Inbound email
mailblastr.Emails.Receiving.list()
mailblastr.Emails.Receiving.list_addresses()    # per-address inbound stats
mailblastr.Emails.Receiving.get(email_id)
mailblastr.Emails.Receiving.list_attachments(email_id)
mailblastr.Emails.Receiving.get_attachment(email_id, attachment_id)  # -> bytes
mailblastr.Emails.Receiving.get_raw(email_id)                            # -> bytes (RFC822)
mailblastr.Emails.Receiving.forward(email_id, {"from": "me@yourdomain.com", "to": "delivered@mailblastr.dev"})
mailblastr.Emails.Receiving.reply(email_id, {"from": "me@yourdomain.com", "html": "<p>Thanks!</p>"})
mailblastr.Emails.Receiving.remove(email_id)

# Domains (incl. claiming a domain verified elsewhere + one-click DNS)
mailblastr.Domains.create({"name": "yourdomain.com"})
mailblastr.Domains.verify(domain_id)
mailblastr.Domains.claim({"name": "yourdomain.com"})
mailblastr.Domains.verify_claim(domain_id)
mailblastr.Domains.detect_dns(domain_id)
mailblastr.Domains.apply_cloudflare_dns(domain_id, {"token": cf_token})
mailblastr.Domains.mx_check("yourdomain.com")   # live MX lookup
mailblastr.Domains.records_csv(domain_id)       # -> bytes (text/csv)

# Contacts (domain-first)
mailblastr.Contacts.create({"domain": "yourdomain.com", "email": "user@example.com", "first_name": "Ada"})
mailblastr.Contacts.list({"domain": "yourdomain.com"})
mailblastr.Contacts.get({"id": contact_id})                                # by id (exact) …
mailblastr.Contacts.get({"id": "user@example.com", "domain": "yourdomain.com"})  # … or by email + domain
mailblastr.Contacts.update({"id": contact_id, "unsubscribed": True})
mailblastr.Contacts.remove({"id": contact_id})
mailblastr.Contacts.batch({"audience_id": aud_id, "contacts": [{"email": "a@b.com"}]})
# Domain-first: import straight into a domain's pool, no audience id needed.
mailblastr.Contacts.batch({"domain": "yourdomain.com", "contacts": [{"email": "a@b.com"}]})
mailblastr.Contacts.import_csv({"audience_id": aud_id, "csv": "email,company\na@b.com,Acme"})
mailblastr.Contacts.create_import_upload({"audience_id": aud_id, "filename": "big.csv", "size": 90_000_000})
mailblastr.Contacts.add_to_segment(contact_id, segment_id)
mailblastr.Contacts.list_segments(contact_id)
mailblastr.Contacts.update_topics(contact_id, {"topics": [{"id": topic_id, "subscription": "opt_in"}]})

# Contact properties (custom fields / merge tags)
mailblastr.ContactProperties.create({"key": "plan", "type": "string"})

# Campaigns & Segments (domain-first)
mailblastr.Campaigns.create({"domain": "yourdomain.com", "from": sender, "subject": subject, "html": html})
mailblastr.Campaigns.send(campaign_id, {"scheduled_at": "tomorrow at 9am"})
mailblastr.Campaigns.stats(campaign_id)
mailblastr.Campaigns.engagement(campaign_id)   # who opened / clicked / replied
mailblastr.Campaigns.ab(campaign_id)
mailblastr.Segments.create({"domain": "yourdomain.com", "name": "VIP", "filter": {"status": "subscribed"}})
mailblastr.Segments.list({"domain": "yourdomain.com"})
mailblastr.Segments.contacts(segment_id)   # preview who matches

# Topics (domain-first)
mailblastr.Topics.create({"domain": "yourdomain.com", "name": "Product updates", "default_subscription": "opt_in"})
mailblastr.Topics.list({"domain": "yourdomain.com"})

# Templates
mailblastr.Templates.create({"name": "Welcome", "subject": "Hi {{first_name}}", "html": html})
mailblastr.Templates.duplicate(template_id)
mailblastr.Templates.publish(template_id)
mailblastr.Emails.send({"from": sender, "to": to, "template_id": tmpl_id, "variables": {"first_name": "Ada"}})

# Audiences
mailblastr.Audiences.list()
mailblastr.Audiences.import_sheet(audience_id, {"url": sheet_url})

# API keys (listing only — creating, re-scoping and revoking is dashboard-only)
mailblastr.ApiKeys.list()

# Logs & Polls
mailblastr.Logs.list({"limit": 100, "method": "POST", "status": 429})
mailblastr.Logs.get(log_id)
mailblastr.Polls.list()
mailblastr.Polls.get(email_id)
```

### API keys are managed in the dashboard

`ApiKeys.list()` is the whole surface: the SDK deliberately exposes no method to
create, re-scope or revoke a key. Key lifecycle belongs to a signed-in dashboard
session, and the API enforces it — `POST /api-keys`, `PATCH /api-keys/:id` and
`DELETE /api-keys/:id` answer `403 dashboard_only` to any API-key caller,
whatever its permission. That is the point: a key that leaks cannot mint itself
a replacement, widen its own access, or revoke the keys you would use to shut it
off. Create and revoke keys at
[mailblastr.com](https://www.mailblastr.com) instead.

### Automations & Events

Every automation belongs to one of your sending domains — `domain` is required
on create, and `Events.send` names the domain it targets.

```python
automation = mailblastr.Automations.create({
    "name": "Welcome series",
    "domain": "yourdomain.com",
    "trigger": "contact.created",
})

mailblastr.Automations.add_step(automation["id"], {
    "type": "send_email",
    "config": {"template_id": "tmpl_welcome"},
})
mailblastr.Automations.update(automation["id"], {"status": "enabled"})

# Fire a custom event — only yourdomain.com's automations are triggered
mailblastr.Events.send({
    "event": "signup.completed",
    "domain": "yourdomain.com",
    "email": "delivered@mailblastr.dev",
    "payload": {"plan": "pro"},
})
mailblastr.Events.create({"name": "signup.completed", "schema": {"plan": "string"}})
mailblastr.Events.update(event_id, {"schema": {"plan": "string", "seats": "number"}})

# Inspect execution
runs = mailblastr.Automations.runs(automation["id"], {"limit": 25, "status": "failed"})
mailblastr.Automations.get_run(automation["id"], runs["data"][0]["id"])
mailblastr.Automations.stop(automation["id"])
```

The step graph is edited while the automation is **disabled** —
`add_step` / `update_step` / `delete_step` (and changing `domain`, `trigger` or
`connections`) all 422 on an enabled automation.
`Automations.create_with_ai` builds or extends the graph from a prompt:

```python
mailblastr.Automations.create_with_ai(automation["id"], {"prompt": "Wait 2 days, then send the welcome email"})
```

### Webhooks

```python
hook = mailblastr.Webhooks.create({
    "endpoint": "https://yourapp.com/hooks/mailblastr",   # must be https://
    "events": ["email.delivered", "email.bounced", "email.unsubscribed"],
})
signing_secret = hook["signing_secret"]   # shown ONCE, only here

mailblastr.Webhooks.list()
mailblastr.Webhooks.update(hook["id"], {"status": "disabled"})
mailblastr.Webhooks.rotate(hook["id"])    # new secret, returned once
mailblastr.Webhooks.test(hook["id"])
```

Verify incoming deliveries locally (no HTTP request) — pass the EXACT raw
request body string, the `svix-*` headers, and your signing secret:

```python
result = mailblastr.Webhooks.verify(raw_body, request.headers, signing_secret)
if not result["valid"]:
    abort(401)   # result["reason"] says why, e.g. 'no_match'
```

### Pagination

`list()` methods accept optional cursor pagination — `{"limit", "after", "before"}`.
`limit` is an integer 1–100 (default 20); `after` and `before` are item ids and
cannot be combined. Responses are `{"object": "list", "has_more": bool, "data": [...]}` —
there is no `total` and no `next_cursor`, so page forward with the last
`data[-1]["id"]`:

```python
page = mailblastr.Campaigns.list({"limit": 25})
while page["has_more"]:
    page = mailblastr.Campaigns.list({"limit": 25, "after": page["data"][-1]["id"]})
```

Called with **no** pagination params, most list endpoints return the whole
collection up to a 1,000-item ceiling (`Campaigns`, `Contacts`, `Segments`,
`ContactProperties`, `Domains`, `ApiKeys`, `Topics`, `Polls`, and the nested
contact/segment/topic lists) — past that the response is truncated and
`has_more` is `True`, so keep paging rather than trusting one call to be
complete. `Audiences`, `Automations`, `Automations.runs`, `Templates`,
`Webhooks` and `Events` cap at 20 instead — pass `limit` explicitly when it
matters.
An unknown cursor is not an error: it returns an empty page with
`has_more: False`.

### Idempotency

Pass an idempotency key to safely retry a send — replaying the same key returns
the original response instead of sending twice:

```python
mailblastr.Emails.send(params, options={"idempotency_key": "order-123"})
mailblastr.Batch.send(payloads, options={"idempotency_key": "orders-2026-08-08"})
```

- The key must be **1 to 255 characters** — measured after the server trims it,
  so 255, not 256 (`mailblastr.IDEMPOTENCY_KEY_MAX_LENGTH`). The SDK sends the
  key verbatim and lets the **server** be the authority: an out-of-range key
  comes back as a `MailblastrError` with `name == "invalid_idempotency_key"`
  (400).
- **`Emails.send`, `Batch.send`, and received-email reply/forward honour it.** Every other endpoint —
  including `Events.send` — accepts and forwards the header but the API ignores
  it, so a retry there creates a second resource. De-duplicate on your side
  instead.
- Reusing a key with a *different* body raises `invalid_idempotent_request`
  (409); reusing it while the first request is still running raises
  `concurrent_idempotent_requests` (409).

## Documentation

Full docs: <https://www.mailblastr.com/docs>

## License

MIT

## Recovery and tracking contracts

Use a stable, unique operation key for each intended send, batch, reply, or
forward. Keep the same key and payload when recovering that operation. These
are the supported idempotent send endpoints; events do not implement this
header. Existing calls without options still work.

```python
mailblastr.Emails.Receiving.reply(email_id, reply, {"idempotency_key": "reply-operation-1"})
mailblastr.Emails.Receiving.forward(email_id, forward, {"idempotency_key": "forward-operation-1"})
health = mailblastr.Domains.tracking_health(domain_id)
```

Automatic retries consider only 429/503. They stop on an original email `id`,
positive `sent_count`, nonempty `sent` or `reserved`, or `batch_incomplete`.
An ordinary rate limit can retry; a generic 503 can retry a read or a send with
the same supported key. Other writes retry only documented pre-processing
rejections (`service_unavailable`, `sending_service_unavailable`,
`sending_configuration_unavailable`, `contacts_busy`, `contacts_timeout`).
No network/body-read failure, 409, 422, or other 5xx is retried automatically.
The default transport refuses redirects; a custom transport/client must enforce
its own policy.

On a failed or unconfirmed send, inspect `id` with the email retrieval method
before creating another send. A 422 with an ID can identify an uncertain
provider handoff; 422 does not always mean nothing happened. For interrupted
batches, `sent` contains confirmed sends, `reserved` contains the original
attempted prefix (including uncertain handoffs), and `unsent_count` counts the
never-attempted tail. Do not resend the full batch or the reserved prefix under
a new key. Reconcile original IDs first, then submit only known unattempted
items as a new operation. Recovery fields remain available in the full error
body as well as language-specific fields/accessors.

Tracking health returns `custom_host`, `status` (`shared`, `ready`, or
`unavailable`), and `checked_at`. Configure custom tracking through the domain
API and check health before relying on it. A healthy endpoint cannot guarantee
an open event: recipients may block images, and coupon redemption alone is not
proof that the tracking pixel loaded. SDKs preserve supplied HTML/text and do
not infer opens or rewrite editor spacing.

Campaign cancellation also stops pending follow-ups for an already-sent
campaign while retaining its sent history. Permanent received-email deletion
acknowledges a durable cleanup request; attachment/object cleanup can finish
asynchronously. Retrying that deletion is safe; it cannot be undone after the
purge request is accepted.
