# Changelog

All nine MailBlastr SDKs release in lockstep — one version, one tag, every registry.
Dates are release dates; entries cover every package unless a language is called out.

## 5.1.1 — 2026-08-30

Documentation, typing, and CLI-flag release — no wire behavior changed in any SDK.

- `emails.list` `source` accepts two new values alongside `individual`:
  `api` (one-off sends made with an API key — plus mail sent before the send
  door was recorded) and `dashboard` (mail composed in the dashboard).
  `emails.sources` rows gained a matching `kind: "api"`; the `individual` row
  now means dashboard-composed one-offs only. The Node types widen their
  unions; every other SDK already carried these as open strings, so nothing
  breaks — docs updated across all nine SDKs and the CLI.
- CLI: `contacts batch` now registers the `--audience-id <id>` flag its own
  error message has always suggested (the audience positional still works);
  a flag value no longer silently loses to the missing positional.
- Ruby and PHP treat an empty-string `audienceId` as absent (routing to the
  flat `/contacts/batch` door like every other SDK) instead of building a
  malformed `/audiences//contacts/batch` URL.
- The `folder` filter is now documented in every SDK README and doc comment,
  with string constants in the typed SDKs (Go, Rust, Java, .NET), and has
  wire-level tests in every package. Heads-up: the API now answers an unknown
  `folder` value with a 422 listing the legal values instead of silently
  returning the unfiltered list.
- Docs state the real one-of semantics for bulk imports: when both `domain`
  and `audienceId` are supplied, `audienceId` wins and `domain` is ignored.

## 5.1.0 — 2026-08-30

First public 5.1.0. Two features, all nine SDKs and the CLI:

- **Domain-first bulk contact import.** `POST /contacts/batch` gets a
  first-class door in every SDK (`contacts.batch({ domain, contacts })` in
  Node, `batch_in_domain` in Rust/Python-style SDKs, `ContactBatchInDomainAsync`
  in .NET, `mailblastr contacts batch --domain` in the CLI): pass the sending
  domain instead of an audience id, up to 10,000 contacts per call,
  `on_conflict: 'upsert' | 'skip'` (default `upsert`). Response is the same
  `contact_import` object as the audience-scoped batch. Prefer this over a
  create-per-contact loop — the batch takes the account's contact-limit lock
  once instead of per contact.
- `emails.list` accepts optional `folder` (`outbox` | `sent` | `scheduled` |
  `failed`) matching GET `/emails?folder=`.

Public `emails.send` is unchanged (still waits for SES). Dashboard
instant-send is a session route, not part of this API.

## 5.0.1 — 2026-08-20

**Every SDK except the Node one was deriving the wrong webhook signing key for some
secrets.** A wrong key does not fail loudly: verification returns `no_match`, so a
correctly configured endpoint silently treats every genuine delivery as forged. If you
verify webhooks in python, php, ruby, go, rust, java or .NET, upgrade.

No API changed. This release is behaviour only.

### What was wrong

The backend derives the key with Node's `Buffer.from(suffix, 'base64')`. Every SDK
reimplemented that, and every reimplementation was written against what base64 *means*
rather than what Node *does*. Node's decoder has five rules, and no SDK had all five:

1. **`=` TERMINATES the input.** Everything from the first `=` onward is discarded — it
   is not padding to be stripped. `YWJj====ZA` is `abc`, not `abcd`.
2. Characters outside the alphabet are **skipped**, never fatal — whitespace,
   punctuation, non-ASCII.
