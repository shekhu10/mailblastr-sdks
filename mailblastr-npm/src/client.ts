import type { MailblastrError, Result, RequestOptions } from './types';

export const DEFAULT_BASE_URL = 'https://www.mailblastr.com/api';

/** Keep in sync with package.json "version". */
export const VERSION = '4.0.0';
export const USER_AGENT = `mailblastr-node/${VERSION}`;

/**
 * Max length of an `Idempotency-Key` accepted by the API (1–255 characters
 * after trimming — the storage column width). A longer or empty key is
 * rejected with `400 invalid_idempotency_key`.
 */
export const IDEMPOTENCY_KEY_MAX_LENGTH = 255;

export interface ClientConfig {
  baseUrl?: string;
  /** Override the fetch implementation (e.g. for tests or older runtimes). */
  fetch?: typeof fetch;
  /** Per-request timeout in milliseconds. Default 30000 (30s). 0 disables it. */
  timeoutMs?: number;
  /**
   * Max automatic retries on a rate-limit (429) or service-unavailable (503)
   * response — the only two the server guarantees were NOT applied, so retrying
   * can't duplicate a side-effect (e.g. a double send). Honors `Retry-After`,
   * else exponential backoff. Default 2 (→ up to 3 attempts). 0 disables retries.
   */
  maxRetries?: number;
}

const DEFAULT_TIMEOUT_MS = 30_000;
const DEFAULT_MAX_RETRIES = 2;
const RETRYABLE_STATUS = new Set([429, 503]);

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

/** Parse a Retry-After header (delta-seconds or HTTP-date) → ms, capped at 30s. */
function retryAfterMs(header: string | null): number | null {
  if (!header) return null;
  const secs = Number(header);
  if (Number.isFinite(secs)) return Math.min(Math.max(0, secs * 1000), 30_000);
  const when = Date.parse(header);
  if (!Number.isNaN(when)) return Math.min(Math.max(0, when - Date.now()), 30_000);
  return null;
}

/** Decode a response body as JSON, falling back to the raw text. */
function parseBody(text: string): unknown {
  if (!text) return null;
  try { return JSON.parse(text); } catch { return text; }
}

/**
 * Normalize an error response into {@link MailblastrError}.
 *
 * The API's envelope is `{ statusCode, name, message }`, but some errors carry
 * additive fields — `limit` on plan/quota rejections, `reputation` on
 * reputation gates, `sent`/`sent_count` on a partial batch failure — which are
 * preserved verbatim. Two non-envelope shapes also exist (`{"error":
 * "csrf_failed"}` and `{"error":"rate_limited"}`); their `error` string is
 * lifted into `name` so callers can still branch on it.
 *
 * `sent_count` falls back to `sent.length` when the body names the emails that
 * went out but omits the count, so a caller deciding what NOT to resend never
 * has to compute it. It stays absent on errors that carry no `sent` list.
 */
function toError(parsed: unknown, status: number): MailblastrError {
  const body = (parsed && typeof parsed === 'object' ? parsed : {}) as Record<string, unknown>;
  const { statusCode, name, message, error, ...rest } = body;
  const sent = rest.sent;
  if (Array.isArray(sent) && typeof rest.sent_count !== 'number') {
    rest.sent_count = sent.length;
  }
  return {
    ...rest,
    // Trust the transport status; `statusCode` in the body always mirrors it.
    statusCode: typeof statusCode === 'number' ? statusCode : status,
    name: typeof name === 'string' ? name : typeof error === 'string' ? error : 'application_error',
    message: typeof message === 'string' ? message : `Request failed with status ${status}`,
  };
}

export class HttpClient {
  private readonly apiKey: string;
  private readonly baseUrl: string;
  private readonly fetchImpl: typeof fetch;
  private readonly timeoutMs: number;
  private readonly maxRetries: number;

