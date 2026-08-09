//! Option and response types for the `emails`, `emails.receiving`, and
//! `batch` services.

use std::collections::HashMap;

use serde::{Deserialize, Serialize};
use serde_json::{Map, Value};

// ── Options ──────────────────────────────────────────────────────────────────

/// A file attached to an outgoing email. Provide the content inline (base64,
/// [`Attachment::from_content`]) OR as a hosted URL fetched at send time
/// ([`Attachment::from_path`]).
#[derive(Debug, Clone, Default, Serialize)]
pub struct Attachment {
    pub filename: String,
    /// Base64-encoded file content (`content` OR `path`).
    #[serde(skip_serializing_if = "Option::is_none")]
    pub content: Option<String>,
    /// A hosted URL to fetch the file from (`content` OR `path`).
    #[serde(skip_serializing_if = "Option::is_none")]
    pub path: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub content_type: Option<String>,
    /// Content-ID for inline/related parts (renders as `cid:` references).
    #[serde(skip_serializing_if = "Option::is_none")]
    pub content_id: Option<String>,
}

impl Attachment {
    /// Attach inline base64-encoded content.
    pub fn from_content(filename: impl Into<String>, base64_content: impl Into<String>) -> Self {
        Self {
            filename: filename.into(),
            content: Some(base64_content.into()),
            ..Default::default()
        }
    }

    /// Attach a file hosted at `url`, fetched at send time.
    pub fn from_path(filename: impl Into<String>, url: impl Into<String>) -> Self {
        Self {
            filename: filename.into(),
            path: Some(url.into()),
            ..Default::default()
        }
    }

    pub fn with_content_type(mut self, content_type: impl Into<String>) -> Self {
        self.content_type = Some(content_type.into());
        self
    }

    pub fn with_content_id(mut self, content_id: impl Into<String>) -> Self {
        self.content_id = Some(content_id.into());
        self
    }
}

/// Nested template reference: select a template by `id` OR `alias` and carry
/// its variable values. Provide `template` OR `html`/`text`, not both.
#[derive(Debug, Clone, Default, Serialize)]
pub struct TemplateRef {
    #[serde(skip_serializing_if = "Option::is_none")]
    pub id: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub alias: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub variables: Option<Map<String, Value>>,
}

impl TemplateRef {
    pub fn by_id(id: impl Into<String>) -> Self {
        Self {
            id: Some(id.into()),
            ..Default::default()
        }
    }

    pub fn by_alias(alias: impl Into<String>) -> Self {
        Self {
            alias: Some(alias.into()),
            ..Default::default()
        }
    }

    pub fn with_variable(mut self, key: impl Into<String>, value: impl Into<Value>) -> Self {
        self.variables
            .get_or_insert_with(Map::new)
            .insert(key.into(), value.into());
        self
    }
}

/// Options for sending an email (`POST /emails`).
///
/// ```no_run
/// use mailblastr::SendEmailOptions;
///
/// let email = SendEmailOptions::new(
///     "Acme <hello@yourdomain.com>",
///     ["user@example.com"],
///     "Hello",
/// )
/// .with_html("<p>Hi 👋</p>");
/// ```
#[derive(Debug, Clone, Default, Serialize)]
pub struct SendEmailOptions {
    pub from: String,
    pub to: Vec<String>,
    pub subject: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub cc: Option<Vec<String>>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub bcc: Option<Vec<String>>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub reply_to: Option<Vec<String>>,
    /// HTML body. Markdown-style `[text](url)` links and bare URLs in the
    /// body text are converted to tracked hyperlinks automatically at send
    /// time (content inside `<a>` tags, attributes, and `<pre>`/`<code>`
    /// blocks is left untouched).
    #[serde(skip_serializing_if = "Option::is_none")]
    pub html: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub text: Option<String>,
    /// Inbox preview text (preheader) shown next to the subject in the
    /// recipient's inbox list. Max 150 characters.
    #[serde(skip_serializing_if = "Option::is_none")]
    pub preview_text: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub headers: Option<HashMap<String, String>>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub attachments: Option<Vec<Attachment>>,
    /// ISO 8601 timestamp to schedule the send.
    #[serde(skip_serializing_if = "Option::is_none")]
    pub scheduled_at: Option<String>,
    /// Drop recipients unsubscribed from this topic (topic gating).
    #[serde(skip_serializing_if = "Option::is_none")]
    pub topic_id: Option<String>,
    /// Send using a saved template; its subject/html/text fill any omitted field.
    #[serde(skip_serializing_if = "Option::is_none")]
    pub template_id: Option<String>,
    /// Nested template reference (`template` OR `html`/`text`, not both).
    #[serde(skip_serializing_if = "Option::is_none")]
    pub template: Option<TemplateRef>,
    /// Values for the template's `{{ placeholder }}` variables.
    #[serde(skip_serializing_if = "Option::is_none")]
    pub variables: Option<Map<String, Value>>,
}

