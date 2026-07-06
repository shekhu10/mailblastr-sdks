'use strict';

/** Error whose message is reported as a CLI usage error (exit 1, no stack). */
class CliError extends Error {}

/**
 * Commander collector for repeatable flags. Each occurrence may itself be
 * comma-separated: `--to a@b.com --to c@d.com,e@f.com` → [a@b.com, c@d.com, e@f.com].
 */
function collect(value, previous) {
  return (previous || []).concat(
    String(value)
      .split(',')
      .map((s) => s.trim())
      .filter(Boolean),
  );
}

/** Parse a JSON-valued flag, failing with a friendly message. */
function parseJson(value, flagName) {
  if (value === undefined) return undefined;
  try {
    return JSON.parse(value);
  } catch {
    throw new CliError(`Invalid JSON for ${flagName}: ${value}`);
  }
}

/** Parse an integer-valued flag. */
function toInt(value, flagName) {
  if (value === undefined) return undefined;
  const n = Number(value);
  if (!Number.isInteger(n)) throw new CliError(`${flagName} must be an integer, got: ${value}`);
  return n;
}

/** Drop undefined values so we only send what the user set. */
function clean(obj) {
  const out = {};
  for (const [k, v] of Object.entries(obj)) {
    if (v !== undefined) out[k] = v;
  }
  return out;
}

/** Add the standard cursor-pagination flags to a list command. */
function withPagination(cmd) {
  return cmd
    .option('--limit <n>', 'max results to return')
    .option('--after <cursor>', 'cursor: results after this id')
    .option('--before <cursor>', 'cursor: results before this id');
}

/** Build the SDK pagination params object from parsed options. */
function pagination(opts) {
  return clean({
    limit: toInt(opts.limit, '--limit'),
    after: opts.after,
    before: opts.before,
  });
}

module.exports = { CliError, collect, parseJson, toInt, clean, withPagination, pagination };
