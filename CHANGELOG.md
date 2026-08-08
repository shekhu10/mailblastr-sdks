# Changelog

All nine MailBlastr SDKs release in lockstep — one version, one tag, every registry.
Dates are release dates; entries cover every package unless a language is called out.

## 2.0.0 — 2026-08-08

A contract-parity release. Every package was audited endpoint-by-endpoint against the
live API, which turned up three classes of problem: **response models that decoded
fields the API never sends**, **endpoints and filters that were simply unreachable**,
and **documentation that promised guarantees the API does not make**. 188 files changed —
186 across the nine packages, plus the two GitHub workflows.

Almost all of that is additive. Two changes are not, and they are why this is a major
rather than a 1.4.0:

1. **The API-key lifecycle methods are removed from all nine packages.** `create` and
   `remove` / `delete` shipped in 1.3.0 on every registry, so deleting them breaks code
   that is running today.
2. **Rust's `Error::Api` changed shape** so it can carry the plan, reputation and
   partial-batch detail the API sends. Any `match` that destructures the variant must
   be updated.

Both are covered below with the exact migration.

---

### Breaking changes

#### 1. API-key create, re-scope and revoke are gone — the lifecycle is dashboard-only

**Creating an API key, changing its permission or domain scoping, and revoking it are
dashboard-only operations, and no SDK exposes a method for any of them.**

