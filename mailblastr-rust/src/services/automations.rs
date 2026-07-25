//! `mailblastr.automations` — multi-step automations triggered by events.
//! Every automation belongs to one of your sending domains (`domain` is
//! REQUIRED on create); only `events.send` calls naming the same `domain`
//! trigger it, so the same event name across products can never double-fire.

use std::sync::Arc;

use reqwest::Method;
use serde::{Deserialize, Serialize};
use serde_json::Value;

use crate::client::{page_query, seg, Config};
use crate::types::{ListResponse, PaginationParams, RemovedResponse, Result};

/// A step in an automation's graph.
#[derive(Debug, Clone, Deserialize)]
pub struct AutomationStep {
    /// Absent on the synthesized trigger step.
    pub id: Option<String>,
    pub key: String,
    #[serde(rename = "type")]
    pub step_type: String,
    pub position: Option<u32>,
    #[serde(default)]
    pub config: Value,
}

/// A typed edge between two step keys.
#[derive(Debug, Clone, Deserialize)]
pub struct AutomationConnection {
    pub from: String,
    pub to: String,
    #[serde(rename = "type")]
    pub connection_type: String,
}

/// Enrollment counts (retrieve only).
#[derive(Debug, Clone, Deserialize)]
pub struct AutomationEnrollments {
    pub active: u64,
    pub completed: u64,
}

/// An automation.
#[derive(Debug, Clone, Deserialize)]
pub struct Automation {
    pub object: String,
    pub id: String,
    pub audience_id: String,
    pub name: String,
    /// The event that starts a run.
    pub trigger: String,
    /// The sending domain this automation belongs to. `None` on pre-domain
    /// rows (treated as the account's single domain when exactly one exists).
    pub domain: Option<String>,
    /// `enabled` | `disabled`.
    pub status: String,
    pub steps: Option<Vec<AutomationStep>>,
    pub connections: Option<Vec<AutomationConnection>>,
    /// Included only on retrieve.
    pub enrollments: Option<AutomationEnrollments>,
    pub created_at: String,
    pub updated_at: String,
}

/// An inline step supplied on create; `key` lets connections reference it.
#[derive(Debug, Clone, Serialize)]
pub struct AutomationStepInput {
    #[serde(skip_serializing_if = "Option::is_none")]
    pub key: Option<String>,
    #[serde(rename = "type")]
    pub step_type: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub config: Option<Value>,
}

impl AutomationStepInput {
    pub fn new(step_type: impl Into<String>) -> Self {
        Self {
            key: None,
            step_type: step_type.into(),
            config: None,
        }
    }

    pub fn with_key(mut self, key: impl Into<String>) -> Self {
        self.key = Some(key.into());
        self
    }

    pub fn with_config(mut self, config: Value) -> Self {
        self.config = Some(config);
        self
    }
}

/// A typed edge between step keys supplied on create/update.
#[derive(Debug, Clone, Serialize)]
pub struct ConnectionInput {
    pub from: String,
    pub to: String,
    #[serde(rename = "type", skip_serializing_if = "Option::is_none")]
    pub connection_type: Option<String>,
}

impl ConnectionInput {
    pub fn new(from: impl Into<String>, to: impl Into<String>) -> Self {
        Self {
            from: from.into(),
            to: to.into(),
            connection_type: None,
        }
    }

    pub fn with_type(mut self, connection_type: impl Into<String>) -> Self {
        self.connection_type = Some(connection_type.into());
        self
    }
}

/// Config for the `mailblastr:schedule` trigger: the automation fires ONCE
/// at `at`, enrolling every contact of its domain's pool. Required with that
/// trigger; not accepted on any other.
#[derive(Debug, Clone, Serialize)]
pub struct AutomationTriggerConfig {
    /// ISO 8601 instant the automation fires (future, at most 366 days ahead).
    pub at: String,
    /// IANA timezone the schedule was picked in (e.g. `America/New_York`).
    pub timezone: String,
}

impl AutomationTriggerConfig {
    pub fn new(at: impl Into<String>, timezone: impl Into<String>) -> Self {
        Self {
            at: at.into(),
            timezone: timezone.into(),
        }
    }
}

