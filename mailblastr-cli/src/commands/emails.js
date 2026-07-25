'use strict';

const fs = require('node:fs');
const path = require('node:path');

const { CliError, collect, parseJson, clean, withPagination, pagination } = require('../helpers');

/** Write a binary API result to disk and return a printable summary. */
function saveBinary(filePath, arrayBuffer) {
  const abs = path.resolve(filePath);
  const buf = Buffer.from(arrayBuffer);
  fs.writeFileSync(abs, buf);
  return { data: { object: 'file', path: abs, bytes: buf.length } };
}

/** Resolve the batch payload array from --file or --data, with friendly errors. */
function readBatchPayloads(opts) {
  if (opts.file && opts.data) throw new CliError('Provide only one of --file or --data.');
  if (!opts.file && !opts.data) throw new CliError('Provide --file <path> or --data <json array>.');
  let payloads;
  if (opts.data) {
    payloads = parseJson(opts.data, '--data');
  } else {
    let raw;
    try {
      raw = fs.readFileSync(opts.file, 'utf8');
    } catch (err) {
      throw new CliError(`Cannot read --file ${opts.file}: ${err.message}`);
    }
    try {
      payloads = JSON.parse(raw);
    } catch {
      throw new CliError(`Invalid JSON in --file ${opts.file}`);
    }
  }
  if (!Array.isArray(payloads)) {
    throw new CliError('Batch payload must be a JSON array of send objects.');
  }
  return payloads;
}