**What to do instead.** Create, re-scope and revoke keys in the MailBlastr dashboard at
[mailblastr.com/app/api-keys](https://www.mailblastr.com/app/api-keys), signed in. There
is no code migration, by design — the dashboard *is* the replacement.

**Why: this is the security property, not a limitation.** A key that leaks — out of a
shell history, a CI log, a `.env` pasted into a chat — cannot mint itself a replacement,
promote itself to `full_access`, add a domain to its own scope, or revoke the keys you
would have used to shut it down. Its blast radius stays fixed at what it could already
do, and containment stays a human action behind a session the attacker does not have.

**The SDKs are mirroring the server, not deciding policy.** `POST /api-keys`,
`PATCH /api-keys/:id` and `DELETE /api-keys/:id` answer **`403 dashboard_only`** to every
caller authenticating with an API key, whatever its permission or domain scoping. Every
call an SDK makes authenticates with an API key, so those three routes could never
succeed from any of these packages again. The methods are removed rather than deprecated
because a method that cannot work is worse than no method: it fails at runtime, in
production, on the call you reached for during an incident. In the typed packages this
turns that production 403 into a build error.

**`GET /api-keys` is unaffected.** `list` authenticates with a key and remains in all
nine packages, untouched — same route, same shape, same behaviour. It returns each key's
non-secret display prefix, permission, domain scoping and `last_used_at`, which is enough
to audit what is live and notice a key being used when it should not be. Revoke it in the
dashboard.

What was removed, per package — all of it published in 1.3.0:

| Package | Removed | Surviving surface |
| --- | --- | --- |
| npm | `apiKeys.create()`, `apiKeys.remove()`, and the `CreateApiKeyOptions` / `CreateApiKeyResponse` types | `apiKeys.list()` |
| CLI | `mailblastr api-keys create`, `mailblastr api-keys delete` | `mailblastr api-keys list` |
| Python | `ApiKeys.create()`, `ApiKeys.remove()` | `ApiKeys.list()` |
| Ruby | `Mailblastr::ApiKeys.create`, `Mailblastr::ApiKeys.delete` | `Mailblastr::ApiKeys.list` |
| PHP | `ApiKeys::create()`, `ApiKeys::remove()` | `$mb->apiKeys->list()` |
| Go | `ApiKeys.Create` / `.CreateWithContext`, `.Remove` / `.RemoveWithContext`, and the `CreateApiKeyRequest` / `CreateApiKeyResponse` types | `ApiKeys.List` / `.ListWithContext` |
| Rust | `api_keys.create()`, `api_keys.remove()`, and the `CreateApiKeyOptions` / `CreateApiKeyResponse` types | `api_keys.list(..)` |
| Java | `apiKeys().create()`, `apiKeys().remove()`, and `CreateApiKeyRequest` | `apiKeys().list()` |
| .NET | `ApiKeyCreateAsync`, `ApiKeyDeleteAsync` (interface and client), and the `ApiKeyCreateOptions` / `ApiKeyCreated` models | `ApiKeyListAsync` |

Each package states the rule positively in its README and on the resource's own doc
comment, and each carries a test asserting the absence, so the surface cannot quietly
grow the methods back:

| Package | Absence enforced by |
| --- | --- |
| npm | `@ts-expect-error` + runtime `undefined` checks (`npm run typecheck:test` is what makes these bite) |
| CLI | the `api-keys` group registers exactly one subcommand |
| Python | `hasattr` assertions |
| Ruby | `respond_to?` assertions |
| PHP | `method_exists` assertions |
| Go | reflection over the service's method set |
| Rust | a `compile_fail` doctest — `api_keys.create(..)` does not build |
| Java | reflection: the public surface is exactly `[list]` |
| .NET | surface guard over the client's public methods |

An `apiKeys.update()` (`PATCH /api-keys/:id`) was added and removed inside this
unreleased cycle. It was never published, so it is not a break against any released
version — and, contrary to an earlier draft of this file, it was never a new capability.

#### 2. Rust: `Error::Api` now carries a boxed `ApiError`

The variant had to carry the plan / reputation / partial-batch detail described under
*Error envelope* below, so it changed from a struct variant to a tuple variant. It is
boxed so `Result<T, Error>` stays cheap to return.

```rust
// 1.3.0
match err {
    Error::Api { status_code, name, message } => {
        eprintln!("{status_code} {name}: {message}");
    }
    other => return Err(other),
}

// 2.0.0
match err {
    Error::Api(e) => {
        eprintln!("{} {}: {}", e.status_code, e.name, e.message);
        if let Some(limit) = &e.limit {
            eprintln!("hit the {} cap: {}/{}", limit.kind, limit.used, limit.limit);
        }
    }
    other => return Err(other),
}
```

If you only need to read the error, prefer the new accessor — it works on any `Error` and
returns `None` for transport and decode failures:

```rust
if let Some(api) = err.api() {       // Option<&ApiError>
    eprintln!("{} {}", api.status_code, api.name);
}
```

`Display` is unchanged: `Error::Api` still formats as
`MailBlastr API error <status> (<name>): <message>`.

#### Other source-level breaks

Most of these are fields that could not have been carrying a correct value before.

**Go**
- `WebhookTestResult.Delivered` → `Ok`
- `UpdateCampaignRequest.Followups` / `.ListTo` → pointers, so an empty value can express
  "clear" rather than being indistinguishable from absent
- `CampaignStats.Links` → typed `[]CampaignStatsLink`
- `ApiKeys.List()` → `List(*ListParams)`
- `Segments.Contacts(id)` → `Contacts(id, *ListParams)`, now returning `[]SegmentContact`
- `Contacts.ListSegments(id)` → `ListSegments(id, *ListParams)`, now returning
  `[]ContactSegmentRef` instead of `[]Segment` — the route only ever sends
  `id`/`name`/`created_at`, so the old type promised an audience, filter and `updated_at`
  that arrived as empty strings
- `Emails.Receiving.ListAttachments(id)` → `ListAttachments(id, *ListParams)`
  (`Emails.ListAttachments` is unchanged — that route is genuinely not paginated)
- `TemplateVariable` lost `Id`, `CreatedAt` and `UpdatedAt`
- `Campaigns.List` → `[]CampaignListItem` and `Templates.List` → `[]TemplateListItem`
  instead of the full `Campaign` / `Template`

**Rust**
- `Error::Api` — see above
- `emails.receiving.list_attachments`, `contacts.list_segments`, `segments.contacts` and
  `api_keys.list` take a pagination argument (pass `None`)
- `campaigns.list` → `CampaignListItem`; `templates.list` → `TemplateListItem`;
  `segments.contacts` → `SegmentContact`; `contacts.list_segments` → `ContactSegmentRef`
- `DeletedContactResponse.contact` → `.id`
- Several timestamp fields → `Option<String>`

**.NET**
- `ApiKeyListAsync` takes `PaginationOptions` instead of only a `CancellationToken`
- `SegmentListContactsAsync` / `ContactListSegmentsAsync` return `SegmentContact` /
  `ContactSegmentRef` instead of `Contact` / `Segment`
- `CampaignListAsync` / `TemplateListAsync` return `CampaignListItem` / `TemplateListItem`
- `TemplateVariable` lost `Id`, `CreatedAt` and `UpdatedAt`; `ReceivedEmailRaw.ExpiresAt`
  is now `string?`
- `ReceivedEmailListAttachmentsAsync` takes a `PaginationOptions?` before its
  `CancellationToken`, and `ContactImportAsync` / `ContactImportStorageKeyAsync` take
  `segmentId` (and, for the inline-CSV overload, `fileName`) before theirs. All are
  optional, so named and default call sites are unaffected — a call that passed the
  `CancellationToken` **positionally** must name it

**npm**
- `contacts.remove()` returns `{ id }`
- the four list methods return their narrower row types
- `TemplateVariable` lost `id`, `created_at` and `updated_at`

**PHP, Python, Ruby, Java, CLI** — nothing beyond the API-key removal.

Deprecated rather than removed: `RemoveContactResponse.Contact`, `Events.SendWithOptions`
and the `Emails` batch aliases (Go); `events.send_with_idempotency_key` and the `emails`
batch aliases (Rust); both `events().send/create(request, key)` overloads and the
`Emails.batch(..)` / `Batch.send(..)` aliases (Java).

---

### Fixed — decode bugs that silently returned nothing

These were invisible until runtime: the request went out fine, the response came back
fine, and the SDK handed you a null.

- **`contacts.remove()` read a `contact` key the route never returns.** The route returns
  `id`. Every delete handed back an unreachable id — in strongly-typed packages the field
  was declared non-nullable and was always null. (npm, Rust, Go, .NET.)
- **`webhooks.test()` read a `delivered` key that does not exist.** The route returns
  `ok`. Because a *failed* test delivery still returns HTTP 200, every test delivery
  looked like a failure. All nine packages — CLI included — now have a regression test
  proving a failed delivery is a 200 that does not read as success.
- **List routes were typed as their full single-resource object.** `campaigns.list`,
  `templates.list`, `segments.contacts` and `contacts.listSegments` each return a
  deliberately narrower row than the corresponding `get`. Fields such as `filter`,
  `properties`, `html` and `variables` typechecked but arrived `undefined`. Each now has
  its own row type — `CampaignListItem`, `TemplateListItem`, `SegmentContact`,
  `ContactSegmentRef` — following the `SentEmailListItem` precedent already in the
  packages. (npm, Rust, .NET, Go.)
- **`TemplateVariable` declared an `id`, `created_at` and `updated_at` the registry never
  emits.** `GET /templates/:id` sends each variable as `{ key, type, fallback_value }` and
  nothing else. (npm, Rust, Go, .NET; the other packages return plain maps.)
- **`ReceivedEmail.raw.expires_at` was declared non-nullable** even though the raw pointer
  carries only `download_url`. (.NET.)
- **`SegmentStatus` had no `members_only` variant** — the exact status the SDK's own
  `audiences.import_sheet()` creates. (Rust.)
- **Nullable timestamps were modelled as required strings.** (Rust.)
- **Response fields dropped on the floor:** `domain_id` / `campaign_id` / `automation_id`
  on emails, `schedule_timezone` / `daily_batch_size` / `failure_reason` on campaigns,
  `trigger_key` / `trigger_config` on automations, all seven `CampaignStats` counters and
  its rates object, both `CampaignAbResult` variant arms plus lift/zScore/pValue/
  confidence, and seven CSV-import counters. (Go, .NET, Rust.)
- **`aws_last_checked_at` / `aws_check_error` were missing from the domain model.** The
  identity re-check timestamp and its last failure are part of the domain object on
  `POST /domains`, `GET /domains` rows and `GET /domains/:id`, and the model carries no
  extension-data catch-all, so both were unreachable. (.NET.)
- **`PATCH /campaigns/:id` could not express "clear".** `followups` and `list_to` were
  `omitempty`, so an empty slice was indistinguishable from absent — despite the docstring
  promising otherwise. (Go.)

### Fixed — a retry that could duplicate a send

`POST /emails/batch` can fail part way through and return **429 with `sent` / `sent_count`
listing the emails that already went out**. The retry loop treated that like any other 429
and re-sent them, which is exactly what the README promised could never happen. A response
reporting a partial send is now never retried, and the raised error carries `sent` /
`sent_count`. (Python.)

Retries remain confined to 429 and 503 everywhere — the two statuses the server guarantees
were not applied.

### Fixed — a User-Agent that could be switched off

Go is the one package that exposes the header as a mutable field (`client.UserAgent`), and
`net/http` omits `User-Agent` entirely when its value is `""`. Setting it blank therefore
produced a client that answered `403 validation_error` on every call, from the gate that
runs before authentication. A blank override now falls back to `mailblastr-go/<version>`;
a real override is still sent verbatim. (Go.)

### Fixed — documentation that would break working code

- **The README webhook example subscribed to `contact.unsubscribed`, which is not a real
  event name.** Copy-pasting it returned a 422. The canonical name is
  `email.unsubscribed`. Fixed in every package; npm now prevents it at compile time via a
  `WebhookEvent` union, and Java documents the full 19-name vocabulary on the builder.
- **The Idempotency-Key was documented as a vague "24h window" with no length rule.** It
  is **1–255 characters — not 256** — and it is honoured **only** by `POST /emails` and
  `POST /emails/batch`. Both facts are now documented in every package.
- **Five npm README examples would have failed verbatim:** `default_subscribed` instead of
  the required-and-immutable `default_subscription`, the wrong `updateTopics` shape, an
  object-valued `trigger`, and a `forward()` missing its required `from`.
- **Pagination defaults were never documented.** Most resources return the *entire*
  collection when called bare; six always cap at 20 — `templates`, `webhooks`,
  `audiences`, `automations`, `automations.runs` and `events`. That asymmetry is now
  written down in every README.
- Also newly documented: `webhooks.test()` returns 200 even on failure, `/templates/:id`
  accepts an alias, topic `default_subscription` is immutable, and the send-shape caps
  (`to` 1–50, cc/bcc ≤ 50, `from` ≤ 320, `preview_text` ≤ 150, batch ≤ 100, attachments
  25 MB each / 40 MB total, `scheduled_at` ≤ 30 days).
- `ApiKey.token` no longer claims an `mb_live_` prefix, which does not exist.

### Added — endpoints and filters that were unreachable

Present in all eight library SDKs, and in the CLI:

| Capability | Route |
|---|---|
| Per-source send metrics | `GET /emails/sources` |
| Inbound address stats | `GET /emails/receiving/addresses` |
| `status` + `search` filters on the email log | `GET /emails` |
| MX readiness probe | `GET /domains/mx-check` |
| DNS records as CSV | `GET /domains/:id/records.csv` |
| Update an event definition | `PATCH /events/:id` |
| Per-campaign engagement | `GET /campaigns/:id/engagement` |
| Update one automation step | `PATCH /automations/:id/steps/:stepId` |
| Build an automation with AI | `POST /automations/:id/ai` |
| Large-CSV import (presigned upload) | `POST /audiences/:aid/contacts/import/upload` |

The two new email-log filters are worth calling out because they are the ones you reach
for most: **`status`** matches the row's latest state (the same value reads expose as
`last_event`, e.g. `delivered`), and **`search`** is a substring match across recipients,
subject and sender. Both compose with the existing `campaign_id` / `automation_id` /
`source` / `domain_id` filters. Ruby and PHP additionally accept the server's `q` alias
for `search`, honoured only when `search` is absent.

