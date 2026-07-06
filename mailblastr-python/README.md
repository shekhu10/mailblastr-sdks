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
    "to": ["user@example.com"],
    "subject": "Hello from MailBlastr",
    "html": "<p>Your first email 🎉</p>",
}

email = mailblastr.Emails.send(params)
print(email["id"])
```

Every method returns the parsed JSON response. On any non-2xx status the SDK raises `mailblastr.MailblastrError` carrying the API error body:

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
    "to": ["user@example.com"],
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
    {"from": "hello@yourdomain.com", "to": ["a@example.com"], "subject": "Hi A", "html": "<p>A</p>"},
    {"from": "hello@yourdomain.com", "to": ["b@example.com"], "subject": "Hi B", "html": "<p>B</p>"},
])  # up to 100 emails per request
```

### Options

```python
mailblastr.base_url = "https://api.mailblastr.com"   # override your API host
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
`Emails` (with nested `Emails.Attachments` and `Emails.Receiving`), `Batch`,
`Domains`, `Audiences`, `Contacts`, `ContactProperties`, `Campaigns`,
`Segments`, `Topics`, `Templates`, `Automations`, `Webhooks`, `Events`,
`ApiKeys`, `Logs`, `Polls`.

```python
# Emails
mailblastr.Emails.send(params)
mailblastr.Emails.list({"limit": 20, "after": cursor})   # cursor pagination
mailblastr.Emails.get(email_id)
mailblastr.Emails.update(email_id, {"scheduled_at": "2026-08-01T09:00:00Z"})  # reschedule
mailblastr.Emails.cancel(email_id)
mailblastr.Emails.Attachments.list(email_id)
mailblastr.Emails.Attachments.get(email_id, attachment_id)

# Inbound email
mailblastr.Emails.Receiving.list()
mailblastr.Emails.Receiving.get(email_id)
mailblastr.Emails.Receiving.attachments(email_id)
mailblastr.Emails.Receiving.get_attachment(email_id, attachment_id)  # -> bytes
mailblastr.Emails.Receiving.raw(email_id)                            # -> bytes (RFC822)
mailblastr.Emails.Receiving.forward(email_id, {"from": "me@yourdomain.com", "to": "team@you.com"})
mailblastr.Emails.Receiving.reply(email_id, {"from": "me@yourdomain.com", "html": "<p>Thanks!</p>"})
mailblastr.Emails.Receiving.remove(email_id)

# Domains (incl. claiming a domain verified elsewhere + one-click DNS)
mailblastr.Domains.create({"name": "yourdomain.com"})
mailblastr.Domains.verify(domain_id)
mailblastr.Domains.claim({"name": "yourdomain.com"})
mailblastr.Domains.verify_claim(domain_id)
mailblastr.Domains.detect_dns(domain_id)
mailblastr.Domains.apply_cloudflare_dns(domain_id, {"token": cf_token})

# Contacts (domain-first)
mailblastr.Contacts.create({"domain": "yourdomain.com", "email": "user@example.com", "first_name": "Ada"})
mailblastr.Contacts.list({"domain": "yourdomain.com"})
mailblastr.Contacts.get({"id": contact_id})                                # by id (exact) …
mailblastr.Contacts.get({"id": "user@example.com", "domain": "yourdomain.com"})  # … or by email + domain
mailblastr.Contacts.update({"id": contact_id, "unsubscribed": True})
mailblastr.Contacts.remove({"id": contact_id})
mailblastr.Contacts.batch({"audience_id": aud_id, "contacts": [{"email": "a@b.com"}]})
mailblastr.Contacts.import_csv({"audience_id": aud_id, "csv": "email,company\na@b.com,Acme"})
mailblastr.Contacts.add_to_segment(contact_id, segment_id)
mailblastr.Contacts.list_segments(contact_id)
mailblastr.Contacts.update_topics(contact_id, {"topics": [{"id": topic_id, "subscription": "opt_in"}]})

# Contact properties (custom fields / merge tags)
mailblastr.ContactProperties.create({"key": "plan", "type": "string"})

# Campaigns & Segments (domain-first)
mailblastr.Campaigns.create({"domain": "yourdomain.com", "from": sender, "subject": subject, "html": html})
mailblastr.Campaigns.send(campaign_id, {"scheduled_at": "tomorrow 9am"})
mailblastr.Campaigns.stats(campaign_id)
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

# API keys
mailblastr.ApiKeys.create({"name": "CI", "permission": "sending_access"})
mailblastr.ApiKeys.list()
mailblastr.ApiKeys.remove(key_id)

# Logs & Polls
mailblastr.Logs.list({"limit": 100, "method": "POST", "status": 429})
mailblastr.Logs.get(log_id)
mailblastr.Polls.list()
mailblastr.Polls.get(email_id)
```

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
    "email": "user@example.com",
    "payload": {"plan": "pro"},
})

# Inspect execution
runs = mailblastr.Automations.runs(automation["id"], {"limit": 25})
mailblastr.Automations.get_run(automation["id"], runs["data"][0]["id"])
mailblastr.Automations.stop(automation["id"])
```

### Webhooks

```python
hook = mailblastr.Webhooks.create({
    "endpoint": "https://yourapp.com/hooks/mailblastr",
    "events": ["email.delivered", "email.bounced", "contact.unsubscribed"],
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

`list()` methods accept optional cursor pagination — `{"limit", "after", "before"}`:

```python
mailblastr.Campaigns.list({"limit": 25, "after": "cursor_abc"})
```

### Idempotency

Pass an idempotency key to safely retry a create (24h window):

```python
mailblastr.Emails.send(params, options={"idempotency_key": "order-123"})
```

## Documentation

Full docs: <https://www.mailblastr.com/docs>

## License

MIT
