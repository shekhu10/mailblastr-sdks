//! `mailblastr.emails` — sending, listing, scheduling, attachments, and
//! inbound (received) email; plus `mailblastr.batch` for batch sends.
//! Option/response types live in [`super::email_types`].

use std::sync::Arc;

use reqwest::Method;
use serde_json::json;

use crate::client::{page_query, seg, Config};
use crate::services::email_types::{
    AttachmentMeta, CreateEmailBaseOptions, CreateEmailResponse, Email,
    ForwardReceivedEmailOptions, ReceivedAttachment, ReceivedEmail, ReplyReceivedEmailOptions,
    SendEmailBatchResponse, SentEmailListItem,
};
use crate::types::{ListResponse, ObjectAck, PaginationParams, RemovedResponse, Result};

/// Inbound (received) email — accessed as `mailblastr.emails.receiving`.
#[derive(Clone, Debug)]
pub struct ReceivingSvc {
    config: Arc<Config>,
}

impl ReceivingSvc {
    pub(crate) fn new(config: Arc<Config>) -> Self {
        Self { config }
    }

    /// List received emails. `GET /emails/receiving`
    pub async fn list(
        &self,
        params: Option<PaginationParams>,
    ) -> Result<ListResponse<ReceivedEmail>> {
        let req = self
            .config
            .request(Method::GET, "/emails/receiving")
            .query(&page_query(params.as_ref()));
        self.config.send(req).await
    }

    /// Retrieve a received email. `GET /emails/receiving/:id`
    pub async fn get(&self, email_id: &str) -> Result<ReceivedEmail> {
        let path = format!("/emails/receiving/{}", seg(email_id));
        self.config
            .send(self.config.request(Method::GET, &path))
            .await
    }

    /// List a received email's attachments. `GET /emails/receiving/:id/attachments`
    pub async fn list_attachments(
        &self,
        email_id: &str,
    ) -> Result<ListResponse<ReceivedAttachment>> {
        let path = format!("/emails/receiving/{}/attachments", seg(email_id));
        self.config
            .send(self.config.request(Method::GET, &path))
            .await
    }

    /// Download one attachment as raw bytes (the route streams the binary
    /// file, not JSON). `GET /emails/receiving/:id/attachments/:attachment_id`
    pub async fn get_attachment(&self, email_id: &str, attachment_id: &str) -> Result<Vec<u8>> {
        let path = format!(
            "/emails/receiving/{}/attachments/{}",
            seg(email_id),
            seg(attachment_id)
        );
        self.config
            .send_raw(self.config.request(Method::GET, &path))
            .await
    }

    /// Download the original RFC822/MIME message as raw bytes.
    /// `GET /emails/receiving/:id/raw`
    pub async fn get_raw(&self, email_id: &str) -> Result<Vec<u8>> {
        let path = format!("/emails/receiving/{}/raw", seg(email_id));
        self.config
            .send_raw(self.config.request(Method::GET, &path))
            .await
    }

    /// Forward a received email. `POST /emails/receiving/:id/forward`
    pub async fn forward(
        &self,
        email_id: &str,
        options: ForwardReceivedEmailOptions,
    ) -> Result<CreateEmailResponse> {
        let path = format!("/emails/receiving/{}/forward", seg(email_id));
        self.config
            .send(self.config.request(Method::POST, &path).json(&options))
            .await
    }

    /// Reply to a received email's sender, threaded into the same conversation
    /// (In-Reply-To; subject defaults to `Re: …`). `POST /emails/receiving/:id/reply`
    pub async fn reply(
        &self,
        email_id: &str,
        options: ReplyReceivedEmailOptions,
    ) -> Result<CreateEmailResponse> {
        let path = format!("/emails/receiving/{}/reply", seg(email_id));
        self.config
            .send(self.config.request(Method::POST, &path).json(&options))
            .await
    }

    /// Delete a received email. `DELETE /emails/receiving/:id`
    pub async fn remove(&self, email_id: &str) -> Result<RemovedResponse> {
        let path = format!("/emails/receiving/{}", seg(email_id));
        self.config
            .send(self.config.request(Method::DELETE, &path))
            .await
    }
}

/// `mailblastr.emails`.
#[derive(Clone, Debug)]
pub struct EmailsSvc {
    config: Arc<Config>,
    /// Inbound email sub-resource.
    pub receiving: ReceivingSvc,
}

impl EmailsSvc {
    pub(crate) fn new(config: Arc<Config>) -> Self {
        Self {
            receiving: ReceivingSvc::new(Arc::clone(&config)),
            config,
        }
    }

