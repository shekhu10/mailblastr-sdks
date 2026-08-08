# mailblastr-go

Official Go SDK for the [MailBlastr](https://www.mailblastr.com) email API — send transactional and marketing email from your own verified domain.

## Install

```bash
go get github.com/shekhu10/mailblastr-sdks/mailblastr-go/v2
```

Requires Go 1.22+. The SDK depends only on the Go standard library.

## Usage

First, get an API key from the MailBlastr dashboard.

```go
package main

import (
	"fmt"

	"github.com/shekhu10/mailblastr-sdks/mailblastr-go/v2"
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

Non-2xx API responses come back as a `*mailblastr.MailblastrError` parsed from the API's `{statusCode, name, message}` error shape:

```go
sent, err := client.Emails.Send(params)
if err != nil {
	var apiErr *mailblastr.MailblastrError
	if errors.As(err, &apiErr) {
		fmt.Println(apiErr.StatusCode, apiErr.Name, apiErr.Message)
	}
}
```

Branch on `Name`, never on `Message` — messages are scrubbed of provider
identifiers and are free to change. Do not infer the status from the name
either: a handler may override it (e.g. `validation_error` is `422` almost
everywhere, but `403` for a missing User-Agent and `409` for a duplicate
contact property or segment name). Read `StatusCode`.

#### Errors that carry more than the envelope

Some errors are a superset of `{statusCode, name, message}`. `MailblastrError`
parses each extra into a field that stays nil/empty on an ordinary error, and
keeps the whole parsed body in `Body` for anything not modelled yet:

| Field | Present on | Tells you |
| --- | --- | --- |
| `Limit` (`*PlanLimitDetail`) | `plan_limit_reached`, `daily_quota_exceeded`, `monthly_quota_exceeded`, `contact_limit_reached`, `ai_credits_exceeded`, `automation_quota_exceeded` | WHICH allowance ran out (`Kind`), `Used` / `Limit` / `Remaining`, the rolling `Period`, the current `Plan`, the cheapest `NextPlan` that would fit (nil when only Enterprise does), and prepaid `Credits` for the email-quota kinds |
| `Reputation` (`*ReputationDetail`) | `reputation_paused`, `reputation_limit_exceeded`, `sending_service_unavailable` | whether it is `Retryable`, the `Scope` (`tenant` / `domain` / `platform`), hourly and daily counters, and `RetryAt` |
| `Sent` / `SentCount` | a `POST /emails/batch` that failed part way through, sent with an `Idempotency-Key` | the emails that DID go out, so a retry does not send them twice. `SentCount` falls back to `len(Sent)` when the server omits it |

```go
_, err := client.Batch.SendEmailsWithOptions(ctx, emails,
	&mailblastr.RequestOptions{IdempotencyKey: key})

var apiErr *mailblastr.MailblastrError
if errors.As(err, &apiErr) {
	if apiErr.Limit != nil {
		// e.g. "emails_daily cap hit: 100/100"
		fmt.Printf("%s cap hit: %d/%d\n", apiErr.Limit.Kind, apiErr.Limit.Used, apiErr.Limit.Limit)
		if p := apiErr.Limit.NextPlan; p != nil {
			fmt.Printf("upgrade to %s\n", p.Name)
		}
	}
	if apiErr.Reputation != nil {
		fmt.Printf("%s sending gated, retryable=%v\n", apiErr.Reputation.Scope, apiErr.Reputation.Retryable)
	}
	for _, already := range apiErr.Sent {
		fmt.Println("already delivered:", already.Id) // skip these on retry
	}
}
```

The SDK always sends a non-empty `User-Agent` (`mailblastr-go/<version>`),
which the API requires on every `/api/*` resource route — a request without one
is rejected with `403 validation_error` before it is even authenticated. If you
override `client.UserAgent`, keep it non-empty.

### Options

```go
client := mailblastr.NewClient("mb_xxxxxxxxx")
client.BaseURL = "https://www.mailblastr.com/api" // override your API host
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
all with `...WithContext` variants. `ApiKeys` is the deliberate exception — it
is read-only, see [API keys are dashboard-only](#api-keys-are-dashboard-only):

`Emails` (with nested `Emails.Receiving`), `Batch`, `Domains`, `Audiences`,
`Contacts`, `ContactProperties`, `Campaigns`, `Segments`, `Topics`,
`Templates`, `Automations`, `Webhooks`, `Events`, `ApiKeys`, `Logs`, `Polls`.

```go
// Emails
client.Emails.Send(&mailblastr.SendEmailRequest{From: from, To: to, Subject: subject, Html: html})
client.Emails.List(&mailblastr.ListParams{Limit: 20, After: cursor}) // cursor pagination
client.Emails.ListFiltered(&mailblastr.ListEmailsRequest{             // server-side filters
	Status: "bounced", Search: "invoice", DomainId: domId, CampaignId: cmpId,
})
client.Emails.Sources() // per-campaign / automation / individual send metrics
client.Emails.Get(id)
client.Emails.ListAttachments(id)
client.Emails.GetAttachment(id, attachmentId)
client.Emails.Update(id, &mailblastr.UpdateEmailRequest{ScheduledAt: at}) // reschedule
client.Emails.Cancel(id)

// Inbound email
client.Emails.Receiving.List(nil)
client.Emails.Receiving.ListAddresses() // inbound volume per receiving address
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
client.Domains.MxCheck("example.com") // are the MX records already ours?
client.Domains.RecordsCSV(id)         // DNS records as CSV bytes
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
client.Contacts.ListSegments(contactId, nil) // nil ⇒ every segment; pass ListParams to page
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
client.Campaigns.Stats(id)      // counts, rates, top 50 links
client.Campaigns.Engagement(id) // who opened / clicked / replied
client.Campaigns.Ab(id)         // A/B winner, lift, confidence
client.Segments.Create(&mailblastr.CreateSegmentRequest{
	Domain: "yourdomain.com", Name: "VIP",
	Filter: &mailblastr.SegmentFilter{Status: "subscribed"},
})
client.Segments.List(&mailblastr.ListSegmentsRequest{Domain: "yourdomain.com"})
client.Segments.Contacts(id, nil) // preview who matches

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

// API keys (read-only — see "API keys are dashboard-only" below)
client.ApiKeys.List(nil)

// Custom events (definitions + ingestion)
client.Events.Create(&mailblastr.CreateEventRequest{
	Name:   "signup.completed",
	Schema: map[string]string{"plan": "string"},
})
client.Events.Update(eventId, &mailblastr.UpdateEventRequest{
	Schema: map[string]string{"plan": "string", "seats": "number"},
})

// Logs & Polls
client.Logs.List(&mailblastr.ListLogsRequest{Limit: 100, Method: "POST", Status: 429})
client.Logs.Get(logId)
client.Polls.List(nil)
client.Polls.Get(emailId)
```

### API keys are dashboard-only

`ApiKeys` is the one read-only service: it has `List` and nothing else.

Keys are created, re-scoped and revoked **only from a signed-in session in the
MailBlastr dashboard**. The API enforces this — `POST /api-keys`,
`PATCH /api-keys/:id` and `DELETE /api-keys/:id` answer `403 dashboard_only` to
any caller authenticating with an API key, whatever its permission. Since every
SDK call authenticates with a key, there is deliberately no Go method for those
routes.

That boundary is worth having: a key that leaks cannot mint itself a
replacement, promote itself to `full_access`, add a domain to its own scope, or
revoke the keys you would have used to shut it down. Containing a leaked key is
a dashboard action taken by a human, and it stays that way.

`ApiKeys.List` still works with a key and is the SDK-side tool for it — it
returns each key's non-secret 8-character display prefix, permission, domain
scoping and `LastUsedAt`, which is enough to audit what is live and spot a key
that is being used when it should not be. Revoke from the dashboard.

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
client.Automations.RunsFiltered(automation.Id, &mailblastr.ListAutomationRunsRequest{
	Status: []string{"failed"},
})
client.Automations.GetRun(automation.Id, runs.Data[0].Id)
client.Automations.Stop(automation.Id)

// Build the graph with AI (the automation must be stopped first)
client.Automations.CreateWithAi(automation.Id, &mailblastr.AutomationAiRequest{
	Prompt: "Wait a day, then send the onboarding template.",
})
```

Editing an automation's `Domain`, `Trigger`, `TriggerConfig`, `Connections` or
steps requires it to be **disabled** — call `Automations.Stop` first, or the
API answers `422`.

### Webhooks

```go
hook, err := client.Webhooks.Create(&mailblastr.CreateWebhookRequest{
	Endpoint: "https://yourapp.com/hooks/mailblastr",
	Events: []string{
		mailblastr.EventEmailDelivered,
		mailblastr.EventEmailBounced,
		mailblastr.EventEmailUnsubscribed,
	},
})
// hook.SigningSecret is shown ONCE — store it now.
```

`mailblastr.WebhookEvents` lists every valid event name; an unknown one is a
`422`. Endpoints must be `https://` and resolve to a public address.
`Webhooks.Test` returns HTTP 200 even when the delivery failed — check the
result's `Ok` field, not the error return.

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

`Limit` must be an integer between 1 and 100 (default 20); `After` and `Before`
are item ids and are mutually exclusive. Responses are
`{Object, HasMore, Data}` — there is no total and no next-cursor, so page by
feeding the last `Data[].Id` back in as `After`. An unknown cursor returns an
empty page, not an error.

Watch the default: some endpoints return **everything** when you pass no
pagination params at all — domains, API keys, topics, campaigns, contacts,
contact properties, segments, segment contacts, contact segments/topics. The
ones that always cap at 20 are templates, webhooks, audiences, automations,
events, and automation runs.

### Idempotency

Pass an idempotency key to safely retry a send:

```go
client.Emails.SendWithOptions(ctx, payload, &mailblastr.RequestOptions{IdempotencyKey: "order-123"})
client.Batch.SendEmailsWithOptions(ctx, payloads, &mailblastr.RequestOptions{IdempotencyKey: "batch-2026-08-08"})
```

The key must be **1–255 characters**, measured after the server trims it — 255,
not 256 (`mailblastr.IdempotencyKeyMaxLen`). The SDK sends the key verbatim and
lets the **server** be the authority: anything outside that range comes back as
a `400 invalid_idempotency_key`. Reusing a key with a different payload is a `409`
`invalid_idempotent_request`, and reusing it while the first request is still
in flight is a `409` `concurrent_idempotent_requests`; once the original
completes, its response is replayed.

Only `POST /emails` and `POST /emails/batch` honour the header — i.e.
`Emails.SendWithOptions`, `Batch.SendEmailsWithOptions` and
`Batch.SendWithOptions`. Everywhere else it is ignored, so a retry creates a
second resource; `Events.SendWithOptions` is deprecated for exactly that reason.
De-duplicate on your side instead.

### Rate limits and retries

`/emails/**` (sends **and** reads, including the inbound subtree) is capped at
30 requests per 60 s per client IP; no other resource route is mount-limited.
The client retries `429` and `503` up to `MaxRetries` times (default 2),
honouring `Retry-After`, and never retries other 5xx, network errors, or
timeouts — so a non-idempotent write is never silently duplicated.

## Documentation

Full docs: <https://www.mailblastr.com/docs>

## License

MIT
