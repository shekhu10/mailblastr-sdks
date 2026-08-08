# mailblastr-cli

Official command-line interface for the [MailBlastr](https://www.mailblastr.com) email API. Wraps the [`mailblastr`](https://www.npmjs.com/package/mailblastr) Node.js SDK.

## Install

```bash
npm i -g mailblastr-cli
```

Requires Node.js 18+.

## Authentication

Set your API key once:

```bash
export MAILBLASTR_API_KEY=mb_xxxxxxxxx
```

Or pass `--api-key mb_xxxxxxxxx` to any command. Use `MAILBLASTR_BASE_URL` (or `--base-url`) to target a different API host.

## Output

Every command prints the API response as pretty-printed JSON. Pass `--json` for raw compact JSON (handy for piping to `jq`). On failure the API error `{ statusCode, name, message }` is printed to stderr and the command exits `1`.

One endpoint reports failure inside a `200` body rather than as an error: `webhooks test` returns `{ ok: false, error }` when the delivery did not land. The CLI prints that body to stdout and still exits `1`, so `mailblastr webhooks test wh_123 && deploy` behaves as you would expect.

`--help` works at every level: `mailblastr --help`, `mailblastr emails --help`, `mailblastr emails send --help`.

### Pagination

List commands take `--limit` (integer `1`–`100`, default `20`) plus one of `--after` / `--before` — cursors are item ids, and passing both is rejected. Responses are `{ "object": "list", "has_more": bool, "data": [...] }`; page forward by feeding the last `data[].id` back as `--after`. Note `domains list`, `api-keys list`, `topics list`, `contacts list`, `segments list`, `campaigns list` and `contact-properties list` return the whole collection when no pagination flag is given, while `templates list`, `webhooks list`, `audiences list`, `automations list`, `automations runs` and `events list` cap at 20.

The nested list commands page the same way: `contacts segments`, `contacts topics`, `segments contacts`, `emails receiving attachments` and `automations runs`. A few endpoints are deliberately unpaginated and take no cursor flags — `emails sources`, `emails receiving addresses` and `campaigns engagement` (whose three lists are each capped at 500 rows server-side).

## Usage

### Emails

```bash
mailblastr emails send --from 'Acme <hi@yourdomain.com>' --to 'a@b.com' --subject 'hello' --html '<p>hi</p>'
mailblastr emails send --from 'Acme <hi@yourdomain.com>' --to 'a@b.com,c@d.com' --subject 'hi' \
  --template-id tmpl_welcome --variables '{"first_name":"Ada"}' --scheduled-at 2026-08-01T09:00:00Z
mailblastr emails list --limit 20
mailblastr emails list --status delivered --search 'invoice'
mailblastr emails sources                      # per-source metrics: one row per campaign/automation + an individual-sends roll-up
mailblastr emails get em_123
mailblastr emails update em_123 --scheduled-at 2026-08-02T09:00:00Z
mailblastr emails cancel em_123
mailblastr emails attachments em_123
mailblastr emails attachment em_123 att_456
```

`--to`, `--cc`, `--bcc` and `--reply-to` are repeatable and accept comma-separated values.

Attach files with `--attachment <path>` (read and base64-encoded locally) or `--attachment-url <url>` (fetched server-side). Both are repeatable; the API caps an attachment at 25 MB and a message at 40 MB decoded, and the CLI checks local files against those limits before sending.

```bash
mailblastr emails send --from hi@yourdomain.com --to a@b.com --subject 'Your invoice' \
  --text 'Attached.' --attachment ./invoice.pdf
mailblastr emails send --from hi@yourdomain.com --to a@b.com --subject 'Welcome' \
  --template-alias welcome --variables '{"first_name":"Ada"}'
```

`--idempotency-key` accepts **1–255 characters**, measured after the server trims the value — 255, not 256 — and is honoured only by `emails send` and `emails batch`. The CLI sends the key verbatim and lets the server be the authority: an out-of-range key comes back as `400 invalid_idempotency_key`. Reusing a key with a different body is rejected (`409 invalid_idempotent_request`); replaying it with the same body returns the original response instead of sending twice. Every other command ignores the header, so a retry there creates a second resource.

Batch-send up to 100 emails in one request from a JSON file (an array of send payloads) or inline JSON:

```bash
mailblastr emails batch --file ./batch.json
mailblastr emails batch --data '[{"from":"hi@yourdomain.com","to":["a@b.com"],"subject":"hi","text":"hello"}]'
```

### Received (inbound) email

```bash
mailblastr emails receiving list --limit 20
mailblastr emails receiving addresses                                         # per-address inbound stats
mailblastr emails receiving get rem_123
mailblastr emails receiving attachments rem_123
mailblastr emails receiving attachment rem_123 att_456 --output invoice.pdf   # default filename: the attachment id
mailblastr emails receiving raw rem_123 --output message.eml                  # default filename: <id>.eml
mailblastr emails receiving forward rem_123 --from you@yourdomain.com --to a@b.com,c@d.com
mailblastr emails receiving reply rem_123 --from you@yourdomain.com --html '<p>thanks!</p>'
mailblastr emails receiving delete rem_123
```

`attachment` and `raw` download binary content: the file is written to `--output` (or the default filename) and the CLI prints where it was saved. `reply` requires at least one of `--html` / `--text`.

### Domains

```bash
mailblastr domains add yourdomain.com
mailblastr domains list
mailblastr domains get dom_123
mailblastr domains verify dom_123
mailblastr domains update dom_123 --click-tracking --tls enforced
mailblastr domains update dom_123 --custom-return-path mail --receiving
mailblastr domains mx-check yourdomain.com
mailblastr domains records-csv dom_123 --output dns.csv   # default filename: <id>-dns-records.csv
mailblastr domains delete dom_123
```

`records-csv` answers `text/csv`, not JSON, so the CLI writes it to `--output` (or the default filename) and prints where it landed — the same shape as the `emails receiving` binary downloads.

One-click DNS — detect the provider, then apply the records via its API (auto-verifies after):

```bash
mailblastr domains dns detect dom_123
mailblastr domains dns cloudflare dom_123 --token cf_api_token
mailblastr domains dns godaddy dom_123 --key gd_key --secret gd_secret
mailblastr domains dns namecheap dom_123 --api-user ncuser --key nc_api_key
```

Claim a domain already verified by another account (start the claim, add the TXT record it returns, then verify):

```bash
mailblastr domains claim start yourdomain.com
mailblastr domains claim get dom_123
mailblastr domains claim verify dom_123
```

### Contacts (domain-first)

Each sending domain has its own contact pool, so `--domain` is required on create/list:

```bash
mailblastr contacts create --domain yourdomain.com --email a@b.com --first-name Ada
mailblastr contacts list --domain yourdomain.com
mailblastr contacts list --domain yourdomain.com --segment-id seg_123
mailblastr contacts list --audience-id aud_123          # plain audiences instead of a domain pool
mailblastr contacts get a@b.com --domain yourdomain.com   # or by contact id, no --domain needed
mailblastr contacts update con_123 --unsubscribed
mailblastr contacts add-to-segment con_123 seg_123
mailblastr contacts remove-from-segment con_123 seg_123
mailblastr contacts topics con_123
mailblastr contacts set-topics con_123 --topics '[{"id":"top_123","subscription":"opt_out"}]'
mailblastr contacts delete con_123
```

Bulk-import a CSV. Files up to 5 MB / 10,000 rows go inline with `--csv`; anything larger is uploaded directly to storage first — mint a presigned URL, `PUT` the file to it, then finish the import with the `storage_key` it returned:

```bash
mailblastr contacts import aud_123 --csv ./contacts.csv --segment-id seg_123
mailblastr contacts import-upload aud_123 --csv ./big-list.csv     # → { storage_key, upload_url, max_bytes, ... }
curl -X PUT --upload-file ./big-list.csv "$UPLOAD_URL"
mailblastr contacts import aud_123 --storage-key "$STORAGE_KEY"
```

`import` takes exactly one of `--csv` / `--storage-key`. The `upload_url` is a short-lived bearer credential — don't log it or paste it into a shared shell history.

### Contact properties & audiences

```bash
mailblastr contact-properties create --key company --type string --fallback-value 'your company'
mailblastr contact-properties list
mailblastr contact-properties update prop_123 --fallback-value 'n/a'
mailblastr contact-properties update prop_123 --clear-fallback
mailblastr contact-properties delete prop_123

mailblastr audiences create --name Newsletter
mailblastr audiences list
mailblastr audiences update aud_123 --name 'Newsletter EU'
mailblastr audiences import-sheet aud_123 --url 'https://docs.google.com/spreadsheets/d/...' --segment-name 'July leads'
mailblastr audiences delete aud_123
```

### Segments & topics

```bash
mailblastr segments create --domain yourdomain.com --name VIP
mailblastr segments create --domain yourdomain.com --name Actives --filter '{"status":"subscribed"}'
mailblastr segments list --domain yourdomain.com
mailblastr segments contacts seg_123

mailblastr topics create --domain yourdomain.com --name 'Product updates' --default-subscription opt_in
mailblastr topics list --domain yourdomain.com
```

### Campaigns

```bash
mailblastr campaigns create --domain yourdomain.com --from 'Acme <hi@yourdomain.com>' \
  --subject 'Summer sale' --html '<p>50% off</p>' --segment-id seg_123
mailblastr campaigns create --domain yourdomain.com --from 'Acme <hi@yourdomain.com>' \
  --subject 'Weekly digest' --html '<p>...</p>' --recurrence weekly --unsubscribe-policy domain \
  --followups '[{"condition":"not_opened","delay":"2 days","html":"<p>Did you see this?</p>"}]'
mailblastr campaigns send camp_123
mailblastr campaigns send camp_123 --scheduled-at 2026-08-01T09:00:00Z
mailblastr campaigns send camp_123 --scheduled-at 'in 1 min'
mailblastr campaigns stats camp_123
mailblastr campaigns engagement camp_123    # who opened, clicked and replied (each list capped at 500 rows)
mailblastr campaigns ab camp_123
mailblastr campaigns cancel camp_123
```

### Templates

```bash
mailblastr templates create --name Welcome --subject 'Hi {{first_name}}' --html '<p>Welcome!</p>'
mailblastr templates list
mailblastr templates duplicate tmpl_123 --name 'Welcome v2'
mailblastr templates publish tmpl_123
```

### Automations & events

```bash
mailblastr automations create --name 'Welcome series' --domain yourdomain.com --trigger contact.created
mailblastr automations add-step auto_123 --type send_email --config '{"template_id":"tmpl_welcome"}'
mailblastr automations update-step auto_123 step_456 --config '{"timeout":"12 hours"}'
mailblastr automations delete-step auto_123 step_456
mailblastr automations update auto_123 --status enabled
mailblastr automations runs auto_123
mailblastr automations runs auto_123 --status failed,skipped   # filtered before paging; repeatable
mailblastr automations run auto_123 run_456
mailblastr automations stop auto_123

# Author (or extend) the step graph from a prompt — the automation must be stopped, and this spends AI credits
mailblastr automations ai auto_123 --prompt 'Welcome new signups, then nudge anyone who has not opened after 2 days'
mailblastr automations ai auto_123 --prompt 'Send the upgrade nudge' --attach-from step_3 --attach-type condition_met

# Fire a custom event — only yourdomain.com's automations are triggered
mailblastr events send --domain yourdomain.com --name signup.completed --email a@b.com --data '{"plan":"pro"}'
mailblastr events list
mailblastr events create --name signup.completed --schema '{"plan":"string"}'
mailblastr events update evt_123 --schema '{"plan":"string","seats":"number"}'   # the name is immutable
```

### Webhooks, API keys, logs, polls

```bash
mailblastr webhooks create --endpoint https://yourapp.com/hooks --events email.delivered,email.bounced
mailblastr webhooks rotate wh_123
mailblastr webhooks test wh_123

mailblastr api-keys list

mailblastr logs list --limit 100 --method POST --status 429
mailblastr logs get log_123

mailblastr polls list
mailblastr polls get em_123
```

#### API keys are read-only from the CLI

`api-keys list` is the whole group. There is no `create`, `update` or `delete`, and that is deliberate: **keys are created, re-scoped and revoked in the [MailBlastr dashboard](https://www.mailblastr.com/app/api-keys)**, behind a signed-in session.

Every CLI invocation authenticates with an API key, so keeping key lifecycle out of the CLI means a key that leaks — from a shell history, a CI log, a `.env` someone pasted — cannot mint itself a replacement, widen its own permission or domain scope, or revoke the keys around it. Its blast radius stays fixed at what it could already do. The API enforces the same rule: `POST /api-keys`, `PATCH /api-keys/:id` and `DELETE /api-keys/:id` answer `403 dashboard_only` to any API-key caller.

`api-keys list` still shows everything you need to audit: each key's non-secret prefix, permission, domain scoping and `last_used_at`.

## Documentation

Full API docs: <https://www.mailblastr.com/docs>

## License

MIT
