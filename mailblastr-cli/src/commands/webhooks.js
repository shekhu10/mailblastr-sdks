'use strict';

const fs = require('node:fs');
const { CliError, collect, clean, withPagination, pagination } = require('../helpers');

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

  // Verify a delivery's signature LOCALLY — pure computation, no HTTP request (SDK-4).
  act(
    leaf(webhooks, 'verify', 'Verify a webhook delivery signature locally (no HTTP request)')
      .requiredOption('--secret <secret>', 'the endpoint signing secret')
      .option('--payload <raw>', 'the raw request body string (exact bytes the server sent)')
      .option('--payload-file <path>', 'path to a file containing the raw request body')
      .requiredOption('--svix-id <id>', 'the svix-id header')
      .requiredOption('--svix-timestamp <ts>', 'the svix-timestamp header')
      .requiredOption('--svix-signature <sig>', 'the svix-signature header')
      .option('--tolerance <seconds>', 'max timestamp skew in seconds (default 300; 0 = skip the freshness check)'),
    ({ client, opts }) => {
      if (opts.payload && opts.payloadFile) throw new CliError('Provide only one of --payload or --payload-file.');
      let payload;
      if (opts.payloadFile) {
        try { payload = fs.readFileSync(opts.payloadFile, 'utf8'); }
        catch (err) { throw new CliError(`Cannot read --payload-file ${opts.payloadFile}: ${err.message}`); }
      } else if (opts.payload != null) {
        payload = opts.payload;
      } else {
        throw new CliError('Provide --payload <raw> or --payload-file <path>.');
      }
      const headers = {
        'svix-id': opts.svixId,
        'svix-timestamp': opts.svixTimestamp,
        'svix-signature': opts.svixSignature,
      };
      const options = {};
      if (opts.tolerance != null) options.toleranceSec = Number(opts.tolerance);
      // verify() is a pure, synchronous local computation returning { valid, reason }.
      // Wrap it in the {data,error} shape the CLI runner prints.
      return { data: client.webhooks.verify(payload, headers, opts.secret, options), error: null };
    },
  );
}

module.exports = { register };