Plus: the `status` filter on `GET /automations/:id/runs`; `storage_key` / `segment_id` /
`file_name` on contact import; pagination on `GET /api-keys`, `GET /segments/:id/contacts`,
`GET /contacts/:id/segments` and two more sub-lists; `custom_return_path`,
`custom_tracking` and `capabilities.receiving` on domain update; and the `engagement`
predicate plus `members_only` on segment filters.

The import gap mattered most: without `storage_key`, **no CSV above the 5 MB /
10,000-row inline cap could be imported at all** from most languages.

### Added — the CLI caught up with the SDK

The CLI shipped 1.3.0 declaring `mailblastr: ^1.0.0` and had no command for several
capabilities the npm SDK had gained. It now declares `^2.0.0` and reaches every one.

Re-derived mechanically against the built SDK: of **116 public methods, 115 have a direct
CLI call site**. The one that does not is `emails.batch()`, which the SDK itself documents
as an alias of `batch.send` — it posts to the same `/emails/batch`, and
`mailblastr emails batch` already routes through `client.batch.send`, so the capability is
present under one command rather than two.

New and extended commands include `api-keys list` (now paginated), `automations ai`,
`automations update-step`, `domains mx-check`, `emails sources`, `receiving addresses`,
`events update`, `--status` / `--search` on `emails list`, `--status` on
`automations runs`, `--storage-key` / `--segment-id` / `--file-name` on `contacts import`,
and pagination flags on five sub-lists that previously had none. `webhooks test` now exits
1 when the delivery failed, instead of exiting 0 on a 200 that reports `ok: false`.

