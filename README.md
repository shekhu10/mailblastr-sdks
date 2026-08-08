# MailBlastr SDKs

Official client libraries and tools for the [MailBlastr](https://www.mailblastr.com) email API — one package per ecosystem, all mirroring the same resource surface (emails, receiving, domains, contacts, segments, topics, campaigns, templates, automations, webhooks, events, API keys, logs, polls) with the platform's domain-first model and Svix-style webhook signature verification.

**v2.0.0 is published on every registry.** Grab your API key from the [MailBlastr dashboard](https://www.mailblastr.com) → API Keys, and you're one snippet away from the inbox.

| Package | Language | Install | Registry |
|---|---|---|---|
| [`mailblastr-npm`](./mailblastr-npm) | Node.js ≥ 18 | `npm install mailblastr` | [npm](https://www.npmjs.com/package/mailblastr) |
| [`mailblastr-python`](./mailblastr-python) | Python ≥ 3.8 | `pip install mailblastr` | [PyPI](https://pypi.org/project/mailblastr/) |
| [`mailblastr-go`](./mailblastr-go) | Go ≥ 1.22 | `go get github.com/shekhu10/mailblastr-sdks/mailblastr-go/v2` | Go modules |
| [`mailblastr-ruby`](./mailblastr-ruby) | Ruby ≥ 2.7 | `gem install mailblastr` | [RubyGems](https://rubygems.org/gems/mailblastr) |
| [`mailblastr-php`](./mailblastr-php) | PHP ≥ 8.1 | `composer require mailblastr/mailblastr` | Packagist |
| [`mailblastr-java`](./mailblastr-java) | Java ≥ 11 | `com.mailblastr:mailblastr:2.0.0` | Maven Central |
| [`mailblastr-dotnet`](./mailblastr-dotnet) | .NET 8 | `dotnet add package Mailblastr` | [NuGet](https://www.nuget.org/packages/Mailblastr) |
| [`mailblastr-rust`](./mailblastr-rust) | Rust ≥ 1.75 | `cargo add mailblastr` | [crates.io](https://crates.io/crates/mailblastr) |
| [`mailblastr-cli`](./mailblastr-cli) | Node CLI | `npm i -g mailblastr-cli` | [npm](https://www.npmjs.com/package/mailblastr-cli) |

Every package: base URL `https://www.mailblastr.com/api`, Bearer `mb_…` keys, the API's `{statusCode, name, message}` error shape, percent-encoded path ids, and `domain`-scoped contacts/segments/topics/campaigns per the [docs](https://www.mailblastr.com/docs/api/introduction).

## Quickstart — send your first email

### Node.js

```bash
npm install mailblastr
```

```ts
import { MailBlastr } from 'mailblastr';

const mb = new MailBlastr('mb_xxxxxxxxx');

const { data, error } = await mb.emails.send({
  from: 'Acme <hello@yourdomain.com>',
  to: ['user@example.com'],
  subject: 'Hello from MailBlastr',
  html: '<p>Your first email 🎉</p>',
});
if (error) console.error(error.name, error.message);
else console.log('sent', data.id);
```

### Python

```bash
pip install mailblastr
```

```python
import mailblastr

mailblastr.api_key = "mb_xxxxxxxxx"

email = mailblastr.Emails.send({
    "from": "Acme <hello@yourdomain.com>",
    "to": ["user@example.com"],
    "subject": "Hello from MailBlastr",
    "html": "<p>Your first email 🎉</p>",
})
print(email["id"])
```

### Go

```bash
go get github.com/shekhu10/mailblastr-sdks/mailblastr-go/v2
```

```go
client := mailblastr.NewClient("mb_xxxxxxxxx")

sent, err := client.Emails.Send(&mailblastr.SendEmailRequest{
    From:    "Acme <hello@yourdomain.com>",
    To:      []string{"user@example.com"},
    Subject: "Hello from MailBlastr",
    Html:    "<p>Your first email 🎉</p>",
})
```

### Ruby

```bash
gem install mailblastr
```

```ruby
require "mailblastr"

Mailblastr.api_key = "mb_xxxxxxxxx"

sent = Mailblastr::Emails.send({
  from: "Acme <hello@yourdomain.com>",
  to: ["user@example.com"],
  subject: "Hello from MailBlastr",
  html: "<p>Your first email 🎉</p>"
})
puts sent["id"]
```

### PHP

```bash
composer require mailblastr/mailblastr
```

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

### Java

```xml
<dependency>
  <groupId>com.mailblastr</groupId>
  <artifactId>mailblastr</artifactId>
  <version>2.0.0</version>
</dependency>
```

```java
Mailblastr mailblastr = new Mailblastr("mb_xxxxxxxxx");

SendEmailRequest request = SendEmailRequest.builder()
        .from("Acme <hello@yourdomain.com>")
        .to("user@example.com")
        .subject("Hello from MailBlastr")
        .html("<p>Your first email 🎉</p>")
        .build();

MailblastrResponse sent = mailblastr.emails().send(request);
System.out.println("sent " + sent.getString("id"));
```

### .NET

```bash
dotnet add package Mailblastr
```

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

### Rust

```bash
cargo add mailblastr
cargo add tokio -F macros,rt-multi-thread
```

```rust
use mailblastr::{CreateEmailBaseOptions, Mailblastr, Result};

#[tokio::main]
async fn main() -> Result<()> {
    let mailblastr = Mailblastr::new("mb_xxxxxxxxx");

    let email = CreateEmailBaseOptions::new(
        "Acme <hello@yourdomain.com>", ["user@example.com"], "Hello from MailBlastr",
    ).with_html("<p>Your first email 🎉</p>");

    let sent = mailblastr.emails.send(email).await?;
    println!("sent {}", sent.id);
    Ok(())
}
```

### CLI

```bash
npm i -g mailblastr-cli
export MAILBLASTR_API_KEY=mb_xxxxxxxxx

mailblastr emails send \
  --from 'Acme <hello@yourdomain.com>' --to 'user@example.com' \
  --subject 'Hello from MailBlastr' --html '<p>Your first email 🎉</p>'
```

The CLI covers the full surface — `mailblastr <resource> <action>` for emails, batches, receiving, domains, contacts, segments, topics, campaigns, templates, automations, webhooks, events, keys, logs, and polls. `mailblastr --help` lists everything.

## Beyond sending

Each package README documents its full surface: batch sends, scheduling (`scheduled_at`), template sends with variables, inbound email, domain management with DNS records, audiences (contacts/segments/topics), campaigns, automations, API keys, logs, polls — and **webhook signature verification** (Svix-compatible `svix-id`/`svix-timestamp`/`svix-signature` headers with `whsec_…` secrets) so your webhook handlers can trust what they receive.

## Repository layout & contributing

This is the development monorepo for all SDKs — issues and PRs for every language belong here.

- [`shekhu10/mailblastr-php`](https://github.com/shekhu10/mailblastr-php) is a **read-only subtree split** of [`mailblastr-php/`](./mailblastr-php), regenerated by CI on every release, because Packagist requires `composer.json` at the repository root. Never edit it directly.
- Go consumes this monorepo directly via the `mailblastr-go/vX.Y.Z` tags.

Releases: tag `vX.Y.Z` here and CI publishes every package (`.github/workflows/release.yml`).

## License

MIT — see the `LICENSE` file in each package.
