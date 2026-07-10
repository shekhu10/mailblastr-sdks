//! HTTP core: the [`Mailblastr`] client, request building, and API-error
//! mapping shared by every service.

use std::fmt;
use std::sync::Arc;
use std::time::{Duration, SystemTime};

use reqwest::header::HeaderMap;
use reqwest::{Method, RequestBuilder};
use serde::de::DeserializeOwned;

use crate::services::*;
use crate::types::{Error, PaginationParams, Result};

/// The production MailBlastr API host.
pub const DEFAULT_BASE_URL: &str = "https://api.mailblastr.com";

/// Default per-request timeout, in seconds. Overridable via
/// [`MailblastrBuilder::timeout`]; `0` / [`MailblastrBuilder::no_timeout`]
/// disables it.
pub const DEFAULT_TIMEOUT_SECS: u64 = 30;

/// Default number of automatic retries (up to `1 + max_retries` total
/// attempts). Overridable via [`MailblastrBuilder::max_retries`]; `0` disables
/// retrying.
pub const DEFAULT_MAX_RETRIES: u32 = 2;

/// Upper bound (seconds) on any single backoff wait, whether it comes from a
/// `Retry-After` header or exponential backoff.
const MAX_BACKOFF_SECS: f64 = 30.0;

/// This crate's version (kept in sync with `Cargo.toml`).
pub const VERSION: &str = env!("CARGO_PKG_VERSION");

/// `User-Agent` header sent with every request.
pub const USER_AGENT: &str = concat!("mailblastr-rust/", env!("CARGO_PKG_VERSION"));

/// Shared request configuration handed to every service (cheap `Arc` clone).
pub(crate) struct Config {
    api_key: String,
    base_url: String,
    client: reqwest::Client,
    /// Max automatic retries on 429/503 (see [`Config::execute`]).
    max_retries: u32,
}

impl fmt::Debug for Config {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        f.debug_struct("Config")
            .field("api_key", &"mb_***redacted***")
            .field("base_url", &self.base_url)
            .finish()
    }
}

impl Config {
    /// Start a request against `path` (leading slash) with auth + UA set.
    pub(crate) fn request(&self, method: Method, path: &str) -> RequestBuilder {
        self.client
            .request(method, format!("{}{}", self.base_url, path))
            .bearer_auth(&self.api_key)
            .header(reqwest::header::USER_AGENT, USER_AGENT)
    }

    /// The single request chokepoint: send `req`, and on HTTP `429`/`503`
    /// retry up to `max_retries` times, honoring `Retry-After` (else
    /// exponential backoff). Both the JSON ([`Config::send`]) and raw
    /// ([`Config::send_raw`]) paths funnel through here, so both inherit the
    /// configured timeout (carried by the [`reqwest::Client`]) and retry
    /// policy. Returns the final response status and fully-read body.
    ///
    /// Only `429` and `503` are retried — never other `5xx`, network errors,
    /// or timeouts — because those are the only statuses the server
    /// guarantees were not applied, so a retry cannot duplicate a
    /// non-idempotent side effect (e.g. sending an email twice).
    async fn execute(&self, req: RequestBuilder) -> Result<(reqwest::StatusCode, Vec<u8>)> {
        // Buffer the request so it can be rebuilt per attempt. `try_clone`
        // yields a fresh builder for the in-memory bodies this SDK sends
        // (`.json(..)`); a non-cloneable streaming body returns `None`, in
        // which case we make a single attempt and cannot retry.
        let mut next = Some(req);
        let mut attempt: u32 = 0;
        loop {
            let current = next
                .take()
                .expect("request builder is present at the start of each attempt");
            // Pre-clone for a possible retry before `send` consumes `current`.
            let spare = current.try_clone();

            let resp = current.send().await?;
            let status = resp.status();
            let code = status.as_u16();
            let retryable = code == 429 || code == 503;

            if !retryable || attempt >= self.max_retries || spare.is_none() {
                let bytes = resp.bytes().await?.to_vec();
                return Ok((status, bytes));
            }

            let wait =
                parse_retry_after(resp.headers()).unwrap_or_else(|| backoff_delay(attempt));
            tokio::time::sleep(wait).await;
            next = spare; // guaranteed `Some` here (checked above)
            attempt += 1;
        }
    }

