//! HTTP core: the [`Mailblastr`] client, request building, and API-error
//! mapping shared by every service.

use std::fmt;
use std::sync::Arc;

use reqwest::{Method, RequestBuilder};
use serde::de::DeserializeOwned;

use crate::services::*;
use crate::types::{Error, PaginationParams, Result};

/// The production MailBlastr API host.
pub const DEFAULT_BASE_URL: &str = "https://api.mailblastr.com";

/// This crate's version (kept in sync with `Cargo.toml`).
pub const VERSION: &str = env!("CARGO_PKG_VERSION");

/// `User-Agent` header sent with every request.
pub const USER_AGENT: &str = concat!("mailblastr-rust/", env!("CARGO_PKG_VERSION"));

/// Shared request configuration handed to every service (cheap `Arc` clone).
pub(crate) struct Config {
    api_key: String,
    base_url: String,
    client: reqwest::Client,
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

    /// Execute a request and decode the JSON response; non-2xx statuses are
    /// mapped to [`Error::Api`] parsed from the standard error body.
    pub(crate) async fn send<T: DeserializeOwned>(&self, req: RequestBuilder) -> Result<T> {
        let resp = req.send().await?;
        let status = resp.status();
        let bytes = resp.bytes().await?;
        if !status.is_success() {
            return Err(api_error(status.as_u16(), &bytes));
        }
        serde_json::from_slice(&bytes).map_err(Error::Json)
    }

    /// Execute a request whose success response streams raw bytes (received
    /// attachment / RFC822 downloads). Errors are still JSON and are mapped
    /// like [`Config::send`].
    pub(crate) async fn send_raw(&self, req: RequestBuilder) -> Result<Vec<u8>> {
        let resp = req.send().await?;
        let status = resp.status();
        let bytes = resp.bytes().await?;
        if !status.is_success() {
            return Err(api_error(status.as_u16(), &bytes));
        }
        Ok(bytes.to_vec())
    }
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
    /// Create a client against the production API (`https://api.mailblastr.com`).
    pub fn new(api_key: impl Into<String>) -> Self {
        Self::with_base_url(api_key, DEFAULT_BASE_URL)
    }

    /// Point the client at a different host (a proxy, a staging environment,
    /// or a local test stub). A trailing `/` on `base_url` is trimmed.
    pub fn with_base_url(api_key: impl Into<String>, base_url: impl Into<String>) -> Self {
        Self::with_client(api_key, base_url, reqwest::Client::new())
    }

    /// Full control: bring your own [`reqwest::Client`] (timeouts, proxies,
    /// connection pools).
    pub fn with_client(
        api_key: impl Into<String>,
        base_url: impl Into<String>,
        client: reqwest::Client,
    ) -> Self {
        let base = base_url.into();
        let config = Arc::new(Config {
            api_key: api_key.into(),
            base_url: base.trim_end_matches('/').to_owned(),
            client,
        });
        Self {
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
}
