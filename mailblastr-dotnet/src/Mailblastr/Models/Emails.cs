using System.Text.Json.Serialization;

namespace Mailblastr;

/// <summary>An email attachment. Provide <see cref="Content"/> (base64) OR <see cref="Path"/> (hosted URL).</summary>
public class EmailAttachment
{
    [JsonPropertyName("filename")]
    public string Filename { get; set; } = null!;

    /// <summary>Base64-encoded file content. Provide <c>Content</c> OR <c>Path</c>.</summary>
    [JsonPropertyName("content")]
    public string? Content { get; set; }

    /// <summary>A hosted URL to fetch the file from. Provide <c>Content</c> OR <c>Path</c>.</summary>
    [JsonPropertyName("path")]
    public string? Path { get; set; }

    [JsonPropertyName("content_type")]
    public string? ContentType { get; set; }

    /// <summary>Content-ID for inline/related parts (renders as <c>cid:</c> references).</summary>
    [JsonPropertyName("content_id")]
    public string? ContentId { get; set; }
}

/// <summary>A name/value tag attached to a sent email.</summary>
public class EmailTag
{
    [JsonPropertyName("name")]
    public string Name { get; set; } = null!;

    [JsonPropertyName("value")]
    public string Value { get; set; } = null!;
}

/// <summary>Nested template reference on send. Provide <c>Id</c> OR <c>Alias</c>.</summary>
public class TemplateReference
{
    [JsonPropertyName("id")]
    public string? Id { get; set; }

    [JsonPropertyName("alias")]
    public string? Alias { get; set; }

    [JsonPropertyName("variables")]
    public Dictionary<string, object?>? Variables { get; set; }
}

/// <summary>Payload for sending a single email (POST /emails) or one entry of a batch send.</summary>
public class EmailMessage
{
    [JsonPropertyName("from")]
    public string From { get; set; } = null!;

    [JsonPropertyName("to")]
    public EmailAddressList To { get; set; } = new();

    [JsonPropertyName("subject")]
    public string Subject { get; set; } = null!;

    [JsonPropertyName("cc")]
    public EmailAddressList? Cc { get; set; }

    [JsonPropertyName("bcc")]
    public EmailAddressList? Bcc { get; set; }

    [JsonPropertyName("reply_to")]
    public EmailAddressList? ReplyTo { get; set; }

    /// <summary>
    /// HTML body. Markdown-style <c>[text](url)</c> links and bare URLs in the
    /// body text are converted to tracked hyperlinks automatically at send time.
    /// </summary>
    [JsonPropertyName("html")]
    public string? HtmlBody { get; set; }

    [JsonPropertyName("text")]
    public string? TextBody { get; set; }

    /// <summary>Inbox preview text (preheader). Max 150 characters.</summary>
    [JsonPropertyName("preview_text")]
    public string? PreviewText { get; set; }

    [JsonPropertyName("headers")]
    public Dictionary<string, string>? Headers { get; set; }

    [JsonPropertyName("attachments")]
    public List<EmailAttachment>? Attachments { get; set; }

    [JsonPropertyName("tags")]
    public List<EmailTag>? Tags { get; set; }

    /// <summary>ISO 8601 timestamp to schedule the send.</summary>
    [JsonPropertyName("scheduled_at")]
    public string? ScheduledAt { get; set; }

    /// <summary>Drop recipients unsubscribed from this topic (topic gating).</summary>
    [JsonPropertyName("topic_id")]
    public string? TopicId { get; set; }

    /// <summary>Send using a saved template; its subject/html/text fill any omitted field.</summary>
    [JsonPropertyName("template_id")]
    public string? TemplateId { get; set; }

    /// <summary>Nested template reference. Provide <c>Template</c> OR <c>HtmlBody</c>/<c>TextBody</c>, not both.</summary>
    [JsonPropertyName("template")]
    public TemplateReference? Template { get; set; }

    /// <summary>Values for the template's <c>{{ placeholder }}</c> variables.</summary>
    [JsonPropertyName("variables")]
    public Dictionary<string, object?>? Variables { get; set; }
}

