'use strict';

const fs = require('node:fs');
const { CliError, parseJson, clean, withPagination, pagination } = require('../helpers');

function register({ group, leaf, act }) {
  const contacts = group('contacts', 'Manage contacts (domain-first: each sending domain has its own pool)');

  act(
    leaf(contacts, 'create', "Create a contact in a domain's pool")
      .requiredOption('--domain <domain>', 'the sending domain whose contact pool to use')
      .requiredOption('--email <email>', "the contact's email address")
      .option('--first-name <name>', 'first name')
      .option('--last-name <name>', 'last name')
      .option('--unsubscribed', 'create as unsubscribed')
      .option('--properties <json>', 'custom property values as a JSON object'),
    ({ client, opts }) =>
      client.contacts.create(
        clean({
          domain: opts.domain,
          email: opts.email,
          first_name: opts.firstName,
          last_name: opts.lastName,
          unsubscribed: opts.unsubscribed,
          properties: parseJson(opts.properties, '--properties'),
        }),
      ),
  );

  act(
    withPagination(
      leaf(contacts, 'list', "List a domain's contacts")
        .requiredOption('--domain <domain>', 'the sending domain whose contact pool to list')
        .option('--segment-id <id>', 'only contacts in this segment'),
    ),
    ({ client, opts }) =>
      client.contacts.list(clean({ domain: opts.domain, segment_id: opts.segmentId, ...pagination(opts) })),
  );

  act(
    leaf(contacts, 'get <id>', 'Retrieve a contact by id or email').option(
      '--domain <domain>',
      'disambiguates an email id across domain pools',
    ),
    ({ client, opts, args: [id] }) => client.contacts.get(clean({ id, domain: opts.domain })),
  );

  act(
    leaf(contacts, 'update <id>', 'Update a contact (id or email)')
      .option('--domain <domain>', 'disambiguates an email id across domain pools')
      .option('--first-name <name>', 'first name')
      .option('--last-name <name>', 'last name')
      .option('--unsubscribed', 'mark unsubscribed')
      .option('--subscribed', 'mark subscribed again')
      .option('--properties <json>', 'custom property values as a JSON object'),
    ({ client, opts, args: [id] }) =>
      client.contacts.update(
        clean({
          id,
          domain: opts.domain,
          first_name: opts.firstName,
          last_name: opts.lastName,
          unsubscribed: opts.unsubscribed ? true : opts.subscribed ? false : undefined,
          properties: parseJson(opts.properties, '--properties'),
        }),
      ),
  );

  act(
    leaf(contacts, 'delete <id>', 'Delete a contact (id or email)').option(
      '--domain <domain>',
      'disambiguates an email id across domain pools',
    ),
    ({ client, opts, args: [id] }) => client.contacts.remove(clean({ id, domain: opts.domain })),
  );

  act(
    leaf(contacts, 'add-to-segment <id> <segmentId>', 'Add a contact to a segment'),
    ({ client, args: [id, segmentId] }) => client.contacts.addToSegment(id, segmentId),
  );

  act(
    leaf(contacts, 'segments <id>', 'List the segments a contact belongs to'),
    ({ client, args: [id] }) => client.contacts.listSegments(id),
  );

  act(
    leaf(contacts, 'remove-from-segment <id> <segmentId>', 'Remove a contact from a segment'),
    ({ client, args: [id, segmentId] }) => client.contacts.removeFromSegment(id, segmentId),
  );

  act(
    leaf(contacts, 'topics <id>', "Get a contact's topic subscriptions"),
    ({ client, args: [id] }) => client.contacts.getTopics(id),
  );

  act(
    leaf(contacts, 'set-topics <id>', "Update a contact's topic subscriptions").requiredOption(
      '--topics <json>',
      'JSON array of {"id","subscription"} objects, e.g. \'[{"id":"top_1","subscription":"opt_in"}]\'',
    ),
    ({ client, opts, args: [id] }) =>
      client.contacts.updateTopics(id, { topics: parseJson(opts.topics, '--topics') }),
  );

  // Bulk import — from a JSON array of contacts, or a CSV file (SDK-3).
  act(
    leaf(contacts, 'batch <audienceId>', 'Bulk-import contacts from a JSON array (upsert by email; max 10,000)')
      .option('--file <path>', 'path to a JSON file containing an array of contact objects')
      .option('--data <json>', 'inline JSON array of contact objects')
      .option('--on-conflict <mode>', "how to resolve an existing email: 'upsert' (default) or 'skip'"),
    ({ client, opts, args: [audienceId] }) => {
      if (opts.file && opts.data) throw new CliError('Provide only one of --file or --data.');
      if (!opts.file && !opts.data) throw new CliError('Provide --file <path> or --data <json array>.');
      let contactsArr;
      if (opts.file) {
        let raw;
        try { raw = fs.readFileSync(opts.file, 'utf8'); }
        catch (err) { throw new CliError(`Cannot read --file ${opts.file}: ${err.message}`); }
        contactsArr = parseJson(raw, `--file ${opts.file}`);
      } else {
        contactsArr = parseJson(opts.data, '--data');
      }
      if (!Array.isArray(contactsArr)) throw new CliError('The contacts payload must be a JSON array.');
      return client.contacts.batch(clean({ audienceId, contacts: contactsArr, on_conflict: opts.onConflict }));
    },
  );

  act(
    leaf(contacts, 'import <audienceId>', 'Bulk-import contacts from a CSV file (email column required; header row optional)')
      .requiredOption('--csv <path>', 'path to a CSV file')
      .option('--on-conflict <mode>', "how to resolve an existing email: 'upsert' (default) or 'skip'")
      .option('--no-create-properties', 'strict mode: keep only columns matching a registered property'),
    ({ client, opts, args: [audienceId] }) => {
      let csv;
      try { csv = fs.readFileSync(opts.csv, 'utf8'); }
      catch (err) { throw new CliError(`Cannot read --csv ${opts.csv}: ${err.message}`); }
      // commander sets opts.createProperties=false for --no-create-properties (default true).
      return client.contacts.import(clean({
        audienceId, csv, on_conflict: opts.onConflict,
        create_properties: opts.createProperties === false ? false : undefined,
      }));
    },
  );
}

module.exports = { register };
