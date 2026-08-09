//! `mailblastr.webhooks` — webhook endpoints plus a pure, local
//! Svix-style signature verification helper
//! ([`verify_webhook_signature`] / `webhooks.verify`).

use std::sync::Arc;
use std::time::{SystemTime, UNIX_EPOCH};

use base64::engine::general_purpose::STANDARD as B64;
use base64::Engine as _;
use hmac::{Hmac, Mac};
use reqwest::Method;
use serde::{Deserialize, Serialize};
use serde_json::{Map, Value};
use sha2::Sha256;

use crate::client::{page_query, seg, Config};
use crate::types::{ListResponse, ObjectAck, PaginationParams, RemovedResponse, Result};

type HmacSha256 = Hmac<Sha256>;

/// Whether a webhook is delivering.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum WebhookStatus {
    Enabled,
    Disabled,
}

/// A webhook endpoint.
#[derive(Debug, Clone, Deserialize)]
pub struct Webhook {
    pub object: String,
    pub id: String,
    pub endpoint: String,
    pub events: Vec<String>,
    pub status: String,
    /// Whether a signing secret is set (the secret itself is returned ONLY
    /// on create + rotate, never on get/list).
    pub has_secret: Option<bool>,
    pub last_delivery_at: Option<String>,
    pub last_delivery_status: Option<u16>,
    /// Consecutive delivery failure count.
    pub failure_count: Option<u64>,
    pub created_at: String,
}

/// Options for `webhooks.create` (`POST /webhooks`).
#[derive(Debug, Clone, Serialize)]
pub struct CreateWebhookOptions {
    pub endpoint: String,
    pub events: Vec<String>,
    /// Optional caller-supplied signing secret; when omitted, MailBlastr
    /// generates one (returned once).
    #[serde(skip_serializing_if = "Option::is_none")]
    pub secret: Option<String>,
}

impl CreateWebhookOptions {
    pub fn new(
        endpoint: impl Into<String>,
        events: impl IntoIterator<Item = impl Into<String>>,
    ) -> Self {
        Self {
            endpoint: endpoint.into(),
            events: events.into_iter().map(Into::into).collect(),
            secret: None,
        }
    }

    pub fn with_secret(mut self, secret: impl Into<String>) -> Self {
        self.secret = Some(secret.into());
        self
    }
}

/// Options for `webhooks.update` (`PATCH /webhooks/:id`).
#[derive(Debug, Clone, Default, Serialize)]
pub struct UpdateWebhookOptions {
    #[serde(skip_serializing_if = "Option::is_none")]
    pub endpoint: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub events: Option<Vec<String>>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub status: Option<WebhookStatus>,
}

impl UpdateWebhookOptions {
    pub fn new() -> Self {
        Self::default()
    }

    pub fn with_endpoint(mut self, endpoint: impl Into<String>) -> Self {
        self.endpoint = Some(endpoint.into());
        self
    }

    pub fn with_events(mut self, events: impl IntoIterator<Item = impl Into<String>>) -> Self {
        self.events = Some(events.into_iter().map(Into::into).collect());
        self
    }

    pub fn with_status(mut self, status: WebhookStatus) -> Self {
        self.status = Some(status);
        self
    }
}

/// `{ object, id, signing_secret }` returned by create + rotate — the
/// plaintext secret is revealed ONCE, only here.
#[derive(Debug, Clone, Deserialize)]
pub struct CreateWebhookResponse {
    pub object: String,
    pub id: String,
    pub signing_secret: String,
}

/// Result of a synchronous test delivery (`POST /webhooks/:id/test`).
///
/// **A FAILED DELIVERY IS STILL HTTP 200.** The route answers `200` whether or
/// not your endpoint accepted the event, so `webhooks.test(..)` returns `Ok`
/// either way — branch on [`ok`](Self::ok), never on the absence of an error:
///
/// ```no_run
/// # async fn run(mb: mailblastr::Mailblastr) -> mailblastr::Result<()> {
/// let result = mb.webhooks.test("42").await?;
/// if !result.ok {
///     eprintln!("delivery failed: {:?}", result.error);
/// }
/// # Ok(())
/// # }
/// ```
#[derive(Debug, Clone, Deserialize)]
pub struct WebhookTestResult {
    pub object: String,
    pub id: String,
    /// Whether the endpoint accepted the test delivery. `false` means the
    /// delivery FAILED even though the HTTP status was 200.
    #[serde(default)]
    pub ok: bool,
    /// The HTTP status your endpoint returned, when it responded at all.
    pub status: Option<u16>,
    /// Why the delivery failed when [`ok`](Self::ok) is `false`, e.g.
    /// `lookup_failed` or `webhook missing or disabled`.
    pub error: Option<String>,
    /// Any further fields the route grows, kept verbatim.
    #[serde(flatten)]
    pub extra: Map<String, Value>,
}