    /// Execute a request and decode the JSON response; non-2xx statuses are
    /// mapped to [`Error::Api`] parsed from the standard error body.
    pub(crate) async fn send<T: DeserializeOwned>(&self, req: RequestBuilder) -> Result<T> {
        let (status, bytes) = self.execute(req).await?;
        if !status.is_success() {
            return Err(api_error(status.as_u16(), &bytes));
        }
        serde_json::from_slice(&bytes).map_err(Error::Json)
    }

    /// Execute a request whose success response streams raw bytes (received
    /// attachment / RFC822 downloads). Errors are still JSON and are mapped
    /// like [`Config::send`].
    pub(crate) async fn send_raw(&self, req: RequestBuilder) -> Result<Vec<u8>> {
        let (status, bytes) = self.execute(req).await?;
        if !status.is_success() {
            return Err(api_error(status.as_u16(), &bytes));
        }
        Ok(bytes)
    }
}

/// Parse a `Retry-After` header into a wait duration, capped at
/// [`MAX_BACKOFF_SECS`]. Accepts an integer/decimal number of seconds
/// (negative treated as `0`) or an HTTP-date; returns `None` when the header
/// is absent or unparseable.
fn parse_retry_after(headers: &HeaderMap) -> Option<Duration> {
    let raw = headers
        .get(reqwest::header::RETRY_AFTER)?
        .to_str()
        .ok()?
        .trim();

    // Numeric seconds (integer or decimal).
    if let Ok(secs) = raw.parse::<f64>() {
        if secs.is_finite() {
            return Some(Duration::from_secs_f64(secs.clamp(0.0, MAX_BACKOFF_SECS)));
        }
    }

    // HTTP-date: wait until that instant (never negative, capped).
    if let Ok(when) = httpdate::parse_http_date(raw) {
        let wait = when
            .duration_since(SystemTime::now())
            .unwrap_or(Duration::ZERO);
        return Some(wait.min(Duration::from_secs_f64(MAX_BACKOFF_SECS)));
    }

    None
}

/// Exponential backoff for retry `attempt` (0-based): `min(30s, 0.5s * 2^attempt)`.
fn backoff_delay(attempt: u32) -> Duration {
    let secs = (0.5 * 2f64.powi(attempt as i32)).min(MAX_BACKOFF_SECS);
    Duration::from_secs_f64(secs)
}

/// Parse the standard `{ statusCode, name, message }` error body, falling
/// back to the HTTP status when the body is absent or malformed.
fn api_error(status: u16, body: &[u8]) -> Error {
    #[derive(Default, serde::Deserialize)]
    struct ErrorBody {
        #[serde(rename = "statusCode")]
        status_code: Option<u16>,
        name: Option<String>,
        message: Option<String>,
    }
    let parsed: ErrorBody = serde_json::from_slice(body).unwrap_or_default();
    Error::Api {
        status_code: parsed.status_code.unwrap_or(status),
        name: parsed
            .name
            .unwrap_or_else(|| "application_error".to_owned()),
        message: parsed
            .message
            .unwrap_or_else(|| format!("Request failed with status {status}")),
    }
}

/// Percent-encode a single URL path segment (RFC 3986 unreserved characters
/// pass through) so an id like `user@example.com` — or a hostile one like
/// `../api-keys` — can never traverse the URL path and hit a different
/// endpoint.
pub(crate) fn seg(s: &str) -> String {
    let mut out = String::with_capacity(s.len());
    for b in s.bytes() {
        match b {
            b'A'..=b'Z' | b'a'..=b'z' | b'0'..=b'9' | b'-' | b'_' | b'.' | b'~' => {
                out.push(b as char);
            }
            _ => out.push_str(&format!("%{b:02X}")),
        }
    }
    out
}

/// Turn optional pagination params into query pairs (empty when `None`).
pub(crate) fn page_query(params: Option<&PaginationParams>) -> Vec<(&'static str, String)> {
    params.map(PaginationParams::to_query).unwrap_or_default()
}

/// The MailBlastr API client. One field per resource:
///
/// ```no_run
/// use mailblastr::Mailblastr;
///
/// let mailblastr = Mailblastr::new("mb_xxxxxxxxx");
/// // mailblastr.emails, mailblastr.domains, mailblastr.contacts, …
/// ```
#[derive(Clone, Debug)]
pub struct Mailblastr {
    pub emails: EmailsSvc,
    pub batch: BatchSvc,
    pub domains: DomainsSvc,
    pub audiences: AudiencesSvc,
    pub contacts: ContactsSvc,
    pub contact_properties: ContactPropertiesSvc,
    pub campaigns: CampaignsSvc,
    pub segments: SegmentsSvc,
    pub topics: TopicsSvc,
    pub templates: TemplatesSvc,
    pub automations: AutomationsSvc,
    pub webhooks: WebhooksSvc,
    pub logs: LogsSvc,
    pub events: EventsSvc,
    pub api_keys: ApiKeysSvc,
    pub polls: PollsSvc,
}

