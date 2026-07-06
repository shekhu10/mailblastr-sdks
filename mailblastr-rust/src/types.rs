//! Shared primitives: the crate error type, pagination, and the common
//! response envelopes used across every resource.

use std::fmt;

use serde::Deserialize;

/// Convenience alias — every service method returns `Result<T, mailblastr::Error>`.
pub type Result<T> = std::result::Result<T, Error>;

/// The unified SDK error.
#[derive(Debug)]
pub enum Error {
    /// The API answered with a non-2xx status. Fields mirror the standard
    /// MailBlastr error body (`{ statusCode, name, message }`); when the body
    /// cannot be parsed, `name` falls back to `application_error` and
    /// `status_code` to the HTTP status.
    Api {
        status_code: u16,
        name: String,
        message: String,
    },
    /// A transport-level failure (connection, TLS, timeout, body read).
    Http(reqwest::Error),
    /// The response body could not be decoded as the expected JSON shape.
    Json(serde_json::Error),
}

impl fmt::Display for Error {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Error::Api {
                status_code,
                name,
                message,
            } => write!(f, "MailBlastr API error {status_code} ({name}): {message}"),
            Error::Http(e) => write!(f, "HTTP transport error: {e}"),
            Error::Json(e) => write!(f, "failed to decode response JSON: {e}"),
        }
    }
}

impl std::error::Error for Error {
    fn source(&self) -> Option<&(dyn std::error::Error + 'static)> {
        match self {
            Error::Http(e) => Some(e),
            Error::Json(e) => Some(e),
            Error::Api { .. } => None,
        }
    }
}

impl From<reqwest::Error> for Error {
    fn from(e: reqwest::Error) -> Self {
        Error::Http(e)
    }
}

impl From<serde_json::Error> for Error {
    fn from(e: serde_json::Error) -> Self {
        Error::Json(e)
    }
}

/// Slim acknowledgement `{ object, id }` returned by several create/update
/// routes (e.g. `templates.create`, `domains.verify`).
#[derive(Debug, Clone, Deserialize)]
pub struct ObjectAck {
    pub object: String,
    pub id: String,
}

/// Minimal `{ id }` acknowledgement returned by several routes
/// (e.g. `campaigns.create`, `contacts.add_to_segment`).
#[derive(Debug, Clone, Deserialize)]
pub struct IdResponse {
    pub id: String,
}

/// The standard `{ object: 'list', has_more, data }` envelope returned by
/// every `list` endpoint.
#[derive(Debug, Clone, Deserialize)]
pub struct ListResponse<T> {
    pub object: String,
    #[serde(default)]
    pub has_more: bool,
    pub data: Vec<T>,
}

/// The standard `{ object, id, deleted: true }` envelope returned by
/// `remove` endpoints.
#[derive(Debug, Clone, Deserialize)]
pub struct RemovedResponse {
    pub object: String,
    pub id: String,
    pub deleted: bool,
}

/// Cursor pagination accepted by most `list` methods, appended as
/// `?limit=&after=&before=`.
///
/// ```
/// use mailblastr::PaginationParams;
/// let page = PaginationParams::new().with_limit(25).with_after("em_123");
/// ```
#[derive(Debug, Clone, Default)]
pub struct PaginationParams {
    /// Page size (most endpoints cap at 100).
    pub limit: Option<u32>,
    /// Cursor: id of the last item on the previous page.
    pub after: Option<String>,
    /// Cursor: id of the first item on the next page.
    pub before: Option<String>,
}

impl PaginationParams {
    pub fn new() -> Self {
        Self::default()
    }

    pub fn with_limit(mut self, limit: u32) -> Self {
        self.limit = Some(limit);
        self
    }

    pub fn with_after(mut self, after: impl Into<String>) -> Self {
        self.after = Some(after.into());
        self
    }

    pub fn with_before(mut self, before: impl Into<String>) -> Self {
        self.before = Some(before.into());
        self
    }

    pub(crate) fn to_query(&self) -> Vec<(&'static str, String)> {
        let mut q = Vec::new();
        if let Some(limit) = self.limit {
            q.push(("limit", limit.to_string()));
        }
        if let Some(after) = &self.after {
            q.push(("after", after.clone()));
        }
        if let Some(before) = &self.before {
            q.push(("before", before.clone()));
        }
        q
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn pagination_to_query_orders_and_skips_unset() {
        let q = PaginationParams::new()
            .with_limit(10)
            .with_after("a")
            .to_query();
        assert_eq!(
            q,
            vec![("limit", "10".to_string()), ("after", "a".to_string())]
        );
        assert!(PaginationParams::new().to_query().is_empty());
    }
}