/// Options for `automations.create` (`POST /automations`).
#[derive(Debug, Clone, Default, Serialize)]
pub struct CreateAutomationOptions {
    pub name: String,
    /// REQUIRED. The sending domain this automation belongs to. Only
    /// `events.send` calls with the same `domain` trigger it.
    pub domain: String,
    /// The event that starts a run: `contact.created`, the built-in scheduled
    /// trigger `mailblastr:schedule` (requires `trigger_config`), an
    /// engagement event (`email.opened` / `email.clicked` / `email.replied`
    /// / `email.bounced` / `email.delivered`), or any custom event name you
    /// send via `events.send`. Usually supplied as a `steps[0]` trigger step
    /// instead.
    #[serde(skip_serializing_if = "Option::is_none")]
    pub trigger: Option<String>,
    /// Schedule for the `mailblastr:schedule` trigger (`{at, timezone}`).
    /// Required with that trigger; not accepted on any other.
    #[serde(skip_serializing_if = "Option::is_none")]
    pub trigger_config: Option<AutomationTriggerConfig>,
    /// Initial status: `enabled` | `disabled` (default `disabled`).
    #[serde(skip_serializing_if = "Option::is_none")]
    pub status: Option<String>,
    /// Optional inline step graph.
    #[serde(skip_serializing_if = "Option::is_none")]
    pub steps: Option<Vec<AutomationStepInput>>,
    /// Optional typed edges between step keys.
    #[serde(skip_serializing_if = "Option::is_none")]
    pub connections: Option<Vec<ConnectionInput>>,
}

impl CreateAutomationOptions {
    pub fn new(name: impl Into<String>, domain: impl Into<String>) -> Self {
        Self {
            name: name.into(),
            domain: domain.into(),
            ..Default::default()
        }
    }

    pub fn with_trigger(mut self, trigger: impl Into<String>) -> Self {
        self.trigger = Some(trigger.into());
        self
    }

    /// Schedule for the `mailblastr:schedule` trigger (`{at, timezone}`).
    pub fn with_trigger_config(mut self, trigger_config: AutomationTriggerConfig) -> Self {
        self.trigger_config = Some(trigger_config);
        self
    }

    pub fn with_status(mut self, status: impl Into<String>) -> Self {
        self.status = Some(status.into());
        self
    }

    pub fn with_step(mut self, step: AutomationStepInput) -> Self {
        self.steps.get_or_insert_with(Vec::new).push(step);
        self
    }

    pub fn with_connection(mut self, connection: ConnectionInput) -> Self {
        self.connections
            .get_or_insert_with(Vec::new)
            .push(connection);
        self
    }
}

/// Options for `automations.update` (`PATCH /automations/:id`).
#[derive(Debug, Clone, Default, Serialize)]
pub struct UpdateAutomationOptions {
    #[serde(skip_serializing_if = "Option::is_none")]
    pub name: Option<String>,
    /// `enabled` | `disabled`.
    #[serde(skip_serializing_if = "Option::is_none")]
    pub status: Option<String>,
    /// Re-point at another of your domains (disabled automations only).
    #[serde(skip_serializing_if = "Option::is_none")]
    pub domain: Option<String>,
    /// Update the `mailblastr:schedule` trigger's schedule (`{at, timezone}`).
    /// Only valid on automations with that trigger.
    #[serde(skip_serializing_if = "Option::is_none")]
    pub trigger_config: Option<AutomationTriggerConfig>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub connections: Option<Vec<ConnectionInput>>,
}

impl UpdateAutomationOptions {
    pub fn new() -> Self {
        Self::default()
    }

    pub fn with_name(mut self, name: impl Into<String>) -> Self {
        self.name = Some(name.into());
        self
    }

    pub fn with_status(mut self, status: impl Into<String>) -> Self {
        self.status = Some(status.into());
        self
    }

    pub fn with_domain(mut self, domain: impl Into<String>) -> Self {
        self.domain = Some(domain.into());
        self
    }

    /// Update the `mailblastr:schedule` trigger's schedule (`{at, timezone}`).
    pub fn with_trigger_config(mut self, trigger_config: AutomationTriggerConfig) -> Self {
        self.trigger_config = Some(trigger_config);
        self
    }

    pub fn with_connections(mut self, connections: Vec<ConnectionInput>) -> Self {
        self.connections = Some(connections);
        self
    }
}

/// Options for `automations.add_step` (`POST /automations/:id/steps`).
#[derive(Debug, Clone, Serialize)]
pub struct AddAutomationStepOptions {
    #[serde(rename = "type")]
    pub step_type: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub config: Option<Value>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub key: Option<String>,
}

impl AddAutomationStepOptions {
    pub fn new(step_type: impl Into<String>) -> Self {
        Self {
            step_type: step_type.into(),
            config: None,
            key: None,
        }
    }

    pub fn with_config(mut self, config: Value) -> Self {
        self.config = Some(config);
        self
    }

    pub fn with_key(mut self, key: impl Into<String>) -> Self {
        self.key = Some(key.into());
        self
    }
}

/// `{ id, deleted }` returned by `automations.delete_step`.
#[derive(Debug, Clone, Deserialize)]
pub struct DeletedStepResponse {
    pub id: String,
    pub deleted: bool,
}

