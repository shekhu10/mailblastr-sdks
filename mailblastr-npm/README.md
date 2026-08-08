# mailblastr

Official Node.js SDK for the [MailBlastr](https://www.mailblastr.com) email API — send transactional and marketing email from your own verified domain.

## Install

```bash
npm install mailblastr
```

## Usage

```ts
import { MailBlastr } from 'mailblastr';

const mb = new MailBlastr('mb_xxxxxxxxx');

const { data, error } = await mb.emails.send({
  from: 'Acme <hello@yourdomain.com>',
  to: ['user@example.com'],
  subject: 'Hello from MailBlastr',
  html: '<p>Your first email 🎉</p>',
});

if (error) {
  console.error(error.name, error.message);
} else {
  console.log('sent', data.id);
}
```

Every method returns `{ data, error }` — `error` is `null` on success, or
`{ statusCode, name, message }` on failure (no exceptions to catch for API
errors). Branch on `error.name`, never on `error.message`: messages are scrubbed
of provider identifiers and may change.

Plan and quota rejections carry an extra `error.limit` describing what was hit
and which plan would fit:

```ts
if (error?.name === 'daily_quota_exceeded') {
  console.log(`${error.limit?.used}/${error.limit?.limit} sent today`);
  console.log('next plan:', error.limit?.next_plan?.name ?? 'Enterprise');
}
```

Reputation gates carry `error.reputation` (a `ReputationDetail`) saying what was
paused and whether waiting helps, and a `batch.send` that failed part way through
carries `error.sent` / `error.sent_count` — the emails that DID go out, so a retry
does not send them twice:

```ts
if (error?.reputation?.retryable) {
  console.log(`${error.reputation.scope} paused; retry at ${error.reputation.retry_at}`);
}
console.log('already sent, do not resend:', error?.sent_count ?? 0);
```

All three read as `undefined` on an ordinary error, and the full parsed body is
preserved, so a field newer than this SDK version is still reachable.

### Attachments

Attach files by hosted URL (`path`, fetched at send time) or inline base64 (`content`):

```ts
await mb.emails.send({
  from: 'Acme <hello@yourdomain.com>',
  to: ['user@example.com'],
  subject: 'Your invoice',
  html: '<p>Invoice attached.</p>',
  attachments: [
    { filename: 'invoice.pdf', path: 'https://yourdomain.com/invoices/invoice.pdf' },
    { filename: 'report.csv', content: base64Content, content_type: 'text/csv' },
  ],
});
```

### Options

```ts
const mb = new MailBlastr('mb_xxxxxxxxx', {
  baseUrl: 'https://www.mailblastr.com/api', // override your API host
  timeoutMs: 30000,  // per-request timeout (default 30s; 0 disables)
  maxRetries: 2,     // auto-retry 429/503 responses (default 2; 0 disables)
});
```

Requests time out after 30 seconds by default. A `429` (rate limited) or `503`
(service unavailable) response is retried up to `maxRetries` times, honoring the
`Retry-After` header (otherwise exponential backoff). Only those two statuses are
retried — never other errors, network failures, or timeouts — so a non-idempotent
request (like sending an email) is never duplicated by a retry.

## Resources

The client exposes one property per resource, each following a consistent
(`create` / `get` / `list` / `update` / `remove`, plus resource-specific verbs) shape:

`emails` (with nested `emails.receiving`), `batch`, `domains`, `audiences`,
`contacts`, `contactProperties`, `campaigns`, `segments`, `topics`,
`templates`, `automations`, `webhooks`, `logs`, `events`, `apiKeys`, `polls`.

`apiKeys` is the one exception: it is list-only by design — see
[API keys](#api-keys).

```ts
// Emails
await mb.emails.send({ from, to, subject, html });   // to: 1-50 recipients
await mb.emails.list({ limit: 20, after });          // cursor pagination
await mb.emails.list({ status: 'bounced', search: 'ada@' });
await mb.emails.get(id);
await mb.emails.sources();                           // per-campaign/automation metrics
await mb.emails.listAttachments(id);
await mb.emails.getAttachment(id, attachmentId);
await mb.emails.update(id, { scheduled_at });        // reschedule
await mb.emails.cancel(id);

// Inbound email — `from` must be one of your verified sending addresses
await mb.emails.receiving.list();
await mb.emails.receiving.get(id);
await mb.emails.receiving.addresses();               // per-address inbound stats
await mb.emails.receiving.forward(id, { from: 'hello@yourdomain.com', to: 'team@you.com' });
await mb.emails.receiving.reply(id, { from: 'hello@yourdomain.com', text: 'Thanks!' });

// Batch send (alias of mb.emails.batch)
await mb.batch.send([ /* up to 100 emails */ ]);

// Domains (incl. claiming a domain verified elsewhere)
await mb.domains.create({ name: 'example.com' });
await mb.domains.verify(id);
await mb.domains.recordsCsv(id);   // DNS records as CSV text
await mb.domains.claim({ name: 'example.com' });
await mb.domains.verifyClaim(id);

// Contacts are DOMAIN-FIRST: each sending domain has its own contact pool
// (the same address on two domains is two records with separate consent).
await mb.contacts.create({ domain: 'example.com', email, first_name });
await mb.contacts.list({ domain: 'example.com' });
await mb.contacts.get({ id });                  // by contact id (exact) …
await mb.contacts.get({ id: email, domain: 'example.com' }); // … or by email + domain
await mb.contacts.update({ id, unsubscribed: true });
await mb.contacts.remove({ id });
await mb.contacts.addToSegment(contactId, segmentId);
await mb.contacts.updateTopics(contactId, {
  topics: [{ id: 'topic_1', subscription: 'opt_in' }],
});

// Contact properties (custom fields)
await mb.contactProperties.create({ name: 'Plan', type: 'string' });

// Campaigns, Segments — also domain-first: `domain` picks the contact pool the
// campaign/segment targets. Segment names are unique per domain (reusable
// across domains), and every domain carries an auto-created "General" segment.
await mb.campaigns.create({ domain: 'example.com', from, subject, html, segment_id });
await mb.campaigns.send(id, { scheduled_at });
await mb.campaigns.stats(id);       // counts, rates, top links
await mb.campaigns.engagement(id);  // who opened / clicked / replied
await mb.segments.create({ domain: 'example.com', name, filter: { status: 'subscribed' } });
await mb.segments.list({ domain: 'example.com' });
await mb.segments.contacts(id);   // preview who matches

// Templates — every :id route also accepts the template's alias.
// Edits land on the DRAFT; sends always use the published snapshot.
await mb.templates.create({ name, subject, html });
await mb.templates.duplicate(id);
await mb.templates.publish(id);
await mb.emails.send({ from, to, template_id, variables: { first_name: 'Ada' } });

// API keys — read-only. Listing shows each key's non-secret prefix, permission,
// domain scoping and last-used time. See "API keys" below.
await mb.apiKeys.list();
```

### API keys

`apiKeys.list()` is the whole resource. There is no `create`, `update` or
`remove`, and that is deliberate: **keys are created, re-scoped and revoked in
the [MailBlastr dashboard](https://www.mailblastr.com/app/api-keys)**,
behind a signed-in session.

Because every SDK call authenticates with an API key, leaving key lifecycle out
of the SDK means a key that leaks cannot mint itself a replacement, widen its
own permission or domain scope, or revoke the keys around it — the blast radius
of a leaked key stays fixed at what that key could already do. The API enforces
the same rule: `POST /api-keys`, `PATCH /api-keys/:id` and `DELETE /api-keys/:id`
answer `403 dashboard_only` to any API-key caller.

```ts
const { data } = await mb.apiKeys.list();
data!.data.forEach((k) => console.log(k.name, k.token, k.permission, k.last_used_at));
```

### Topics

Topics let contacts manage granular subscriptions (e.g. "Product updates").

```ts
const { data: topic } = await mb.topics.create({
  domain: 'example.com', // topics belong to a sending domain
  name: 'Product updates',
  description: 'New features and releases',
  default_subscription: 'opt_in', // required, and immutable afterwards
});

await mb.topics.list({ domain: 'example.com', limit: 50 });
await mb.topics.update(topic!.id, { visibility: 'public' });
await mb.topics.remove(topic!.id);

// Subscribe/unsubscribe a contact per-topic. The whole list is validated
// before anything is written, so one bad entry rejects the call outright.
await mb.contacts.updateTopics(contactId, {
  topics: [{ id: topic!.id, subscription: 'opt_out' }],
});
```

### Webhooks

The endpoint must be `https://` and must not resolve to a private address.
Event names come from the `WebhookEvent` union (`WEBHOOK_EVENTS` lists them all).

```ts
const { data: hook } = await mb.webhooks.create({
  endpoint: 'https://yourapp.com/hooks/mailblastr',
  events: ['email.delivered', 'email.bounced', 'email.unsubscribed'],
});
// hook.signing_secret is shown ONCE, here and on rotate() — store it now.

await mb.webhooks.list();
await mb.webhooks.update(hook!.id, { status: 'disabled' }); // `events` replaces, never merges
await mb.webhooks.remove(hook!.id);

// A test delivery answers 200 even when it failed — check `ok`, not `error`.
const { data: probe } = await mb.webhooks.test(hook!.id);
if (!probe!.ok) console.error('endpoint unreachable:', probe!.error);

// Verify an incoming delivery against the RAW request body.
const { valid } = mb.webhooks.verify(rawBody, req.headers, signingSecret);
```

### Automations

Build multi-step automations triggered by events, then inspect their runs.

Every automation belongs to one of your sending domains — `domain` is required
on create, and `events.send` names the domain it targets, so the same event
name across several products can never trigger the wrong automation.

```ts
const { data: automation } = await mb.automations.create({
  name: 'Welcome series',
  domain: 'yourdomain.com',
  trigger: 'contact.created', // a plain event name, not an object
});

// Steps can only be added or edited while the automation is `disabled`.
await mb.automations.addStep(automation!.id, {
  type: 'send_email',
  config: { template_id: 'tmpl_welcome' },
});
await mb.automations.update(automation!.id, { status: 'enabled' });

// Or let AI draft the graph from a prompt (automation must be disabled)
await mb.automations.ai(automation!.id, { prompt: 'Wait 2 days, then send the welcome email' });

// Fire a custom event — only yourdomain.com's automations are triggered
await mb.events.send({ name: 'signup.completed', domain: 'yourdomain.com', email: 'user@example.com', data: { plan: 'pro' } });

// Inspect execution
const { data: runs } = await mb.automations.runs(automation!.id, { limit: 25, status: 'failed' });
await mb.automations.getRun(automation!.id, runs!.data[0].id);
```

`automations.list()` omits `steps` and `connections` — retrieve a single
automation to get its graph.

### Logs

```ts
await mb.logs.list({ limit: 100, after });
await mb.logs.get(logId);
```

### Pagination

`list()` methods accept optional cursor pagination — `{ limit?, after?, before? }`.
`limit` is an integer 1–100 (default 20), and `after`/`before` are item ids and
are mutually exclusive. Responses are `{ object: 'list', has_more, data }`;
there is no `total` and no `next_cursor` — page forward with the last row's `id`:

```ts
let after: string | undefined;
do {
  const { data } = await mb.campaigns.list({ limit: 25, after });
  after = data!.data.at(-1)?.id;
} while (/* data!.has_more */ after);
```

Some list endpoints return **everything** when you pass no pagination params at
all: `domains`, `apiKeys`, `topics`, `campaigns`, `contacts`, `segments`,
`contactProperties` and `polls`. Pass `limit` if you want a bounded page.
`templates`, `webhooks`, `audiences`, `automations`, `automations.runs` and
`events` always default to 20.

An unknown cursor is not an error — it returns an empty page with
`has_more: false`.

### Idempotency

Pass an idempotency key so a retried send is de-duplicated instead of delivered
twice.

```ts
await mb.emails.send(payload, { idempotencyKey: 'order-123' });
await mb.batch.send(payloads, { idempotencyKey: 'nightly-digest-2026-08-08' });
```

- The key must be **1–255 characters**, measured after the server trims it —
  255, not 256. The exported `IDEMPOTENCY_KEY_MAX_LENGTH` carries that number.
  The SDK sends the key verbatim and lets the **server** be the authority: an
  out-of-range key comes back as `400 invalid_idempotency_key`.
- **Only `emails.send` and `batch.send` honour the header.** Everywhere else —
  including `events.send` — it is accepted and forwarded but ignored, so a retry
  there creates a second resource. De-duplicate on your side instead.
- Reusing a key with a different payload is a `409 invalid_idempotent_request`;
  reusing it while the first call is still in flight is
  `409 concurrent_idempotent_requests`; reusing it after the first call
  completed replays the original response.

## Requirements

Node.js 18+ (uses the global `fetch`). For older runtimes pass a `fetch` implementation: `new MailBlastr(key, { fetch })`.

The API rejects any request without a `User-Agent` header (`403
validation_error`). The SDK always sends one — if you supply your own `fetch`,
do not strip it.

## Documentation

Full docs: <https://www.mailblastr.com/docs>

## License

MIT