/// The Svix-style delivery headers MailBlastr sends with each webhook
/// (`svix-id`, `svix-timestamp`, `svix-signature`). Extract them from your
/// framework's request and pass them to the verify helper.
#[derive(Debug, Clone, Default)]
pub struct WebhookHeaders {
    pub svix_id: String,
    pub svix_timestamp: String,
    pub svix_signature: String,
}

impl WebhookHeaders {
    pub fn new(
        svix_id: impl Into<String>,
        svix_timestamp: impl Into<String>,
        svix_signature: impl Into<String>,
    ) -> Self {
        Self {
            svix_id: svix_id.into(),
            svix_timestamp: svix_timestamp.into(),
            svix_signature: svix_signature.into(),
        }
    }
}

/// Outcome of verifying a webhook delivery signature.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct VerifyWebhookResult {
    /// True when a signature matches and (when checked) the timestamp is fresh.
    pub valid: bool,
    /// Machine reason when `valid` is false: `missing_headers`,
    /// `missing_secret`, `invalid_timestamp`, `timestamp_out_of_tolerance`,
    /// or `no_match`.
    pub reason: Option<&'static str>,
}

impl VerifyWebhookResult {
    fn invalid(reason: &'static str) -> Self {
        Self {
            valid: false,
            reason: Some(reason),
        }
    }
}

/// Options for signature verification (timestamp tolerance).
#[derive(Debug, Clone, Copy)]
pub struct VerifyWebhookOptions {
    /// Max allowed clock skew in seconds (default 300). `0` skips the check.
    pub tolerance_secs: u64,
}

impl Default for VerifyWebhookOptions {
    fn default() -> Self {
        Self {
            tolerance_secs: 300,
        }
    }
}

/// Derive the HMAC key from a `whsec_`-prefixed secret (base64-decode the
/// suffix); a secret without the prefix is used as raw UTF-8 bytes.
fn secret_to_key(secret: &str) -> Vec<u8> {
    if let Some(b64) = secret.strip_prefix("whsec_") {
        if let Ok(bytes) = B64.decode(b64) {
            if !bytes.is_empty() {
                return bytes;
            }
        }
    }
    secret.as_bytes().to_vec()
}

/// Constant-time byte comparison (xor-fold; length mismatch short-circuits,
/// which leaks only the length — signatures are fixed-size anyway).
fn ct_eq(a: &[u8], b: &[u8]) -> bool {
    if a.len() != b.len() {
        return false;
    }
    let mut diff = 0u8;
    for (x, y) in a.iter().zip(b.iter()) {
        diff |= x ^ y;
    }
    diff == 0
}

