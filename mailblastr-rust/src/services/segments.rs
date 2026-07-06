//! `mailblastr.segments` — DOMAIN-REQUIRED segments. A segment belongs to a
//! sending domain's contact pool; names are unique WITHIN a domain (freely
//! reusable across domains), and every domain carries an auto-created
//! "General" (all contacts) segment.

use std::sync::Arc;

use reqwest::Method;
use serde::{Deserialize, Serialize};
use serde_json::Value;

use crate::client::{seg, Config};
use crate::services::contact_types::Contact;
use crate::types::{ListResponse, RemovedResponse, Result};

/// Subscription-status filter of a segment.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum SegmentStatus {
    All,
    Subscribed,
    Unsubscribed,
}

/// Operator for a custom-property segment predicate.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum PropertyOperator {
    Eq,
    Contains,
    Exists,
}

/// A single custom-property predicate. `value` is required for
/// `eq`/`contains` (use the constructors).
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PropertyFilter {
    pub key: String,
    pub operator: PropertyOperator,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub value: Option<Value>,
}

impl PropertyFilter {
    /// `property == value`
    pub fn eq(key: impl Into<String>, value: impl Into<Value>) -> Self {
        Self {
            key: key.into(),
            operator: PropertyOperator::Eq,
            value: Some(value.into()),
        }
    }

    /// `property contains value`
    pub fn contains(key: impl Into<String>, value: impl Into<Value>) -> Self {
        Self {
            key: key.into(),
            operator: PropertyOperator::Contains,
            value: Some(value.into()),
        }
    }

    /// `property exists`
    pub fn exists(key: impl Into<String>) -> Self {
        Self {
            key: key.into(),
            operator: PropertyOperator::Exists,
            value: None,
        }
    }
}

/// A segment's filter as returned by the API.
#[derive(Debug, Clone, Deserialize)]
pub struct SegmentFilter {
    pub status: SegmentStatus,
    pub email_contains: Option<String>,
    #[serde(default)]
    pub property_filters: Vec<PropertyFilter>,
}

/// A segment.
#[derive(Debug, Clone, Deserialize)]
pub struct Segment {
    pub object: String,
    pub id: String,
    pub audience_id: String,
    pub name: String,
    pub filter: SegmentFilter,
    pub created_at: String,
    pub updated_at: String,
}

/// Filter accepted on segment create/update.
#[derive(Debug, Clone, Default, Serialize)]
pub struct SegmentFilterOptions {
    #[serde(skip_serializing_if = "Option::is_none")]
    pub status: Option<SegmentStatus>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub email_contains: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub property_filters: Option<Vec<PropertyFilter>>,
}

impl SegmentFilterOptions {
    pub fn new() -> Self {
        Self::default()
    }

    pub fn with_status(mut self, status: SegmentStatus) -> Self {
        self.status = Some(status);
        self
    }

    pub fn with_email_contains(mut self, email_contains: impl Into<String>) -> Self {
        self.email_contains = Some(email_contains.into());
        self
    }

    pub fn with_property_filter(mut self, filter: PropertyFilter) -> Self {
        self.property_filters
            .get_or_insert_with(Vec::new)
            .push(filter);
        self
    }
}

/// Options for `segments.create` (`POST /segments`).
///
/// ```no_run
/// use mailblastr::CreateSegmentOptions;
/// let segment = CreateSegmentOptions::new("yourdomain.com", "VIP");
/// ```
#[derive(Debug, Clone, Serialize)]
pub struct CreateSegmentOptions {
    /// REQUIRED. The sending domain this segment belongs to (one of your
    /// domains — replaces the retired `audience_id`).
    pub domain: String,
    pub name: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub filter: Option<SegmentFilterOptions>,
}

impl CreateSegmentOptions {
    pub fn new(domain: impl Into<String>, name: impl Into<String>) -> Self {
        Self {
            domain: domain.into(),
            name: name.into(),
            filter: None,
        }
    }

    pub fn with_filter(mut self, filter: SegmentFilterOptions) -> Self {
        self.filter = Some(filter);
        self
    }
}

/// Options for `segments.update` (`PATCH /segments/:id`).
#[derive(Debug, Clone, Default, Serialize)]
pub struct UpdateSegmentOptions {
    #[serde(skip_serializing_if = "Option::is_none")]
    pub name: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub filter: Option<SegmentFilterOptions>,
}

impl UpdateSegmentOptions {
    pub fn new() -> Self {
        Self::default()
    }

    pub fn with_name(mut self, name: impl Into<String>) -> Self {
        self.name = Some(name.into());
        self
    }

    pub fn with_filter(mut self, filter: SegmentFilterOptions) -> Self {
        self.filter = Some(filter);
        self
    }
}

/// Params for `segments.list` — domain-scoped (`domain` is REQUIRED).
#[derive(Debug, Clone)]
pub struct ListSegmentsParams {
    /// REQUIRED. The sending domain whose segments to list.
    pub domain: String,
    pub limit: Option<u32>,
    pub after: Option<String>,
    pub before: Option<String>,
}

impl ListSegmentsParams {
    pub fn new(domain: impl Into<String>) -> Self {
        Self {
            domain: domain.into(),
            limit: None,
            after: None,
            before: None,
        }
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
}

/// `mailblastr.segments`.
#[derive(Clone, Debug)]
pub struct SegmentsSvc {
    config: Arc<Config>,
}

impl SegmentsSvc {
    pub(crate) fn new(config: Arc<Config>) -> Self {
        Self { config }
    }

    /// Create a segment on a sending domain (`domain` required). `POST /segments`
    pub async fn create(&self, options: CreateSegmentOptions) -> Result<Segment> {
        self.config
            .send(
                self.config
                    .request(Method::POST, "/segments")
                    .json(&options),
            )
            .await
    }

    /// Retrieve a segment. `GET /segments/:id`
    pub async fn get(&self, segment_id: &str) -> Result<Segment> {
        let path = format!("/segments/{}", seg(segment_id));
        self.config
            .send(self.config.request(Method::GET, &path))
            .await
    }

    /// List a domain's segments (includes its auto-created "General"
    /// segment). `GET /segments?domain=`
    pub async fn list(&self, params: ListSegmentsParams) -> Result<ListResponse<Segment>> {
        let mut query: Vec<(&str, String)> = vec![("domain", params.domain.clone())];
        if let Some(limit) = params.limit {
            query.push(("limit", limit.to_string()));
        }
        if let Some(after) = &params.after {
            query.push(("after", after.clone()));
        }
        if let Some(before) = &params.before {
            query.push(("before", before.clone()));
        }
        self.config
            .send(self.config.request(Method::GET, "/segments").query(&query))
            .await
    }

    /// Preview the contacts a segment currently resolves to.
    /// `GET /segments/:id/contacts`
    pub async fn contacts(&self, segment_id: &str) -> Result<ListResponse<Contact>> {
        let path = format!("/segments/{}/contacts", seg(segment_id));
        self.config
            .send(self.config.request(Method::GET, &path))
            .await
    }

    /// Update a segment. `PATCH /segments/:id`
    pub async fn update(&self, segment_id: &str, options: UpdateSegmentOptions) -> Result<Segment> {
        let path = format!("/segments/{}", seg(segment_id));
        self.config
            .send(self.config.request(Method::PATCH, &path).json(&options))
            .await
    }

    /// Delete a segment. `DELETE /segments/:id`
    pub async fn remove(&self, segment_id: &str) -> Result<RemovedResponse> {
        let path = format!("/segments/{}", seg(segment_id));
        self.config
            .send(self.config.request(Method::DELETE, &path))
            .await
    }
}
