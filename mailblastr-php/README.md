# mailblastr-php

Official PHP SDK for the [MailBlastr](https://www.mailblastr.com) email API — send transactional and marketing email from your own verified domain.

## Requirements

PHP 8.1+ with the `curl` and `json` extensions. No other runtime dependencies.

## Install

```bash
composer require mailblastr/mailblastr
```

## Usage

```php
use Mailblastr\Mailblastr;

$mailblastr = Mailblastr::client('mb_xxxxxxxxx');

$sent = $mailblastr->emails->send([
    'from' => 'Acme <hello@yourdomain.com>',
    'to' => ['user@example.com'],
    'subject' => 'Hello from MailBlastr',
    'html' => '<p>Your first email 🎉</p>',
]);

echo 'sent ' . $sent['id'];
```

Every method returns the decoded JSON response as an associative array. Any non-2xx
response throws `Mailblastr\Exceptions\MailblastrException` carrying the API error
shape:

```php
use Mailblastr\Exceptions\MailblastrException;

try {
    $mailblastr->emails->send($payload);
} catch (MailblastrException $e) {
    echo $e->getStatusCode(); // 422
    echo $e->getName();       // 'validation_error'
    echo $e->getMessage();    // 'domain is required'
}
```

### Attachments

Attach files by hosted URL (`path`, fetched at send time) or inline base64 (`content`):

```php
$mailblastr->emails->send([
    'from' => 'Acme <hello@yourdomain.com>',
    'to' => ['user@example.com'],
    'subject' => 'Your invoice',
    'html' => '<p>Invoice attached.</p>',
    'attachments' => [
        ['filename' => 'invoice.pdf', 'path' => 'https://yourdomain.com/invoices/invoice.pdf'],
        ['filename' => 'report.csv', 'content' => $base64Content, 'content_type' => 'text/csv'],
    ],
]);
```

### Options

```php
$mailblastr = Mailblastr::client('mb_xxxxxxxxx', [
    'baseUrl' => 'https://api.mailblastr.com', // override your API host
]);
```

## Resources

The client exposes one property per resource, each following a consistent
(`create` / `get` / `list` / `update` / `remove`, plus resource-specific verbs) shape:

`emails` (with nested `emails->attachments` and `emails->receiving`), `batch`,
`domains`, `audiences`, `contacts`, `contactProperties`, `campaigns`, `segments`,
`topics`, `templates`, `automations`, `webhooks`, `logs`, `events`, `apiKeys`, `polls`.

```php
// Emails
$mailblastr->emails->send(['from' => …, 'to' => …, 'subject' => …, 'html' => …]);
$mailblastr->emails->list(['limit' => 20, 'after' => $cursor]); // cursor pagination
$mailblastr->emails->get($id);
$mailblastr->emails->attachments->list(emailId: $id);
$mailblastr->emails->attachments->get(emailId: $id, attachmentId: $attachmentId);
$mailblastr->emails->update($id, ['scheduled_at' => $ts]);      // reschedule
$mailblastr->emails->cancel($id);

// Inbound email
$mailblastr->emails->receiving->list();
$mailblastr->emails->receiving->get($id);
$mailblastr->emails->receiving->forward($id, ['from' => 'me@yourdomain.com', 'to' => 'team@you.com']);
$mailblastr->emails->receiving->reply($id, ['from' => 'me@yourdomain.com', 'text' => 'Thanks!']);
$bytes = $mailblastr->emails->receiving->getAttachment($id, $attachmentId); // raw bytes
$mime = $mailblastr->emails->receiving->getRaw($id);                        // raw RFC822

// Batch send (alias of $mailblastr->emails->batch())
$mailblastr->batch->send([ /* up to 100 email payloads */ ]);

// Domains (incl. claiming a domain verified elsewhere + one-click DNS)
$mailblastr->domains->create(['name' => 'example.com']);
$mailblastr->domains->verify($id);
$mailblastr->domains->claim(['name' => 'example.com']);
$mailblastr->domains->verifyClaim($id);
$mailblastr->domains->detectDns($id);
$mailblastr->domains->applyCloudflareDns($id, ['token' => $cfToken]);

// Contacts are DOMAIN-FIRST: each sending domain has its own contact pool
// (the same address on two domains is two records with separate consent).
$mailblastr->contacts->create(['domain' => 'example.com', 'email' => $email, 'first_name' => 'Ada']);
$mailblastr->contacts->list(['domain' => 'example.com']);
$mailblastr->contacts->get(['id' => $contactId]);                            // by contact id (exact) …
$mailblastr->contacts->get(['id' => $email, 'domain' => 'example.com']);     // … or by email + domain
$mailblastr->contacts->update(['id' => $contactId, 'unsubscribed' => true]);
$mailblastr->contacts->remove(['id' => $contactId]);
$mailblastr->contacts->batch(['audienceId' => $audienceId, 'contacts' => [ … ]]);
$mailblastr->contacts->import(['audienceId' => $audienceId, 'csv' => $csvText]);
$mailblastr->contacts->addToSegment($contactId, $segmentId);
$mailblastr->contacts->updateTopics($contactId, ['topics' => [['id' => 'top_1', 'subscription' => 'opt_in']]]);

// Contact properties (custom fields)
$mailblastr->contactProperties->create(['key' => 'plan', 'type' => 'string']);

// Campaigns, Segments — also domain-first: 'domain' picks the contact pool the
// campaign/segment targets. Segment names are unique per domain (reusable
// across domains), and every domain carries an auto-created "General" segment.
$mailblastr->campaigns->create(['domain' => 'example.com', 'from' => …, 'subject' => …, 'html' => …, 'segment_id' => $segmentId]);
$mailblastr->campaigns->send($id, ['scheduled_at' => $ts]);
$mailblastr->segments->create(['domain' => 'yourdomain.com', 'name' => 'VIP']);
$mailblastr->segments->list(['domain' => 'yourdomain.com']);
$mailblastr->segments->contacts($id);   // preview who matches

// Templates
$mailblastr->templates->create(['name' => $name, 'subject' => $subject, 'html' => $html]);
$mailblastr->templates->duplicate($id);
$mailblastr->templates->publish($id);
$mailblastr->emails->send(['from' => …, 'to' => …, 'template_id' => $templateId, 'variables' => ['first_name' => 'Ada']]);

// Audiences (incl. Google Sheet import)
$mailblastr->audiences->importSheet($audienceId, ['url' => $sheetUrl]);

// API keys
$mailblastr->apiKeys->create(['name' => 'CI', 'permission' => 'sending_access']);
$mailblastr->apiKeys->list();
$mailblastr->apiKeys->remove($id);

// Polls (read-only results of the in-email poll widget)
$mailblastr->polls->list();
$mailblastr->polls->get($emailId);
```

### Topics

Topics let contacts manage granular subscriptions (e.g. "Product updates").

```php
$topic = $mailblastr->topics->create([
    'domain' => 'example.com', // topics belong to a sending domain
    'name' => 'Product updates',
    'description' => 'New features and releases',
    'default_subscription' => 'opt_in',
]);

$mailblastr->topics->list(['domain' => 'example.com', 'limit' => 50]);
$mailblastr->topics->update($topic['id'], ['visibility' => 'private']);
$mailblastr->topics->remove($topic['id']);
```

### Automations

Build multi-step automations triggered by events, then inspect their runs.

Every automation belongs to one of your sending domains — `domain` is REQUIRED
on create, and `events->send()` names the domain it targets, so the same event
name across several products can never trigger the wrong automation.

```php
$automation = $mailblastr->automations->create([
    'name' => 'Welcome series',
    'domain' => 'yourdomain.com',
    'trigger' => 'contact.created',
]);

$mailblastr->automations->addStep($automation['id'], [
    'type' => 'send_email',
    'config' => ['template_id' => 'tmpl_welcome'],
]);
$mailblastr->automations->update($automation['id'], ['status' => 'enabled']);

// Fire a custom event — only yourdomain.com's automations are triggered
$mailblastr->events->send([
    'event' => 'signup.completed',
    'domain' => 'yourdomain.com',
    'email' => 'user@example.com',
    'payload' => ['plan' => 'pro'],
]);

// Inspect execution
$runs = $mailblastr->automations->runs($automation['id'], ['limit' => 25]);
$mailblastr->automations->getRun($automation['id'], $runs['data'][0]['id']);
```

### Webhooks

```php
$hook = $mailblastr->webhooks->create([
    'endpoint' => 'https://yourapp.com/hooks/mailblastr',
    'events' => ['email.delivered', 'email.bounced', 'contact.unsubscribed'],
]);
// $hook['signing_secret'] is shown ONCE — store it now.

$mailblastr->webhooks->list();
$mailblastr->webhooks->update($hook['id'], ['status' => 'disabled']);
$mailblastr->webhooks->rotate($hook['id']); // new signing_secret, revealed once
```

Verify a delivery's Svix-style signature in your endpoint (a pure local
HMAC-SHA256 computation — no HTTP request). Pass the EXACT raw request body:

