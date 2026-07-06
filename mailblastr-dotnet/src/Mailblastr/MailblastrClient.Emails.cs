namespace Mailblastr;

public partial class MailblastrClient
{
    // ---- Emails ----

    public async Task<EmailCreated> EmailSendAsync(EmailMessage message, string? idempotencyKey = null, CancellationToken cancellationToken = default)
    {
        ArgumentNullException.ThrowIfNull(message);
        return await RequestAsync<EmailCreated>(HttpMethod.Post, "/emails", message, idempotencyKey, cancellationToken).ConfigureAwait(false);
    }

    public async Task<List<EmailCreated>> EmailBatchAsync(IEnumerable<EmailMessage> messages, string? idempotencyKey = null, CancellationToken cancellationToken = default)
    {
        ArgumentNullException.ThrowIfNull(messages);
        var payload = messages as IList<EmailMessage> ?? messages.ToList();
        var envelope = await RequestAsync<DataEnvelope<EmailCreated>>(HttpMethod.Post, "/emails/batch", payload, idempotencyKey, cancellationToken).ConfigureAwait(false);
        return envelope.Data;
    }

    public Task<ListResponse<SentEmailListItem>> EmailListAsync(PaginationOptions? pagination = null, CancellationToken cancellationToken = default)
        => RequestAsync<ListResponse<SentEmailListItem>>(HttpMethod.Get, "/emails" + Paginate(pagination), null, null, cancellationToken);

    public Task<Email> EmailRetrieveAsync(string emailId, CancellationToken cancellationToken = default)
        => RequestAsync<Email>(HttpMethod.Get, $"/emails/{E(emailId)}", null, null, cancellationToken);

    public Task<ListResponse<AttachmentMeta>> EmailListAttachmentsAsync(string emailId, CancellationToken cancellationToken = default)
        => RequestAsync<ListResponse<AttachmentMeta>>(HttpMethod.Get, $"/emails/{E(emailId)}/attachments", null, null, cancellationToken);

    public Task<AttachmentMeta> EmailRetrieveAttachmentAsync(string emailId, string attachmentId, CancellationToken cancellationToken = default)
        => RequestAsync<AttachmentMeta>(HttpMethod.Get, $"/emails/{E(emailId)}/attachments/{E(attachmentId)}", null, null, cancellationToken);

    public Task<ObjectRef> EmailRescheduleAsync(string emailId, string scheduledAt, CancellationToken cancellationToken = default)
    {
        ArgumentNullException.ThrowIfNull(scheduledAt);
        var body = new Dictionary<string, string> { ["scheduled_at"] = scheduledAt };
        return RequestAsync<ObjectRef>(HttpMethod.Patch, $"/emails/{E(emailId)}", body, null, cancellationToken);
    }

    public Task<ObjectRef> EmailCancelAsync(string emailId, CancellationToken cancellationToken = default)
        => RequestAsync<ObjectRef>(HttpMethod.Post, $"/emails/{E(emailId)}/cancel", null, null, cancellationToken);

    // ---- Received (inbound) emails ----

    public Task<ListResponse<ReceivedEmail>> ReceivedEmailListAsync(PaginationOptions? pagination = null, CancellationToken cancellationToken = default)
        => RequestAsync<ListResponse<ReceivedEmail>>(HttpMethod.Get, "/emails/receiving" + Paginate(pagination), null, null, cancellationToken);

    public Task<ReceivedEmail> ReceivedEmailRetrieveAsync(string receivedEmailId, CancellationToken cancellationToken = default)
        => RequestAsync<ReceivedEmail>(HttpMethod.Get, $"/emails/receiving/{E(receivedEmailId)}", null, null, cancellationToken);

    public Task<ListResponse<ReceivedAttachment>> ReceivedEmailListAttachmentsAsync(string receivedEmailId, CancellationToken cancellationToken = default)
        => RequestAsync<ListResponse<ReceivedAttachment>>(HttpMethod.Get, $"/emails/receiving/{E(receivedEmailId)}/attachments", null, null, cancellationToken);

    public Task<byte[]> ReceivedEmailDownloadAttachmentAsync(string receivedEmailId, string attachmentId, CancellationToken cancellationToken = default)
        => RequestBytesAsync(HttpMethod.Get, $"/emails/receiving/{E(receivedEmailId)}/attachments/{E(attachmentId)}", cancellationToken);

    public Task<byte[]> ReceivedEmailDownloadRawAsync(string receivedEmailId, CancellationToken cancellationToken = default)
        => RequestBytesAsync(HttpMethod.Get, $"/emails/receiving/{E(receivedEmailId)}/raw", cancellationToken);

    public Task<EmailCreated> ReceivedEmailForwardAsync(string receivedEmailId, ReceivedEmailForwardOptions options, CancellationToken cancellationToken = default)
    {
        ArgumentNullException.ThrowIfNull(options);
        return RequestAsync<EmailCreated>(HttpMethod.Post, $"/emails/receiving/{E(receivedEmailId)}/forward", options, null, cancellationToken);
    }

    public Task<EmailCreated> ReceivedEmailReplyAsync(string receivedEmailId, ReceivedEmailReplyOptions options, CancellationToken cancellationToken = default)
    {
        ArgumentNullException.ThrowIfNull(options);
        return RequestAsync<EmailCreated>(HttpMethod.Post, $"/emails/receiving/{E(receivedEmailId)}/reply", options, null, cancellationToken);
    }

    public Task<RemovedResponse> ReceivedEmailDeleteAsync(string receivedEmailId, CancellationToken cancellationToken = default)
        => RequestAsync<RemovedResponse>(HttpMethod.Delete, $"/emails/receiving/{E(receivedEmailId)}", null, null, cancellationToken);
}