impl SendEmailOptions {
    pub fn new(
        from: impl Into<String>,
        to: impl IntoIterator<Item = impl Into<String>>,
        subject: impl Into<String>,
    ) -> Self {
        Self {
            from: from.into(),
            to: to.into_iter().map(Into::into).collect(),
            subject: subject.into(),
            ..Default::default()
        }
    }

    pub fn with_cc(mut self, cc: impl IntoIterator<Item = impl Into<String>>) -> Self {
        self.cc = Some(cc.into_iter().map(Into::into).collect());
        self
    }

    pub fn with_bcc(mut self, bcc: impl IntoIterator<Item = impl Into<String>>) -> Self {
        self.bcc = Some(bcc.into_iter().map(Into::into).collect());
        self
    }

    pub fn with_reply_to(mut self, reply_to: impl IntoIterator<Item = impl Into<String>>) -> Self {
        self.reply_to = Some(reply_to.into_iter().map(Into::into).collect());
        self
    }

    pub fn with_html(mut self, html: impl Into<String>) -> Self {
        self.html = Some(html.into());
        self
    }

    pub fn with_text(mut self, text: impl Into<String>) -> Self {
        self.text = Some(text.into());
        self
    }

    pub fn with_preview_text(mut self, preview_text: impl Into<String>) -> Self {
        self.preview_text = Some(preview_text.into());
        self
    }

    pub fn with_header(mut self, name: impl Into<String>, value: impl Into<String>) -> Self {
        self.headers
            .get_or_insert_with(HashMap::new)
            .insert(name.into(), value.into());
        self
    }

    pub fn with_attachment(mut self, attachment: Attachment) -> Self {
        self.attachments
            .get_or_insert_with(Vec::new)
            .push(attachment);
        self
    }

    pub fn with_scheduled_at(mut self, scheduled_at: impl Into<String>) -> Self {
        self.scheduled_at = Some(scheduled_at.into());
        self
    }

    pub fn with_topic_id(mut self, topic_id: impl Into<String>) -> Self {
        self.topic_id = Some(topic_id.into());
        self
    }

    pub fn with_template_id(mut self, template_id: impl Into<String>) -> Self {
        self.template_id = Some(template_id.into());
        self
    }

    pub fn with_template(mut self, template: TemplateRef) -> Self {
        self.template = Some(template);
        self
    }

    pub fn with_variable(mut self, key: impl Into<String>, value: impl Into<Value>) -> Self {
        self.variables
            .get_or_insert_with(Map::new)
            .insert(key.into(), value.into());
        self
    }
}