/// One executed step within a run trace.
#[derive(Debug, Clone, Deserialize)]
pub struct AutomationRunStep {
    pub key: String,
    #[serde(rename = "type")]
    pub step_type: String,
    /// `completed` | `failed` | `skipped`.
    pub status: String,
    pub started_at: Option<String>,
    pub completed_at: Option<String>,
    pub output: Option<Value>,
    pub error: Option<String>,
}

/// A single enrollment (run) of a contact through an automation.
#[derive(Debug, Clone, Deserialize)]
pub struct AutomationRun {
    pub object: String,
    pub id: String,
    pub contact_id: String,
    /// Email of the contact the run is for; `None` if that contact was deleted.
    pub contact_email: Option<String>,
    /// `running` | `completed` | `failed` | `cancelled` | `skipped`.
    pub status: String,
    pub started_at: String,
    pub completed_at: Option<String>,
    pub created_at: String,
    /// Present on retrieve only.
    pub automation_id: Option<String>,
    pub steps: Option<Vec<AutomationRunStep>>,
    pub error: Option<String>,
}

/// `mailblastr.automations`.
#[derive(Clone, Debug)]
pub struct AutomationsSvc {
    config: Arc<Config>,
}

impl AutomationsSvc {
    pub(crate) fn new(config: Arc<Config>) -> Self {
        Self { config }
    }

    /// Create an automation (`domain` required). `POST /automations`
    pub async fn create(&self, options: CreateAutomationOptions) -> Result<Automation> {
        self.config
            .send(
                self.config
                    .request(Method::POST, "/automations")
                    .json(&options),
            )
            .await
    }

    /// Retrieve an automation (includes enrollments). `GET /automations/:id`
    pub async fn get(&self, automation_id: &str) -> Result<Automation> {
        let path = format!("/automations/{}", seg(automation_id));
        self.config
            .send(self.config.request(Method::GET, &path))
            .await
    }

    /// List automations. `GET /automations`
    pub async fn list(&self, params: Option<PaginationParams>) -> Result<ListResponse<Automation>> {
        let req = self
            .config
            .request(Method::GET, "/automations")
            .query(&page_query(params.as_ref()));
        self.config.send(req).await
    }

    /// Update an automation. `PATCH /automations/:id`
    pub async fn update(
        &self,
        automation_id: &str,
        options: UpdateAutomationOptions,
    ) -> Result<Automation> {
        let path = format!("/automations/{}", seg(automation_id));
        self.config
            .send(self.config.request(Method::PATCH, &path).json(&options))
            .await
    }

    /// Append a step; returns the created step. `POST /automations/:id/steps`
    pub async fn add_step(
        &self,
        automation_id: &str,
        options: AddAutomationStepOptions,
    ) -> Result<AutomationStep> {
        let path = format!("/automations/{}/steps", seg(automation_id));
        self.config
            .send(self.config.request(Method::POST, &path).json(&options))
            .await
    }

    /// Delete a step. `DELETE /automations/:id/steps/:step_id`
    pub async fn delete_step(
        &self,
        automation_id: &str,
        step_id: &str,
    ) -> Result<DeletedStepResponse> {
        let path = format!("/automations/{}/steps/{}", seg(automation_id), seg(step_id));
        self.config
            .send(self.config.request(Method::DELETE, &path))
            .await
    }

    /// List an automation's runs. `GET /automations/:id/runs`
    pub async fn runs(
        &self,
        automation_id: &str,
        params: Option<PaginationParams>,
    ) -> Result<ListResponse<AutomationRun>> {
        let path = format!("/automations/{}/runs", seg(automation_id));
        let req = self
            .config
            .request(Method::GET, &path)
            .query(&page_query(params.as_ref()));
        self.config.send(req).await
    }

    /// Retrieve a single run trace. `GET /automations/:id/runs/:run_id`
    pub async fn get_run(&self, automation_id: &str, run_id: &str) -> Result<AutomationRun> {
        let path = format!("/automations/{}/runs/{}", seg(automation_id), seg(run_id));
        self.config
            .send(self.config.request(Method::GET, &path))
            .await
    }

    /// Stop an automation — prevents new runs; in-progress runs finish.
    /// `POST /automations/:id/stop`
    pub async fn stop(&self, automation_id: &str) -> Result<Automation> {
        let path = format!("/automations/{}/stop", seg(automation_id));
        self.config
            .send(self.config.request(Method::POST, &path))
            .await
    }

    /// Delete an automation. `DELETE /automations/:id`
    pub async fn remove(&self, automation_id: &str) -> Result<RemovedResponse> {
        let path = format!("/automations/{}", seg(automation_id));
        self.config
            .send(self.config.request(Method::DELETE, &path))
            .await
    }
}