### Changed — the error envelope now carries the API's additive fields

The error object previously kept only `{statusCode, name, message}` and discarded
everything else. **All eight library SDKs** now also surface:

- **`limit`** — the `{kind, used, limit, remaining, plan, next_plan, credits}` object on
  plan and quota rejections. A caller hitting a daily cap can finally tell *which* quota,
  how much is used, and which plan would fit.
- **`reputation`** — on reputation gates.
- **`sent` / `sent_count`** — on a partial batch failure, naming the emails that already
  went out so a retry does not send them twice.
- the full parsed **`body`**, so extras a later version models are never lost.

Each in its own idiom: typed detail structs in Go, Rust, .NET and npm —
`PlanLimitDetail` **and** `ReputationDetail` in each of the four, so
`error.reputation.retryable` needs no cast; map/array-returning accessors in Python, Ruby,
PHP and Java.

Non-envelope bodies (`{"error": "csrf_failed"}`, `{"error": "rate_limited"}`) are lifted
into `name` instead of collapsing to a generic application error. The real HTTP status is
authoritative over the body's mirrored copy, since handlers may override the name→status
map.

The semantics are the same everywhere, with one deliberate split on how "absent" is
spelled:

- `limit` and `reputation` are optional in all eight and read as absent (`null` / `nil` /
  `None`) on an ordinary error — never as an empty object a caller might act on.
