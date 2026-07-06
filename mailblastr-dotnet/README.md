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
    To = "user@example.com",
    Subject = "Hello from MailBlastr",
    HtmlBody = "<p>Your first email 🎉</p>",
});

Console.WriteLine($"sent {sent.Id}");
```

Non-2xx API responses throw `MailblastrException` with the parsed error shape:

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

### Attachments

Attach files by hosted URL (`Path`, fetched at send time) or inline base64 (`Content`):

```csharp
await mailblastr.EmailSendAsync(new EmailMessage
{
    From = "Acme <hello@yourdomain.com>",
    To = "user@example.com",
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
    BaseUrl = "https://api.mailblastr.com",   // override your API host
    HttpClient = myPooledHttpClient,          // e.g. from IHttpClientFactory
});
```

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
await mailblastr.EmailBatchAsync(messages);                    // up to 100 per request
await mailblastr.EmailListAsync(new PaginationOptions { Limit = 20 });
await mailblastr.EmailRetrieveAsync(id);
await mailblastr.EmailListAttachmentsAsync(id);
await mailblastr.EmailRescheduleAsync(id, "2026-08-01T09:00:00Z");
await mailblastr.EmailCancelAsync(id);

// Inbound email
await mailblastr.ReceivedEmailListAsync();
await mailblastr.ReceivedEmailRetrieveAsync(id);
byte[] file = await mailblastr.ReceivedEmailDownloadAttachmentAsync(id, attachmentId);
await mailblastr.ReceivedEmailForwardAsync(id, new ReceivedEmailForwardOptions
{
    From = "you@yourdomain.com", To = "team@you.com",
});

// Domains (incl. claiming a domain verified elsewhere)
var domain = await mailblastr.DomainCreateAsync(new DomainCreateOptions { Name = "example.com" });
await mailblastr.DomainVerifyAsync(domain.Id);
await mailblastr.DomainClaimAsync(new DomainClaimOptions { Name = "example.com" });
await mailblastr.DomainVerifyClaimAsync(domain.Id);
await mailblastr.DomainDetectDnsAsync(domain.Id);              // one-click DNS options
await mailblastr.DomainApplyCloudflareDnsAsync(domain.Id, cloudflareToken);
```

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

var segment = await mailblastr.SegmentCreateAsync(new SegmentCreateOptions
{
    Domain = "example.com",
    Name = "Subscribed",
    Filter = new SegmentFilter { Status = "subscribed" },
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
    From = from, To = to, Subject = "ignored-when-template-sets-it",
    TemplateId = tmpl.Id,
    Variables = new() { ["first_name"] = "Ada" },
});
```

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

// Inspect execution
var runs = await mailblastr.AutomationListRunsAsync(automation.Id, new PaginationOptions { Limit = 25 });
await mailblastr.AutomationRetrieveRunAsync(automation.Id, runs.Data[0].Id);
```

### Webhooks & signature verification

```csharp
var hook = await mailblastr.WebhookCreateAsync(new WebhookCreateOptions
{
    Endpoint = "https://yourapp.com/hooks/mailblastr",
    Events = { "email.delivered", "email.bounced", "contact.unsubscribed" },
});
// hook.SigningSecret is shown ONCE — store it now.
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

(Also available as `mailblastr.WebhookVerifySignature(...)` on the client.)

### Logs, API keys, polls

```csharp
await mailblastr.LogListAsync(new LogListOptions { Limit = 100, Method = "POST", Status = 429 });
await mailblastr.ApiKeyCreateAsync(new ApiKeyCreateOptions { Name = "CI", Permission = "sending_access" });
await mailblastr.PollListAsync();
await mailblastr.PollRetrieveAsync(emailId);
```

### Pagination

`*ListAsync` methods accept optional cursor pagination:

```csharp
await mailblastr.CampaignListAsync(new PaginationOptions { Limit = 25, After = "cursor_abc" });
```

### Idempotency

Pass an idempotency key to safely retry a create (24h window):

```csharp
await mailblastr.EmailSendAsync(message, idempotencyKey: "order-123");
```

## Documentation

Full docs: <https://www.mailblastr.com/docs>

## License

MIT