/// A single email in a batch send (`POST /emails/batch`). Identical to
/// [`SendEmailOptions`] minus `attachments` and `scheduled_at` — the
/// batch endpoint rejects both per item; send those individually via
/// `emails.send`.
#[derive(Debug, Clone, Default, Serialize)]
pub struct BatchEmailOptions {
    pub from: String,
    pub to: Vec<String>,
    pub subject: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub cc: Option<Vec<String>>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub bcc: Option<Vec<String>>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub reply_to: Option<Vec<String>>,
    /// HTML body. Markdown-style `[text](url)` links and bare URLs in the
    /// body text are converted to tracked hyperlinks automatically at send
    /// time.
    #[serde(skip_serializing_if = "Option::is_none")]
    pub html: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub text: Option<String>,
    /// Inbox preview text (preheader). Max 150 characters.
    #[serde(skip_serializing_if = "Option::is_none")]
    pub preview_text: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub headers: Option<HashMap<String, String>>,
    /// Drop recipients unsubscribed from this topic (topic gating).
    #[serde(skip_serializing_if = "Option::is_none")]
    pub topic_id: Option<String>,
    /// Send using a saved template; its subject/html/text fill any omitted field.
    #[serde(skip_serializing_if = "Option::is_none")]
    pub template_id: Option<String>,
    /// Nested template reference (`template` OR `html`/`text`, not both).
    #[serde(skip_serializing_if = "Option::is_none")]
    pub template: Option<TemplateRef>,
    /// Values for the template's `{{ placeholder }}` variables.
    #[serde(skip_serializing_if = "Option::is_none")]
    pub variables: Option<Map<String, Value>>,
}

impl BatchEmailOptions {
    pub fn new(
        from: impl Into<String>,
        to: impl IntoIterator<Item = impl Into<String>>,
        subject: impl Into<String>,
    ) -> Self {
        Self {
            from: from.into(),
            to: to.into_iter().map(Into::into).collect(),
            subject: subject.into(),
            ..Default::default()
        }
    }

    pub fn with_cc(mut self, cc: impl IntoIterator<Item = impl Into<String>>) -> Self {
        self.cc = Some(cc.into_iter().map(Into::into).collect());
        self
    }

    pub fn with_bcc(mut self, bcc: impl IntoIterator<Item = impl Into<String>>) -> Self {
        self.bcc = Some(bcc.into_iter().map(Into::into).collect());
        self
    }

    pub fn with_reply_to(mut self, reply_to: impl IntoIterator<Item = impl Into<String>>) -> Self {
        self.reply_to = Some(reply_to.into_iter().map(Into::into).collect());
        self
    }

    pub fn with_html(mut self, html: impl Into<String>) -> Self {
        self.html = Some(html.into());
        self
    }

    pub fn with_text(mut self, text: impl Into<String>) -> Self {
        self.text = Some(text.into());
        self
    }

    pub fn with_preview_text(mut self, preview_text: impl Into<String>) -> Self {
        self.preview_text = Some(preview_text.into());
        self
    }

    pub fn with_header(mut self, name: impl Into<String>, value: impl Into<String>) -> Self {
        self.headers
            .get_or_insert_with(HashMap::new)
            .insert(name.into(), value.into());
        self
    }

    pub fn with_topic_id(mut self, topic_id: impl Into<String>) -> Self {
        self.topic_id = Some(topic_id.into());
        self
    }

    pub fn with_template_id(mut self, template_id: impl Into<String>) -> Self {
        self.template_id = Some(template_id.into());
        self
    }

    pub fn with_template(mut self, template: TemplateRef) -> Self {
        self.template = Some(template);
        self
    }

    pub fn with_variable(mut self, key: impl Into<String>, value: impl Into<Value>) -> Self {
        self.variables
            .get_or_insert_with(Map::new)
            .insert(key.into(), value.into());
        self
    }
}

/// Reply to a received email's sender, threaded into the same conversation.
#[derive(Debug, Clone, Default, Serialize)]
pub struct ReplyReceivedEmailOptions {
    /// A verified sending address to reply from.
    pub from: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub html: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub text: Option<String>,
    /// Defaults to `Re: …` (keeps the thread).
    #[serde(skip_serializing_if = "Option::is_none")]
    pub subject: Option<String>,
}

impl ReplyReceivedEmailOptions {
    pub fn new(from: impl Into<String>) -> Self {
        Self {
            from: from.into(),
            ..Default::default()
        }
    }

    pub fn with_html(mut self, html: impl Into<String>) -> Self {
        self.html = Some(html.into());
        self
    }

