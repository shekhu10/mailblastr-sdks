//! `mailblastr.api_keys` — API key management. The full secret token is
//! returned only ONCE, at creation.

use std::sync::Arc;

use reqwest::Method;
use serde::{Deserialize, Serialize};

use crate::client::{seg, Config};
use crate::types::{ListResponse, RemovedResponse, Result};

/// What an API key may do.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum ApiKeyPermission {
    FullAccess,
    SendingAccess,
}

/// An API key as returned by `api_keys.list()`.
#[derive(Debug, Clone, Deserialize)]
pub struct ApiKey {
    pub id: String,
    pub name: String,
    /// Non-secret display prefix (e.g. `mb_live_abcd…`); `None` for legacy
    /// keys. The full secret is returned only at creation.
    pub token: Option<String>,
    /// Derived from the key's scopes.
    pub permission: ApiKeyPermission,
    /// Set when the key is scoped to exactly one sending domain (legacy).
    pub domain_id: Option<String>,
    /// Domains the key is scoped to; `None` when unscoped.
    pub domain_ids: Option<Vec<String>>,
    pub created_at: String,
    /// Last time the key authenticated a request; `None` if never used.
    pub last_used_at: Option<String>,
}

/// Options for `api_keys.create` (`POST /api-keys`).
#[derive(Debug, Clone, Serialize)]
pub struct CreateApiKeyOptions {
    pub name: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub permission: Option<ApiKeyPermission>,
    /// Scope a `sending_access` key to one domain (legacy; prefer
    /// `domain_ids`).
    #[serde(skip_serializing_if = "Option::is_none")]
    pub domain_id: Option<String>,
    /// Scope the key to one or more domains. Only valid with `sending_access` —
/// full-access keys always work across all your domains (the API rejects the
/// combination with a validation_error).
    /// Mutually exclusive with `domain_id` — providing both is a 422.
    #[serde(skip_serializing_if = "Option::is_none")]
    pub domain_ids: Option<Vec<String>>,
}

impl CreateApiKeyOptions {
    pub fn new(name: impl Into<String>) -> Self {
        Self {
            name: name.into(),
            permission: None,
            domain_id: None,
            domain_ids: None,
        }
    }

    pub fn with_permission(mut self, permission: ApiKeyPermission) -> Self {
        self.permission = Some(permission);
        self
    }

    pub fn with_domain_id(mut self, domain_id: impl Into<String>) -> Self {
        self.domain_id = Some(domain_id.into());
        self
    }

    pub fn with_domain_ids(mut self, domain_ids: impl IntoIterator<Item = impl Into<String>>) -> Self {
        self.domain_ids = Some(domain_ids.into_iter().map(Into::into).collect());
        self
    }
}

/// `{ object, id, token, domain_id, domain_ids }` — `token` is the full
/// secret, revealed ONCE.
#[derive(Debug, Clone, Deserialize)]
pub struct CreateApiKeyResponse {
    pub object: String,
    pub id: String,
    pub token: String,
    pub domain_id: Option<String>,
    pub domain_ids: Option<Vec<String>>,
}

/// `mailblastr.api_keys`.
#[derive(Clone, Debug)]
pub struct ApiKeysSvc {
    config: Arc<Config>,
}

impl ApiKeysSvc {
    pub(crate) fn new(config: Arc<Config>) -> Self {
        Self { config }
    }

    /// Create an API key (secret returned once). `POST /api-keys`
    pub async fn create(&self, options: CreateApiKeyOptions) -> Result<CreateApiKeyResponse> {
        self.config
            .send(
                self.config
                    .request(Method::POST, "/api-keys")
                    .json(&options),
            )
            .await
    }

    /// List API keys (non-secret prefixes only). `GET /api-keys`
    pub async fn list(&self) -> Result<ListResponse<ApiKey>> {
        self.config
            .send(self.config.request(Method::GET, "/api-keys"))
            .await
    }

    /// Revoke an API key. `DELETE /api-keys/:id`
    pub async fn remove(&self, api_key_id: &str) -> Result<RemovedResponse> {
        let path = format!("/api-keys/{}", seg(api_key_id));
        self.config
            .send(self.config.request(Method::DELETE, &path))
            .await
    }
}
