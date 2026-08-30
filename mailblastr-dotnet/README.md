# Mailblastr .NET SDK

Official .NET SDK for the [MailBlastr](https://www.mailblastr.com) email API — send transactional and marketing email from your own verified domain.

## Install

```bash
dotnet add package Mailblastr
```

Targets **.NET 8.0+**. No dependencies beyond the BCL (`HttpClient` + `System.Text.Json`).

## Usage

```csharp
using Mailblastr;

IMailblastr mailblastr = MailblastrClient.Create("mb_xxxxxxxxx");

var sent = await mailblastr.EmailSendAsync(new EmailMessage
{
    From = "Acme <hello@yourdomain.com>",
    To = "delivered@mailblastr.dev",
    Subject = "Hello from MailBlastr",
    HtmlBody = "<p>Your first email 🎉</p>",
});

Console.WriteLine($"sent {sent.Id}");
```

Non-2xx API responses throw `MailblastrException` carrying the API's
`{ statusCode, name, message }` error envelope:

```csharp
try
{
    await mailblastr.EmailSendAsync(message);
}
catch (MailblastrException ex)
{
    Console.Error.WriteLine($"{ex.StatusCode} {ex.Name}: {ex.Message}");
}
```

Branch on `ex.Name` (e.g. `validation_error`, `not_found`,
`daily_quota_exceeded`), and read the real HTTP status from `ex.StatusCode` —
a handler may return a name with a status other than its usual one. Never
pattern-match `ex.Message`: the API scrubs provider identifiers out of it.

Some errors carry more than the envelope. Those extras are typed on the
exception and are `null` on an ordinary error:

```csharp
catch (MailblastrException ex)
{
    // WHICH quota ran out, and what would clear it.
    if (ex.Limit is { } limit)
    {
        Console.Error.WriteLine($"{limit.Kind}: {limit.Used}/{limit.Limit} used" +
                                $" over {limit.Period}, upgrade to {limit.NextPlan?.Name}");
    }

    // Reputation gates: whether waiting helps, and until when.
    if (ex.Reputation is { } rep && rep.Retryable)
    {
        Console.Error.WriteLine($"paused on {rep.Scope} until {rep.RetryAt}");
    }

    // A batch that failed part way through — do NOT resend these.
    if (ex.Sent is { } sent)
    {
        Console.Error.WriteLine($"{ex.SentCount} already went out: {string.Join(", ", sent.Select(e => e.Id))}");
    }
}
```

`ex.Extra` holds every additive field unparsed, so a field newer than this SDK
version is still reachable.

Every request carries `Authorization: Bearer <your key>` and a non-empty
`User-Agent` (`mailblastr-dotnet/<version>`). The API rejects a request with a
missing User-Agent with HTTP 403 `validation_error`, so the header is not
configurable.

### Attachments

Attach files by hosted URL (`Path`, fetched at send time) or inline base64 (`Content`):

```csharp
await mailblastr.EmailSendAsync(new EmailMessage
{
    From = "Acme <hello@yourdomain.com>",
    To = "delivered@mailblastr.dev",
    Subject = "Your invoice",
    HtmlBody = "<p>Invoice attached.</p>",
    Attachments = new List<EmailAttachment>
    {
        new() { Filename = "invoice.pdf", Path = "https://yourdomain.com/invoices/invoice.pdf" },
        new() { Filename = "report.csv", Content = base64Content, ContentType = "text/csv" },
    },
});
```

### Options

```csharp
IMailblastr mailblastr = MailblastrClient.Create("mb_xxxxxxxxx", new MailblastrClientOptions
{
    BaseUrl = "https://www.mailblastr.com/api",   // override your API host
    HttpClient = myPooledHttpClient,              // e.g. from IHttpClientFactory
    Timeout = TimeSpan.FromSeconds(30),           // per attempt; Zero disables
    MaxRetries = 2,                               // 429/503 only; 0 disables
});
```

`MailblastrClient` is thread-safe and meant to be created once and reused for the
process lifetime. It implements `IDisposable`, but disposes the `HttpClient` only
when it created it — a client you supply via `HttpClient` stays yours.

### Timeouts, retries & rate limits

`Timeout` (default 30s) applies to **each attempt**, so a retry gets a fresh
budget. Exceeding it throws a `MailblastrException` with `Name == "timeout"` and
`StatusCode == 0`; a connection failure throws `Name == "network_error"`, also
with `StatusCode == 0`. Neither is retried — the request may already have been
applied.

The client retries **only HTTP 429 and 503**, up to `MaxRetries` times (default
2, so 3 attempts total). It honours the response's `Retry-After` header when
present — delta-seconds or an HTTP-date — and otherwise backs off exponentially
(`0.5s * 2^attempt`); either way a single wait is capped at 30s. Every other
status, including 5xx other than 503, is surfaced immediately.

The `/emails` **send** routes are capped at 30 requests per 60 seconds per client
IP, returning 429 `rate_limit_exceeded`. Reads are not subject to that
per-request cap, so paging a large list will not trip it. Responses on the capped
mount carry `RateLimit-Limit` / `RateLimit-Remaining` / `RateLimit-Reset`, plus
`Retry-After` on the 429 itself — read them to throttle ahead of time instead of
discovering the cap through rejections. Sustained volume is governed separately
by your plan's send quota, which arrives as a limit error (see `ex.Limit`) rather
than a 429 rate-limit rejection.

## Resources

One flat async method per operation, following a consistent
`<Resource><Verb>Async` naming (`Create` / `Retrieve` / `List` / `Update` /
`Delete`, plus resource-specific verbs):

emails (incl. batch, attachments and inbound `ReceivedEmail*`), domains
(incl. claim + one-click DNS), audiences, contacts (domain-first flat +
nested + batch + CSV/Sheet import + segment membership + topics), contact
properties, campaigns, segments, topics, templates, automations, webhooks,
events, api keys, logs, polls.

```csharp
// Emails
await mailblastr.EmailSendAsync(message);
await mailblastr.EmailBatchAsync(messages);                    // List<BatchEmailMessage>, max 100
await mailblastr.EmailBatchSendAsync(messages);                // same call, full response (Queued)
await mailblastr.EmailListAsync(new PaginationOptions { Limit = 20 });
await mailblastr.EmailListAsync(new EmailListOptions            // server-side filters
{
    Status = "bounced", Search = "invoice", DomainId = domainId,
});
await mailblastr.EmailListAsync(new EmailListOptions            // mailbox folder: one of the
{                                                               // EmailFolders constants (Outbox /
    Folder = EmailFolders.Scheduled,                            // Sent / Scheduled / Failed);
});                                                             // any other value is rejected
await mailblastr.EmailListSourcesAsync();                       // per-origin send counters
await mailblastr.EmailRetrieveAsync(id);
await mailblastr.EmailListAttachmentsAsync(id);
await mailblastr.EmailUpdateAsync(id, "2026-08-01T09:00:00Z"); // reschedule a queued send
await mailblastr.EmailCancelAsync(id);

// Inbound email
await mailblastr.ReceivedEmailListAsync();
await mailblastr.ReceivedEmailListAddressesAsync();             // per-address inbound stats
await mailblastr.ReceivedEmailRetrieveAsync(id);
byte[] file = await mailblastr.ReceivedEmailGetAttachmentAsync(id, attachmentId);
await mailblastr.ReceivedEmailForwardAsync(id, new ReceivedEmailForwardOptions
{
    From = "you@yourdomain.com", To = "delivered@mailblastr.dev",
});

// Domains (incl. claiming a domain verified elsewhere)
var domain = await mailblastr.DomainCreateAsync(new DomainCreateOptions { Name = "example.com" });
await mailblastr.DomainVerifyAsync(domain.Id);
await mailblastr.DomainClaimAsync(new DomainClaimOptions { Name = "example.com" });
await mailblastr.DomainVerifyClaimAsync(domain.Id);
await mailblastr.DomainDetectDnsAsync(domain.Id);              // one-click DNS options
await mailblastr.DomainApplyCloudflareDnsAsync(domain.Id, cloudflareToken);
await mailblastr.DomainMxCheckAsync("example.com");            // MX preflight for inbound
string csv = await mailblastr.DomainRetrieveRecordsCsvAsync(domain.Id);
```

Pass `BatchEmailMessage` items to `EmailBatchAsync` — the batch endpoint rejects
`Attachments` and `ScheduledAt` per item, and that type simply omits them so the
rule holds at compile time. The older `IEnumerable<EmailMessage>` overload is
`[Obsolete]` for exactly that reason; send attachments or scheduled messages one
at a time through `EmailSendAsync`.

A batch is served one of two ways, chosen by SIZE alone, and both are success —
so if you need to know whether the mail is actually out, call
`EmailBatchSendAsync` and read `Queued`:

```csharp
BatchSendResponse batch = await mailblastr.EmailBatchSendAsync(messages, idempotencyKey: "nightly-2026-08-19");

if (batch.Queued)
{
    // 41-100 emails: accepted (HTTP 202) and delivered in the BACKGROUND. The
    // ids are real — EmailRetrieveAsync(id) works — but they start at
    // `scheduled` and nothing has been transmitted yet.
    Console.WriteLine($"{batch.QueuedCount} emails queued");
}
else
{
    // 1-40 emails: already handed to the mail service before the call returned.
    Console.WriteLine($"{batch.Data.Count} emails sent");
}
```

`EmailBatchAsync` returns `batch.Data` and nothing else, so it cannot report the
queued case. An inline batch near the 40 boundary can also take ~100s
server-side, past the 30s default `MailblastrClientOptions.Timeout` — raise it
for batches that large, and always pass an idempotency key so a client that
gives up mid-request can replay the recorded answer instead of re-sending.

### Contacts are DOMAIN-FIRST

Each sending domain has its own contact pool — the same address on two
domains is two records with separate consent. `Domain` is required on the
flat contacts API (and on segments, topics, campaigns, automations and
`EventSendAsync`):

```csharp
await mailblastr.ContactCreateAsync(new ContactCreateOptions
{
    Domain = "example.com",              // the pool the contact lands in
    Email = "user@example.com",
    FirstName = "Ada",
});
await mailblastr.ContactListAsync(new ContactListOptions { Domain = "example.com" });
await mailblastr.ContactRetrieveAsync(contactId);                        // by id (exact) …
await mailblastr.ContactRetrieveAsync("user@example.com", "example.com"); // … or by email + domain
await mailblastr.ContactUpdateAsync(new ContactUpdateOptions { Id = contactId, Unsubscribed = true });
await mailblastr.ContactDeleteAsync(contactId);

// Bulk import (array or CSV), segment membership and per-topic subscriptions
await mailblastr.ContactBatchAsync(audienceId, contacts, onConflict: "skip");
// Domain-first: import straight into a domain's pool, no audience id needed.
await mailblastr.ContactBatchInDomainAsync("yourdomain.com", contacts);
await mailblastr.ContactImportAsync(audienceId, csvText);
await mailblastr.ContactAddToSegmentAsync(contactId, segmentId);
await mailblastr.ContactUpdateTopicsAsync(contactId, new ContactTopicsUpdateOptions
{
    Topics = { new ContactTopicSetting { Id = topicId, Subscription = "opt_in" } },
});

// Contact properties (custom fields / merge tags)
await mailblastr.ContactPropertyCreateAsync(new ContactPropertyCreateOptions { Key = "plan", Type = "string" });
```

### Campaigns, segments, topics (domain-required)

```csharp
var campaign = await mailblastr.CampaignCreateAsync(new CampaignCreateOptions
{
    Domain = "example.com",              // REQUIRED: the contact pool it targets
    From = "Acme <hello@example.com>",
    Subject = "Product news",
    HtmlBody = "<p>…</p>",
    SegmentId = segmentId,               // optional subset
});
await mailblastr.CampaignSendAsync(campaign.Id);                  // or scheduledAt: "2026-08-01T09:00:00Z"
await mailblastr.CampaignRetrieveStatsAsync(campaign.Id);
await mailblastr.CampaignRetrieveEngagementAsync(campaign.Id);    // who opened/clicked/replied

var segment = await mailblastr.SegmentCreateAsync(new SegmentCreateOptions
{
    Domain = "example.com",
    Name = "Subscribed",
    Filter = new SegmentFilterOptions { Status = "subscribed" },
});
await mailblastr.SegmentListAsync("example.com");
await mailblastr.SegmentListContactsAsync(segment.Id);            // preview who matches

await mailblastr.TopicCreateAsync(new TopicCreateOptions
{
    Domain = "example.com",
    Name = "Product updates",
    DefaultSubscription = "opt_in",
});
```

### Templates

```csharp
var tmpl = await mailblastr.TemplateCreateAsync(new TemplateCreateOptions
{
    Name = "Welcome", Subject = "Welcome, {{first_name}}!", HtmlBody = "<p>Hi {{first_name}}</p>",
});
await mailblastr.TemplatePublishAsync(tmpl.Id);
await mailblastr.EmailSendAsync(new EmailMessage
{
    From = from, To = to,                // omit Subject -> the template's is used
    TemplateId = tmpl.Id,
    Variables = new() { ["first_name"] = "Ada" },
});
```

The template only fills fields you leave out: omit `From` / `ReplyTo` / `Subject`
and its own values apply, but a `Subject` you *do* pass wins over the template's.
Templates must be published before they can be sent (an unpublished one is a
422 `validation_error`), and pairing `TemplateId`/`Template` with `HtmlBody` or
`TextBody` is rejected — pick one.

### Automations & events

Every automation belongs to one of your sending domains — `Domain` is
required on create, and `EventSendAsync` names the domain it targets, so the
same event name across several products can never trigger the wrong automation.

```csharp
var automation = await mailblastr.AutomationCreateAsync(new AutomationCreateOptions
{
    Name = "Welcome series",
    Domain = "yourdomain.com",
    Trigger = "contact.created",
});
await mailblastr.AutomationAddStepAsync(automation.Id, new AutomationAddStepOptions
{
    Type = "send_email",
    Config = new() { ["template_id"] = "tmpl_welcome" },
});

// Fire a custom event — only yourdomain.com's automations are triggered
await mailblastr.EventSendAsync(new EventSendOptions
{
    Name = "signup.completed",
    Domain = "yourdomain.com",
    Email = "user@example.com",
    Data = new() { ["plan"] = "pro" },
});

// Inspect execution (optionally filtered by run status)
var runs = await mailblastr.AutomationListRunsAsync(automation.Id, new AutomationRunListOptions
{
    Limit = 25, Status = new[] { "failed", "running" },
});
await mailblastr.AutomationRetrieveRunAsync(automation.Id, runs.Data[0].Id);
```

### Webhooks & signature verification

```csharp
var hook = await mailblastr.WebhookCreateAsync(new WebhookCreateOptions
{
    Endpoint = "https://yourapp.com/hooks/mailblastr",
    Events = { "email.delivered", "email.bounced", "email.unsubscribed" },
});
// hook.SigningSecret is shown ONCE — store it now.
```

Endpoints must be `https://` and must not resolve to a private address. Event
names come from a fixed vocabulary — `email.sent`, `email.delivered`,
`email.delivery_delayed`, `email.bounced`, `email.complained`, `email.opened`,
`email.clicked`, `email.failed`, `email.scheduled`, `email.suppressed`,
`email.received`, `email.replied`, `email.unsubscribed`, `contact.created`,
`contact.updated`, `contact.deleted`, `domain.created`, `domain.updated`,
`domain.deleted` — anything else is a 422.

`WebhookTestAsync` returns HTTP 200 even when the delivery failed — it does not
throw. The outcome is `result.Ok`, with `result.Status` (your endpoint's HTTP
status, when it responded) and `result.Error` (e.g. `lookup_failed`):

```csharp
var test = await mailblastr.WebhookTestAsync(webhookId);
if (!test.Ok)
{
    Console.Error.WriteLine($"test delivery failed: {test.Error} (status {test.Status})");
}
```

Verify a delivery in your endpoint (pure local HMAC-SHA256; pass the EXACT raw
request body, not re-serialized JSON):

```csharp
var result = WebhookSignature.Verify(rawBody, new Dictionary<string, string>
{
    ["svix-id"] = Request.Headers["svix-id"]!,
    ["svix-timestamp"] = Request.Headers["svix-timestamp"]!,
    ["svix-signature"] = Request.Headers["svix-signature"]!,
}, signingSecret);

if (!result.Valid) return Unauthorized(result.Reason);
```

(Also available as `mailblastr.WebhookVerify(...)` on the client.)

### Logs, API keys, polls

```csharp
await mailblastr.LogListAsync(new LogListOptions { Limit = 100, Method = "POST", Status = 429 });
await mailblastr.ApiKeyListAsync();
await mailblastr.PollListAsync();
await mailblastr.PollRetrieveAsync(emailId);
```

**API keys are created, re-scoped and revoked in the dashboard**, at
[mailblastr.com/app/api-keys](https://www.mailblastr.com/app/api-keys). Those
routes accept a signed-in dashboard session only, so `ApiKeyListAsync` is the
entire API-key surface of this SDK. That is the point: a key that leaks cannot
mint itself a replacement, widen its own permission, or revoke the keys around
it — the blast radius stays whatever the leaked key could already do.

An API key's `Token` is only ever the 8-character display prefix (e.g.
`mb_ab12`); the full secret is shown once, in the dashboard, at creation.

### Pagination

`*ListAsync` methods accept optional cursor pagination:

```csharp
await mailblastr.CampaignListAsync(new PaginationOptions { Limit = 25, After = "cursor_abc" });
```

`limit` is an integer 1–100 (default 20); `After` and `Before` are item ids and
are mutually exclusive. Responses are `{ object: "list", has_more, data }` —
there is no total and no next-cursor: page forward with the last row's `Id`.
Note that several list endpoints (domains, api keys, topics, campaigns,
contacts, segments, contact properties, polls) ignore the default of 20 when you
pass no pagination options at all and return up to **1,000** rows in one
response, while templates, webhooks, audiences, automations, automation runs and
events cap at 20. That 1,000 is a ceiling, not a promise: `HasMore` stays
truthful when it bites, so always page with `After` rather than assuming one
call drained the collection.

### Idempotency

`EmailSendAsync` and `EmailBatchAsync` accept an idempotency key so a retry
cannot send twice:

```csharp
await mailblastr.EmailSendAsync(message, idempotencyKey: "order-123");
```

- The key is **1–255 characters**, measured after the server trims it — 255, not
  256. `MailblastrClient.MaxIdempotencyKeyLength` carries that number. The client
  sends the key verbatim and lets the **server** be the authority: an
  out-of-range key comes back as 400 `invalid_idempotency_key`
  (a `MailblastrException`), not a local `ArgumentException`.
- Replaying a completed key returns the original response. Reusing it with a
  different payload is 409 `invalid_idempotent_request`; while the first request
  is still in flight it is 409 `concurrent_idempotent_requests`.
- **Only `POST /emails` and `POST /emails/batch` honour the header.** The
  `idempotencyKey` parameter on `EventSendAsync` / `EventCreateAsync` is sent but
  ignored by the API — retrying those creates a second record. De-duplicate on
  your side instead.

## Documentation

Full docs: <https://www.mailblastr.com/docs>

## License

MIT