    pub fn with_text(mut self, text: impl Into<String>) -> Self {
        self.text = Some(text.into());
        self
    }

    pub fn with_subject(mut self, subject: impl Into<String>) -> Self {
        self.subject = Some(subject.into());
        self
    }
}

/// Forward a received email to other recipients.
#[derive(Debug, Clone, Serialize)]
pub struct ForwardReceivedEmailOptions {
    /// A verified sending address to forward from (required by the backend).
    pub from: String,
    pub to: Vec<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub subject: Option<String>,
}

impl ForwardReceivedEmailOptions {
    pub fn new(from: impl Into<String>, to: impl IntoIterator<Item = impl Into<String>>) -> Self {
        Self {
            from: from.into(),
            to: to.into_iter().map(Into::into).collect(),
            subject: None,
        }
    }

    pub fn with_subject(mut self, subject: impl Into<String>) -> Self {
        self.subject = Some(subject.into());
        self
    }
}

// ── Responses ────────────────────────────────────────────────────────────────

/// `{ id }` returned by send/forward/reply.
#[derive(Debug, Clone, Deserialize)]
pub struct CreateEmailResponse {
    pub id: String,
}

/// `{ data: [{ id }, …] }` returned by a batch send.
#[derive(Debug, Clone, Deserialize)]
pub struct SendEmailBatchResponse {
    pub data: Vec<CreateEmailResponse>,
}

#[derive(Debug, Clone, Deserialize)]
pub struct EmailEvent {
    #[serde(rename = "type")]
    pub event_type: String,
    pub created_at: String,
}

/// A sent email with its full event timeline (`GET /emails/:id`).
#[derive(Debug, Clone, Deserialize)]
pub struct Email {
    pub object: String,
    pub id: String,
    /// Provider message id; `None` until the send is accepted.
    pub message_id: Option<String>,
    pub from: String,
    pub to: Vec<String>,
    /// Id of the sending domain the message went out on.
    pub domain_id: Option<String>,
    /// On the retrieve route these are ARRAYS (empty when unset), unlike the
    /// list rows in [`SentEmailListItem`], where they are `null`.
    pub cc: Option<Vec<String>>,
    pub bcc: Option<Vec<String>>,
    pub reply_to: Option<Vec<String>>,
    pub subject: Option<String>,
    pub html: Option<String>,
    pub text: Option<String>,
    pub status: String,
    pub last_event: Option<String>,
    /// Plain-language failure reason; non-null only when the row is `failed`
    /// or `scheduled`.
    pub error: Option<String>,
    pub scheduled_at: Option<String>,
    pub created_at: String,
    /// Oldest-first event timeline.
    pub events: Option<Vec<EmailEvent>>,
}

/// Trimmed row returned by `emails.list()` — no `status`/`html`/`text`/
/// `events`; use `emails.get(id)` for the full [`Email`].
#[derive(Debug, Clone, Deserialize)]
pub struct SentEmailListItem {
    pub object: String,
    pub id: String,
    pub message_id: Option<String>,
    pub from: String,
    pub to: Vec<String>,
    /// Id of the sending domain the message went out on.
    pub domain_id: Option<String>,
    /// `None` (JSON `null`) when unset — list rows never use `[]` here.
    pub cc: Option<Vec<String>>,
    pub bcc: Option<Vec<String>>,
    pub reply_to: Option<Vec<String>>,
    pub subject: Option<String>,
    /// Latest recorded event/state, e.g. `sent`, `delivered`, `bounced`.
    pub last_event: String,
    pub scheduled_at: Option<String>,
    /// The campaign this send belongs to (`None` for one-off API sends).
    pub campaign_id: Option<String>,
    /// The automation this send belongs to (`None` for one-off API sends).
    pub automation_id: Option<String>,
    pub created_at: String,
}