impl Mailblastr {
    /// Create a client against the production API (`https://api.mailblastr.com`)
    /// with the default 30s timeout and up to 2 retries.
    pub fn new(api_key: impl Into<String>) -> Self {
        Self::builder(api_key).build()
    }

    /// Point the client at a different host (a proxy, a staging environment,
    /// or a local test stub). A trailing `/` on `base_url` is trimmed. Keeps
    /// the default timeout and retry policy.
    pub fn with_base_url(api_key: impl Into<String>, base_url: impl Into<String>) -> Self {
        Self::builder(api_key).base_url(base_url).build()
    }

    /// Full control: bring your own [`reqwest::Client`] (timeouts, proxies,
    /// connection pools). The supplied client's own timeout applies; the
    /// builder's `timeout` is not imposed on a user-supplied client. The
    /// default retry policy still applies — use [`Mailblastr::builder`] to
    /// change it.
    pub fn with_client(
        api_key: impl Into<String>,
        base_url: impl Into<String>,
        client: reqwest::Client,
    ) -> Self {
        Self::builder(api_key)
            .base_url(base_url)
            .client(client)
            .build()
    }

    /// Start a [`MailblastrBuilder`] to configure the base URL, request
    /// timeout, retry count, or an entirely custom [`reqwest::Client`].
    pub fn builder(api_key: impl Into<String>) -> MailblastrBuilder {
        MailblastrBuilder::new(api_key)
    }
}

/// Builder for a [`Mailblastr`] client. Configures the base URL, the
/// per-request timeout (default 30s), the automatic-retry count (default 2),
/// and optionally a bring-your-own [`reqwest::Client`].
pub struct MailblastrBuilder {
    api_key: String,
    base_url: String,
    timeout: Option<Duration>,
    max_retries: u32,
    client: Option<reqwest::Client>,
}

impl fmt::Debug for MailblastrBuilder {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        f.debug_struct("MailblastrBuilder")
            .field("api_key", &"mb_***redacted***")
            .field("base_url", &self.base_url)
            .field("timeout", &self.timeout)
            .field("max_retries", &self.max_retries)
            .field("client", &self.client.is_some())
            .finish()
    }
}

impl MailblastrBuilder {
    fn new(api_key: impl Into<String>) -> Self {
        Self {
            api_key: api_key.into(),
            base_url: DEFAULT_BASE_URL.to_owned(),
            timeout: Some(Duration::from_secs(DEFAULT_TIMEOUT_SECS)),
            max_retries: DEFAULT_MAX_RETRIES,
            client: None,
        }
    }

    /// Override the API host. A trailing `/` is trimmed at build time.
    pub fn base_url(mut self, base_url: impl Into<String>) -> Self {
        self.base_url = base_url.into();
        self
    }

    /// Set the per-request timeout applied to every call (JSON and raw). A
    /// zero duration disables it (equivalent to [`Self::no_timeout`]). Ignored
    /// when a client is supplied via [`Self::client`].
    pub fn timeout(mut self, timeout: Duration) -> Self {
        self.timeout = if timeout.is_zero() { None } else { Some(timeout) };
        self
    }

    /// Disable the request timeout entirely.
    pub fn no_timeout(mut self) -> Self {
        self.timeout = None;
        self
    }

    /// Set the maximum number of automatic retries on HTTP `429`/`503`
    /// (`0` disables retrying; total attempts are `1 + max_retries`).
    pub fn max_retries(mut self, max_retries: u32) -> Self {
        self.max_retries = max_retries;
        self
    }

    /// Escape hatch: use a fully custom [`reqwest::Client`]. Its own timeout,
    /// proxy, and pool settings apply as-is; the builder's [`Self::timeout`]
    /// is not imposed on it.
    pub fn client(mut self, client: reqwest::Client) -> Self {
        self.client = Some(client);
        self
    }

