'use strict';

const { collect, clean, withPagination, pagination } = require('../helpers');

function register({ group, leaf, act }) {
  const webhooks = group('webhooks', 'Manage webhooks');

  act(
    leaf(webhooks, 'create', 'Create a webhook (the signing secret is shown once)')
      .requiredOption('--endpoint <url>', 'the HTTPS endpoint to deliver events to')
      .requiredOption(
        '--events <event>',
        "events to subscribe to (repeatable or comma-separated), e.g. 'email.delivered,email.bounced'",
        collect,
      )
      .option('--secret <secret>', 'caller-supplied signing secret (otherwise generated)'),
    ({ client, opts }) =>
      client.webhooks.create(clean({ endpoint: opts.endpoint, events: opts.events, secret: opts.secret })),
  );

  act(leaf(webhooks, 'get <id>', 'Retrieve a webhook'), ({ client, args: [id] }) =>
    client.webhooks.get(id),
  );

  act(withPagination(leaf(webhooks, 'list', 'List webhooks')), ({ client, opts }) =>
    client.webhooks.list(pagination(opts)),
  );

  act(
    leaf(webhooks, 'update <id>', 'Update a webhook')
      .option('--endpoint <url>', 'new endpoint URL')
      .option('--events <event>', 'new event list (repeatable or comma-separated)', collect)
      .option('--status <status>', "'enabled' | 'disabled'"),
    ({ client, opts, args: [id] }) =>
      client.webhooks.update(
        id,
        clean({ endpoint: opts.endpoint, events: opts.events, status: opts.status }),
      ),
  );

  act(
    leaf(webhooks, 'rotate <id>', 'Rotate the signing secret (new secret shown once)'),
    ({ client, args: [id] }) => client.webhooks.rotate(id),
  );

  act(
    leaf(webhooks, 'test <id>', 'Send a synchronous test delivery'),
    ({ client, args: [id] }) => client.webhooks.test(id),
  );

  act(leaf(webhooks, 'delete <id>', 'Delete a webhook'), ({ client, args: [id] }) =>
    client.webhooks.remove(id),
  );
}

module.exports = { register };
