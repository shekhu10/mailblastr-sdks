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
  <version>1.2.0</version>
</dependency>
```

### Gradle

```groovy
implementation 'com.mailblastr:mailblastr:1.2.0'
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

Any non-2xx response throws `MailblastrException` carrying the API error body:
`getStatusCode()`, `getName()`, `getMessage()`.

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

```java
// Emails
mailblastr.emails().send(request);
mailblastr.emails().list(ListParams.builder().limit(20).after(cursor).build());
mailblastr.emails().get(id);
mailblastr.emails().listAttachments(id);
mailblastr.emails().getAttachment(id, attachmentId);
mailblastr.emails().update(id, "2026-08-05T11:52:01.858Z"); // reschedule
mailblastr.emails().cancel(id);

// Inbound email
mailblastr.emails().receiving().list();
mailblastr.emails().receiving().get(id);
byte[] file = mailblastr.emails().receiving().getAttachment(id, attachmentId);
byte[] mime = mailblastr.emails().receiving().getRaw(id);
mailblastr.emails().receiving().forward(id,
        ForwardEmailRequest.builder().from("you@yourdomain.com").to("team@you.com").build());
mailblastr.emails().receiving().reply(id,
        ReplyEmailRequest.builder().from("you@yourdomain.com").html("<p>Thanks!</p>").build());

// Batch send (alias of mailblastr.emails().batch(...))
mailblastr.batch().send(List.of(request1, request2)); // up to 100 emails

// Domains (incl. claiming a domain verified elsewhere + one-click DNS applies)
mailblastr.domains().create(CreateDomainRequest.builder().name("example.com").build());
mailblastr.domains().verify(id);
mailblastr.domains().claim(ClaimDomainRequest.builder().name("example.com").build());
mailblastr.domains().verifyClaim(id);
mailblastr.domains().detectDns(id);
mailblastr.domains().applyCloudflareDns(id, cloudflareToken);
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
mailblastr.campaigns().send(id, "2026-08-05T11:00:00Z"); // scheduled
mailblastr.campaigns().stats(id);

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
mailblastr.automations().getRun(automationId, runId);
mailblastr.automations().stop(automationId);
```

### Webhooks

```java
MailblastrResponse hook = mailblastr.webhooks().create(CreateWebhookRequest.builder()
        .endpoint("https://yourapp.com/hooks/mailblastr")
        .events("email.delivered", "email.bounced", "contact.unsubscribed")
        .build());
String secret = hook.getString("signing_secret"); // shown ONCE — store it

mailblastr.webhooks().rotate(id); // new secret, also shown once
mailblastr.webhooks().test(id);
```

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

mailblastr.apiKeys().create(CreateApiKeyRequest.builder()
        .name("CI").permission("sending_access").build());
mailblastr.apiKeys().list();

mailblastr.polls().list();
mailblastr.polls().get(emailId);
```

### Pagination

`list()` methods accept optional cursor pagination:

```java
mailblastr.campaigns().list(ListParams.builder().limit(25).after("cursor_abc").build());
```

### Idempotency

Pass an idempotency key to safely retry a create (24h window):

```java
mailblastr.emails().send(request, "order-123");
mailblastr.events().send(eventRequest, "signup-42");
```

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
