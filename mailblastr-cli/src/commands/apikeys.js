'use strict';

const { clean } = require('../helpers');

function register({ group, leaf, act }) {
  const apiKeys = group('api-keys', 'Manage API keys');

  act(
    leaf(apiKeys, 'create', 'Create an API key (the token is shown once)')
      .requiredOption('--name <name>', 'key name')
      .option('--permission <permission>', "'full_access' | 'sending_access'")
      .option('--domain-id <id>', 'scope a sending_access key to one domain'),
    ({ client, opts }) =>
      client.apiKeys.create(
        clean({ name: opts.name, permission: opts.permission, domain_id: opts.domainId }),
      ),
  );

  act(leaf(apiKeys, 'list', 'List API keys'), ({ client }) => client.apiKeys.list());

  act(leaf(apiKeys, 'delete <id>', 'Delete (revoke) an API key'), ({ client, args: [id] }) =>
    client.apiKeys.remove(id),
  );
}

module.exports = { register };