/// One row of `emails.sources()` — aggregate send metrics per campaign /
/// automation / one-off ("individual") origin.
#[derive(Debug, Clone, Deserialize)]
pub struct EmailSource {
    /// `campaign` | `automation` | `individual`.
    pub kind: String,
    /// `None` for the `individual` row.
    pub id: Option<String>,
    pub name: Option<String>,
    /// `None` for `automation` and `individual` rows.
    pub subject: Option<String>,
    pub status: Option<String>,
    #[serde(default)]
    pub total: u64,
    #[serde(default)]
    pub sent: u64,
    #[serde(default)]
    pub delivered: u64,
    #[serde(default)]
    pub opened: u64,
    #[serde(default)]
    pub clicked: u64,
    #[serde(default)]
    pub replied: u64,
    #[serde(default)]
    pub failed: u64,
    pub last_sent_at: Option<String>,
}

/// One row of `emails.receiving.list_addresses()` — inbound volume per
/// receiving address.
#[derive(Debug, Clone, Deserialize)]
pub struct ReceivingAddressStats {
    pub address: String,
    #[serde(default)]
    pub total: u64,
    #[serde(default)]
    pub replies: u64,
    #[serde(default)]
    pub interested: u64,
    pub last_received_at: Option<String>,
}

/// Metadata for one attachment of a SENT email.
///
/// Sent attachment bytes are not re-hosted, so `download_url` and
/// `expires_at` are always `None` here — keep your own copy of anything you
/// attach. (Received attachments DO stream: see
/// [`ReceivingSvc::get_attachment`](crate::services::emails::ReceivingSvc::get_attachment).)
#[derive(Debug, Clone, Deserialize)]
pub struct AttachmentMeta {
    pub object: String,
    /// The stored attachment id, or the array index as a string.
    pub id: String,
    pub filename: Option<String>,
    pub content_type: Option<String>,
    /// `inline` when a `content_id` is set, else `attachment`.
    pub content_disposition: String,
    pub content_id: Option<String>,
    /// Decoded byte length, when known.
    pub size: Option<u64>,
    /// Always `None` for sent attachments.
    pub download_url: Option<String>,
    /// Always `None` for sent attachments.
    pub expires_at: Option<String>,
}

/// One attachment of a RECEIVED email.
#[derive(Debug, Clone, Deserialize)]
pub struct ReceivedAttachment {
    pub id: Option<String>,
    pub filename: Option<String>,
    pub content_type: Option<String>,
    #[serde(default)]
    pub size: u64,
    pub content_id: Option<String>,
    pub content_disposition: Option<String>,
    #[serde(default)]
    pub downloadable: bool,
    pub download_url: Option<String>,
    pub expires_at: Option<String>,
    /// Backward-compatible relative-path alias.
    pub url: Option<String>,
}

/// Raw (RFC822) download pointer of a received email. Only `download_url` is
/// guaranteed; `expires_at` is omitted by the API today.
#[derive(Debug, Clone, Deserialize)]
pub struct RawDownload {
    pub download_url: String,
    pub expires_at: Option<String>,
}

/// An inbound email (`GET /emails/receiving/:id`).
#[derive(Debug, Clone, Deserialize)]
pub struct ReceivedEmail {
    pub object: String,
    pub id: String,
    pub from: String,
    pub to: Vec<String>,
    /// Id of the receiving domain (always present, may be `null`).
    pub domain_id: Option<String>,
    pub cc: Option<Vec<String>>,
    pub bcc: Option<Vec<String>>,
    pub received_for: Option<Vec<String>>,
    pub subject: Option<String>,
    pub html: Option<String>,
    pub text: Option<String>,
    /// Parsed message headers as a name → value map.
    pub headers: Option<HashMap<String, String>>,
    pub message_id: Option<String>,
    /// AI reply intent (replies only): `interested` | `neutral` | `not_interested`.
    pub category: Option<String>,
    /// The sent email this message replied to, when the reply was recognised.
    pub reply_to_email_id: Option<String>,
    pub spf: Option<String>,
    pub verdicts: Option<Value>,
    pub attachments: Option<Vec<ReceivedAttachment>>,
    #[serde(default)]
    pub raw_available: bool,
    pub raw: Option<RawDownload>,
    pub raw_url: Option<String>,
    pub created_at: String,
}
