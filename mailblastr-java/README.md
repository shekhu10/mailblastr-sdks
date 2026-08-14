# mailblastr-java

Official Java SDK for the [MailBlastr](https://www.mailblastr.com) email API — send transactional and marketing email from your own verified domain.

- **Zero dependencies** — built on `java.net.http.HttpClient` with hand-rolled JSON.
- **Java 11+**.

## Install

### Maven

```xml
<dependency>
  <groupId>com.mailblastr</groupId>
  <artifactId>mailblastr</artifactId>
  <version>3.0.1</version>
</dependency>
```

### Gradle

```groovy
implementation 'com.mailblastr:mailblastr:3.0.1'
```

## Usage

First, get an API key from your MailBlastr dashboard.

```java
import com.mailblastr.Mailblastr;
import com.mailblastr.MailblastrException;
import com.mailblastr.MailblastrResponse;
import com.mailblastr.requests.SendEmailRequest;

public class Main {
    public static void main(String[] args) {
        Mailblastr mailblastr = new Mailblastr("mb_xxxxxxxxx");

        SendEmailRequest request = SendEmailRequest.builder()
                .from("Acme <hi@yourdomain.com>")
                .to("user@example.com")
                .subject("Hello from MailBlastr")
                .html("<p>Your first email 🎉</p>")
                .build();

        try {
            MailblastrResponse response = mailblastr.emails().send(request);
            System.out.println("sent " + response.getString("id"));
        } catch (MailblastrException e) {
            System.err.println(e.getStatusCode() + " " + e.getName() + ": " + e.getMessage());
        }
    }
}
```

### Responses and errors

Every method returns a `MailblastrResponse` wrapping the raw JSON body with
dotted-path helper getters — `getString("id")`, `getBoolean("has_more")`,
`getList("data")`, `getString("data.0.id")` (numeric segments index into
lists), `asMap()`, and `raw()` for the exact body text. Binary downloads
(received-email attachments, raw MIME) return `byte[]` instead.

Any non-2xx response throws `MailblastrException` carrying the API error
envelope: `getStatusCode()`, `getName()`, `getMessage()`.

Branch on `getName()` plus `getStatusCode()` — never on message text, which is
scrubbed server-side and is not a stable contract. A name does not imply a
status either: `validation_error` is 422 most of the time, but 403 for a
missing `User-Agent` and 409 for a duplicate contact-property key.

Some errors add fields on top of the envelope, reachable through `getBody()`,
the dotted-path `get(...)`, or the `getLimit()` shortcut:

```java
try {
    mailblastr.emails().send(request);
} catch (MailblastrException e) {
    if (e.getLimit() != null) {                       // plan / quota rejection
        System.out.println("used " + e.get("limit.used") + " of " + e.get("limit.limit"));
        System.out.println("upgrade to " + e.get("limit.next_plan.name"));
    }
    // A partial batch failure made with an Idempotency-Key also carries
    // `sent` and `sent_count`; reputation errors carry `reputation`.
}
```

The SDK always sends a non-empty `User-Agent`. If you build an `ApiClient`
yourself, keep it non-empty — the API rejects requests without one with a 403
`validation_error` before authentication even runs.

### Attachments

Attach files by hosted URL (`path`, fetched at send time) or inline base64 (`content`):

```java
SendEmailRequest request = SendEmailRequest.builder()
        .from("Acme <hi@yourdomain.com>")
        .to("user@example.com")
        .subject("Your invoice")
        .html("<p>Invoice attached.</p>")
        .attachment(Attachment.builder()
                .filename("invoice.pdf")
                .path("https://yourdomain.com/invoices/invoice.pdf")
                .build())
        .attachment(Attachment.builder()
                .filename("report.csv")
                .content(base64Content)
                .contentType("text/csv")
                .build())
        .build();
```

### Options

```java
// Override the API host:
Mailblastr mailblastr = new Mailblastr("mb_xxxxxxxxx", "https://www.mailblastr.com/api");
// Inject a custom transport (e.g. for tests):
Mailblastr mailblastr = new Mailblastr("mb_xxxxxxxxx", "https://www.mailblastr.com/api", myTransport);
```

## Resources

One accessor per resource, each following a consistent
(`create` / `get` / `list` / `update` / `remove`, plus resource-specific verbs) shape:

`emails()` (with nested `emails().receiving()`), `batch()`, `domains()`,
`audiences()`, `contacts()`, `contactProperties()`, `campaigns()`,
`segments()`, `topics()`, `templates()`, `automations()`, `webhooks()`,
`logs()`, `events()`, `apiKeys()`, `polls()`.

Read-only resources expose only the verbs they support — `apiKeys()` is
list-only, and `polls()` is list/get.

```java
// Emails
mailblastr.emails().send(request);
mailblastr.emails().list(ListParams.builder().limit(20).after(cursor).build());
mailblastr.emails().list(ListEmailsParams.builder()   // server-side filters
        .status("bounced").search("acme.com").domainId(domainId).build());
mailblastr.emails().sources();   // per-campaign / automation send metrics
mailblastr.emails().get(id);
mailblastr.emails().listAttachments(id);
mailblastr.emails().getAttachment(id, attachmentId);
mailblastr.emails().update(id, "2026-08-05T11:52:01.858Z"); // reschedule
mailblastr.emails().cancel(id);

// Inbound email
mailblastr.emails().receiving().list();
mailblastr.emails().receiving().listAddresses(); // per-address inbound stats
mailblastr.emails().receiving().get(id);
byte[] file = mailblastr.emails().receiving().getAttachment(id, attachmentId);
byte[] mime = mailblastr.emails().receiving().getRaw(id);
mailblastr.emails().receiving().forward(id,
        ForwardEmailRequest.builder().from("you@yourdomain.com").to("team@you.com").build());
mailblastr.emails().receiving().reply(id,
        ReplyEmailRequest.builder().from("you@yourdomain.com").html("<p>Thanks!</p>").build());

// Batch send — up to 100 emails. Items reject `attachments` and
// `scheduled_at`; send those individually via emails().send(...).
mailblastr.batch().sendEmails(List.of(batchRequest1, batchRequest2));

// Domains (incl. claiming a domain verified elsewhere + one-click DNS applies)
mailblastr.domains().create(CreateDomainRequest.builder().name("example.com").build());
mailblastr.domains().verify(id);
mailblastr.domains().claim(ClaimDomainRequest.builder().name("example.com").build());
mailblastr.domains().verifyClaim(id);
mailblastr.domains().detectDns(id);
mailblastr.domains().applyCloudflareDns(id, cloudflareToken);
mailblastr.domains().mxCheck("example.com");   // inbound MX pre-flight
String csv = mailblastr.domains().recordsCsv(id); // DNS records as text/csv
```

### Contacts are DOMAIN-FIRST

Each sending domain has its own contact pool — the same address on two domains
is two records with separate consent. `domain` is required on the flat
`/contacts` API (pass `audienceId` instead to use the nested audience API):

```java
mailblastr.contacts().create(CreateContactRequest.builder()
        .domain("example.com")
        .email("ada@lovelace.dev")
        .firstName("Ada")
        .property("plan", "pro")
        .build());

mailblastr.contacts().list("example.com");
mailblastr.contacts().get(contactId);                          // by id (exact)
mailblastr.contacts().get("ada@lovelace.dev", "example.com");  // by email + domain
mailblastr.contacts().update(UpdateContactRequest.builder()
        .id(contactId).unsubscribed(true).build());
mailblastr.contacts().remove(contactId);

// Bulk import (array or CSV) + segment membership + topics
mailblastr.contacts().batch(BatchContactsRequest.builder()
        .audienceId(audienceId)
        .contact(ContactInput.builder().email("a@b.com").build())
        .onConflict("skip")
        .build());
mailblastr.contacts().importCsv(ImportContactsRequest.builder()
        .audienceId(audienceId).csv("email,company\na@b.com,Acme").build());

// Inline CSV is capped at 5 MB / 10,000 rows. For bigger files, mint a
// presigned URL, PUT the file to it yourself, then import by storage key:
MailblastrResponse upload = mailblastr.contacts()
        .createImportUpload(audienceId, "contacts.csv", fileSizeBytes);
// ... PUT the bytes to upload.getString("upload_url") — do not log that URL ...
mailblastr.contacts().importCsv(ImportContactsRequest.builder()
        .audienceId(audienceId)
        .storageKey(upload.getString("storage_key"))
        .build());

mailblastr.contacts().addToSegment(contactId, segmentId);
mailblastr.contacts().listSegments(contactId);
mailblastr.contacts().updateTopics(contactId, UpdateContactTopicsRequest.builder()
        .optIn(topicId).build());

// Contact properties (custom merge-tag fields)
mailblastr.contactProperties().create(CreateContactPropertyRequest.builder()
        .key("plan").type("string").build());
```

### Campaigns, segments, topics — also domain-first

`domain` picks the contact pool the campaign/segment/topic targets and is
REQUIRED on create (and on segment/topic list):

```java
mailblastr.campaigns().create(CreateCampaignRequest.builder()
        .domain("example.com")
        .from("Acme <hi@example.com>")
        .subject("Launch day")
        .html("<h1>We shipped!</h1>")
        .segmentId(segmentId)
        .build());
mailblastr.campaigns().send(id);                        // now
mailblastr.campaigns().send(id, "2026-08-05T11:00:00Z"); // scheduled (max 30 days out)
mailblastr.campaigns().stats(id);
mailblastr.campaigns().engagement(id); // who opened / clicked / replied
mailblastr.campaigns().ab(id);         // A/B winner evaluation

mailblastr.segments().create(CreateSegmentRequest.builder()
        .domain("example.com")
        .name("Pro users")
        .filter(SegmentFilter.builder().status("subscribed")
                .propertyFilter("plan", "eq", "pro").build())
        .build());
mailblastr.segments().list("example.com");
mailblastr.segments().contacts(id); // preview who matches

mailblastr.topics().create(CreateTopicRequest.builder()
        .domain("example.com")
        .name("Product updates")
        .defaultSubscription("opt_in")
        .build());
mailblastr.topics().list("example.com");
```

### Templates

```java
mailblastr.templates().create(CreateTemplateRequest.builder()
        .name("Welcome").subject("Hi {{first_name}}").html("<p>Welcome!</p>")
        .variable(TemplateVariable.of("first_name", "string", "there"))
        .build());
mailblastr.templates().duplicate(id);
mailblastr.templates().publish(id);
```

### Automations & events

Every automation belongs to one of your sending domains — `domain` is REQUIRED
on create, and `events().send(...)` names the domain it targets, so the same
event name across several products can never trigger the wrong automation:

```java
MailblastrResponse automation = mailblastr.automations().create(
        CreateAutomationRequest.builder()
                .name("Welcome series")
                .domain("yourdomain.com")
                .trigger("contact.created")
                .build());

mailblastr.automations().addStep(automation.getString("id"), AutomationStep.builder()
        .type("send_email")
        .config("template_id", "tmpl_welcome")
        .build());
mailblastr.automations().update(automation.getString("id"),
        UpdateAutomationRequest.builder().status("enabled").build());

// Fire a custom event — only yourdomain.com's automations are triggered
mailblastr.events().send(SendEventRequest.builder()
        .event("signup.completed")
        .domain("yourdomain.com")        // REQUIRED
        .email("user@example.com")
        .payload("plan", "pro")
        .build());

// Inspect execution
mailblastr.automations().runs(automationId, ListParams.builder().limit(25).build());
mailblastr.automations().runs(automationId, ListAutomationRunsParams.builder()
        .status("failed", "running").limit(50).build());
mailblastr.automations().getRun(automationId, runId);
mailblastr.automations().stop(automationId);

// Editing steps requires a stopped automation
mailblastr.automations().updateStep(automationId, stepId, AutomationStep.builder()
        .type("delay").config("duration", "3 days").build());
```

`events().send(...)` does **not** honour `Idempotency-Key` — a retry ingests a
second event and can enroll the contact twice. Dedupe before you call.

### Webhooks

```java
MailblastrResponse hook = mailblastr.webhooks().create(CreateWebhookRequest.builder()
        .endpoint("https://yourapp.com/hooks/mailblastr") // must be https://
        .events("email.delivered", "email.bounced", "email.unsubscribed")
        .build());
String secret = hook.getString("signing_secret"); // shown ONCE — store it

mailblastr.webhooks().rotate(id); // new secret, also shown once

// A failed test delivery is still HTTP 200 — check `ok`, not the status.
MailblastrResponse probe = mailblastr.webhooks().test(id);
if (!Boolean.TRUE.equals(probe.getBoolean("ok"))) {
    System.err.println("endpoint unreachable: " + probe.getString("error"));
}
```

Event names are `email.sent`, `email.delivered`, `email.delivery_delayed`,
`email.bounced`, `email.complained`, `email.opened`, `email.clicked`,
`email.failed`, `email.scheduled`, `email.suppressed`, `email.received`,
`email.replied`, `email.unsubscribed`, `contact.created`, `contact.updated`,
`contact.deleted`, `domain.created`, `domain.updated` and `domain.deleted`.
Anything else is a 422 — note there is no `contact.unsubscribed`.

Verify deliveries locally (no HTTP call) — pass the EXACT raw request body and
the `svix-*` headers:

```java
import com.mailblastr.resources.Webhooks;
import com.mailblastr.resources.VerifyWebhookResult;

VerifyWebhookResult result = Webhooks.verifyWebhookSignature(rawBody, headers, secret);
if (!result.isValid()) {
    System.err.println("rejected: " + result.getReason());
}
```

`headers` is a `Map<String, String>` containing `svix-id`, `svix-timestamp`
and `svix-signature` (read case-insensitively). A fourth argument sets the
timestamp tolerance in seconds (default 300; pass 0 to disable the check).

### Logs, API keys, polls

```java
mailblastr.logs().list(ListLogsParams.builder().limit(100).method("POST").status(429).build());
mailblastr.logs().get(logId);

mailblastr.apiKeys().list();   // display prefixes, permission, domain scoping

mailblastr.polls().list();
mailblastr.polls().get(emailId);
```

**API keys are created, re-scoped and revoked in the dashboard**, at
[mailblastr.com/app/api-keys](https://www.mailblastr.com/app/api-keys). Those
routes accept a signed-in dashboard session only, so `apiKeys()` deliberately
offers nothing but `list()`. That is the point: a key that leaks cannot mint
itself a replacement, widen its own permission, or revoke the keys around it —
containing the blast radius to whatever the leaked key could already do.
`token` on a listed key is only the 8-character display prefix; the full secret
is shown once, in the dashboard, at creation.

### Pagination

`list()` methods accept optional cursor pagination — `limit` is an integer
1–100 and `after`/`before` are item ids (supplying both is a 422):

```java
mailblastr.campaigns().list(ListParams.builder().limit(25).after("cursor_abc").build());
```

Responses are `{ "object": "list", "has_more": bool, "data": [...] }`. There is
no `total` and no `next_cursor` — page forward with the last `data[].id` as
`after`, and stop when `has_more` is false. An unknown cursor returns an empty
page rather than an error.

**The unpaged default is not uniform.** Called with no pagination params,
`domains()`, `apiKeys()`, `topics()`, `campaigns()`, `contacts()`,
`contactProperties()`, `segments()` and the contact/segment sub-lists return
the **whole collection**, while `templates()`, `webhooks()`, `audiences()`,
`automations()`, `automations().runs(...)` and `events()` cap at **20**. Pass an
explicit `limit` whenever you care which you get.

### Rate limits

Every `/emails/**` route — including the reads and the whole `receiving`
subtree — shares one limit of **30 requests per 60 seconds per client IP**.
Over the cap you get a 429 `rate_limit_exceeded`. Those responses carry
`RateLimit-Limit` / `RateLimit-Remaining` / `RateLimit-Reset` and a
`Retry-After`, which the default transport already honours: it retries 429 and
503 up to twice, waiting out `Retry-After` (capped at 30s) or falling back to
exponential backoff. Tune or disable it per client:

```java
Mailblastr mailblastr = new Mailblastr(
        "mb_xxxxxxxxx", "https://www.mailblastr.com/api", Duration.ofSeconds(30), 0);
```

No other resource is rate-limited at the mount level; `automations().createWithAi(...)`
allows 20 requests/60s per account.

### Idempotency

`POST /emails` and `POST /emails/batch` are the **only** endpoints that read
`Idempotency-Key`. Pass one to make a retry replay the first response instead
of sending twice:

```java
mailblastr.emails().send(request, "order-123");
mailblastr.batch().sendEmails(requests, "digest-2026-08-08");
```

- The key must be **1–255 characters**, measured after the server trims it —
  255, not 256 (`Mailblastr.IDEMPOTENCY_KEY_MAX_LENGTH`). The SDK sends the key
  verbatim and lets the **server** be the authority: anything else is a
  `400 invalid_idempotency_key`.
- It is bound to the request body: reusing it with a different payload is a
  `409 invalid_idempotent_request`, and reusing it while the first call is
  still in flight is `409 concurrent_idempotent_requests`.
- Every other endpoint ignores the header, so a retry there creates a second
  resource. `events().send(request, key)` and `events().create(request, key)`
  are deprecated for exactly that reason — dedupe those on your side.

## Building from source

Plain `javac` is enough — there are no dependencies:

```bash
javac -d out $(find src/main/java -name '*.java')
```

The test suite is a set of plain `main()` runner classes (no JUnit):

```bash
javac -d out $(find src -name '*.java')
java -cp out com.mailblastr.tests.AllTests
```

## Documentation

Full docs: <https://www.mailblastr.com/docs>

## License

MIT