    /// Send a single email. `POST /emails`
    pub async fn send(&self, email: CreateEmailBaseOptions) -> Result<CreateEmailResponse> {
        self.config
            .send(self.config.request(Method::POST, "/emails").json(&email))
            .await
    }

    /// Like [`send`](Self::send), with an `Idempotency-Key` header so the
    /// create can be retried safely (24h window).
    pub async fn send_with_idempotency_key(
        &self,
        email: CreateEmailBaseOptions,
        idempotency_key: &str,
    ) -> Result<CreateEmailResponse> {
        self.config
            .send(
                self.config
                    .request(Method::POST, "/emails")
                    .header("Idempotency-Key", idempotency_key)
                    .json(&email),
            )
            .await
    }

    /// Send up to 100 emails in one request (alias of `mailblastr.batch.send`).
    /// `POST /emails/batch`
    pub async fn batch(
        &self,
        emails: Vec<CreateEmailBaseOptions>,
    ) -> Result<SendEmailBatchResponse> {
        self.config
            .send(
                self.config
                    .request(Method::POST, "/emails/batch")
                    .json(&emails),
            )
            .await
    }

    /// List sent emails — trimmed [`SentEmailListItem`] rows. `GET /emails`
    pub async fn list(
        &self,
        params: Option<PaginationParams>,
    ) -> Result<ListResponse<SentEmailListItem>> {
        let req = self
            .config
            .request(Method::GET, "/emails")
            .query(&page_query(params.as_ref()));
        self.config.send(req).await
    }

    /// Retrieve a sent email and its events. `GET /emails/:id`
    pub async fn get(&self, email_id: &str) -> Result<Email> {
        let path = format!("/emails/{}", seg(email_id));
        self.config
            .send(self.config.request(Method::GET, &path))
            .await
    }

    /// List a sent email's attachments. `GET /emails/:id/attachments`
    pub async fn list_attachments(&self, email_id: &str) -> Result<ListResponse<AttachmentMeta>> {
        let path = format!("/emails/{}/attachments", seg(email_id));
        self.config
            .send(self.config.request(Method::GET, &path))
            .await
    }

    /// Retrieve one attachment of a sent email (metadata + presigned URL).
    /// `GET /emails/:id/attachments/:attachment_id`
    pub async fn get_attachment(
        &self,
        email_id: &str,
        attachment_id: &str,
    ) -> Result<AttachmentMeta> {
        let path = format!(
            "/emails/{}/attachments/{}",
            seg(email_id),
            seg(attachment_id)
        );
        self.config
            .send(self.config.request(Method::GET, &path))
            .await
    }

    /// Reschedule a scheduled email. `PATCH /emails/:id`
    pub async fn update(&self, email_id: &str, scheduled_at: &str) -> Result<ObjectAck> {
        let path = format!("/emails/{}", seg(email_id));
        self.config
            .send(
                self.config
                    .request(Method::PATCH, &path)
                    .json(&json!({ "scheduled_at": scheduled_at })),
            )
            .await
    }

    /// Cancel a scheduled email. `POST /emails/:id/cancel`
    pub async fn cancel(&self, email_id: &str) -> Result<ObjectAck> {
        let path = format!("/emails/{}/cancel", seg(email_id));
        self.config
            .send(self.config.request(Method::POST, &path))
            .await
    }
}

/// Batch send — `mailblastr.batch.send(vec![...])`.
#[derive(Clone, Debug)]
pub struct BatchSvc {
    config: Arc<Config>,
}

impl BatchSvc {
    pub(crate) fn new(config: Arc<Config>) -> Self {
        Self { config }
    }

    /// Send up to 100 emails in one request. `POST /emails/batch`
    pub async fn send(
        &self,
        emails: Vec<CreateEmailBaseOptions>,
    ) -> Result<SendEmailBatchResponse> {
        self.config
            .send(
                self.config
                    .request(Method::POST, "/emails/batch")
                    .json(&emails),
            )
            .await
    }

    /// Like [`send`](Self::send), with an `Idempotency-Key` header.
    pub async fn send_with_idempotency_key(
        &self,
        emails: Vec<CreateEmailBaseOptions>,
        idempotency_key: &str,
    ) -> Result<SendEmailBatchResponse> {
        self.config
            .send(
                self.config
                    .request(Method::POST, "/emails/batch")
                    .header("Idempotency-Key", idempotency_key)
                    .json(&emails),
            )
            .await
    }
}
