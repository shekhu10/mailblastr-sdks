# mailblastr

Official Ruby SDK for the [MailBlastr](https://www.mailblastr.com) email API — send transactional and marketing email from your own verified domain.

Zero runtime dependencies: the gem uses only Ruby's standard library (`Net::HTTP`, `JSON`, `OpenSSL`).

## Install

```bash
gem install mailblastr
```

Or in your Gemfile:

```ruby
gem "mailblastr"
```

## Setup

```ruby
require "mailblastr"

Mailblastr.api_key = "mb_xxxxxxxxx"

# or
Mailblastr.configure do |config|
  config.api_key = ENV["MAILBLASTR_API_KEY"]
  # config.base_url = "https://www.mailblastr.com/api" # override your API host
end
```

## Usage

```ruby
sent = Mailblastr::Emails.send({
  from: "Acme <hello@yourdomain.com>",
  to: ["user@example.com"],
  subject: "Hello from MailBlastr",
  html: "<p>Your first email 🎉</p>"
})
puts sent["id"]
```

Params are plain hashes with snake_case keys, passed through as JSON. Successful calls return the parsed response (a Hash, or a raw String for binary downloads). Any non-2xx response raises `Mailblastr::Error`:

```ruby
begin
  Mailblastr::Emails.send(params)
rescue Mailblastr::Error => e
  puts e.status_code # => 422
  puts e.name        # => "validation_error"
  puts e.message     # => human-readable explanation
end
```

## Domain-first model

MailBlastr is **domain-first**: each sending domain has its own pool of contacts. The same email address on two domains is two records with separate consent, so unsubscribes on one product never leak into another.

That means `domain` (the sending domain, e.g. `"yourdomain.com"` — one of your verified domains) is **required** on:

- `Contacts.create` / `Contacts.list` (the flat `/contacts` API — pass `audience_id:` to use the nested audience routes instead)
- `Segments.create` / `Segments.list`
- `Topics.create` / `Topics.list`
- `Campaigns.create` (picks the contact pool the campaign targets; `from` may be a different verified domain)
- `Automations.create` and `Events.send` (only automations belonging to that domain are triggered)

## Emails

```ruby
Mailblastr::Emails.send({ from: from, to: to, subject: subject, html: html })
Mailblastr::Emails.list({ limit: 20, after: cursor })   # cursor pagination
Mailblastr::Emails.get(email_id)
Mailblastr::Emails.list_attachments(email_id)
Mailblastr::Emails.get_attachment(email_id, attachment_id)
Mailblastr::Emails.update(email_id, { scheduled_at: "2026-08-01T09:00:00Z" }) # reschedule
Mailblastr::Emails.cancel(email_id)

# Batch send — up to 100 emails in one request
Mailblastr::Batch.send([
  { from: from, to: ["a@example.com"], subject: "Hi A", html: "<p>A</p>" },
  { from: from, to: ["b@example.com"], subject: "Hi B", html: "<p>B</p>" }
])

# Attachments: hosted URL (path) or inline base64 (content)
Mailblastr::Emails.send({
  from: from, to: to, subject: "Your invoice", html: "<p>Attached.</p>",
  attachments: [
    { filename: "invoice.pdf", path: "https://yourdomain.com/invoices/invoice.pdf" },
    { filename: "report.csv", content: base64_content, content_type: "text/csv" }
  ]
})
```

### Inbound email

```ruby
Mailblastr::Emails::Receiving.list
Mailblastr::Emails::Receiving.get(id)
Mailblastr::Emails::Receiving.list_attachments(id)
Mailblastr::Emails::Receiving.get_attachment(id, attachment_id) # => raw bytes (String)
Mailblastr::Emails::Receiving.raw(id)                           # => original RFC822 message
Mailblastr::Emails::Receiving.forward(id, { from: "you@yourdomain.com", to: "team@you.com" })
Mailblastr::Emails::Receiving.reply(id, { from: "you@yourdomain.com", html: "<p>Thanks!</p>" })
Mailblastr::Emails::Receiving.delete(id)
```

## Domains

```ruby
Mailblastr::Domains.create({ name: "yourdomain.com" })
Mailblastr::Domains.get(id)
Mailblastr::Domains.list
Mailblastr::Domains.update(id, { click_tracking: true })
Mailblastr::Domains.verify(id)
Mailblastr::Domains.delete(id)

# Claim a domain verified in another account
Mailblastr::Domains.claim({ name: "yourdomain.com" })
Mailblastr::Domains.get_claim(id)
Mailblastr::Domains.verify_claim(id)

# One-click DNS setup
Mailblastr::Domains.detect_dns(id)
Mailblastr::Domains.apply_cloudflare_dns(id, { token: cf_token })
Mailblastr::Domains.apply_godaddy_dns(id, { key: key, secret: secret })
Mailblastr::Domains.apply_namecheap_dns(id, { apiUser: user, apiKey: key })
```

## Contacts (domain-first)

```ruby
Mailblastr::Contacts.create({ domain: "yourdomain.com", email: "user@example.com", first_name: "Ada" })
Mailblastr::Contacts.list({ domain: "yourdomain.com" })
Mailblastr::Contacts.get({ id: contact_id })                                # by id (exact) …
Mailblastr::Contacts.get({ id: "user@example.com", domain: "yourdomain.com" }) # … or email + domain
Mailblastr::Contacts.update({ id: contact_id, unsubscribed: true })
Mailblastr::Contacts.delete({ id: contact_id })

# Nested audience variants
Mailblastr::Contacts.create({ audience_id: aud_id, email: "user@example.com" })
Mailblastr::Contacts.list({ audience_id: aud_id, segment_id: seg_id })

# Bulk import
Mailblastr::Contacts.batch({ audience_id: aud_id, contacts: [{ email: "a@b.com" }], on_conflict: "skip" })
Mailblastr::Contacts.import({ audience_id: aud_id, csv: "email,company\na@b.com,Acme" })

# Segments & topics per contact
Mailblastr::Contacts.add_to_segment(contact_id, segment_id)
Mailblastr::Contacts.remove_from_segment(contact_id, segment_id)
Mailblastr::Contacts.list_segments(contact_id)
Mailblastr::Contacts.get_topics(contact_id)
Mailblastr::Contacts.update_topics(contact_id, { topics: [{ id: topic_id, subscription: "opt_in" }] })

# Custom contact properties ({{merge_tags}})
Mailblastr::ContactProperties.create({ key: "plan", type: "string", fallback_value: "free" })
```

## Audiences

```ruby
Mailblastr::Audiences.create({ name: "Newsletter" })
Mailblastr::Audiences.get(id)
Mailblastr::Audiences.list
Mailblastr::Audiences.update(id, { name: "Weekly newsletter" })
Mailblastr::Audiences.delete(id)

# Import from a link-shared Google Sheet
Mailblastr::Audiences.import_sheet(id, { url: sheet_url, segment_name: "June leads" })
```

## Segments & Topics (domain-first)

```ruby
Mailblastr::Segments.create({ domain: "yourdomain.com", name: "VIP", filter: { status: "subscribed" } })
Mailblastr::Segments.list({ domain: "yourdomain.com" })
Mailblastr::Segments.get(id)
Mailblastr::Segments.contacts(id) # preview who matches
Mailblastr::Segments.update(id, { name: "VIP customers" })
Mailblastr::Segments.delete(id)

Mailblastr::Topics.create({ domain: "yourdomain.com", name: "Product updates", default_subscription: "opt_in" })
Mailblastr::Topics.list({ domain: "yourdomain.com" })
Mailblastr::Topics.update(id, { description: "New features" })
Mailblastr::Topics.delete(id)
```

## Campaigns (domain-first)

```ruby
campaign = Mailblastr::Campaigns.create({
  domain: "yourdomain.com", # REQUIRED — the contact pool this campaign targets
  from: "Acme <hello@yourdomain.com>",
  subject: "Big news",
  html: "<p>Hello {{first_name}}</p>",
  segment_id: seg_id # optional — subset instead of everyone
})

Mailblastr::Campaigns.send(campaign["id"])                                    # send now
Mailblastr::Campaigns.send(campaign["id"], { scheduled_at: "2026-08-01T09:00:00Z" }) # or schedule
Mailblastr::Campaigns.cancel(campaign["id"])
Mailblastr::Campaigns.stats(campaign["id"])
Mailblastr::Campaigns.ab(campaign["id"]) # A/B winner evaluation
Mailblastr::Campaigns.get(campaign["id"])
Mailblastr::Campaigns.list({ limit: 25 })
Mailblastr::Campaigns.update(campaign["id"], { subject: "Bigger news" })
Mailblastr::Campaigns.delete(campaign["id"])
```

## Templates

```ruby
tmpl = Mailblastr::Templates.create({ name: "Welcome", subject: "Welcome!", html: "<p>Hi {{first_name}}</p>" })
Mailblastr::Templates.publish(tmpl["id"])
Mailblastr::Templates.duplicate(tmpl["id"], { name: "Welcome v2" })
Mailblastr::Templates.get(tmpl["id"])
Mailblastr::Templates.list
Mailblastr::Templates.update(tmpl["id"], { subject: "Welcome aboard!" })
Mailblastr::Templates.delete(tmpl["id"])

# Send with a template
Mailblastr::Emails.send({ from: from, to: to, template_id: tmpl["id"], variables: { first_name: "Ada" } })
```

## Automations & Events (domain-first)

```ruby
automation = Mailblastr::Automations.create({
  name: "Welcome series",
  domain: "yourdomain.com", # REQUIRED
  trigger: "contact.created"
})

Mailblastr::Automations.add_step(automation["id"], { type: "send_email", config: { template_id: tmpl_id } })
Mailblastr::Automations.update(automation["id"], { status: "enabled" })

# Fire a custom event — only yourdomain.com's automations are triggered
Mailblastr::Events.send({
  event: "signup.completed",
  domain: "yourdomain.com", # REQUIRED
  email: "user@example.com",
  payload: { plan: "pro" }
})

# Event definitions
Mailblastr::Events.create({ name: "signup.completed", schema: { plan: "string" } })
Mailblastr::Events.list
Mailblastr::Events.delete(event_id)

# Inspect execution
runs = Mailblastr::Automations.runs(automation["id"], { limit: 25 })
Mailblastr::Automations.get_run(automation["id"], runs["data"].first["id"])
Mailblastr::Automations.delete_step(automation["id"], step_id)
Mailblastr::Automations.stop(automation["id"])
Mailblastr::Automations.delete(automation["id"])
```

## Webhooks

```ruby
hook = Mailblastr::Webhooks.create({
  endpoint: "https://yourapp.com/hooks/mailblastr",
  events: ["email.delivered", "email.bounced", "contact.unsubscribed"]
})
hook["signing_secret"] # shown ONCE — store it

Mailblastr::Webhooks.list
Mailblastr::Webhooks.update(hook["id"], { status: "disabled" })
Mailblastr::Webhooks.rotate(hook["id"]) # new secret returned once
Mailblastr::Webhooks.test(hook["id"])
Mailblastr::Webhooks.delete(hook["id"])
```

### Verifying deliveries

`verify_signature` checks the Svix-style HMAC-SHA256 signature locally (no HTTP request). Pass the **exact raw request body** — re-serializing parsed JSON breaks the signature.

```ruby
result = Mailblastr::Webhooks.verify_signature(
  request.raw_post,          # raw body string
  {
    "svix-id" => request.headers["svix-id"],
    "svix-timestamp" => request.headers["svix-timestamp"],
    "svix-signature" => request.headers["svix-signature"]
  },
  signing_secret             # the whsec_... secret from create/rotate
)

head :unauthorized unless result[:valid]
# result => { valid: true } or { valid: false, reason: "no_match" | "timestamp_out_of_tolerance" | ... }
```

Pass `tolerance: 0` to skip the timestamp freshness check (default 300 seconds).

## API keys, Logs & Polls

```ruby
Mailblastr::ApiKeys.create({ name: "CI", permission: "sending_access" })
Mailblastr::ApiKeys.list
Mailblastr::ApiKeys.delete(key_id)

Mailblastr::Logs.list({ limit: 100, method: "POST", status: 429 })
Mailblastr::Logs.get(log_id)

Mailblastr::Polls.list
Mailblastr::Polls.get(email_id) # aggregated answer breakdown
```

## Pagination

`list` methods accept cursor pagination — `{ limit:, after:, before: }` — appended as a query string:

```ruby
Mailblastr::Campaigns.list({ limit: 25, after: "cursor_abc" })
```

## Idempotency

Pass an idempotency key to safely retry a create (24h window):

```ruby
Mailblastr::Emails.send(payload, { idempotency_key: "order-123" })
```

## Documentation

Full docs: <https://www.mailblastr.com/docs>

## License

MIT