/// Verify a webhook delivery's Svix-style signature against your endpoint's
/// signing secret. Mirrors the MailBlastr signing scheme:
/// `"{id}.{timestamp}.{body}"` → base64 HMAC-SHA256, tagged `v1,`.
///
/// `payload` MUST be the exact raw request body string the server sent — do
/// not re-serialize parsed JSON (whitespace differences break the signature).
/// The `svix-signature` header may carry multiple space-separated
/// signatures; any one matching makes the delivery valid.
///
/// This is a pure local computation — it makes no HTTP request.
pub fn verify_webhook_signature(
    payload: &str,
    headers: &WebhookHeaders,
    secret: &str,
    options: &VerifyWebhookOptions,
) -> VerifyWebhookResult {
    if headers.svix_id.is_empty()
        || headers.svix_timestamp.is_empty()
        || headers.svix_signature.is_empty()
    {
        return VerifyWebhookResult::invalid("missing_headers");
    }
    if secret.is_empty() {
        return VerifyWebhookResult::invalid("missing_secret");
    }

    // Optional timestamp freshness check (default 5-minute tolerance).
    if options.tolerance_secs > 0 {
        let ts: i64 = match headers.svix_timestamp.trim().parse() {
            Ok(ts) => ts,
            Err(_) => return VerifyWebhookResult::invalid("invalid_timestamp"),
        };
        let now = SystemTime::now()
            .duration_since(UNIX_EPOCH)
            .map(|d| d.as_secs() as i64)
            .unwrap_or(0);
        if (now - ts).unsigned_abs() > options.tolerance_secs {
            return VerifyWebhookResult::invalid("timestamp_out_of_tolerance");
        }
    }

    let key = secret_to_key(secret);
    let mut mac = HmacSha256::new_from_slice(&key).expect("HMAC-SHA256 accepts keys of any length");
    mac.update(format!("{}.{}.{}", headers.svix_id, headers.svix_timestamp, payload).as_bytes());
    let expected = mac.finalize().into_bytes();

    // The header may contain multiple space-separated `v1,<sig>` entries;
    // any match wins.
    for part in headers.svix_signature.split(' ') {
        let trimmed = part.trim();
        if trimmed.is_empty() {
            continue;
        }
        let sig_b64 = trimmed.strip_prefix("v1,").unwrap_or(trimmed);
        if let Ok(candidate) = B64.decode(sig_b64) {
            if ct_eq(&candidate, expected.as_slice()) {
                return VerifyWebhookResult {
                    valid: true,
                    reason: None,
                };
            }
        }
    }
    VerifyWebhookResult::invalid("no_match")
}

/// `mailblastr.webhooks`.
#[derive(Clone, Debug)]
pub struct WebhooksSvc {
    config: Arc<Config>,
}

impl WebhooksSvc {
    pub(crate) fn new(config: Arc<Config>) -> Self {
        Self { config }
    }

    /// Create a webhook. The signing secret is shown ONCE, only here.
    /// `POST /webhooks`
    pub async fn create(&self, options: CreateWebhookOptions) -> Result<CreateWebhookResponse> {
        self.config
            .send(
                self.config
                    .request(Method::POST, "/webhooks")
                    .json(&options),
            )
            .await
    }

    /// Retrieve a webhook. `GET /webhooks/:id`
    pub async fn get(&self, webhook_id: &str) -> Result<Webhook> {
        let path = format!("/webhooks/{}", seg(webhook_id));
        self.config
            .send(self.config.request(Method::GET, &path))
            .await
    }

    /// List webhooks. `GET /webhooks`
    pub async fn list(&self, params: Option<PaginationParams>) -> Result<ListResponse<Webhook>> {
        let req = self
            .config
            .request(Method::GET, "/webhooks")
            .query(&page_query(params.as_ref()));
        self.config.send(req).await
    }

    /// Update a webhook. Returns the slim `{ object, id }` ack.
    /// `PATCH /webhooks/:id`
    pub async fn update(
        &self,
        webhook_id: &str,
        options: UpdateWebhookOptions,
    ) -> Result<ObjectAck> {
        let path = format!("/webhooks/{}", seg(webhook_id));
        self.config
            .send(self.config.request(Method::PATCH, &path).json(&options))
            .await
    }

    /// Rotate the signing secret. The new plaintext secret is returned ONCE;
    /// the old secret stops verifying immediately. `POST /webhooks/:id/rotate`
    pub async fn rotate(&self, webhook_id: &str) -> Result<CreateWebhookResponse> {
        let path = format!("/webhooks/{}/rotate", seg(webhook_id));
        self.config
            .send(self.config.request(Method::POST, &path))
            .await
    }

    /// Send a synchronous test delivery and return the endpoint's live
    /// result. `POST /webhooks/:id/test`
    ///
    /// A failed delivery is still HTTP 200, so this returns `Ok` even then —
    /// inspect [`WebhookTestResult::ok`].
    pub async fn test(&self, webhook_id: &str) -> Result<WebhookTestResult> {
        let path = format!("/webhooks/{}/test", seg(webhook_id));
        self.config
            .send(self.config.request(Method::POST, &path))
            .await
    }

    /// Delete a webhook. `DELETE /webhooks/:id`
    pub async fn remove(&self, webhook_id: &str) -> Result<RemovedResponse> {
        let path = format!("/webhooks/{}", seg(webhook_id));
        self.config
            .send(self.config.request(Method::DELETE, &path))
            .await
    }

