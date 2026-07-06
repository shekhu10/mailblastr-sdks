namespace Mailblastr;

/// <summary>
/// The MailBlastr API client surface — one flat async method per operation.
/// <code>
/// IMailblastr mailblastr = MailblastrClient.Create("mb_xxxxxxxxx");
/// await mailblastr.EmailSendAsync(new EmailMessage
/// {
///     From = "Acme &lt;hello@yourdomain.com&gt;",
///     To = "user@example.com",
///     Subject = "Hello",
///     HtmlBody = "&lt;p&gt;Hi&lt;/p&gt;",
/// });
/// </code>
/// Non-2xx API responses throw <see cref="MailblastrException"/> with the
/// parsed StatusCode / Name / Message.
/// </summary>
public partial interface IMailblastr
{
    // ---- Emails ----

    /// <summary>Send a single email. POST /emails</summary>
    /// <param name="message">The email to send.</param>
    /// <param name="idempotencyKey">Optional Idempotency-Key for safely retrying the create (24h window).</param>
    /// <param name="cancellationToken">Token to cancel the request.</param>
    Task<EmailCreated> EmailSendAsync(EmailMessage message, string? idempotencyKey = null, CancellationToken cancellationToken = default);

    /// <summary>Send up to 100 emails in one request. POST /emails/batch</summary>
    Task<List<EmailCreated>> EmailBatchAsync(IEnumerable<EmailMessage> messages, string? idempotencyKey = null, CancellationToken cancellationToken = default);

    /// <summary>List sent emails (trimmed rows — see <see cref="SentEmailListItem"/>). GET /emails</summary>
    Task<ListResponse<SentEmailListItem>> EmailListAsync(PaginationOptions? pagination = null, CancellationToken cancellationToken = default);

    /// <summary>Retrieve a sent email and its events. GET /emails/:id</summary>
    Task<Email> EmailRetrieveAsync(string emailId, CancellationToken cancellationToken = default);

    /// <summary>List a sent email's attachments. GET /emails/:id/attachments</summary>
    Task<ListResponse<AttachmentMeta>> EmailListAttachmentsAsync(string emailId, CancellationToken cancellationToken = default);

    /// <summary>Retrieve one attachment of a sent email. GET /emails/:id/attachments/:attachmentId</summary>
    Task<AttachmentMeta> EmailRetrieveAttachmentAsync(string emailId, string attachmentId, CancellationToken cancellationToken = default);

    /// <summary>Reschedule a scheduled email. PATCH /emails/:id</summary>
    /// <param name="emailId">Id of the scheduled email.</param>
    /// <param name="scheduledAt">ISO 8601 timestamp of the new send time.</param>
    /// <param name="cancellationToken">Token to cancel the request.</param>
    Task<ObjectRef> EmailRescheduleAsync(string emailId, string scheduledAt, CancellationToken cancellationToken = default);

    /// <summary>Cancel a scheduled email. POST /emails/:id/cancel</summary>
    Task<ObjectRef> EmailCancelAsync(string emailId, CancellationToken cancellationToken = default);

    // ---- Received (inbound) emails ----

    /// <summary>List received emails. GET /emails/receiving</summary>
    Task<ListResponse<ReceivedEmail>> ReceivedEmailListAsync(PaginationOptions? pagination = null, CancellationToken cancellationToken = default);

    /// <summary>Retrieve a received email. GET /emails/receiving/:id</summary>
    Task<ReceivedEmail> ReceivedEmailRetrieveAsync(string receivedEmailId, CancellationToken cancellationToken = default);

    /// <summary>List a received email's attachments. GET /emails/receiving/:id/attachments</summary>
    Task<ListResponse<ReceivedAttachment>> ReceivedEmailListAttachmentsAsync(string receivedEmailId, CancellationToken cancellationToken = default);

    /// <summary>Download one attachment of a received email as raw bytes. GET /emails/receiving/:id/attachments/:attachmentId</summary>
    Task<byte[]> ReceivedEmailDownloadAttachmentAsync(string receivedEmailId, string attachmentId, CancellationToken cancellationToken = default);

    /// <summary>Download the original RFC822/MIME message as raw bytes. GET /emails/receiving/:id/raw</summary>
    Task<byte[]> ReceivedEmailDownloadRawAsync(string receivedEmailId, CancellationToken cancellationToken = default);

    /// <summary>Forward a received email. POST /emails/receiving/:id/forward</summary>
    Task<EmailCreated> ReceivedEmailForwardAsync(string receivedEmailId, ReceivedEmailForwardOptions options, CancellationToken cancellationToken = default);

    /// <summary>Reply to a received email's sender, threaded into the same conversation. POST /emails/receiving/:id/reply</summary>
    Task<EmailCreated> ReceivedEmailReplyAsync(string receivedEmailId, ReceivedEmailReplyOptions options, CancellationToken cancellationToken = default);

    /// <summary>Delete a received email. DELETE /emails/receiving/:id</summary>
    Task<RemovedResponse> ReceivedEmailDeleteAsync(string receivedEmailId, CancellationToken cancellationToken = default);
}
