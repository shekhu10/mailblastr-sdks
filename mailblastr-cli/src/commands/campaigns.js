'use strict';

const { collect, clean, withPagination, pagination } = require('../helpers');

function register({ group, leaf, act }) {
  const campaigns = group('campaigns', "Manage campaigns (bulk sends to a domain's contact pool)");

  act(
    leaf(campaigns, 'create', 'Create a draft campaign')
      .requiredOption('--domain <domain>', 'the sending domain whose contact pool this campaign targets')
      .requiredOption('--from <from>', "sender, e.g. 'Acme <hi@yourdomain.com>'")
      .requiredOption('--subject <subject>', 'subject line')
      .option('--html <html>', 'HTML body')
      .option('--text <text>', 'plain-text body')
      .option('--name <name>', 'internal campaign name')
      .option('--segment-id <id>', 'target a segment instead of the whole pool')
      .option('--topic-id <id>', 'gate recipients by a topic subscription')
      .option('--reply-to <address>', 'reply-to address (repeatable or comma-separated)', collect)
      .option('--preview-text <text>', 'inbox preview text'),
    ({ client, opts }) =>
      client.campaigns.create(
        clean({
          domain: opts.domain,
          from: opts.from,
          subject: opts.subject,
          html: opts.html,
          text: opts.text,
          name: opts.name,
          segment_id: opts.segmentId,
          topic_id: opts.topicId,
          reply_to: opts.replyTo,
          preview_text: opts.previewText,
        }),
      ),
  );

  act(
    leaf(campaigns, 'send <id>', 'Send a campaign now, or schedule it').option(
      '--scheduled-at <iso>',
      'ISO 8601 timestamp to schedule the send',
    ),
    ({ client, opts, args: [id] }) =>
      client.campaigns.send(id, opts.scheduledAt ? { scheduled_at: opts.scheduledAt } : undefined),
  );

  act(leaf(campaigns, 'get <id>', 'Retrieve a campaign'), ({ client, args: [id] }) =>
    client.campaigns.get(id),
  );

  act(withPagination(leaf(campaigns, 'list', 'List campaigns')), ({ client, opts }) =>
    client.campaigns.list(pagination(opts)),
  );

  act(
    leaf(campaigns, 'update <id>', 'Update a draft campaign')
      .option('--from <from>', 'sender')
      .option('--subject <subject>', 'subject line')
      .option('--html <html>', 'HTML body')
      .option('--text <text>', 'plain-text body')
      .option('--name <name>', 'internal campaign name')
      .option('--segment-id <id>', 're-target a segment')
      .option('--topic-id <id>', 're-target a topic gate')
      .option('--preview-text <text>', 'inbox preview text'),
    ({ client, opts, args: [id] }) =>
      client.campaigns.update(
        id,
        clean({
          from: opts.from,
          subject: opts.subject,
          html: opts.html,
          text: opts.text,
          name: opts.name,
          segment_id: opts.segmentId,
          topic_id: opts.topicId,
          preview_text: opts.previewText,
        }),
      ),
  );

  act(
    leaf(campaigns, 'cancel <id>', 'Cancel a scheduled campaign (returns it to draft)'),
    ({ client, args: [id] }) => client.campaigns.cancel(id),
  );

  act(leaf(campaigns, 'stats <id>', 'Per-campaign analytics'), ({ client, args: [id] }) =>
    client.campaigns.stats(id),
  );

  act(
    leaf(campaigns, 'ab <id>', 'A/B winner evaluation for an A/B campaign'),
    ({ client, args: [id] }) => client.campaigns.ab(id),
  );

  act(leaf(campaigns, 'delete <id>', 'Delete a campaign'), ({ client, args: [id] }) =>
    client.campaigns.remove(id),
  );
}

module.exports = { register };