/// <summary>
/// A single email in a batch send (POST /emails/batch). Identical to
/// <see cref="EmailMessage"/> minus <c>Attachments</c> and <c>ScheduledAt</c> —
/// the batch endpoint rejects both per item; send those individually via
/// <c>EmailSendAsync</c>.
/// </summary>
public class BatchEmailMessage
{
    [JsonPropertyName("from")]
    public string From { get; set; } = null!;

    [JsonPropertyName("to")]
    public EmailAddressList To { get; set; } = new();

    [JsonPropertyName("subject")]
    public string Subject { get; set; } = null!;

    [JsonPropertyName("cc")]
    public EmailAddressList? Cc { get; set; }

    [JsonPropertyName("bcc")]
    public EmailAddressList? Bcc { get; set; }

    [JsonPropertyName("reply_to")]
    public EmailAddressList? ReplyTo { get; set; }

    /// <summary>
    /// HTML body. Markdown-style <c>[text](url)</c> links and bare URLs in the
    /// body text are converted to tracked hyperlinks automatically at send time.
    /// </summary>
    [JsonPropertyName("html")]
    public string? HtmlBody { get; set; }

    [JsonPropertyName("text")]
    public string? TextBody { get; set; }

    /// <summary>Inbox preview text (preheader). Max 150 characters.</summary>
    [JsonPropertyName("preview_text")]
    public string? PreviewText { get; set; }

    [JsonPropertyName("headers")]
    public Dictionary<string, string>? Headers { get; set; }

    [JsonPropertyName("tags")]
    public List<EmailTag>? Tags { get; set; }

    /// <summary>Drop recipients unsubscribed from this topic (topic gating).</summary>
    [JsonPropertyName("topic_id")]
    public string? TopicId { get; set; }

    /// <summary>Send using a saved template; its subject/html/text fill any omitted field.</summary>
    [JsonPropertyName("template_id")]
    public string? TemplateId { get; set; }

    /// <summary>Nested template reference. Provide <c>Template</c> OR <c>HtmlBody</c>/<c>TextBody</c>, not both.</summary>
    [JsonPropertyName("template")]
    public TemplateReference? Template { get; set; }

    /// <summary>Values for the template's <c>{{ placeholder }}</c> variables.</summary>
    [JsonPropertyName("variables")]
    public Dictionary<string, object?>? Variables { get; set; }
}

/// <summary>
/// Options for listing sent emails (GET /emails): cursor pagination plus
/// optional server-side <c>campaign_id</c> / <c>automation_id</c> /
/// <c>source</c> / <c>domain_id</c> filters.
/// </summary>
public class EmailListOptions
{
    /// <summary>Page size (most endpoints cap at 100).</summary>
    public int? Limit { get; set; }

    /// <summary>Cursor: id of the last item on the previous page.</summary>
    public string? After { get; set; }

    /// <summary>Cursor: id of the first item on the next page.</summary>
    public string? Before { get; set; }

    /// <summary>Only emails sent by this campaign (takes precedence over the other source filters).</summary>
    public string? CampaignId { get; set; }

    /// <summary>Only emails sent by this automation.</summary>
    public string? AutomationId { get; set; }

    /// <summary><c>individual</c> restricts to one-off API sends (no campaign/automation).</summary>
    public string? Source { get; set; }

    /// <summary>Only emails sent from this sending domain (domain id); composes with the source filters.</summary>
    public string? DomainId { get; set; }
}

/// <summary>Response of a send: <c>{ id }</c>.</summary>
public class EmailCreated
{
    [JsonPropertyName("id")]
    public string Id { get; set; } = null!;
}

/// <summary>One entry of a sent email's event timeline.</summary>
public class EmailEvent
{
    [JsonPropertyName("type")]
    public string Type { get; set; } = null!;

    [JsonPropertyName("created_at")]
    public string CreatedAt { get; set; } = null!;
}

/// <summary>A sent email with its full detail and event timeline (GET /emails/:id).</summary>
public class Email
{
    [JsonPropertyName("object")]
    public string Object { get; set; } = "email";

