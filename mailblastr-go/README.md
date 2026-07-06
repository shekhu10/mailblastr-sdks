# mailblastr-go

Official Go SDK for the [MailBlastr](https://www.mailblastr.com) email API — send transactional and marketing email from your own verified domain.

## Install

```bash
go get github.com/mailblastr/mailblastr-go
```

Requires Go 1.22+. The SDK depends only on the Go standard library.

## Usage

First, get an API key from the MailBlastr dashboard.

```go
package main

import (
	"fmt"

	"github.com/mailblastr/mailblastr-go"
)

func main() {
	client := mailblastr.NewClient("mb_xxxxxxxxx")

	sent, err := client.Emails.Send(&mailblastr.SendEmailRequest{
		From:    "Acme <hello@yourdomain.com>",
		To:      []string{"user@example.com"},
		Subject: "Hello from MailBlastr",
		Html:    "<p>Your first email 🎉</p>",
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("sent", sent.Id)
}
```

Every method has a context-aware variant:

```go
sent, err := client.Emails.SendWithContext(ctx, params)
```

### Errors

Non-2xx API responses come back as a `*mailblastr.MailblastrError` parsed from the API's error shape:

```go
sent, err := client.Emails.Send(params)
if err != nil {
	var apiErr *mailblastr.MailblastrError
	if errors.As(err, &apiErr) {
		fmt.Println(apiErr.StatusCode, apiErr.Name, apiErr.Message)
	}
}
```

### Options

```go
client := mailblastr.NewClient("mb_xxxxxxxxx")
client.BaseURL = "https://api.mailblastr.com" // override your API host
client.HTTPClient = &http.Client{Timeout: 30 * time.Second}
```

## Domain-first model

MailBlastr is **domain-first**: each of your verified sending domains has its own contact pool. The same email address on two domains is two records with separate consent.

- **Contacts** — `Domain` is REQUIRED on the flat `/contacts` API (create/list) whenever `AudienceId` is omitted; pass `Domain` with an email `Id` to disambiguate across pools.
- **Segments / Topics** — `Domain` is REQUIRED on create and list. Segment names are unique within a domain (reusable across domains); every domain carries an auto-created "General" segment.
- **Campaigns** — `Domain` is REQUIRED on create: it names the contact pool the campaign targets (orthogonal to `From`, which may be a different verified domain).
- **Automations + Events** — `Domain` is REQUIRED on `Automations.Create` and `Events.Send`. Only automations belonging to that domain are triggered, so the same event name across several products can never double-fire.

## Resources

The client exposes one service per resource, each following a consistent
`Create` / `Get` / `List` / `Update` / `Remove` (plus resource-specific verbs) shape,
all with `...WithContext` variants:

`Emails` (with nested `Emails.Receiving`), `Batch`, `Domains`, `Audiences`,
`Contacts`, `ContactProperties`, `Campaigns`, `Segments`, `Topics`,
`Templates`, `Automations`, `Webhooks`, `Events`, `ApiKeys`, `Logs`, `Polls`.

```go
// Emails
client.Emails.Send(&mailblastr.SendEmailRequest{From: from, To: to, Subject: subject, Html: html})
client.Emails.List(&mailblastr.ListParams{Limit: 20, After: cursor}) // cursor pagination
client.Emails.Get(id)
client.Emails.ListAttachments(id)
client.Emails.GetAttachment(id, attachmentId)
client.Emails.Update(id, &mailblastr.UpdateEmailRequest{ScheduledAt: at}) // reschedule
client.Emails.Cancel(id)

// Inbound email
client.Emails.Receiving.List(nil)
client.Emails.Receiving.Get(id)
client.Emails.Receiving.GetAttachment(id, attachmentId) // raw []byte
client.Emails.Receiving.GetRaw(id)                      // original RFC822 message
client.Emails.Receiving.Forward(id, &mailblastr.ForwardReceivedEmailRequest{From: from, To: []string{"team@you.com"}})
client.Emails.Receiving.Reply(id, &mailblastr.ReplyReceivedEmailRequest{From: from, Html: "<p>Thanks!</p>"})

// Batch send (up to 100 emails; alias: client.Emails.Batch)
client.Batch.Send([]*mailblastr.SendEmailRequest{ /* ... */ })

// Domains (incl. claiming a domain verified elsewhere + one-click DNS)
client.Domains.Create(&mailblastr.CreateDomainRequest{Name: "example.com"})
client.Domains.Verify(id)
client.Domains.Claim(&mailblastr.ClaimDomainRequest{Name: "example.com"})
client.Domains.VerifyClaim(id)
client.Domains.DetectDns(id)
client.Domains.ApplyCloudflareDns(id, &mailblastr.CloudflareDnsRequest{Token: token})

// Contacts (domain-first)
client.Contacts.Create(&mailblastr.CreateContactRequest{Domain: "example.com", Email: email, FirstName: "Ada"})
client.Contacts.List(&mailblastr.ListContactsRequest{Domain: "example.com"})
client.Contacts.Get(&mailblastr.GetContactRequest{Id: id})                              // by contact id (exact) …
client.Contacts.Get(&mailblastr.GetContactRequest{Id: email, Domain: "example.com"})     // … or by email + domain
client.Contacts.Update(&mailblastr.UpdateContactRequest{Id: id, Unsubscribed: mailblastr.Bool(true)})
client.Contacts.Remove(&mailblastr.RemoveContactRequest{Id: id})
client.Contacts.Batch(&mailblastr.BatchContactsRequest{AudienceId: audId, Contacts: contacts})
client.Contacts.Import(&mailblastr.ImportContactsRequest{AudienceId: audId, Csv: csv})
client.Contacts.AddToSegment(contactId, segmentId)
client.Contacts.ListSegments(contactId)
client.Contacts.GetTopics(contactId)
client.Contacts.UpdateTopics(contactId, &mailblastr.UpdateContactTopicsRequest{
	Topics: []mailblastr.TopicSubscriptionUpdate{{Id: topicId, Subscription: "opt_in"}},
})

// Contact properties (custom fields / merge tags)
client.ContactProperties.Create(&mailblastr.CreateContactPropertyRequest{Key: "plan", Type: "string"})

// Campaigns & Segments (domain-first)
client.Campaigns.Create(&mailblastr.CreateCampaignRequest{
	Domain: "example.com", From: from, Subject: subject, Html: html, SegmentId: segId,
})
client.Campaigns.Send(id, &mailblastr.SendCampaignRequest{ScheduledAt: at})
client.Campaigns.Stats(id)
client.Segments.Create(&mailblastr.CreateSegmentRequest{
	Domain: "yourdomain.com", Name: "VIP",
	Filter: &mailblastr.SegmentFilter{Status: "subscribed"},
})
client.Segments.List(&mailblastr.ListSegmentsRequest{Domain: "yourdomain.com"})
client.Segments.Contacts(id) // preview who matches

// Topics (domain-first)
client.Topics.Create(&mailblastr.CreateTopicRequest{
	Domain: "example.com", Name: "Product updates", DefaultSubscription: "opt_in",
})
client.Topics.List(&mailblastr.ListTopicsRequest{Domain: "example.com"})

// Templates
client.Templates.Create(&mailblastr.CreateTemplateRequest{Name: name, Subject: subject, Html: html})
client.Templates.Duplicate(id, nil)
client.Templates.Publish(id)
client.Emails.Send(&mailblastr.SendEmailRequest{
	From: from, To: to, TemplateId: tmplId,
	Variables: map[string]any{"first_name": "Ada"},
})

// API keys
client.ApiKeys.Create(&mailblastr.CreateApiKeyRequest{Name: "CI", Permission: "sending_access"})
client.ApiKeys.List()
client.ApiKeys.Remove(id)

// Logs & Polls
client.Logs.List(&mailblastr.ListLogsRequest{Limit: 100, Method: "POST", Status: 429})
client.Logs.Get(logId)
client.Polls.List(nil)
client.Polls.Get(emailId)
```

### Automations

Build multi-step automations triggered by events, then inspect their runs.
Every automation belongs to one of your sending domains — `Domain` is required
on create, and `Events.Send` names the domain it targets.

```go
automation, err := client.Automations.Create(&mailblastr.CreateAutomationRequest{
	Name:    "Welcome series",
	Domain:  "yourdomain.com",
	Trigger: "contact.created",
})

client.Automations.AddStep(automation.Id, &mailblastr.AddAutomationStepRequest{
	Type:   "send_email",
	Config: map[string]any{"template_id": "tmpl_welcome"},
})
client.Automations.Update(automation.Id, &mailblastr.UpdateAutomationRequest{Status: "enabled"})

// Fire a custom event — only yourdomain.com's automations are triggered
client.Events.Send(&mailblastr.SendEventRequest{
	Event:  "signup.completed",
	Domain: "yourdomain.com",
	Email:  "user@example.com",
	Data:   map[string]any{"plan": "pro"},
})

// Inspect execution
runs, err := client.Automations.Runs(automation.Id, &mailblastr.ListParams{Limit: 25})
client.Automations.GetRun(automation.Id, runs.Data[0].Id)
client.Automations.Stop(automation.Id)
```

### Webhooks

```go
hook, err := client.Webhooks.Create(&mailblastr.CreateWebhookRequest{
	Endpoint: "https://yourapp.com/hooks/mailblastr",
	Events:   []string{"email.delivered", "email.bounced", "contact.unsubscribed"},
})
// hook.SigningSecret is shown ONCE — store it now.
```

Verify deliveries in your handler with the local (no HTTP) signature check.
Pass the **exact raw request body** — re-serializing parsed JSON breaks the
signature:

```go
func handler(w http.ResponseWriter, r *http.Request) {
	payload, _ := io.ReadAll(r.Body)
	res := mailblastr.VerifyWebhookSignature(payload, r.Header, signingSecret, nil)
	if !res.Valid {
		http.Error(w, res.Reason, http.StatusUnauthorized)
		return
	}
	// process the event ...
}
```

`client.Webhooks.Verify(...)` is an equivalent method form. The timestamp
freshness check defaults to 5 minutes; pass
`&mailblastr.VerifyWebhookOptions{ToleranceSec: -1}` to skip it.

### Pagination

`List` methods accept optional cursor pagination — `&mailblastr.ListParams{Limit, After, Before}` — appended as a query string. Pass `nil` for defaults.

### Idempotency

Pass an idempotency key to safely retry a create:

```go
client.Emails.SendWithOptions(ctx, payload, &mailblastr.RequestOptions{IdempotencyKey: "order-123"})
```

## Documentation

Full docs: <https://www.mailblastr.com/docs>

## License

MIT