    /// Verify a delivery's signature — see [`verify_webhook_signature`].
    /// Pure local computation; no HTTP request.
    pub fn verify(
        &self,
        payload: &str,
        headers: &WebhookHeaders,
        secret: &str,
        options: &VerifyWebhookOptions,
    ) -> VerifyWebhookResult {
        verify_webhook_signature(payload, headers, secret, options)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    const SECRET: &str = "whsec_c2VjcmV0"; // base64("secret")
    const NO_TOLERANCE: VerifyWebhookOptions = VerifyWebhookOptions { tolerance_secs: 0 };

    fn sign(secret: &str, id: &str, ts: &str, payload: &str) -> String {
        let key = secret_to_key(secret);
        let mut mac = HmacSha256::new_from_slice(&key).unwrap();
        mac.update(format!("{id}.{ts}.{payload}").as_bytes());
        B64.encode(mac.finalize().into_bytes())
    }

    #[test]
    fn accepts_a_valid_v1_signature() {
        let payload = r#"{"type":"email.delivered","data":{"id":"em_1"}}"#;
        let sig = sign(SECRET, "msg_1", "1700000000", payload);
        let headers = WebhookHeaders::new("msg_1", "1700000000", format!("v1,{sig}"));
        let res = verify_webhook_signature(payload, &headers, SECRET, &NO_TOLERANCE);
        assert!(res.valid, "expected valid, got {:?}", res.reason);
    }

    #[test]
    fn accepts_when_any_of_multiple_signatures_matches() {
        let payload = "{}";
        let good = sign(SECRET, "msg_2", "1700000000", payload);
        let headers = WebhookHeaders::new(
            "msg_2",
            "1700000000",
            format!("v1,AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA= v1,{good}"),
        );
        assert!(verify_webhook_signature(payload, &headers, SECRET, &NO_TOLERANCE).valid);
    }

    #[test]
    fn rejects_a_tampered_payload() {
        let sig = sign(SECRET, "msg_3", "1700000000", "{\"a\":1}");
        let headers = WebhookHeaders::new("msg_3", "1700000000", format!("v1,{sig}"));
        let res = verify_webhook_signature("{\"a\":2}", &headers, SECRET, &NO_TOLERANCE);
        assert!(!res.valid);
        assert_eq!(res.reason, Some("no_match"));
    }

    #[test]
    fn rejects_missing_headers_and_missing_secret() {
        let headers = WebhookHeaders::new("", "1700000000", "v1,abc");
        let res = verify_webhook_signature("{}", &headers, SECRET, &NO_TOLERANCE);
        assert_eq!(res.reason, Some("missing_headers"));

        let headers = WebhookHeaders::new("msg", "1700000000", "v1,abc");
        let res = verify_webhook_signature("{}", &headers, "", &NO_TOLERANCE);
        assert_eq!(res.reason, Some("missing_secret"));
    }

    #[test]
    fn rejects_a_stale_timestamp_under_default_tolerance() {
        let payload = "{}";
        let sig = sign(SECRET, "msg_4", "1700000000", payload);
        let headers = WebhookHeaders::new("msg_4", "1700000000", format!("v1,{sig}"));
        let res =
            verify_webhook_signature(payload, &headers, SECRET, &VerifyWebhookOptions::default());
        assert!(!res.valid);
        assert_eq!(res.reason, Some("timestamp_out_of_tolerance"));
    }

    #[test]
    fn raw_secret_without_whsec_prefix_is_used_as_utf8() {
        let payload = "hello";
        let sig = sign("rawsecret", "msg_5", "1700000000", payload);
        let headers = WebhookHeaders::new("msg_5", "1700000000", format!("v1,{sig}"));
        assert!(verify_webhook_signature(payload, &headers, "rawsecret", &NO_TOLERANCE).valid);
    }

    #[test]
    fn untagged_signature_is_accepted() {
        let payload = "x";
        let sig = sign(SECRET, "msg_6", "1700000000", payload);
        let headers = WebhookHeaders::new("msg_6", "1700000000", sig);
        assert!(verify_webhook_signature(payload, &headers, SECRET, &NO_TOLERANCE).valid);
    }
}
