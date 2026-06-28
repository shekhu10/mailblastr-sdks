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

Every method returns `{ data, error }` — `error` is `null` on success, or `{ statusCode, name, message }` on failure (no exceptions to catch for API errors).

### Options

```ts
const mb = new MailBlastr('mb_xxxxxxxxx', {
  baseUrl: 'https://api.mailblastr.com', // override your API host
});
```

## Resources

The client exposes one property per resource, each following a consistent
(`create` / `get` / `list` / `update` / `remove`, plus resource-specific verbs) shape:

`emails` (with nested `emails.receiving`), `batch`, `domains`, `audiences`,
`contacts`, `contactProperties`, `broadcasts`, `segments`, `topics`,
`templates`, `automations`, `webhooks`, `logs`, `events`, `apiKeys`.

```ts
// Emails
await mb.emails.send({ from, to, subject, html });
await mb.emails.list({ limit: 20, after });         // cursor pagination
await mb.emails.get(id);
await mb.emails.listAttachments(id);
await mb.emails.getAttachment(id, attachmentId);
await mb.emails.update(id, { scheduled_at });        // reschedule
await mb.emails.cancel(id);

// Inbound email
await mb.emails.receiving.list();
await mb.emails.receiving.get(id);
await mb.emails.receiving.forward(id, { to: 'team@you.com' });

// Batch send (alias of mb.emails.batch)
await mb.batch.send([ /* up to 100 emails */ ]);

// Domains (incl. claiming a domain verified elsewhere)
await mb.domains.create({ name: 'example.com' });
await mb.domains.verify(id);
await mb.domains.claim({ name: 'example.com' });
await mb.domains.verifyClaim(id);

// Audiences & Contacts
await mb.audiences.create({ name: 'Newsletter' });
await mb.contacts.create({ audienceId, email, first_name });
await mb.contacts.list({ audienceId });
await mb.contacts.addToSegment(contactId, segmentId);
await mb.contacts.updateTopics(contactId, { subscribed: ['topic_1'] });

// Contact properties (custom fields)
await mb.contactProperties.create({ name: 'Plan', type: 'string' });

// Broadcasts, Segments
await mb.broadcasts.create({ audience_id, from, subject, html, segment_id });
await mb.broadcasts.send(id, { scheduled_at });
await mb.segments.create({ audience_id, name, filter: { status: 'subscribed' } });
await mb.segments.contacts(id);   // preview who matches

// Templates
await mb.templates.create({ name, subject, html });
await mb.templates.duplicate(id);
await mb.templates.publish(id);
await mb.emails.send({ from, to, template_id, variables: { first_name: 'Ada' } });

// API keys
await mb.apiKeys.create({ name: 'CI', permission: 'sending_access' });
await mb.apiKeys.list();
await mb.apiKeys.remove(id);
```

### Topics

Topics let contacts manage granular subscriptions (e.g. "Product updates").

```ts
const { data: topic } = await mb.topics.create({
  name: 'Product updates',
  description: 'New features and releases',
  default_subscribed: true,
});

await mb.topics.list({ limit: 50 });
await mb.topics.update(topic!.id, { default_subscribed: false });
await mb.topics.remove(topic!.id);

// Subscribe/unsubscribe a contact per-topic
await mb.contacts.updateTopics(contactId, {
  subscribed: [topic!.id],
  unsubscribed: [],
});
```

### Webhooks

```ts
const { data: hook } = await mb.webhooks.create({
  endpoint: 'https://yourapp.com/hooks/mailblastr',
  events: ['email.delivered', 'email.bounced', 'contact.unsubscribed'],
});

await mb.webhooks.list();
await mb.webhooks.update(hook!.id, { status: 'disabled' });
await mb.webhooks.remove(hook!.id);
```

### Automations

Build multi-step automations triggered by events, then inspect their runs.

```ts
const { data: automation } = await mb.automations.create({
  name: 'Welcome series',
  trigger: { type: 'contact.created' },
});

await mb.automations.addStep(automation!.id, {
  type: 'send_email',
  config: { template_id: 'tmpl_welcome' },
});
await mb.automations.update(automation!.id, { status: 'active' });

// Fire a custom event automations can trigger on
await mb.events.send({ name: 'signup.completed', email: 'user@example.com', data: { plan: 'pro' } });

// Inspect execution
const { data: runs } = await mb.automations.runs(automation!.id, { limit: 25 });
await mb.automations.getRun(automation!.id, runs!.data[0].id);
```

### Logs

```ts
await mb.logs.list({ limit: 100, after });
await mb.logs.get(logId);
```

### Pagination

`list()` methods accept optional cursor pagination — `{ limit?, after?, before? }` —
appended as a query string:

```ts
await mb.broadcasts.list({ limit: 25, after: 'cursor_abc' });
```

### Idempotency

Pass an idempotency key to safely retry a create:

```ts
await mb.emails.send(payload, { idempotencyKey: 'order-123' });
```

## Requirements

Node.js 18+ (uses the global `fetch`). For older runtimes pass a `fetch` implementation: `new MailBlastr(key, { fetch })`.

## Documentation

Full docs: <https://www.mailblastr.com/docs>

## License

MIT