```php
$payload = file_get_contents('php://input'); // do NOT re-serialize parsed JSON

$result = $mailblastr->webhooks->verify($payload, [
    'svix-id' => $_SERVER['HTTP_SVIX_ID'] ?? '',
    'svix-timestamp' => $_SERVER['HTTP_SVIX_TIMESTAMP'] ?? '',
    'svix-signature' => $_SERVER['HTTP_SVIX_SIGNATURE'] ?? '',
], $signingSecret);

if (!$result['valid']) {
    http_response_code(401); // $result['reason'] says why
}
```

Also available without a client: `Mailblastr\WebhookSignature::verify(...)`.

### Logs

```php
$mailblastr->logs->list(['limit' => 100, 'method' => 'POST', 'status' => 429]);
$mailblastr->logs->get($logId);
```

### Pagination

`list()` methods accept optional cursor pagination — `['limit' => …, 'after' => …, 'before' => …]` —
appended as a query string:

```php
$mailblastr->campaigns->list(['limit' => 25, 'after' => 'cursor_abc']);
```

### Idempotency

Pass an idempotency key to safely retry a create:

```php
$mailblastr->emails->send($payload, ['idempotencyKey' => 'order-123']);
```

## Testing

The HTTP transport is swappable — pass any `Mailblastr\Transport\TransportInterface`
implementation as `'transport'` to fake responses in your tests. The SDK's own
test suite (no framework needed) runs with:

```bash
php tests/run.php
```

## Documentation

Full docs: <https://www.mailblastr.com/docs>

## License

MIT