- `sent` / `sent_count` are absent (`undefined` / `nil` / `null`) in npm, Ruby, PHP, Java
  and .NET, and are the language's empty value — an empty list and `0` — in Go, Rust and
  Python, where an `Option`-wrapped slice would be unidiomatic. Each of those three states
  it on the field or accessor. Either way an error that carries no `sent` list reports
  nothing sent, so the "do not resend these" decision is the same.
- `sent_count` falls back to the length of `sent` when the body carries the list but omits
  the count, so it can always be trusted.
- An additive field arriving in a shape this SDK version does not model costs **only that
  field**, never the envelope. The typed packages get this from a tolerant decode (Go and
  Rust decode each extra separately, .NET catches `JsonException` per field); the accessor
  packages get it from a type guard on every reader (`is_a?(Hash)`, `is_array`,
  `instanceof Map`, `isinstance(.., dict)`). A malformed `limit` therefore reads as absent
  everywhere, rather than being handed back raw for `e.limit["kind"]` to raise on inside
  the caller's error handler.

Rust is the one language where carrying these required a shape change to a public type —
see *Breaking changes* above.

### Changed — idempotency on event ingestion

`events.send()` / `events.create()` accepted an idempotency key and documented it as a
safe-retry guarantee. Only `/emails` and `/emails/batch` read that header, so the SDKs
were promising exactly-once on an endpoint that gives at-least-once: a retry after a
timeout ingests a **second** event and can double-enroll a contact in an automation.

Mid-cycle this had been resolved four different ways across the packages — removed in one,
deprecated in some, silently kept in others. It is now one rule, applied to all nine:

- The argument is still accepted and still forwarded as `Idempotency-Key`. **Nothing
  changes at runtime**; the server ignores the header on these routes, as it always has.
