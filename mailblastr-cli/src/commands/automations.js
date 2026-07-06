'use strict';

const { parseJson, clean, withPagination, pagination } = require('../helpers');

function register({ group, leaf, act }) {
  const automations = group('automations', 'Manage automations (scoped to a sending domain)');

  act(
    leaf(automations, 'create', 'Create an automation')
      .requiredOption('--name <name>', 'automation name')
      .requiredOption('--domain <domain>', 'the sending domain this automation belongs to')
      .option('--trigger <event>', "trigger event, e.g. 'contact.created' or a custom event name")
      .option('--status <status>', "'enabled' | 'disabled' (default: disabled)"),
    ({ client, opts }) =>
      client.automations.create(
        clean({ name: opts.name, domain: opts.domain, trigger: opts.trigger, status: opts.status }),
      ),
  );

  act(leaf(automations, 'get <id>', 'Retrieve an automation'), ({ client, args: [id] }) =>
    client.automations.get(id),
  );

  act(withPagination(leaf(automations, 'list', 'List automations')), ({ client, opts }) =>
    client.automations.list(pagination(opts)),
  );

  act(
    leaf(automations, 'update <id>', 'Update an automation')
      .option('--name <name>', 'new name')
      .option('--status <status>', "'enabled' | 'disabled'"),
    ({ client, opts, args: [id] }) =>
      client.automations.update(id, clean({ name: opts.name, status: opts.status })),
  );

  act(
    leaf(automations, 'add-step <id>', 'Append a step to an automation')
      .requiredOption('--type <type>', "step type, e.g. 'send_email' or 'wait'")
      .option('--config <json>', "step config as JSON, e.g. '{\"template_id\":\"tmpl_x\"}'")
      .option('--key <key>', 'step key (for connections)'),
    ({ client, opts, args: [id] }) =>
      client.automations.addStep(
        id,
        clean({ type: opts.type, config: parseJson(opts.config, '--config'), key: opts.key }),
      ),
  );

  act(
    leaf(automations, 'delete-step <id> <stepId>', 'Delete a step from an automation'),
    ({ client, args: [id, stepId] }) => client.automations.deleteStep(id, stepId),
  );

  act(
    withPagination(leaf(automations, 'runs <id>', "List an automation's runs")),
    ({ client, opts, args: [id] }) => client.automations.runs(id, pagination(opts)),
  );

  act(
    leaf(automations, 'run <id> <runId>', 'Retrieve a single automation run'),
    ({ client, args: [id, runId] }) => client.automations.getRun(id, runId),
  );

  act(
    leaf(automations, 'stop <id>', 'Stop an automation (in-progress runs finish)'),
    ({ client, args: [id] }) => client.automations.stop(id),
  );

  act(leaf(automations, 'delete <id>', 'Delete an automation'), ({ client, args: [id] }) =>
    client.automations.remove(id),
  );
}

module.exports = { register };