    [JsonPropertyName("id")]
    public string Id { get; set; } = null!;

    /// <summary>Provider message id; null until the send is accepted.</summary>
    [JsonPropertyName("message_id")]
    public string? MessageId { get; set; }

    [JsonPropertyName("from")]
    public string From { get; set; } = null!;

    [JsonPropertyName("to")]
    public List<string> To { get; set; } = new();

    [JsonPropertyName("cc")]
    public List<string>? Cc { get; set; }

    [JsonPropertyName("bcc")]
    public List<string>? Bcc { get; set; }

    [JsonPropertyName("reply_to")]
    public List<string>? ReplyTo { get; set; }

    [JsonPropertyName("subject")]
    public string? Subject { get; set; }

    [JsonPropertyName("html")]
    public string? HtmlBody { get; set; }

    [JsonPropertyName("text")]
    public string? TextBody { get; set; }

    /// <summary>Tags echoed back on retrieve (empty array when none were set).</summary>
    [JsonPropertyName("tags")]
    public List<EmailTag>? Tags { get; set; }

    [JsonPropertyName("status")]
    public string? Status { get; set; }

    [JsonPropertyName("last_event")]
    public string? LastEvent { get; set; }

    /// <summary>Plain-language failure reason; non-null only when the send failed.</summary>
    [JsonPropertyName("error")]
    public string? Error { get; set; }

    [JsonPropertyName("scheduled_at")]
    public string? ScheduledAt { get; set; }

    [JsonPropertyName("created_at")]
    public string CreatedAt { get; set; } = null!;

    [JsonPropertyName("events")]
    public List<EmailEvent>? Events { get; set; }
}

/// <summary>
/// Lightweight row returned by EmailListAsync (GET /emails). The list endpoint
/// trims the full <see cref="Email"/>: no status/html/text/events/tags, and
/// unset cc/bcc/reply_to come back as null (not []). Use EmailRetrieveAsync
/// for the full email with its event timeline.
/// </summary>
public class SentEmailListItem
{
    [JsonPropertyName("object")]
    public string Object { get; set; } = "email";

    [JsonPropertyName("id")]
    public string Id { get; set; } = null!;

    [JsonPropertyName("message_id")]
    public string? MessageId { get; set; }

    [JsonPropertyName("from")]
    public string From { get; set; } = null!;

    [JsonPropertyName("to")]
    public List<string> To { get; set; } = new();

    [JsonPropertyName("cc")]
    public List<string>? Cc { get; set; }

    [JsonPropertyName("bcc")]
    public List<string>? Bcc { get; set; }

    [JsonPropertyName("reply_to")]
    public List<string>? ReplyTo { get; set; }

    [JsonPropertyName("subject")]
    public string? Subject { get; set; }

    /// <summary>Latest recorded event/state, e.g. <c>sent</c>, <c>delivered</c>, <c>bounced</c>.</summary>
    [JsonPropertyName("last_event")]
    public string? LastEvent { get; set; }

    [JsonPropertyName("scheduled_at")]
    public string? ScheduledAt { get; set; }

    [JsonPropertyName("created_at")]
    public string CreatedAt { get; set; } = null!;
}

/// <summary>Metadata for an attachment of a SENT email (GET /emails/:id/attachments).</summary>
public class AttachmentMeta
{
    [JsonPropertyName("object")]
    public string Object { get; set; } = "attachment";

    [JsonPropertyName("id")]
    public string Id { get; set; } = null!;

    [JsonPropertyName("filename")]
    public string? Filename { get; set; }

    [JsonPropertyName("content_type")]
    public string? ContentType { get; set; }

    [JsonPropertyName("content_disposition")]
    public string? ContentDisposition { get; set; }

    [JsonPropertyName("content_id")]
    public string? ContentId { get; set; }

    [JsonPropertyName("size")]
    public long? Size { get; set; }

    [JsonPropertyName("download_url")]
    public string? DownloadUrl { get; set; }

    [JsonPropertyName("expires_at")]
    public string? ExpiresAt { get; set; }
}
