# MailBlastr SDKs

Official client libraries and tools for the [MailBlastr](https://www.mailblastr.com) email API — one package per ecosystem, all mirroring the same resource surface (emails, receiving, domains, contacts, segments, topics, campaigns, templates, automations, webhooks, events, API keys, logs, polls) with the platform's domain-first model and Svix-style webhook signature verification.

| Package | Language | Install | Tests |
|---|---|---|---|
| [`mailblastr-ruby`](./mailblastr-ruby) | Ruby | `gem install mailblastr` | 52 runs / 552 assertions |
| [`mailblastr-php`](./mailblastr-php) | PHP ≥ 8.1 | `composer require mailblastr/mailblastr` | 176 assertions |
| [`mailblastr-python`](./mailblastr-python) | Python ≥ 3.8 | `pip install mailblastr` | 105 tests |
| [`mailblastr-go`](./mailblastr-go) | Go ≥ 1.22 | `go get github.com/mailblastr/mailblastr-go` | 34 tests |
| [`mailblastr-rust`](./mailblastr-rust) | Rust | `cargo add mailblastr` | 22 tests |
| [`mailblastr-java`](./mailblastr-java) | Java ≥ 11 | `com.mailblastr:mailblastr` (Maven/Gradle) | 168 checks |
| [`mailblastr-dotnet`](./mailblastr-dotnet) | .NET 8 | `dotnet add package Mailblastr` | 21 tests |
| [`mailblastr-cli`](./mailblastr-cli) | Node CLI | `npm i -g mailblastr-cli` | 42 tests |

The Node.js SDK lives in its own repo: [`mailblastr-npm`](https://github.com/shekhu10/mailblastr-npm) (`npm install mailblastr`) — the CLI here wraps it.

Every package: base URL `https://api.mailblastr.com`, Bearer `mb_…` keys, the API's `{statusCode, name, message}` error shape, percent-encoded path ids, and `domain`-scoped contacts/segments/topics/campaigns per the [docs](https://www.mailblastr.com/docs/api/introduction).
