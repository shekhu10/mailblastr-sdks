import type { MailBlastrError, Result, RequestOptions } from './types';

export const DEFAULT_BASE_URL = 'https://api.mailblastr.com';

/** Keep in sync with package.json "version". */
export const VERSION = '1.0.0';
export const USER_AGENT = `mailblastr-node/${VERSION}`;

export interface ClientConfig {
  baseUrl?: string;
  /** Override the fetch implementation (e.g. for tests or older runtimes). */
  fetch?: typeof fetch;
}

export class HttpClient {
  private readonly apiKey: string;
  private readonly baseUrl: string;
  private readonly fetchImpl: typeof fetch;

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
      res = await this.fetchImpl(`${this.baseUrl}${path}`, {
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
      res = await this.fetchImpl(`${this.baseUrl}${path}`, { method, headers });
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
