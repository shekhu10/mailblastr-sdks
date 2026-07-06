'use strict';

const { Command } = require('commander');
const { CliError } = require('./helpers');

const VERSION = require('../package.json').version;

const RESOURCES = [
  require('./commands/emails'),
  require('./commands/domains'),
  require('./commands/contacts'),
  require('./commands/contact_properties'),
  require('./commands/audiences'),
  require('./commands/segments'),
  require('./commands/topics'),
  require('./commands/campaigns'),
  require('./commands/templates'),
  require('./commands/automations'),
  require('./commands/webhooks'),
  require('./commands/apikeys'),
  require('./commands/events'),
  require('./commands/logs'),
  require('./commands/polls'),
];

/**
 * Build the execution context. Everything side-effectful (client construction,
 * writing, environment) is injectable so tests can run the command layer with
 * a mocked SDK client.
 */
function createContext(overrides = {}) {
  return {
    createClient(apiKey, options) {
      const { MailBlastr } = require('mailblastr');
      return new MailBlastr(apiKey, options);
    },
    stdout: (s) => process.stdout.write(s),
    stderr: (s) => process.stderr.write(s),
    env: process.env,
    exitCode: 0,
    ...overrides,
  };
}

/** Resolve auth/base-url flags (falling back to env) into an SDK client. */
function getClient(ctx, opts) {
  const apiKey = opts.apiKey || ctx.env.MAILBLASTR_API_KEY;
  if (!apiKey) {
    throw new CliError('Missing API key. Set the MAILBLASTR_API_KEY environment variable or pass --api-key.');
  }
  const baseUrl = opts.baseUrl || ctx.env.MAILBLASTR_BASE_URL;
  return ctx.createClient(apiKey, baseUrl ? { baseUrl } : {});
}

/** Print a success payload: pretty JSON by default, raw compact with --json. */
function print(ctx, data, opts) {
  const s = opts.json ? JSON.stringify(data) : JSON.stringify(data, null, 2);
  ctx.stdout(`${s}\n`);
}

/** Print an API/CLI error to stderr and mark the run failed. */
function fail(ctx, error) {
  ctx.stderr(`${JSON.stringify(error, null, 2)}\n`);
  ctx.exitCode = 1;
}

/**
 * Build the toolkit passed to each resource module: `group` creates a
 * top-level resource command, `leaf` a subcommand carrying the shared
 * auth/output flags, and `act` wires a handler that receives the SDK client
 * and returns its `{ data, error }` result.
 */
function makeToolkit(program, ctx) {
  function group(name, description) {
    return program.command(name).description(description);
  }
  function leaf(parent, spec, description) {
    return parent
      .command(spec)
      .description(description)
      .option('--api-key <key>', 'API key (default: MAILBLASTR_API_KEY env var)')
      .option('--base-url <url>', 'API base URL (default: MAILBLASTR_BASE_URL env var)')
      .option('--json', 'raw compact JSON output (default: pretty-printed)');
  }
  function act(cmd, handler) {
    cmd.action(async (...cbArgs) => {
      const command = cbArgs[cbArgs.length - 1];
      const args = cbArgs.slice(0, -1);
      const opts = command.opts();
      try {
        const client = getClient(ctx, opts);
        const result = await handler({ client, opts, args });
        if (result && result.error) return fail(ctx, result.error);
        return print(ctx, result ? result.data : null, opts);
      } catch (err) {
        if (err instanceof CliError) {
          return fail(ctx, { statusCode: null, name: 'cli_error', message: err.message });
        }
        throw err;
      }
    });
    return cmd;
  }
  return { group, leaf, act };
}

/** Assemble the full commander program for the given context. */
function buildProgram(ctx) {
  const program = new Command();
  program
    .name('mailblastr')
    .description('Official MailBlastr CLI — send email, manage domains, contacts, campaigns and more.')
    .version(VERSION)
    .exitOverride()
    .configureOutput({
      writeOut: (s) => ctx.stdout(s),
      writeErr: (s) => ctx.stderr(s),
    })
    .addHelpText(
      'after',
      '\nAuthentication:\n  Set MAILBLASTR_API_KEY (or pass --api-key to any command).\n  Set MAILBLASTR_BASE_URL (or --base-url) to target a different API host.\n\nExamples:\n  mailblastr emails send --from \'Acme <hi@yourdomain.com>\' --to a@b.com --subject hello --html \'<p>hi</p>\'\n  mailblastr contacts list --domain yourdomain.com\n  mailblastr events send --domain yourdomain.com --name signup.completed --email a@b.com --data \'{"plan":"pro"}\'',
    );

  const toolkit = makeToolkit(program, ctx);
  for (const resource of RESOURCES) resource.register(toolkit);
  return program;
}

/**
 * Parse and execute argv (user args only, no node/script prefix).
 * Returns the exit code; never calls process.exit itself.
 */
async function run(ctx, argv) {
  const program = buildProgram(ctx);
  try {
    await program.parseAsync(argv, { from: 'user' });
  } catch (err) {
    // Commander throws for --help/--version (exitCode 0) and usage errors
    // (exitCode 1) because of exitOverride; the message is already printed.
    if (err && typeof err.exitCode === 'number' && String(err.code || '').startsWith('commander.')) {
      if (err.exitCode !== 0) ctx.exitCode = err.exitCode;
      return ctx.exitCode;
    }
    throw err;
  }
  return ctx.exitCode;
}

/** Entry point used by bin/mailblastr.js. */
function main(argv) {
  return run(createContext(), argv);
}

module.exports = { createContext, buildProgram, run, main };
