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

`--help` works at every level: `mailblastr --help`, `mailblastr emails --help`, `mailblastr emails send --help`.

## Usage

### Emails

```bash
mailblastr emails send --from 'Acme <hi@yourdomain.com>' --to 'a@b.com' --subject 'hello' --html '<p>hi</p>'
mailblastr emails send --from 'Acme <hi@yourdomain.com>' --to 'a@b.com,c@d.com' --subject 'hi' \
  --template-id tmpl_welcome --variables '{"first_name":"Ada"}' --scheduled-at 2026-08-01T09:00:00Z
mailblastr emails list --limit 20
mailblastr emails get em_123
mailblastr emails update em_123 --scheduled-at 2026-08-02T09:00:00Z
mailblastr emails cancel em_123
mailblastr emails attachments em_123
mailblastr emails attachment em_123 att_456
```

`--to`, `--cc`, `--bcc` and `--reply-to` are repeatable and accept comma-separated values.

Batch-send up to 100 emails in one request from a JSON file (an array of send payloads) or inline JSON:

```bash
mailblastr emails batch --file ./batch.json
mailblastr emails batch --data '[{"from":"hi@yourdomain.com","to":["a@b.com"],"subject":"hi","text":"hello"}]'
```

### Received (inbound) email

```bash
mailblastr emails receiving list --limit 20
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
mailblastr domains delete dom_123
```

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
mailblastr contacts get a@b.com --domain yourdomain.com   # or by contact id, no --domain needed
mailblastr contacts update con_123 --unsubscribed
mailblastr contacts add-to-segment con_123 seg_123
mailblastr contacts remove-from-segment con_123 seg_123
mailblastr contacts topics con_123
mailblastr contacts set-topics con_123 --topics '[{"id":"top_123","subscription":"opt_out"}]'
mailblastr contacts delete con_123
```

### Contact properties & audiences

```bash
mailblastr contact-properties create --key company --type string --fallback-value 'your company'
mailblastr contact-properties list
mailblastr contact-properties update prop_123 --fallback-value 'n/a'
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
mailblastr campaigns send camp_123
mailblastr campaigns send camp_123 --scheduled-at 2026-08-01T09:00:00Z
mailblastr campaigns stats camp_123
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
mailblastr automations delete-step auto_123 step_456
mailblastr automations update auto_123 --status enabled
mailblastr automations runs auto_123
mailblastr automations run auto_123 run_456
mailblastr automations stop auto_123

# Fire a custom event — only yourdomain.com's automations are triggered
mailblastr events send --domain yourdomain.com --name signup.completed --email a@b.com --data '{"plan":"pro"}'
mailblastr events list
```

### Webhooks, API keys, logs, polls

```bash
mailblastr webhooks create --endpoint https://yourapp.com/hooks --events email.delivered,email.bounced
mailblastr webhooks rotate wh_123
mailblastr webhooks test wh_123

mailblastr api-keys create --name CI --permission sending_access
mailblastr api-keys list
mailblastr api-keys delete key_123

mailblastr logs list --limit 100 --method POST --status 429
mailblastr logs get log_123

mailblastr polls list
mailblastr polls get em_123
```

## Documentation

Full API docs: <https://www.mailblastr.com/docs>

## License

MIT
