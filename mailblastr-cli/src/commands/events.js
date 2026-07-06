'use strict';

const { parseJson, clean, withPagination, pagination } = require('../helpers');

function register({ group, leaf, act }) {
  const events = group('events', 'Send custom events and manage event definitions');

  act(
    leaf(events, 'send', "Send a custom event to trigger a domain's automations")
      .requiredOption('--domain <domain>', 'the sending domain this event belongs to')
      .requiredOption('--name <name>', "event name, e.g. 'signup.completed'")
      .option('--email <email>', 'identify the contact by email')
      .option('--contact-id <id>', 'identify the contact by id')
      .option('--data <json>', "event payload as JSON, e.g. '{\"plan\":\"pro\"}'"),
    ({ client, opts }) =>
      client.events.send(
        clean({
          name: opts.name,
          domain: opts.domain,
          email: opts.email,
          contact_id: opts.contactId,
          data: parseJson(opts.data, '--data'),
        }),
      ),
  );

  act(
    leaf(events, 'create', 'Create a custom-event definition')
      .requiredOption('--name <name>', 'event name')
      .option('--schema <json>', "flat key→type schema as JSON, e.g. '{\"plan\":\"string\"}'"),
    ({ client, opts }) =>
      client.events.create(clean({ name: opts.name, schema: parseJson(opts.schema, '--schema') })),
  );

  act(withPagination(leaf(events, 'list', 'List custom-event definitions')), ({ client, opts }) =>
    client.events.list(pagination(opts)),
  );

  act(leaf(events, 'delete <id>', 'Delete a custom-event definition'), ({ client, args: [id] }) =>
    client.events.remove(id),
  );
}

module.exports = { register };