  constructor(apiKey: string, config: ClientConfig = {}) {
    if (!apiKey || typeof apiKey !== 'string') {
      throw new Error('MailBlastr: an API key is required, e.g. new Mailblastr("mb_...").');
    }
    this.apiKey = apiKey;
    this.baseUrl = (config.baseUrl ?? DEFAULT_BASE_URL).replace(/\/$/, '');
    const f = config.fetch ?? (globalThis.fetch as typeof fetch | undefined);
    if (!f) {
      throw new Error('MailBlastr: no global fetch available. Use Node 18+ or pass { fetch } in the options.');
    }
    this.fetchImpl = f;
    this.timeoutMs = config.timeoutMs ?? DEFAULT_TIMEOUT_MS;
    this.maxRetries = Math.max(0, config.maxRetries ?? DEFAULT_MAX_RETRIES);
  }

  /** Issue the fetch with a timeout, retrying only 429/503 (Retry-After aware). */
  private async send(url: string, init: RequestInit): Promise<Response> {
    for (let attempt = 0; ; attempt++) {
      const signal = this.timeoutMs > 0 ? AbortSignal.timeout(this.timeoutMs) : undefined;
      const res = await this.fetchImpl(url, { ...init, signal });
      if (!RETRYABLE_STATUS.has(res.status) || attempt >= this.maxRetries) return res;
      const wait = retryAfterMs(res.headers.get('retry-after')) ?? Math.min(30_000, 500 * 2 ** attempt);
      await sleep(wait);
    }
  }

  async request<T>(method: string, path: string, body?: unknown, options: RequestOptions = {}): Promise<Result<T>> {
    const headers: Record<string, string> = {
      Authorization: `Bearer ${this.apiKey}`,
      'Content-Type': 'application/json',
      'User-Agent': USER_AGENT,
    };
    if (options.idempotencyKey) headers['Idempotency-Key'] = options.idempotencyKey;

    let res: Response;
    try {
      res = await this.send(`${this.baseUrl}${path}`, {
        method,
        headers,
        body: body === undefined ? undefined : JSON.stringify(body),
      });
    } catch (err) {
      return { data: null, error: { statusCode: 0, name: 'network_error', message: (err as Error).message } };
    }

    const parsed = parseBody(await res.text());
    if (!res.ok) return { data: null, error: toError(parsed, res.status) };
    return { data: parsed as T, error: null };
  }

  /**
   * Like `request`, but for endpoints that return plain text rather than JSON
   * (e.g. a domain's DNS records as CSV). Errors are still the JSON envelope.
   */
  async requestText(method: string, path: string): Promise<Result<string>> {
    const headers: Record<string, string> = {
      Authorization: `Bearer ${this.apiKey}`,
      'User-Agent': USER_AGENT,
    };

    let res: Response;
    try {
      res = await this.send(`${this.baseUrl}${path}`, { method, headers });
    } catch (err) {
      return { data: null, error: { statusCode: 0, name: 'network_error', message: (err as Error).message } };
    }

    const text = await res.text();
    if (!res.ok) return { data: null, error: toError(parseBody(text), res.status) };
    return { data: text, error: null };
  }

  /**
   * Like `request`, but for endpoints that stream raw binary bytes (e.g. a
   * received-email attachment download). On success returns the response body
   * as an ArrayBuffer; on error parses the JSON error body like `request`.
   */
  async requestRaw(method: string, path: string, options: RequestOptions = {}): Promise<Result<ArrayBuffer>> {
    const headers: Record<string, string> = {
      Authorization: `Bearer ${this.apiKey}`,
      'User-Agent': USER_AGENT,
    };
    if (options.idempotencyKey) headers['Idempotency-Key'] = options.idempotencyKey;

    let res: Response;
    try {
      res = await this.send(`${this.baseUrl}${path}`, { method, headers });
    } catch (err) {
      return { data: null, error: { statusCode: 0, name: 'network_error', message: (err as Error).message } };
    }

    if (!res.ok) return { data: null, error: toError(parseBody(await res.text()), res.status) };

    const buf = await res.arrayBuffer();
    return { data: buf, error: null };
  }
}