    /// Build the [`Mailblastr`] client.
    pub fn build(self) -> Mailblastr {
        let client = self.client.unwrap_or_else(|| {
            let mut builder = reqwest::Client::builder();
            if let Some(timeout) = self.timeout {
                builder = builder.timeout(timeout);
            }
            builder.build().expect("failed to build reqwest client")
        });

        let config = Arc::new(Config {
            api_key: self.api_key,
            base_url: self.base_url.trim_end_matches('/').to_owned(),
            client,
            max_retries: self.max_retries,
        });
        Mailblastr {
            emails: EmailsSvc::new(Arc::clone(&config)),
            batch: BatchSvc::new(Arc::clone(&config)),
            domains: DomainsSvc::new(Arc::clone(&config)),
            audiences: AudiencesSvc::new(Arc::clone(&config)),
            contacts: ContactsSvc::new(Arc::clone(&config)),
            contact_properties: ContactPropertiesSvc::new(Arc::clone(&config)),
            campaigns: CampaignsSvc::new(Arc::clone(&config)),
            segments: SegmentsSvc::new(Arc::clone(&config)),
            topics: TopicsSvc::new(Arc::clone(&config)),
            templates: TemplatesSvc::new(Arc::clone(&config)),
            automations: AutomationsSvc::new(Arc::clone(&config)),
            webhooks: WebhooksSvc::new(Arc::clone(&config)),
            logs: LogsSvc::new(Arc::clone(&config)),
            events: EventsSvc::new(Arc::clone(&config)),
            api_keys: ApiKeysSvc::new(Arc::clone(&config)),
            polls: PollsSvc::new(config),
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn seg_passes_unreserved_and_encodes_the_rest() {
        assert_eq!(seg("em_abc123"), "em_abc123");
        assert_eq!(seg("user@example.com"), "user%40example.com");
        assert_eq!(seg("../api-keys"), "..%2Fapi-keys");
        assert_eq!(seg("a b"), "a%20b");
    }

    #[test]
    fn api_error_parses_standard_body_and_falls_back() {
        let e = api_error(
            422,
            br#"{"statusCode":422,"name":"validation_error","message":"domain is required"}"#,
        );
        match e {
            Error::Api {
                status_code,
                name,
                message,
            } => {
                assert_eq!(status_code, 422);
                assert_eq!(name, "validation_error");
                assert_eq!(message, "domain is required");
            }
            other => panic!("expected Api error, got {other:?}"),
        }

        let e = api_error(500, b"not json");
        match e {
            Error::Api {
                status_code, name, ..
            } => {
                assert_eq!(status_code, 500);
                assert_eq!(name, "application_error");
            }
            other => panic!("expected Api error, got {other:?}"),
        }
    }

    #[test]
    fn backoff_is_exponential_and_capped_at_30s() {
        assert_eq!(backoff_delay(0), Duration::from_millis(500));
        assert_eq!(backoff_delay(1), Duration::from_secs(1));
        assert_eq!(backoff_delay(2), Duration::from_secs(2));
        assert_eq!(backoff_delay(3), Duration::from_secs(4));
        // 0.5 * 2^20 far exceeds the cap → clamped to 30s.
        assert_eq!(backoff_delay(20), Duration::from_secs(30));
    }

    #[test]
    fn retry_after_parses_seconds_and_dates_with_cap() {
        use reqwest::header::{HeaderMap, HeaderValue, RETRY_AFTER};

        let parse = |v: &str| {
            let mut h = HeaderMap::new();
            h.insert(RETRY_AFTER, HeaderValue::from_str(v).unwrap());
            parse_retry_after(&h)
        };

        // Integer / decimal seconds.
        assert_eq!(parse("5"), Some(Duration::from_secs(5)));
        assert_eq!(parse("0.5"), Some(Duration::from_millis(500)));
        // Negative is treated as 0; oversized is capped at 30s.
        assert_eq!(parse("-3"), Some(Duration::ZERO));
        assert_eq!(parse("120"), Some(Duration::from_secs(30)));
        // Unparseable / absent → None (falls back to exponential backoff).
        assert_eq!(parse("soon"), None);
        assert_eq!(parse_retry_after(&HeaderMap::new()), None);

        // HTTP-date in the past → no wait.
        let past = httpdate::fmt_http_date(std::time::UNIX_EPOCH + Duration::from_secs(1));
        assert_eq!(parse(&past), Some(Duration::ZERO));

        // HTTP-date far in the future → capped at 30s.
        let future = httpdate::fmt_http_date(SystemTime::now() + Duration::from_secs(600));
        assert_eq!(parse(&future), Some(Duration::from_secs(30)));
    }
}
