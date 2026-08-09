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
    /// <param name="idempotencyKey">
    /// Optional <c>Idempotency-Key</c> for safely retrying the send. Must be
    /// 1–255 characters after the server trims it
    /// (<see cref="MailblastrClient.MaxIdempotencyKeyLength"/>); the server —
    /// not this client — rejects anything outside that range with
    /// 400 <c>invalid_idempotency_key</c>. Replaying the key returns the
    /// original result; reusing it with a different payload is a
    /// 409 <c>invalid_idempotent_request</c>.
    /// </param>
    /// <param name="cancellationToken">Token to cancel the request.</param>
    Task<EmailCreated> EmailSendAsync(EmailMessage message, string? idempotencyKey = null, CancellationToken cancellationToken = default);

    /// <summary>Send up to 100 emails in one request. POST /emails/batch</summary>
    [Obsolete("Use the BatchEmailMessage overload — batch items reject attachments and scheduled_at (send those individually via EmailSendAsync), which BatchEmailMessage enforces at compile time.")]
    Task<List<EmailCreated>> EmailBatchAsync(IEnumerable<EmailMessage> messages, string? idempotencyKey = null, CancellationToken cancellationToken = default);

    /// <summary>
    /// Send up to 100 emails in one request. Batch items reject
    /// <c>attachments</c> and <c>scheduled_at</c> — send those individually via
    /// <see cref="EmailSendAsync"/>. <paramref name="idempotencyKey"/> follows the
    /// same 1–255 character rule as <see cref="EmailSendAsync"/>.
    /// POST /emails/batch
    /// </summary>
    Task<List<EmailCreated>> EmailBatchAsync(IEnumerable<BatchEmailMessage> messages, string? idempotencyKey = null, CancellationToken cancellationToken = default);

    /// <summary>List sent emails (trimmed rows — see <see cref="SentEmailListItem"/>). GET /emails</summary>
    Task<ListResponse<SentEmailListItem>> EmailListAsync(PaginationOptions? pagination = null, CancellationToken cancellationToken = default);

    /// <summary>
    /// List sent emails with optional server-side <c>campaign_id</c> /
    /// <c>automation_id</c> / <c>source</c> / <c>domain_id</c> / <c>status</c> /
    /// <c>search</c> filters. GET /emails
    /// </summary>
    Task<ListResponse<SentEmailListItem>> EmailListAsync(EmailListOptions options, CancellationToken cancellationToken = default);

    /// <summary>
    /// Send counters rolled up per origin (campaign / automation / individual).
    /// Not paginated. GET /emails/sources
    /// </summary>
    Task<ListResponse<EmailSource>> EmailListSourcesAsync(CancellationToken cancellationToken = default);

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
    Task<ObjectRef> EmailUpdateAsync(string emailId, string scheduledAt, CancellationToken cancellationToken = default);

    /// <summary>Cancel a scheduled email. POST /emails/:id/cancel</summary>
    Task<ObjectRef> EmailCancelAsync(string emailId, CancellationToken cancellationToken = default);

    // ---- Received (inbound) emails ----

    /// <summary>List received emails. GET /emails/receiving</summary>
    Task<ListResponse<ReceivedEmail>> ReceivedEmailListAsync(PaginationOptions? pagination = null, CancellationToken cancellationToken = default);

    /// <summary>
    /// List received emails with an optional <c>received_for</c> filter (only
    /// messages received for that address). GET /emails/receiving
    /// </summary>
    Task<ListResponse<ReceivedEmail>> ReceivedEmailListAsync(ReceivedEmailListOptions options, CancellationToken cancellationToken = default);

    /// <summary>
    /// Inbound counters per receiving address. Not paginated.
    /// GET /emails/receiving/addresses
    /// </summary>
    Task<ListResponse<ReceivingAddressStats>> ReceivedEmailListAddressesAsync(CancellationToken cancellationToken = default);

    /// <summary>Retrieve a received email. GET /emails/receiving/:id</summary>
    Task<ReceivedEmail> ReceivedEmailRetrieveAsync(string receivedEmailId, CancellationToken cancellationToken = default);

    /// <summary>
    /// List a received email's attachments. Pass <paramref name="pagination"/> to
    /// page; with no <c>Limit</c> and no <c>After</c> cursor the route returns ALL
    /// attachments with <c>HasMore = false</c>.
    /// GET /emails/receiving/:id/attachments
    /// </summary>
    Task<ListResponse<ReceivedAttachment>> ReceivedEmailListAttachmentsAsync(string receivedEmailId, PaginationOptions? pagination = null, CancellationToken cancellationToken = default);

    /// <summary>Download one attachment of a received email as raw bytes. GET /emails/receiving/:id/attachments/:attachmentId</summary>
    Task<byte[]> ReceivedEmailGetAttachmentAsync(string receivedEmailId, string attachmentId, CancellationToken cancellationToken = default);

    /// <summary>Download the original RFC822/MIME message as raw bytes. GET /emails/receiving/:id/raw</summary>
    Task<byte[]> ReceivedEmailGetRawAsync(string receivedEmailId, CancellationToken cancellationToken = default);

    /// <summary>Forward a received email. POST /emails/receiving/:id/forward</summary>
    Task<EmailCreated> ReceivedEmailForwardAsync(string receivedEmailId, ReceivedEmailForwardOptions options, CancellationToken cancellationToken = default);

    /// <summary>Reply to a received email's sender, threaded into the same conversation. POST /emails/receiving/:id/reply</summary>
    Task<EmailCreated> ReceivedEmailReplyAsync(string receivedEmailId, ReceivedEmailReplyOptions options, CancellationToken cancellationToken = default);

    /// <summary>Delete a received email. DELETE /emails/receiving/:id</summary>
    Task<RemovedResponse> ReceivedEmailDeleteAsync(string receivedEmailId, CancellationToken cancellationToken = default);
}