function register({ group, leaf, act }) {
  const emails = group('emails', 'Send and manage emails');

  act(
    leaf(emails, 'send', 'Send an email')
      .requiredOption('--from <from>', "sender, e.g. 'Acme <hi@yourdomain.com>'")
      .requiredOption('--to <address>', 'recipient (repeatable or comma-separated)', collect)
      .requiredOption('--subject <subject>', 'subject line')
      .option('--html <html>', 'HTML body')
      .option('--text <text>', 'plain-text body')
      .option('--cc <address>', 'cc recipient (repeatable or comma-separated)', collect)
      .option('--bcc <address>', 'bcc recipient (repeatable or comma-separated)', collect)
      .option('--reply-to <address>', 'reply-to address (repeatable or comma-separated)', collect)
      .option('--preview-text <text>', 'inbox preview text (preheader)')
      .option('--template-id <id>', 'send using a saved template')
      .option('--variables <json>', 'template variables as a JSON object')
      .option('--headers <json>', 'custom headers as a JSON object')
      .option('--scheduled-at <iso>', 'ISO 8601 timestamp to schedule the send')
      .option('--topic-id <id>', 'drop recipients unsubscribed from this topic')
      .option('--idempotency-key <key>', 'idempotency key for safe retries'),
    ({ client, opts }) =>
      client.emails.send(
        clean({
          from: opts.from,
          to: opts.to,
          subject: opts.subject,
          html: opts.html,
          text: opts.text,
          cc: opts.cc,
          bcc: opts.bcc,
          reply_to: opts.replyTo,
          preview_text: opts.previewText,
          template_id: opts.templateId,
          variables: parseJson(opts.variables, '--variables'),
          headers: parseJson(opts.headers, '--headers'),
          scheduled_at: opts.scheduledAt,
          topic_id: opts.topicId,
        }),
        opts.idempotencyKey ? { idempotencyKey: opts.idempotencyKey } : undefined,
      ),
  );

  act(
    leaf(emails, 'batch', 'Send up to 100 emails in one request')
      .option('--file <path>', 'path to a JSON file containing an array of send payloads')
      .option('--data <json>', 'inline JSON array of send payloads')
      .option('--idempotency-key <key>', 'idempotency key for safe retries'),
    ({ client, opts }) =>
      client.batch.send(
        readBatchPayloads(opts),
        opts.idempotencyKey ? { idempotencyKey: opts.idempotencyKey } : undefined,
      ),
  );

  act(leaf(emails, 'get <id>', 'Retrieve a sent email and its events'), ({ client, args: [id] }) =>
    client.emails.get(id),
  );

  act(
    withPagination(leaf(emails, 'list', 'List sent emails'))
      .option('--campaign-id <id>', 'only emails sent by this campaign')
      .option('--automation-id <id>', 'only emails sent by this automation')
      .option('--source <source>', "'individual' restricts to one-off API sends")
      .option('--domain-id <id>', 'only emails sent from this sending domain'),
    ({ client, opts }) =>
      client.emails.list(
        clean({
          ...pagination(opts),
          campaign_id: opts.campaignId,
          automation_id: opts.automationId,
          source: opts.source,
          domain_id: opts.domainId,
        }),
      ),
  );

  act(
    leaf(emails, 'update <id>', 'Reschedule a scheduled email').requiredOption(
      '--scheduled-at <iso>',
      'new ISO 8601 send time',
    ),
    ({ client, opts, args: [id] }) => client.emails.update(id, { scheduled_at: opts.scheduledAt }),
  );

  act(leaf(emails, 'cancel <id>', 'Cancel a scheduled email'), ({ client, args: [id] }) =>
    client.emails.cancel(id),
  );

  act(
    leaf(emails, 'attachments <emailId>', "List a sent email's attachments"),
    ({ client, args: [emailId] }) => client.emails.listAttachments(emailId),
  );

  act(
    leaf(emails, 'attachment <emailId> <attachmentId>', 'Retrieve one attachment of a sent email'),
    ({ client, args: [emailId, attachmentId] }) => client.emails.getAttachment(emailId, attachmentId),
  );

  // ---- inbound (received) email ----

  const receiving = emails.command('receiving').description('Manage inbound (received) email');

  act(
    withPagination(leaf(receiving, 'list', 'List received emails')).option(
      '--received-for <address>',
      'only messages received for this address',
    ),
    ({ client, opts }) =>
      client.emails.receiving.list(
        clean({ ...pagination(opts), received_for: opts.receivedFor }),
      ),
  );

  act(leaf(receiving, 'get <id>', 'Retrieve a received email'), ({ client, args: [id] }) =>
    client.emails.receiving.get(id),
  );

  act(
    leaf(receiving, 'attachments <id>', "List a received email's attachments"),
    ({ client, args: [id] }) => client.emails.receiving.listAttachments(id),
  );

  act(
    leaf(
      receiving,
      'attachment <id> <attachmentId>',
      'Download one attachment of a received email to a file',
    ).option('--output <path>', 'file to write (default: the attachment id)'),
    async ({ client, opts, args: [id, attachmentId] }) => {
      const result = await client.emails.receiving.getAttachment(id, attachmentId);
      if (result.error) return result;
      return saveBinary(opts.output || attachmentId, result.data);
    },
  );

  act(
    leaf(receiving, 'raw <id>', 'Download the original RFC822/MIME message to a file').option(
      '--output <path>',
      'file to write (default: <id>.eml)',
    ),
    async ({ client, opts, args: [id] }) => {
      const result = await client.emails.receiving.getRaw(id);
      if (result.error) return result;
      return saveBinary(opts.output || `${id}.eml`, result.data);
    },
  );

  act(
    leaf(receiving, 'forward <id>', 'Forward a received email')
      .requiredOption('--from <from>', 'a verified sending address to forward from')
      .requiredOption('--to <address>', 'recipient (repeatable or comma-separated)', collect)
      .option('--subject <subject>', 'override the subject line'),
    ({ client, opts, args: [id] }) =>
      client.emails.receiving.forward(
        id,
        clean({ from: opts.from, to: opts.to, subject: opts.subject }),
      ),
  );

  act(
    leaf(receiving, 'reply <id>', "Reply to a received email's sender (threaded)")
      .requiredOption('--from <from>', 'a verified sending address to reply from')
      .option('--html <html>', 'HTML body')
      .option('--text <text>', 'plain-text body')
      .option('--subject <subject>', "override the subject (default: 'Re: …')"),
    ({ client, opts, args: [id] }) => {
      if (opts.html === undefined && opts.text === undefined) {
        throw new CliError('Provide at least one of --html or --text.');
      }
      return client.emails.receiving.reply(
        id,
        clean({ from: opts.from, html: opts.html, text: opts.text, subject: opts.subject }),
      );
    },
  );

  act(leaf(receiving, 'delete <id>', 'Delete a received email'), ({ client, args: [id] }) =>
    client.emails.receiving.remove(id),
  );
}

module.exports = { register };