- Where the argument has its own deprecable surface — a distinct method or overload that
  exists *only* to carry the key — it carries the language's deprecation marker:
  `#[deprecated]` on Rust's `events.send_with_idempotency_key`, `// Deprecated:` on Go's
  `Events.SendWithOptions`, `@Deprecated` on Java's two `events().send/create(request,
  key)` overloads.
- Where the argument is one member of the shared per-request options bag (npm, Python,
  Ruby, PHP, .NET), it cannot be deprecated without deprecating the whole options path
  that `/emails` legitimately needs — so it is documented on the method instead, in the
  same words.
- **On `events.create` the surface was never nine-wide, and this release deliberately does
  not widen it.** Go and Rust have never accepted a key there, at 1.3.0 or now. Adding one
  would mean minting a brand-new public method in two languages that is deprecated the
  moment it exists and that the server ignores. The no-effect note is documented on
  `create` in all eight libraries instead; the CLI exposes no idempotency flag on `events`
  at all.
- **PHP's `array $options` parameter is intact.** An earlier draft of this cycle had
  removed it from `Events::send()` and `Events::create()` — the only package to delete a
  public parameter. It was restored before release, so PHP has no signature change.

### Changed — no package validates the idempotency-key length client-side

Mid-cycle, Python, Ruby and .NET grew a pre-check that raised before the request; the
other six sent the key and let the API answer. Three-of-nine is the worst of both worlds,
and the three did not even agree on the failure: `MailblastrError`, `ArgumentError` and
`ArgumentException` respectively.

All nine now **send the key verbatim and let the server be the authority.** The bound
belongs to the `api_idempotency.key` column, so a client-side copy can only drift — if the
column ever widens, three SDKs would reject keys the API accepts. An out-of-range key is
still a hard failure, reported as the API's own `400 invalid_idempotency_key`.

**This is not a break against a released SDK.** No published version — 1.3.0 included —
has ever length-checked the key locally; the pre-checks were added and removed inside this
cycle, so the shipped behaviour is unchanged.

Two small alignments came with it, in the direction the other packages already went:
Python and Ruby no longer trim the key before sending (the server trims before measuring,
so the effective key is identical), and Ruby and .NET now treat an **empty** key as "no
idempotency" — sending no header — rather than sending an empty one the server would only
reject.

### Added — the 255 bound is an exported constant in every package

Discoverable without changing any runtime behaviour:

| Package | Constant |
| --- | --- |
| npm | `IDEMPOTENCY_KEY_MAX_LENGTH` |
| Python | `mailblastr.IDEMPOTENCY_KEY_MAX_LENGTH` |
| Ruby | `Mailblastr::Client::IDEMPOTENCY_KEY_MAX_LENGTH` |
| Go | `mailblastr.IdempotencyKeyMaxLen` |
| .NET | `MailblastrClient.MaxIdempotencyKeyLength` |
| **Rust** | `mailblastr::IDEMPOTENCY_KEY_MAX_LEN` *(new)* |
| **Java** | `Mailblastr.IDEMPOTENCY_KEY_MAX_LENGTH` *(new)* |
| **PHP** | `Mailblastr\Client::IDEMPOTENCY_KEY_MAX_LENGTH` *(new)* |
| **CLI** | *(new)* not a library, so the bound lives in the `--idempotency-key` help text — which both `emails send` and `emails batch` now build from one constant, so the two cannot drift apart |

The eight libraries each have a test asserting the constant is 255 **and** that a
256-character key still reaches the wire, so nobody re-adds a local check by accident. The
CLI asserts the help text on both subcommands instead.

### Changed — release pipeline

The publish workflow now gates the two least reversible publishes on the same suite every
other registry job runs, and fixes a race:

- **PHP and Go are tested before they publish.** Pushing to the PHP mirror *is* the
  Packagist publish, and pushing the `mailblastr-go/vX.Y.Z` tag *is* the Go publish — and
  a version cached by `proxy.golang.org` can never be replaced or deleted, only retracted
  by a later release. Both jobs now run their build and test first.
- **The Go sub-tag existence check asks the remote, not the local clone.**
  `actions/checkout` fetches only the ref being built, so a local `rev-parse` never saw an
  existing sub-tag and a re-run would fail on the push.
- **The CLI job waits for the exact version just published**, not merely for the package
  to exist. `npm view mailblastr version` succeeds against whatever is already on the
  registry, so the old check returned instantly and the CLI's `npm install` could still
  race propagation.
- **`npm run typecheck:test` runs on the release path.** `npm test` goes through `tsx`,
  which strips types without checking them, so the suite's `@ts-expect-error` absence
  assertions — the ones that enforce the API-key removal above — were inert without it.
  The CI workflow does not run on tags, so if the release job does not type-check the
  tests, nothing on the release path does.

### Known gaps — deliberately not closed in 2.0.0

These are real and tracked; they are listed here rather than quietly omitted.

- **`retry_after` is exposed on the error object in Python only.** Every package parses the
  `Retry-After` header internally to drive its retry loop, but the other eight do not hand
  it to the caller, so code reacting to a 429 that outlived the retry budget cannot read
  the server's own backoff hint. One-of-nine is the same shape as the idempotency
  pre-check this release removed, and it should be settled the same way — uniformly, in
  one pass — rather than piecemeal on the eve of a tag.
- **`RateLimit-Limit` / `RateLimit-Remaining` / `RateLimit-Reset`** are the documented way
  to throttle client-side and ride on 2xx responses, but no package exposes them —
  surfacing them changes the return shape of every method.
- **`GET /domains/dns/callback` is intentionally not modelled** in any SDK. It is a browser
  redirect target with no auth and no JSON body.
- **`UpdateCampaignRequest` cannot send explicit nulls** for its plain string fields (Go,
  .NET), so clearing `segment_id`, `topic_id` or `schedule_timezone` is not expressible.

### Verification

Every package was built and its suite run from this tree, using the same command CI uses.
Counts were read off those runs.

| Package | Command | Result |
|---|---|---|
| mailblastr-npm | `npm run build` + `npm run typecheck:test` + `npm test` | 70 pass / 0 fail |
| mailblastr-cli | `npm test` | 71 pass / 0 fail |
| mailblastr-python | `python -m unittest discover -s tests -t .` | 138 pass / 0 fail |
| mailblastr-ruby | `ruby -Ilib -Itest -e '…'` | 78 runs, 715 assertions, 0 failures |
| mailblastr-php | `php -l` + `php tests/run.php` | 267 pass / 0 fail |
| mailblastr-go | `go build ./...`, `go vet ./...`, `go test ./... -count=1` | 71 pass / 0 fail; vet and gofmt clean |
| mailblastr-rust | `cargo build` + `cargo test` | 60 pass / 0 fail (25 unit + 26 integration + 9 doc) |
| mailblastr-java | `mvn -B compile test-compile` + `java -cp … AllTests` | 264 pass / 0 fail |
| mailblastr-dotnet | `dotnet test tests/Mailblastr.Tests/Mailblastr.Tests.csproj` | 54 pass / 0 fail |

Two harness details worth recording, because getting them wrong makes a suite look green
when it never ran:

- **Java's `pom.xml` sets `<skipTests>true</skipTests>`.** The tests are plain `main()`
  runners, not JUnit, so `mvn test` compiles and reports BUILD SUCCESS while executing
  **zero** tests. The real entrypoint is `com.mailblastr.tests.AllTests`, which
  `System.exit(1)`s on failure — this is what CI runs and what the count above reflects.
- **`dotnet test` must be given the test csproj explicitly**; run from the package root it
  fails with MSB1003 rather than running anything.

No test in any package makes a live API call — all drive an injected mock transport.
Response-shape fixes are therefore verified against the contract documents and replayed
fixtures, not observed traffic.

---

## 1.3.0 — 2026-07-27

- **Breaking:** the `tags` field was removed from the email send shape product-wide.
  The API now returns 422 if you send it.

## 1.2.0 — 2026-07-26

- Campaign scheduling (`schedule_timezone`, `daily_batch_size`, 30-day cap), the
  `mailblastr:schedule` automation trigger, and log/analytics surfaces.

## 1.1.0 — 2026-07-25

- Multi-domain API key scoping via `domain_ids`; reply/forward gaps closed.

## 1.0.0 — 2026-07-24

- First public release of all nine packages.
