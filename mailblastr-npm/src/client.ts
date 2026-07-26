import type { MailBlastrError, Result, RequestOptions } from './types';

export const DEFAULT_BASE_URL = 'https://www.mailblastr.com/api';

/** Keep in sync with package.json "version". */
export const VERSION = '1.3.0';
export const USER_AGENT = `mailblastr-node/${VERSION}`;

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

export class HttpClient {
  private readonly apiKey: string;
  private readonly baseUrl: string;
  private readonly fetchImpl: typeof fetch;
  private readonly timeoutMs: number;
  private readonly maxRetries: number;

  constructor(apiKey: string, config: ClientConfig = {}) {
    if (!apiKey || typeof apiKey !== 'string') {
      throw new Error('MailBlastr: an API key is required, e.g. new MailBlastr("mb_...").');
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

    let parsed: unknown = null;
    const textBody = await res.text();
    if (textBody) {
      try { parsed = JSON.parse(textBody); } catch { parsed = textBody; }
    }

    if (!res.ok) {
      const e = (parsed && typeof parsed === 'object' ? parsed : {}) as Partial<MailBlastrError>;
      return {
        data: null,
        error: {
          statusCode: e.statusCode ?? res.status,
          name: e.name ?? 'application_error',
          message: e.message ?? `Request failed with status ${res.status}`,
        },
      };
    }
    return { data: parsed as T, error: null };
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

    if (!res.ok) {
      let parsed: unknown = null;
      const textBody = await res.text();
      if (textBody) {
        try { parsed = JSON.parse(textBody); } catch { parsed = textBody; }
      }
      const e = (parsed && typeof parsed === 'object' ? parsed : {}) as Partial<MailBlastrError>;
      return {
        data: null,
        error: {
          statusCode: e.statusCode ?? res.status,
          name: e.name ?? 'application_error',
          message: e.message ?? `Request failed with status ${res.status}`,
        },
      };
    }

    const buf = await res.arrayBuffer();
    return { data: buf, error: null };
  }
}