3. `-` and `_` are the URL-safe spellings of `+` and `/` and are **translated**, not
   dropped. (Ruby's `unpack1("m")` dropped them.)
4. A trailing group of **one** character contributes no byte.
5. **The unit is the low 8 bits of each UTF-16 code unit, not the codepoint.** Node masks
   every code unit with `0xFF` before the table lookup, so `Ł` (U+0141) is read as `A`
   and `Ľ` (U+013D) is read as `=` — and therefore TERMINATES the value. An astral
   character contributes its two surrogate halves' low bytes, never its UTF-8 bytes.

5.0.0 fixed rules 1–4 in five SDKs by porting python's implementation, which is how
python's own residual divergence propagated. Rule 5 was in none of them.

And the caller matters as much as the decoder: when the decode yields **zero** bytes, the
key is the UTF-8 bytes of the **whole** secret, `whsec_` prefix included. Ten of the
conformance vectors exercise that path.

None of this is hypothetical. `POST /webhooks` accepts a caller-supplied `secret`
verbatim, with no shape validation — so any of these shapes can be a live endpoint's key.
An auto-generated secret is base64 of 24 random bytes and is pure ASCII with no padding,
which is why this went unnoticed: the common case works, and only a hand-picked secret
diverges.

### How it is prevented from recurring

`scripts/webhook-b64-corpus.mjs` generates the conformance corpus **from Node itself**,
not from anyone's reading of base64, and `scripts/webhook-b64-corpus.json` holds the 41
resulting vectors (10 raw-fallback). Every SDK embeds them in its own suite.

The corpus is what the fix was measured against, and the measurements are the reason to
trust it: all eight verifying SDKs went from 25–27/31 to 31/31 on rules 1–4, then from
**1300/3000 to 3000/3000** once vectors above U+00FF exposed rule 5. Final state is
**41/41 on the canonical corpus and 20,000/20,000 on a differential fuzz against Node**
(15,583 of those containing codepoints above 0xFF, 3,448 exercising the raw fallback),
plus an independent oracle and several thousand more vectors per package during review.

The lesson is in the numbers: every SDK passed 31/31 **and** 2,000/2,000 on an ASCII-only
fuzz while rule 5 was still wrong in all of them. An ASCII test cannot see this bug.

### Also fixed

- **rust** — `cargo fmt` drift introduced in `services/webhooks.rs`, and two comments that
  described `中文` as masking to bytes outside the alphabet. U+4E2D masks to `0x2D`, which
  IS the URL-safe `-`; the decode empties because one usable character cannot encode a
  byte, not because the byte was skipped.
- **ruby** — the `rescue` in `utf16_low_bytes` returned the raw UTF-8 bytes, which is
  precisely the pre-5.0.1 behaviour, so a pathological input would have silently
  reinstated the bug it was added to guard.

---

## 5.0.0 — 2026-08-19

**Every package's SOURCE was audited against the live route handlers — not its README — and 60 defects came back.** Four of them were making correctly written integrations do the wrong thing on the wire: a Python webhook receiver rejecting genuine deliveries as forged, a PHP send that could deliver the same email twice on a retry, a Go template send that either 422'd or went out with a blank subject, and a CLI subcommand whose documented invocation could never succeed. Those are below, first.

This is the opposite of 4.0.0, which changed nothing and said so. If you are on Python or PHP and you verify webhook signatures, upgrade for that alone.

Because this is a major, your dependency ranges will **not** pick it up automatically: `^4.0.0` (npm), `~> 4.0` (RubyGems), `"4"` (Cargo), `^4.0` (Composer) and the pinned Maven / NuGet coordinates all stay on 4.0.0 until you bump the constraint yourself. Go is the exception, and not in your favour — an unchanged `/v4` import keeps resolving forever and keeps serving 4.0.0, template-send bug included. See breaking change 1.

The CLI ships pinned to `mailblastr: ^5.0.0`.

---

### Breaking changes

Seven changes break something. The Go module path is the only one nothing will tell you about.

#### 1. Go: the module path is now `.../mailblastr-go/v5`

Required by Go's module rules for any major above v1.

```bash
go get github.com/shekhu10/mailblastr-sdks/mailblastr-go/v5
```

```go
import "github.com/shekhu10/mailblastr-sdks/mailblastr-go/v5"
```

The package name is still `mailblastr`, so only the import line changes — nothing else in the Go API moved on account of the major. Same migration shape as the v3 and v4 bumps. The release workflow's proxy warm-up derives the module path from `go.mod` (`awk '/^module /{print $2}'`) rather than hard-coding it, so the v5 path is warmed without a workflow edit.

#### 2. CLI: `automations update-step --type` is required, and `--key` is gone

`PATCH /automations/:id/steps/:stepId` re-validates the **whole** step through the same validator the add path uses, so a body without `type` is a `422 validation_error` ("type must be one of: …") before the config is ever looked at. `--type` was optional, which made the form the CLI README itself documented a request that could not succeed.

```bash
# 4.0.0 — 422, every time
mailblastr automations update-step auto_123 step_456 --config '{"timeout":"12 hours"}'

# 5.0.0 — resend the step's current type even when only the config changes
mailblastr automations update-step auto_123 step_456 \
  --type wait_for_event --config '{"event":"email.opened","timeout":"12 hours"}'
```

`--key` is removed rather than fixed: the route forwards only `type` and `config` to storage, deliberately, so the connections referencing a step keep working. A `--key` was accepted and silently discarded, leaving the caller believing the step had been re-keyed. Delete and re-add the step with `add-step --key` to change it. A script still passing `--key` now fails with an unknown-option error instead of quietly doing nothing.

#### 3. Rust: `UpdateAutomationStepOptions` takes the step type in `new()`

Same defect, same fix, one language over: `step_type` was `Option<String>` and skipped when `None`.

```rust
// 4.0.0 — serialized without `type`, so the API rejected it
let opts = UpdateAutomationStepOptions::new().with_config(config);

// 5.0.0
let opts = UpdateAutomationStepOptions::new("wait_for_event").with_config(config);
```

`step_type` is now a plain `String`, `new()` requires it, and `with_key()` is **deleted** — a builder method whose only effect was to add a field the server ignores. `with_type()` still exists for readability.

#### 4. npm: `Campaign.statistics` no longer promises `links`

`GET /campaigns/:id` embeds `getCampaignStats()` verbatim, which computes counts and rates and nothing else. Only `GET /campaigns/:id/stats` adds the per-link breakdown. `Campaign.statistics` was typed `Omit<CampaignStats, 'object' | 'campaign_id'>`, so `statistics.links` typechecked clean and was `undefined` at runtime.

```ts
// 4.0.0 — compiles, throws
const top = campaign.statistics?.links[0];

// 5.0.0 — does not compile; ask the endpoint that computes it
const { data: stats } = await mb.campaigns.stats(campaign.id);
const top = stats?.links[0];
```

The counts-and-rates core is now the exported `CampaignStatistics`, and `CampaignStats extends CampaignStatistics` with `object`, `campaign_id` and `links`. Anyone who wrote that `Omit<…>` by hand gets the same narrowing, which is the intent.

#### 5. .NET: `ReceivedAttachment.Size` is `long?`, and `IMailblastr` gained a method

`GET /emails/receiving/:id/attachments` serializes `size` as `size ?? null`, so a stored attachment whose metadata predates the field arrives as an explicit null against a non-nullable `long`. Read it as `att.Size ?? 0`. The copy embedded on `ReceivedEmail.Attachments` substitutes `0` server-side and is unaffected.

`EmailBatchSendAsync` was added to `IMailblastr`. Anyone **implementing** the interface — a hand-rolled test double — must add it; callers and anything holding a `MailblastrClient` are unaffected. The internal `DataEnvelope<T>` is deleted.

#### 6. Go: `UpdateAutomationStepRequest.Type` is always serialized

`json:"type,omitempty"` → `json:"type"`. Not a compile error: a request that left `Type` empty was already a guaranteed 422, and now fails as `"type":""` instead of as a missing key. `Config` replaces the step's configuration wholesale — anything you leave out is dropped, not preserved.

#### 7. CLI: three more invocations that exited `0` now exit `1`

None of these was a documented change, and each one succeeds today, so a script that runs
unattended will start failing rather than start behaving differently. All three are
argument-parse failures — nothing reaches the API.

- **`webhooks verify --api-key` / `--base-url` are gone.** The command was re-registered as
  local-only: it computes an HMAC and makes no HTTP request, so it resolves no client and no
  longer accepts either flag. This is the one most likely to bite, because every other
  subcommand takes them and the 4.0.0 README said this one needed a key. Drop the flags; the
  command needs nothing.
- **`webhooks verify --tolerance` rejects a non-integer.** It was coerced with `Number()`, so
  `--tolerance 5m` became `NaN` — and a `NaN` comparison is always false, which silently
  **disabled the freshness check** and reported an arbitrarily old signature as valid. That is
  a replay window, not a formatting nit, which is why it now hard-fails instead of warning.
  Pass seconds: `--tolerance 300`.
- **`contacts update --unsubscribed --subscribed` is refused.** The pair used to resolve
  silently in favour of `--unsubscribed`. Anything assembling flags from a template could emit
  both and get a consent state it never chose; guessing is the wrong behaviour for consent.
  Pass exactly one.

**A footnote for Rust, and 3.0.0 predicted it.** `SendEmailBatchResponse` gained `queued` and `queued_count`, which breaks a struct literal that constructs it (a test fixture, typically) because the public structs still carry no `#[non_exhaustive]`. That is exactly the mechanism 3.0.0's *Deliberately not changed* section described — an additive API field forcing a major — arriving on schedule. It stays open for the same reason: adding the attribute is itself a break, and it deserves its own decision.

---

### Fixed — the two that were corrupting live traffic

#### `verify_webhook_signature` rejected genuine deliveries (Python)

Two independent defects, either of which alone made a correctly configured endpoint treat every real event as forged. A webhook SDK silently dropping real deliveries is the worst failure mode this library has.

**1 — the secret was decoded strictly.** The backend derives its signing key with Node's `Buffer.from(suffix, 'base64')`, which is lenient: it ignores characters outside the alphabet, accepts the URL-safe `-` / `_` spellings, needs no `=` padding, and drops a lone trailing character that encodes no byte. `base64.b64decode` does none of that — it raises `binascii.Error` on any length that is not a multiple of four, and silently discards `-` / `_`. So a `whsec_` secret you chose yourself (`POST /webhooks` accepts a caller-supplied one) whose suffix was unpadded or URL-safe produced a **different key here than at the signer**, and every valid delivery came back `{"valid": false, "reason": "no_match"}`. The new `_b64_lenient` reproduces Node byte for byte; Ruby's `unpack1("m")` and the npm SDK were already lenient.

**2 — it never read Flask's headers at all.** `_read_header` did `for key in headers`, which assumes a mapping. Werkzeug's `EnvironHeaders.__iter__` yields `(key, value)` **tuples**, so every lookup missed and every delivery read as `missing_headers` — including the exact `Webhooks.verify(body, request.headers, secret)` call this package's README shows. `_header_pairs` now prefers `.items()`, the one accessor `dict`, `http.client.HTTPMessage`, Django's `request.headers`, Starlette/FastAPI's `Headers` and Werkzeug's all agree on, and falls back to pair detection.

Nothing in your code changes. If you worked around this by trusting unverified deliveries or re-implementing the HMAC yourself, you can drop it.

**PHP had the same base64 half.** `WebhookSignature::secretToKey()` called `base64_decode($suffix, true)` — strict mode, returns `false` on anything non-standard — then fell back to the raw string as the key and rejected every genuine delivery as `no_match`. Now `strtr($suffix, '-_', '+/')` followed by a non-strict decode.

#### PHP dropped an Idempotency-Key of `'0'`, and a retry could send a second email

The header was gated on `!empty($options['idempotencyKey'])`. **PHP's `empty()` is true for the string `'0'` and the integer `0`** — both valid 1-character keys, since the API accepts 1–255. The header was silently omitted, turning an idempotent send into a plain one, and this client retries 429 and 503 automatically. A retried send keyed `'0'` therefore delivered the email **twice**. Anyone numbering keys from a zero-based counter, or passing an integer, was exposed; nothing in the response said so.

The gate is now "present, scalar, non-boolean, and non-empty once stringified": `'0'`, `0` and `12345` go out verbatim; `null`, `''`, `[]` and booleans send no header, since the API rejects an empty one with `400 invalid_idempotency_key` anyway. No caller change.

---

### Fixed — requests that could never succeed

- **go** — a template-based send could never use the **template's own** `from` / `reply_to` / `subject`. `SendEmailRequest` and `BatchEmailRequest` always serialized `"from":""` and `"subject":""`, and the API reads a key that is present-but-empty as "the caller supplied this". A template send with no explicit `From` was rejected `422 missing_required_field`; one with no explicit `Subject` **went out with a blank subject** rather than the template's. `MarshalJSON` now omits an empty `From` / `Subject` only when `TemplateId` or `Template` references a template. Non-template payloads still always send both, so a deliberately empty subject keeps working. The new `templateRef` helper mirrors the server's extraction exactly: a non-nil nested `Template` shadows `TemplateId`, so a `Template` carrying neither `Id` nor `Alias` references nothing even when `TemplateId` is set.
- **cli**, **rust** — `automations update-step` / `update_step` omitted the required `type`, so the call 422'd unconditionally. See breaking changes 2 and 3.
- **cli** — `webhooks verify` demanded an API key for a command that makes **no HTTP request**. It recomputes the signature locally, and requiring `MAILBLASTR_API_KEY` pushed people into putting a send-capable key on the very receiver whose job is to distrust its input. The command is now registered `local`: it resolves no client, advertises no `--api-key` / `--base-url`, and calls the SDK's standalone `verifyWebhookSignature` export directly.

### Fixed — input accepted, then silently discarded or misread

- **cli** — `webhooks verify --tolerance` was coerced with `Number()`, so `'5m'` became `NaN`. The SDK's freshness gate is `if (toleranceSec > 0)`, which `NaN` fails, so a typo **skipped the replay-window check entirely and reported a stale delivery as valid**. It is parsed now, and a negative or non-integer value is a usage error.
- **cli** — `contacts update --unsubscribed --subscribed` resolved silently in favour of `--unsubscribed`, writing the opposite of what the operator meant onto a consent state. Contradictory flags are refused, like every other mutually exclusive pair in this CLI.
- **cli** — `--key` on `update-step` (breaking change 2) was forwarded and dropped by the server.
- **ruby** — `apply_namecheap_dns` documented `api_user` / `api_key` / **`user_name`**. The API reads `apiUser` / `apiKey` / `userName`, accepts `api_user` / `api_key` / `username` as aliases, and has **no `user_name` alias at all** — a username sent under that key was ignored, and the DNS call ran without it. The docstring now names the camelCase spelling every other SDK uses, and a test asserts the gem transmits these keys byte for byte instead of rewriting them.
- **cli** — `--json` asked for compact output but `fail()` always pretty-printed, so the flag was a half-truth on the error stream. Both streams honour it now.

### Fixed — responses that decoded wrong, or could not be read at all

- **rust** — a single `null` `size` on one received attachment made the **entire attachment page undecodable**. `#[serde(default)]` covers an absent key only; an explicit null fails the whole response. A `null_as_zero` deserializer reads it as `0`. (**dotnet** took the nullable route instead — breaking change 5.)
- **npm** — `Campaign.statistics.links` typechecked and was undefined at runtime. Breaking change 4.
- **npm**, **go**, **rust**, **dotnet** — a batch above **40** emails is not sent inline. `POST /emails/batch` writes the emails as due-now sends, answers **202** with `{ queued: true, queued_count }`, and lets the worker deliver them — and every one of these packages typed the response as `{ data }` alone, so **a caller could not tell a sent batch from a queued one**. The ids are real from that moment (`emails.get(id)` works) but start at `scheduled` with nothing transmitted. npm and .NET now return `BatchSendResponse`, Go's `BatchSendResponse` gained `Queued` / `QueuedCount` alongside a `BatchSyncMax = 40` constant, Rust's `SendEmailBatchResponse` gained `queued` / `queued_count`. Both paths are success — branch on `queued`, not on the absence of an error. A batch carrying a `@mailblastr.dev` simulator recipient stays inline at any size. npm's `emails.batch` / `batch.send` return type widened structurally, so existing call sites still compile; .NET kept `EmailBatchAsync` returning the ids and added `EmailBatchSendAsync` for the full answer.
- **python** — an unparseable `Retry-After` raised a bare `ValueError` **from inside the code building the `MailblastrError`**, replacing the API's `{statusCode, name, message}` with an unrelated exception and destroying the error name callers branch on. `parsedate_to_datetime` raises rather than returning `None`, and a non-finite float would have reached `time.sleep()` as `nan`/`inf`. Nothing in that path can raise now; an unrecognised value simply means "no advice".
- **npm** — `ReceivedEmail.domain_id` was missing entirely, so inbound mail could not be attributed when an account receives on more than one domain. `ReceivedAttachment` now models both shapes the API serves: embedded on `ReceivedEmail.attachments` (relative `url` alias, no `object`) versus a `listAttachments()` row (`object: 'attachment'` and an id, no `url`).

### Added — request fields and predicates that were unreachable

- **java** — `SegmentFilter.Builder` had no way to express the campaign-engagement predicate at all. `engagement(event, campaignId)` and `clearEngagement()` are new (clearing needs an explicit null; an omitted key keeps the stored predicate), and `status` documents `members_only` — the value the CSV and Google-Sheet importers create.
- **java**, **python** — `PATCH /automations/:id` accepts a new `trigger` and the `trigger_key` connections reference, on a disabled automation. Neither package modelled either, so re-pointing a trigger was not expressible.
- **go** — `UpdateDomainRequest` had no `custom_return_path`, so the MAIL FROM subdomain could not be changed from Go. It must be a single DNS label; changing it re-issues the domain's DNS records.
- **cli** — `templates create` / `update --variables` sets the declared-variable registry (max 50 `{key, type?, fallback_value?}` entries; `[]` clears it). The send path prefers a declared fallback when a caller omits that variable, so a template with no registry has no fallbacks — which is why leaving this to the dashboard was wrong.

### Fixed — documentation that promised guarantees the API does not make

Nothing here changes a byte on the wire. Each one was a sentence a reader could act on and be wrong.

- **all nine** — the unpaginated list methods claimed "every row is returned with `has_more: false`". The server caps an unpaginated list at `UNPAGINATED_MAX = 1000` and sets `has_more` truthfully. `segments.contacts` is the dangerous one — a segment easily resolves to more than 1,000 contacts, and code treating one call as the whole membership was quietly operating on a prefix. Corrected on `domains`, `api-keys`, `topics`, `campaigns`, `contacts`, `contact-properties`, `segments.contacts`, `contacts.listSegments`, `contacts.getTopics`, `polls` and `receiving.listAttachments` in every package that documents them, plus `ListResponse` / `PaginationParams` in Rust and the pagination help text in the CLI. The CLI README also moves `emails receiving list` out of the cap-20 group, and Go's `Receiving.List` now documents the same divergent default.
- **npm**, **dotnet** — `campaigns.cancel` was described as "returns it to draft", flatly. It depends on how far the send got: `scheduled` / `recurring` / `paused` → back to `draft`, editable and re-sendable; `queued`, already fanning out → **`canceled`, which is terminal**. Part of the audience has been mailed and those copies cannot be recalled; what stops is every remaining recipient, and for a staggered campaign every future batch-day. Read `status` rather than assuming.
- **npm**, **dotnet** — the campaign status vocabulary was missing `canceled` outright, on both `Campaign` and `CampaignListItem`.
- **npm** — `CampaignAbResult.rate` and `.lift` are **fractions** (`0.25` = 25%), while `CampaignStats.rates` are already **percentages** (`25` = 25%). Multiply one, not the other. `zScore` is signed toward the winner and `0` when undefined; `pValue` is one-sided and `1` when undefined.
- **npm** — the error parser claimed a `{"error":"rate_limited"}` non-envelope body exists. It does not; the CSRF gate's `{"error":"csrf_failed"|"csrf_origin"}` is the only one, and a Bearer request never trips it.
- **go** — `CampaignAbResult.Status` is always empty (this endpoint returns the evaluation only) and is now `Deprecated:` in favour of `Campaign.AbTest.Status`.
- **go** — `Client.Timeout = 0` was documented as disabling the timeout. `NewClient` also sets `HTTPClient.Timeout`, so requests stayed bounded at 30s; clear both, or supply your own `HTTPClient`.
- **dotnet** — `EmailUpdateAsync` accepts a relative phrase (`in 1 min`) as well as an ISO 8601 timestamp, must be in the future, and is capped at 30 days; it works only while the email is still `scheduled`.
- **npm**, **dotnet** — an inline batch near the 40-email boundary can take ~100s server-side, past the 30s default client timeout. Raise it, and always pass an idempotency key, because a client that gives up mid-request cannot tell what was already sent.
- **python** — `contact_properties.create` had `key` and `type` backwards: `key` is required, `type` defaults to `string` and is immutable once created.
- **python** — `emails.batch` documented the partial-failure contract only for keyed calls. A failure part way through **always** names the emails that already went out on `sent` / `sent_count`. The key changes the status, not the data: with one, the partial answer is recorded and replayed as its canonical 429/503 so this SDK's retry sends nothing; without one, the same body comes back as a 422 precisely so the retry cannot re-run the batch.
- **java** — `AutomationConnection.of(from, to, type)` said "e.g. `next`, `yes`, `no`". `yes` and `no` are not real: the API accepts exactly `next` (alias `default`), `condition_met`, `condition_not_met`, `event_received` and `timeout`. Also documented: `CreateTopicRequest.defaultSubscription` is required and immutable (which is why `UpdateTopicRequest` has no setter for it), and which `PATCH /automations/:id` fields require a disabled automation.
- **cli** — `emails batch` now says in its own description that batch items reject `attachments` and `scheduled_at`.
- **go**, **ruby** — the top-of-package quickstarts sent to `user@example.com`, which the API refuses with `422 reserved_recipient`. They use `delivered@mailblastr.dev`, the mailbox simulator, and say to swap in a real recipient. The CLI's help examples were corrected the same way.

---

### Breaking — `update-step` now has its own type in four SDKs

`PATCH /automations/:id/steps/:stepId` re-validates the whole step and forwards only
`type` and `config`. Four SDKs handed it the CREATE type, which advertised a `key` the
server silently discards and made `type` look optional when it is required. Each now has
a dedicated update type, so the compiler enforces the real contract:

| SDK | was | now |
|---|---|---|
| npm | `updateStep(id, stepId, AddAutomationStepOptions)` | `UpdateAutomationStepOptions` (`type` required, `key?: never`) |
| .NET | `AutomationUpdateStepAsync(…, AutomationAddStepOptions, …)` | `AutomationUpdateStepOptions` |
| java | `updateStep(id, stepId, AutomationStep)` | `UpdateAutomationStepRequest` |
| rust | (see breaking change 3) | `UpdateAutomationStepOptions::new(type)` |

Passing the create type no longer compiles. Move `type` and `config` across and drop
`key` — it never did anything on this route.

---

### Fixed — the same defects, in the SDKs that had been missed

The 4.0.0-era audit ran one agent per package, so a defect found in one language was
fixed there and left standing everywhere else. This release re-checked every corrected
contract across all nine. **The two worst were each present in five more SDKs than the
release that "fixed" them said.**

#### The webhook secret was decoded strictly in five more SDKs

The `whsec_` suffix must be base64-decoded the way Node's `Buffer.from(suffix,'base64')`
does — ignoring characters outside the alphabet, accepting the URL-safe `-`/`_`
spellings, needing no `=` padding, and dropping a lone trailing character that encodes no
byte. A strict decoder derives a **different key than the signer**, so every genuine
delivery comes back `no_match`. 4.0.0 fixed this in python and php. It was still wrong in:

- **go** — `base64.StdEncoding.DecodeString` (strict padding *and* alphabet).
- **rust** — `general_purpose::STANDARD`, whose `RequireCanonical` padding mode rejects
  an unpadded suffix outright.
- **dotnet** — `Convert.FromBase64String`, then a `catch (FormatException)` that fell
  through to using the raw string — including the `whsec_` prefix — as the key.
- **java** — `Base64.getDecoder()`, which tolerates missing padding but throws on `-`/`_`
  and then fell back to raw UTF-8 bytes the same way.
- **ruby** — `unpack1("m")` tolerates padding but **discards** `-` and `_` instead of
  translating them, so a URL-safe suffix produced a silently truncated key.

All five now derive the same bytes as the server. Each carries a test that signs with the
key the server would derive from an unpadded and from a URL-safe secret and requires a
match, so this cannot regress in one language again.

#### Webhook headers were unreadable from the frameworks people actually use

`readHeader` in **npm** enumerated with `Object.keys()`. A WHATWG `Headers` instance has
no own enumerable properties, so that returns `[]` — and a WHATWG `Headers` is exactly
what `request.headers` **is** in Next.js App Router route handlers, Remix, Hono and
Cloudflare Workers. Verification reported `missing_headers` for every delivery on the
most common way to receive one. It now reads `entries()` first, and falls back through
the other shapes.

**ruby** had the same class of bug from the other direction: `read_header` opened with
`return nil unless headers.is_a?(Hash)`, and Rails' `request.headers` is an
`ActionDispatch::Http::Headers`, not a Hash — so every Rails receiver failed the type
guard before a single lookup.

#### Other cross-SDK gaps closed

- **npm** — `ReceivedAttachment.size` was `number` while the list route serializes
  `size ?? null`. TypeScript said it was always a number, so `a.size.toFixed()`
  type-checked and threw at runtime. Now `number | null`, matching rust and dotnet, and
  matching npm's own sibling `AttachmentMeta.size`, which was already correct.
- **python** — the `Idempotency-Key` gate was a bare truthiness test while the line below
  it stringified the value, so `0` was dropped exactly as php's `!empty()` did. A retried
  send keyed `0` could deliver twice.
- **rust** — `SendEmailOptions.from`/`.subject` (and the batch equivalents) are plain
  `String` with no `skip_serializing_if`, so a template send serialized empty strings over
  the template's own from and subject — the defect go fixed in 4.0.0.
- **cli** — `emails send` declared `--from` and `--subject` as `requiredOption`s, so
  commander aborted before the request even when `--template-id`/`--template-alias` was
  supplied. Sending from a template was impossible from the CLI.

#### Documentation corrected against the route handlers

- **go** — five list methods promised "returns every row"; the server caps an unpaginated
  list at `UNPAGINATED_MAX = 1000` and reports it on `has_more`. `Segments.Contacts` is
  the dangerous one: a segment easily exceeds 1,000 and callers were treating a prefix as
  the whole membership.
- **ruby, php, go, rust, java, cli** — `cancel` was documented as "returns it to draft"
  everywhere. A campaign already fanning out becomes the terminal `canceled` instead;
  `canceled` was also missing from go's and rust's status vocabularies.
- **go** — the only A/B fixture in the package was written in percentages, which the
  server never emits: `rate` and `lift` are fractions while campaign stats `rates` are
  percentages. A reader copying it shipped a 10,000% rate.
- **python, php, java** — the 202 `{ queued, queued_count }` answer for a batch over the
  sync threshold appeared nowhere; callers could not tell a sent batch from a queued one.

---

### Deliberately not changed

- **Go keeps `UpdateAutomationStepRequest.Key` and `CampaignAbResult.Status`, marked `Deprecated:`.** Both are ignored by the API, so deleting them buys nothing at runtime and breaks code that compiles today. Rust's `with_key()` was deleted instead — the asymmetry is deliberate: a struct field a caller may be setting from a map is not the same liability as a builder method whose entire purpose was to produce a no-op.
- **No compatibility shim for the npm `Campaign.statistics` narrowing.** A `links` that keeps typechecking and keeps being `undefined` is the defect, not the migration.

### Known gaps — carried forward

Unchanged from 4.0.0 and listed so they are not mistaken for closed.

- **`PATCH /campaigns/:id` still cannot clear every clearable field everywhere.** Go can clear all six; Rust and npm cannot clear `reply_to` or `preview_text`; .NET's `CampaignUpdateOptions` can clear none.
- **`retry_after` is exposed on the error object in Python only.**
- **`RateLimit-Limit` / `RateLimit-Remaining` / `RateLimit-Reset` are not surfaced** by any package.
- **`GET /domains/dns/callback` is intentionally not modelled** — a browser redirect target with no auth and no JSON body.
- **Rust still has no `#[non_exhaustive]`** on its public structs — and this release is the first one where that cost a caller something (see the footnote under *Breaking changes*).

### Verification

Every package was built and its suite run from this tree with the commands CI uses, by the lead rather than self-reported by the agent that wrote the fix. Counts were read off those runs.

| Package | Command | Result |
|---|---|---|
| mailblastr-npm | `npm run build` + `npm run typecheck:test` + `npm test` | 77 pass / 0 fail |
| mailblastr-cli | `npm test` (against the local SDK, not the registry) | 79 pass / 0 fail |
| mailblastr-python | `python3 -m unittest discover -s tests -t .` | 149 pass / 0 fail |
| mailblastr-ruby | every `test/*_test.rb` loaded under `ruby -Ilib -Itest` | 82 runs, 728 assertions, 0 failures |
| mailblastr-php | `php -l` over all 34 files in `src` + `tests`, then `php tests/run.php` | 276 pass / 0 fail |
| mailblastr-go | `go build ./...`, `go vet ./...`, `gofmt -l .`, `go test -count=1 ./...` | 87 pass / 0 fail; vet and gofmt clean |
| mailblastr-rust | `cargo test` + `cargo package --list` | 77 pass / 0 fail (26 unit, 42 integration, 9 doc) |
| mailblastr-dotnet | `dotnet test tests/Mailblastr.Tests/Mailblastr.Tests.csproj` | 67 pass / 0 fail |
| mailblastr-java | `mvn -B compile test-compile` + `com.mailblastr.tests.AllTests` | 283 pass / 0 fail |

1,177 checks across the nine, every one run from this tree by the lead rather than
self-reported by the agent that wrote the fix. `go test` was forced with `-count=1`: a
bare run reports `(cached)` and would happily pass against pre-change artifacts.

The defects that fail silently are pinned by tests that assert the **request body, query string or header list on the wire**, not merely that the call returned:

- **python** — `test_flask_style_headers_that_iterate_pairs`, `test_unpadded_whsec_secret_verifies`, `test_url_safe_whsec_secret_verifies`, `test_lenient_base64_matches_node_byte_for_byte`, `test_unrecognized_secret_still_falls_back_to_raw_bytes`, `test_unparseable_retry_after_keeps_the_error_envelope`, `test_non_finite_retry_after_is_ignored`, `test_http_date_retry_after_is_honoured`.
- **php** — the idempotency cases assert the exact `Idempotency-Key:` header list for `'0'`, `0` and `12345`, and its absence for `null`, `''` and an array; `webhook.verify: non-strict whsec_ suffix matches the server key` signs with the key the server would derive and requires a match.
- **go** — `TestTemplateSendOmitsEmptyFromAndSubject`, `TestTemplateIdSendOmitsEmptyFromAndSubject`, `TestTemplateSendKeepsExplicitFromAndSubject`, `TestNonTemplateSendAlwaysSendsFromAndSubject`, `TestEmptyTemplateRefShadowsTemplateId`, `TestBatchTemplateItemOmitsEmptyFromAndSubject`, `TestBatchQueuedResponse`, `TestBatchInlineResponseIsNotQueued`, `TestDomainsUpdateCustomReturnPath`.
- **cli** — `webhooks verify needs no API key — it makes no request`, `webhooks verify rejects a non-numeric --tolerance instead of skipping the freshness check`, `automations update-step maps to updateStep and requires --type`, `contacts update refuses contradictory consent flags`, `templates create / update carry the declared-variable registry`, `--json applies to the error stream too`.
- **npm** — `a queued batch (202) surfaces queued + queued_count as success`, `an inline batch (200) reports no queued flag` (absent, not `false`), `campaign.statistics carries counts only — links is stats-only`, `a received email exposes domain_id and list rows expose object/id`.
- **ruby** — `test_namecheap_credentials_are_transmitted_verbatim`.

No test in any package makes a live API call — all drive an injected mock transport. Response-shape claims are verified against the route handlers and replayed fixtures, not observed traffic. Java's `pom.xml` still sets `<skipTests>true</skipTests>` (the tests are plain `main()` runners; `com.mailblastr.tests.AllTests` is the real entrypoint), and `dotnet test` must still be given the test csproj explicitly or it fails with MSB1003. Running the Java suite on macOS needs `JAVA_HOME` exported explicitly — no JDK is registered with the OS, so `mvn` fails with "Unable to locate a Java Runtime" until you point it at one.

---

## 4.0.0 — 2026-08-19

**No API surface, request shape or behaviour changes in any SDK.** Every client method,
argument and return type is identical to 3.0.1. The major version was chosen
deliberately for release hygiene, not because anything broke — with one real
consequence for Go, below.

Because this is a major, your dependency ranges will NOT pick it up automatically:
`^3.0.1` (npm), `~> 3.0` (RubyGems), `"3"` (Cargo) and friends all stay on 3.0.1
until you bump the constraint yourself. Nothing forces you to.

#### Go: the module path is now `.../mailblastr-go/v4`

Go encodes the major version in the import path for any major ≥ 2, so v4 is a genuine
source change for Go users:

```bash
go get github.com/shekhu10/mailblastr-sdks/mailblastr-go/v4
```

```go
import "github.com/shekhu10/mailblastr-sdks/mailblastr-go/v4"
```

Nothing else in the Go API moved — only the path. (Same migration shape as the v3 bump.)

#### Corrected: the `/emails` rate limit covers SENDS only

The php, ruby, rust, go and java READMEs each claimed the 30 requests/60s per-IP cap
covered reads as well as sends. That described the old server behaviour, where the send
limiter matched by mount prefix and therefore also caught the GETs. **The server changed:
the enforced limiter now depends on the HTTP method**, so reads are not capped —
`GET /emails`, `GET /emails/:id`, the `receiving` subtree and attachment listings.

Practically: paging a large list no longer risks a 429, and any client-side throttling
you added for reads can go. npm and python only ever documented retry handling
(scope-agnostic, still accurate); the CLI and .NET READMEs never mentioned rate limits.

## 3.0.1 — 2026-08-14

Documentation only. No API surface, request shape or behaviour changes in any SDK — you
can upgrade or skip this one freely.

**What actually changed is on the server, and it changes what your recipients receive.**
MailBlastr rewrites the links in your email into tracked redirects at send time. That
rewrite previously covered the `html` body only: the `text` body was transmitted verbatim,
so a URL you put in it reached the recipient untracked. A click from a mail client
rendering the plain-text alternative was therefore recorded as nothing, and your click
rates were undercounts of an unknown size.

URLs in the `text` body are now rewritten the same way as the ones in `html`. Nothing in
your integration needs to change — you keep sending ordinary links, and the SDK keeps
sending them to the API untouched:

```ts
await mb.emails.send({
  from: 'you@yourdomain.com',
  to: 'them@example.com',
  subject: 'Spring sale',
  html: '<p>Our <a href="https://acme.com/sale">deals</a> are live.</p>',
  text: 'Our deals are live: https://acme.com/sale',   // now tracked too
});
```

Two properties worth stating explicitly, because both are load-bearing and neither is
obvious:

- **Your content is never modified.** What you send is what `GET /emails/:id`, the
  dashboard and your campaign editor keep showing you, forever. The tracked URL exists
  only in the copy handed to the recipient.
- **Every recipient gets their own link,** so "which of these people clicked?" has an
  answer — and a delivered link keeps resolving to your destination indefinitely, on the
  hundredth click as on the first, well past your log-retention window.

The docstrings on `html` and `text` in every SDK now say all of this.

## 3.0.0 — 2026-08-09

**3.0.0 lands one day after 2.0.0, and it is worth saying plainly why.** It ships no new
endpoint, no new capability, and no fix for a bug that is losing anyone data today.

It exists because the 2.0.0 audit left one endpoint's signature wrong in three languages —
`GET /contacts/:id/topics` took no pagination argument in Go, Rust and .NET while its three
structural siblings took one in every package — and that cannot be fixed without breaking a
signature published on three registries. Once a major was unavoidable, the question stopped
being "is this worth a break?" and became "what else is in this category?" The answer was
every remaining place where the nine packages spell the same endpoint differently, and every
place where a PATCH field the API lets you clear could not be cleared from a typed SDK.

There are no customers on the platform yet. A break costs approximately nothing today and
compounds forever after. That is the whole justification — there is not a second one — and
it is why this release closes the naming and PATCH-shape work in one pass rather than
letting it leak out over the next several minors.

**Migrating is mechanical.** Every change below is a rename or an added optional argument;
none of it changes what any method sends, except the PATCH fields that could not previously
express "clear" at all. There is no behavioural change on the wire for code that already
compiles.

---

### Breaking changes

#### 1. The defect that forced the major: `contacts` topic subscriptions are paginated

`GET /contacts/:id/topics` is a paginating endpoint — contract-c §0.10 lists it with
`forceLimit: false`, meaning an unpaged call returns everything and supplying
`limit`/`after`/`before` restores paging, exactly like `GET /contacts/:id/segments` and
`GET /segments/:id/contacts`. npm, Python, Ruby, PHP, Java and the CLI already passed those
params. Go, Rust and .NET did not accept them at all, so from those three languages the
query string was unreachable.

```go
// Go — 2.0.0
topics, err := client.Contacts.GetTopics(contactID)

// Go — 3.0.0 (nil ⇒ every topic, exactly as before)
topics, err := client.Contacts.GetTopics(contactID, nil)
topics, err = client.Contacts.GetTopics(contactID, &mailblastr.ListParams{Limit: 50})
```

```rust
// Rust — 2.0.0
let topics = mailblastr.contacts.get_topics(contact_id).await?;

// Rust — 3.0.0
let topics = mailblastr.contacts.get_topics(contact_id, None).await?;
```

```csharp
// .NET — 2.0.0
var topics = await mailblastr.ContactRetrieveTopicsAsync(contactId);

// .NET — 3.0.0 — pagination is optional and goes BEFORE the CancellationToken.
var topics = await mailblastr.ContactRetrieveTopicsAsync(contactId);
var page   = await mailblastr.ContactRetrieveTopicsAsync(contactId, new PaginationOptions { Limit = 50 });
```

The .NET insertion has the same hazard 2.0.0 documented for its other list methods: a call
site that passed the `CancellationToken` **positionally** must now name it
(`cancellationToken: ct`). Named and default call sites are unaffected.

Paging you cannot drive is not a fix, so the response model was completed at the same time:
**`ContactTopics` gained `has_more`** in Rust (`pub has_more: bool`) and .NET
(`public bool HasMore`). npm and Go already modelled it; the route returns the standard
`{ object, has_more, data }` list envelope.

#### 2. One name per endpoint, across all nine packages

Six endpoints and one local helper had two, three or four public names across the libraries.
Each now has one.

The rule, applied uniformly, is that a method returning a list is `list<Noun>`, a method
fetching one thing is `get<Noun>`, and a method minting something is `create<Noun>`. Where
that agreed with the majority spelling — `getRaw` (5 of 8), `listAttachments` (7 of 8),
`createImportUpload` (4 of 8, the largest of four spellings), `verify` (6 of 8) — the
majority simply won. Twice it did not, and the rule won anyway: `ai` was a dead 4/4 tie, and
`addresses` was the majority at 6 of 8 but is a list method. Picking the rule over the head
count in those two is the only way the table below is a rule rather than a poll, and a rule
is what stops the next endpoint from re-opening the argument.

| Endpoint | 3.0.0 name | Was, and where |
| --- | --- | --- |
| `POST /automations/:id/ai` | `createWithAi` | `ai` in npm, Python, PHP, Java |
| `GET /emails/receiving/addresses` | `listAddresses` | `addresses` in npm, Python, Ruby, PHP, Java |
| `GET /emails/receiving/:id/attachments` | `listAttachments` | `attachments` in Python |
| `GET /emails/receiving/:id/raw` | `getRaw` | `raw` in Python and Ruby; `ReceivedEmailDownloadRawAsync` in .NET |
| `GET /emails/receiving/:id/attachments/:aid` | `getAttachment` | `ReceivedEmailDownloadAttachmentAsync` in .NET |
| `POST /audiences/:id/contacts/import/upload` | `createImportUpload` | `import_upload` in Python and Ruby, `ImportUpload` / `ImportUploadWithContext` in Go, `uploadUrl` in PHP |
| webhook signature verification | `verify` | `verify_signature` in Ruby and Rust; `WebhookVerifySignature` in .NET |

Each package keeps its own casing convention — `createWithAi` in npm/PHP/Java,
`create_with_ai` in Python/Ruby/Rust, `CreateWithAi` in Go,
`AutomationCreateWithAiAsync` in .NET — so this is one name per endpoint, not one
identifier. .NET in particular keeps `Retrieve` as its verb for fetching a modelled object
(`ReceivedEmailRetrieveAsync`, `EmailRetrieveAttachmentAsync`, `ContactRetrieveTopicsAsync`
are all unchanged); the two rows it moved to `Get` are the raw-**byte** downloads, which is
the same distinction the `getRaw` / `getAttachment` spelling draws in the other eight.

```ts
// npm
await mb.emails.receiving.addresses();          // 2.0.0
await mb.emails.receiving.listAddresses();      // 3.0.0

await mb.automations.ai(id, { prompt });        // 2.0.0
await mb.automations.createWithAi(id, { prompt }); // 3.0.0
```

```python
# Python
mailblastr.Emails.Receiving.raw(email_id)            # 2.0.0
mailblastr.Emails.Receiving.get_raw(email_id)        # 3.0.0

mailblastr.Emails.Receiving.attachments(email_id)    # 2.0.0
mailblastr.Emails.Receiving.list_attachments(email_id)  # 3.0.0

mailblastr.Contacts.import_upload({...})             # 2.0.0
mailblastr.Contacts.create_import_upload({...})      # 3.0.0
```

```ruby
# Ruby
Mailblastr::Webhooks.verify_signature(payload, headers, secret)  # 2.0.0
Mailblastr::Webhooks.verify(payload, headers, secret)            # 3.0.0
```

```php
// PHP
$slot = $mailblastr->contacts->uploadUrl(          // 2.0.0
    ['audienceId' => $audienceId, 'filename' => 'leads.csv', 'size' => $bytes]);
$slot = $mailblastr->contacts->createImportUpload( // 3.0.0
    ['audienceId' => $audienceId, 'filename' => 'leads.csv', 'size' => $bytes]);
```

```rust
// Rust — 2.0.0
let result = mailblastr.webhooks.verify_signature(raw_body, &headers, secret, &VerifyWebhookOptions::default());
// Rust — 3.0.0
let result = mailblastr.webhooks.verify(raw_body, &headers, secret, &VerifyWebhookOptions::default());
```

The free function `verify_webhook_signature` (Rust) and the static
`Webhooks.verifyWebhookSignature` (Java) are unchanged — only the method hanging off the
resource was renamed, so the two spellings no longer describe the same call.

**The CLI's command names are deliberately unchanged.** `emails receiving addresses`,
`emails receiving raw <id>`, `automations ai <id>` and `contacts import-upload` all still
work; only the SDK calls behind them moved. A subcommand is already scoped by its resource
path, so the shell surface stays short. That convention is now written down in the CLI
README as a table mapping each command to the SDK method it invokes, so the difference reads
as a rule rather than as drift.

#### 3. Four .NET methods renamed to the name the other eight already used

Same principle as above, but these were .NET-only outliers on endpoints where the other
eight packages already agreed.

| Endpoint | 2.0.0 (.NET) | 3.0.0 (.NET) | The other eight |
| --- | --- | --- | --- |
| `PATCH /events/:id` | `EventUpdateSchemaAsync` | `EventUpdateAsync` | `update` |
| `POST /webhooks/:id/rotate` | `WebhookRotateSecretAsync` | `WebhookRotateAsync` | `rotate` |
| `GET /domains/mx-check` | `DomainCheckMxAsync` | `DomainMxCheckAsync` | `mxCheck` / `mx_check` / `MxCheck` |
| `PATCH /emails/:id` | `EmailRescheduleAsync` | `EmailUpdateAsync` | `update` |

```csharp
// .NET
await mailblastr.EmailRescheduleAsync(id, "2026-08-01T09:00:00Z"); // 2.0.0
await mailblastr.EmailUpdateAsync(id, "2026-08-01T09:00:00Z");     // 3.0.0
```

#### 4. npm: the client is `Mailblastr`, not `MailBlastr`

npm was the only package spelling the brand with an internal capital B in its published
symbols. Every other package — and every package name on every registry — is lowercase
`mailblastr`. Three exported identifiers changed:

```ts
// 2.0.0
import { MailBlastr, type MailBlastrError, type MailBlastrOptions } from 'mailblastr';
const mb = new MailBlastr('mb_xxxxxxxxx');

// 3.0.0
import { Mailblastr, type MailblastrError, type MailblastrOptions } from 'mailblastr';
const mb = new Mailblastr('mb_xxxxxxxxx');
```

`Result<T>` still resolves to `{ data: null, error: MailblastrError }`, so `const { data,
error } = await …` call sites do not change; only an explicit type annotation or a named
import does. The default export is unchanged in shape (`export default Mailblastr`), so
`import Mailblastr from 'mailblastr'` keeps working under any local name. **No compatibility
alias for the old casing is exported** — a silently-working `MailBlastr` would leave the
split in place, which is the thing this release is spending a major to remove. "MailBlastr"
remains the prose spelling of the product in every README.

The CLI was repointed to the new class name in the same change and now declares
`mailblastr: ^3.0.0`.

#### 5. Python and PHP: sent-email attachments are flat methods, not a sub-resource

Both packages reached a *received* email's attachments through a flat `listAttachments` but
a *sent* email's through a nested sub-resource — internally inconsistent, and the only two
packages doing it. The other six were already flat.

```python
# Python — 2.0.0
mailblastr.Emails.Attachments.list(email_id)
mailblastr.Emails.Attachments.get(email_id, attachment_id)

# Python — 3.0.0
mailblastr.Emails.list_attachments(email_id)
mailblastr.Emails.get_attachment(email_id, attachment_id)
```

```php
// PHP — 2.0.0
$mailblastr->emails->attachments->list(emailId: $id);
$mailblastr->emails->attachments->get(emailId: $id, attachmentId: $attachmentId);

// PHP — 3.0.0  (note the first parameter is now named $id)
$mailblastr->emails->listAttachments(id: $id);
$mailblastr->emails->getAttachment(id: $id, attachmentId: $attachmentId);
```

`Mailblastr\Resources\EmailAttachments` is deleted and `$mailblastr->emails->attachments`
no longer exists; `mailblastr.Emails.Attachments` is likewise gone. `emails->receiving` /
`Emails.Receiving` are untouched — those group a genuinely different resource, not a verb.

#### 6. PATCH fields that clear: Go, Rust and .NET can finally send an explicit JSON null

The API patches on key **presence**: a key that is absent leaves the field alone, and a key
present with `null` clears it (contract-b §4.5 for templates and §5.5 for topics, contract-c
§1.7 for campaigns and §5.1 for segment filters). A plain `string` with `omitempty`, or a
`string?` under `WhenWritingNull`, can only express two of those three states — so from Go,
Rust and .NET, "clear this field" was simply not sendable. npm, Python, Ruby, PHP and Java
were already fine, either because they type the field `| null` or because they pass a map
through untouched.

Each language got the idiomatic three-state carrier:

**Go — `Null[T]` with `Set` / `Clear`** (new, exported from the package root):

```go
// nil        -> key omitted, server leaves the field alone
// Set(v)     -> key sent with v
// Clear[T]() -> key sent as JSON null, server clears the field

client.Campaigns.Update(id, &mailblastr.UpdateCampaignRequest{
    SegmentId: mailblastr.Clear[string](),   // untarget the campaign
    ReplyTo:   mailblastr.Set([]string{"support@yourdomain.com"}),
})
```

**Rust — `Option<Option<T>>` plus a `clear_*` builder per field:**

```rust
// None            -> omitted
// Some(Some(v))   -> sent with v          (with_alias(..))
// Some(None)      -> sent as null         (clear_alias())
let opts = UpdateTemplateOptions::new().clear_alias().with_subject("Your receipt");
```

**.NET — `Patch<T>` with an implicit conversion from `T`:**

```csharp
await mailblastr.TemplateUpdateAsync(id, new TemplateUpdateOptions
{
    Subject = "Your receipt",        // set (implicit conversion)
    Alias   = Patch.Clear<string>(), // cleared server-side
                                     // Name omitted => unchanged
});
```

Which fields changed type:

| Package | Type | Fields whose declared type changed |
| --- | --- | --- |
| Go | `UpdateCampaignRequest` | `ReplyTo`, `PreviewText`, `SegmentId`, `TopicId`, `ScheduleTimezone`, `DailyBatchSize` |
| Go | `UpdateTemplateRequest` | `Alias`, `Subject`, `From`, `ReplyTo`, `Html`, `Text` |
| Go | `UpdateTopicRequest` | `Description` |
| Go | `SegmentFilterInput` | `Engagement` (`*Null[SegmentEngagement]`), `PropertyFilters` (`*[]PropertyFilter` — point it at an empty slice to clear) |
| Rust | `UpdateTemplateOptions` | `alias`, `subject`, `from`, `reply_to`, `html`, `text` |
| Rust | `UpdateTopicOptions` | `description` |
| Rust | `SegmentFilterOptions` | `engagement` (plus `clear_engagement()` / `clear_property_filters()`) |
| .NET | `TemplateUpdateOptions` | `Alias`, `Subject`, `From`, `ReplyTo`, `HtmlBody`, `TextBody` |
| .NET | `TopicUpdateOptions` | `Description` |
| .NET | `SegmentFilterOptions` | `Engagement` |

Setting a value still reads the same in Rust and .NET — `with_alias("x")` and
`Alias = "x"` both compile unchanged, because the builders were updated and `Patch<T>` takes
an implicit conversion from `T`. Go is the one language where a plain assignment must become
`mailblastr.Set(v)`.

Property predicates clear with an **empty array**, not a null: the contract replaces them
wholesale, so `[]` is a valid replacement meaning "none". Go needed a pointer for this
(`&[]mailblastr.PropertyFilter{}`) because a bare empty slice is dropped by `omitempty`;
Rust's `Some(vec![])` and .NET's empty `List` already serialized correctly and now have a
`clear_property_filters()` / documented empty-list contract to make it findable.

#### 7. Go and .NET: the segment filter is two types, not one

Both packages used a single `SegmentFilter` as the request body *and* the response body. The
two directions genuinely differ: going in, every field is optional and `property_filters` /
`engagement` are three-state; coming out, `filter.status` is always present and
`property_filters` is always an array. One type has to pick a side, and both had picked the
request reading — so the response model declared `omitempty` / `WhenWritingNull` on fields
the API always sends. npm and Rust had already split them.

```go
// Go — 2.0.0
Filter: &mailblastr.SegmentFilter{Status: "subscribed"}

// Go — 3.0.0
Filter: &mailblastr.SegmentFilterInput{Status: "subscribed"}
```

```csharp
// .NET — 2.0.0
Filter = new SegmentFilter { Status = "subscribed" }

// .NET — 3.0.0
Filter = new SegmentFilterOptions { Status = "subscribed" }
```

`SegmentFilter` survives in both as the **response** type carried on `Segment`, and it
tightened accordingly: Go's `Status` / `EmailContains` / `PropertyFilters` / `Engagement`
lost `omitempty`, and .NET's `Status` is now non-nullable (`= "all"`) with
`PropertyFilters` non-nullable (`= new()`).

#### 8. Model renames

| Was | Now | Packages | Already correct in |
| --- | --- | --- | --- |
| `ReceivedAddressStats` (Go), `ReceivedEmailAddressStats` (.NET) | `ReceivingAddressStats` | Go, .NET | npm, Rust |
| `WebhookVerificationResult` | `VerifyWebhookResult` | .NET | npm, Go, Rust, Java |
| `CreateEmailBaseOptions` | `SendEmailOptions` | Rust | npm |

The Rust one is worth a note: `Base` was an implementation detail — it existed only because
`BatchEmailOptions` is the same shape minus `attachments` and `scheduled_at` — leaking into
the type users name most often.

```rust
// Rust — 2.0.0
use mailblastr::CreateEmailBaseOptions;
let email = CreateEmailBaseOptions::new(from, to, subject).with_html("<p>Hi</p>");

// Rust — 3.0.0
use mailblastr::SendEmailOptions;
let email = SendEmailOptions::new(from, to, subject).with_html("<p>Hi</p>");
```

The five methods that take it — `emails.send`, `emails.send_with_idempotency_key`, the
deprecated `emails.batch` alias, `batch.send` and `batch.send_with_idempotency_key` — are
otherwise unchanged. `batch.send_emails` / `batch.send_emails_with_idempotency_key` still
take `BatchEmailOptions`, which keeps its name: there, "batch" is the shape, not a
leaked base class.

#### 9. Go: the module path is now `.../mailblastr-go/v3`

Required by Go's module rules for a major above v1, and it is the one change here you cannot
discover from a compiler error — an unchanged `/v2` import keeps resolving, and keeps serving
2.0.0.

```bash
go get github.com/shekhu10/mailblastr-sdks/mailblastr-go/v3
```

```go
import "github.com/shekhu10/mailblastr-sdks/mailblastr-go/v3"
```

The package name is still `mailblastr`, so only the import line changes. The release
workflow's proxy warm-up derives the module path from `go.mod` rather than hard-coding it,
so the v3 path is warmed without a workflow edit.

---

### Deliberately not changed

Listed because each was considered and rejected, not because it was missed.

- **.NET keeps its flat `IMailblastr` interface** with `<Entity><Verb>Async` methods, rather
  than being reshaped into `client.<resource>.<verb>()` like the other eight. That is a
  whole-surface rewrite of every call site in every .NET consumer, and it is a different
  argument from "one name per endpoint" — it is "one *shape* across languages", which trades
  away .NET's own conventions. It is not obviously right, so it is not being done under cover
  of a major that was already happening.
- **The request-payload naming schemes stay split** — `…Options` in npm and Rust, `…Request`
  in Go and Java, `<Noun><Verb>Options` in .NET; `PaginationParams` / `ListParams` /
  `PaginationOptions` for the pagination bag. Every one of those matches its own ecosystem's
  house style. Renaming them would touch far more public surface than the endpoint-name work
  above while making each SDK read slightly foreign to its own users.
- **Rust still has no `#[non_exhaustive]`** on its public structs (verified: zero occurrences
  in `mailblastr-rust/src`). Today's field sets are complete against the contract, so this is
  not a present-day gap — but it is the mechanism that will turn the *next* additive API field
  into a forced major. It stays open because adding it is itself a breaking change for
  callers who construct or destructure with struct literals, and that call deserves its own
  decision rather than being smuggled in here.

### Known gaps — carried forward

- **`PATCH /campaigns/:id` still cannot clear every clearable field everywhere.** Go can now
  clear all six (`reply_to`, `preview_text`, `segment_id`, `topic_id`, `schedule_timezone`,
  `daily_batch_size`). Rust and npm can clear `segment_id`, `topic_id`, `schedule_timezone`
  and `daily_batch_size` but not `reply_to` or `preview_text`. .NET's `CampaignUpdateOptions`
  can clear none of them, and says so in its own doc comments. Campaign PATCH was the one
  body where the fix was applied unevenly, and finishing it is the first thing to do in the
  next cycle.
- **`retry_after` is exposed on the error object in Python only**, unchanged from 2.0.0.
- **`RateLimit-Limit` / `RateLimit-Remaining` / `RateLimit-Reset` are not surfaced** by any
  package; doing so changes the return shape of every method.
- **`GET /domains/dns/callback` is intentionally not modelled** — a browser redirect target
  with no auth and no JSON body.

### Verification

Every package was built and its suite run from this tree, with the same commands CI uses.
The counts below were read off those runs.

| Package | Command | Result |
|---|---|---|
| mailblastr-npm | `npm run build` + `npm run typecheck:test` + `npm test` | 70 pass / 0 fail |
| mailblastr-cli | `npm test` | 71 pass / 0 fail |
| mailblastr-python | `python3 -m unittest discover -s tests -t .` | 138 pass / 0 fail |
| mailblastr-ruby | every `test/**/*_test.rb` loaded under `ruby -Ilib -Itest` | 78 runs, 715 assertions, 0 failures |
| mailblastr-php | `php -l` over `src` + `tests`, then `php tests/run.php` | 267 pass / 0 fail |
| mailblastr-go | `go build ./...`, `go vet ./...`, `gofmt -l .`, `go test ./... -count=1` | 77 pass / 0 fail; vet and gofmt clean |
| mailblastr-rust | `cargo build` + `cargo test` | 64 pass / 0 fail (25 unit + 30 integration + 9 doc) |
| mailblastr-java | `mvn -B compile test-compile` + `java -cp target/classes:target/test-classes com.mailblastr.tests.AllTests` | 264 pass / 0 fail |
| mailblastr-dotnet | `dotnet test tests/Mailblastr.Tests/Mailblastr.Tests.csproj` | 60 pass / 0 fail |

The three-state PATCH work and the topics pagination are the changes most able to fail
silently — a field that serializes as absent instead of null does not raise, it just does
nothing — so each is pinned by a test that asserts the **request body or query string on the
wire**, not merely that the call returned:

- **Go** — `TestContactsGetTopicsPagination` and `TestContactsGetTopicsNilParamsSendsNoQuery`;
  `TestCampaignsUpdateClearsFieldsWithExplicitNull` and `TestCampaignsUpdateSendsSetValues`;
  `TestTemplatesUpdateClearsFieldsWithExplicitNull`;
  `TestSegmentsUpdateClearsEngagementAndPropertyFilters`.
- **Rust** — `contact_topics_are_paginated_like_their_siblings`,
  `template_update_can_clear_fields_with_an_explicit_null`,
  `segment_update_can_clear_the_engagement_predicate`,
  `topic_update_can_clear_the_description`.
- **.NET** — `ContactRetrieveTopics_ForwardsPaginationAndReadsHasMore`,
  `ContactRetrieveTopics_WithoutPagination_SendsNoQuery`,
  `TemplateUpdate_ClearsFieldsWithAnExplicitNull`,
  `TemplateUpdate_OmitsUntouchedFieldsAndSendsSetValues`, `TopicUpdate_ClearsTheDescription`,
  `SegmentUpdate_ClearsTheEngagementPredicate`, `EventUpdate_PatchesAndCanClearTheSchema`,
  `DomainMxCheck_PassesTheNameQuery`.

The renames are covered by each package's existing suite plus Java's `ContractParityTest`
and the CLI's command-to-method assertions, which now name `listAddresses` and
`createWithAi` explicitly.

Two harness details still bite, unchanged from 2.0.0: Java's `pom.xml` sets
`<skipTests>true</skipTests>` (the tests are plain `main()` runners, so `mvn test` reports
BUILD SUCCESS having run zero of them — `com.mailblastr.tests.AllTests` is the real
entrypoint), and `dotnet test` must be given the test csproj explicitly or it fails with
MSB1003.

No test in any package makes a live API call — all drive an injected mock transport.
Response-shape claims are verified against the contract documents and replayed fixtures, not
observed traffic.

---

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
