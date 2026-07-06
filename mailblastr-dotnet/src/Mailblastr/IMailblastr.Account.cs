namespace Mailblastr;

public partial interface IMailblastr
{
    // ---- API keys ----

    /// <summary>Create an API key. The full token is returned ONCE. POST /api-keys</summary>
    Task<ApiKeyCreated> ApiKeyCreateAsync(ApiKeyCreateOptions options, CancellationToken cancellationToken = default);

    /// <summary>List API keys (non-secret display prefixes only). GET /api-keys</summary>
    Task<ListResponse<ApiKey>> ApiKeyListAsync(CancellationToken cancellationToken = default);

    /// <summary>Revoke an API key. DELETE /api-keys/:id</summary>
    Task<RemovedResponse> ApiKeyDeleteAsync(string apiKeyId, CancellationToken cancellationToken = default);

    // ---- Logs ----

    /// <summary>
    /// List API request logs. Cursor-paginated with optional server-side
    /// method/status filters. GET /logs
    /// </summary>
    Task<ListResponse<LogEntry>> LogListAsync(LogListOptions? options = null, CancellationToken cancellationToken = default);

    /// <summary>Retrieve a log entry with request/response bodies. GET /logs/:id</summary>
    Task<LogEntry> LogRetrieveAsync(string logId, CancellationToken cancellationToken = default);

    // ---- Polls (read-only results of the in-email poll widget) ----

    /// <summary>One summary row per email that has poll responses. GET /polls</summary>
    Task<ListResponse<Poll>> PollListAsync(PaginationOptions? pagination = null, CancellationToken cancellationToken = default);

    /// <summary>The aggregated answer breakdown for one email. GET /polls/:emailId</summary>
    Task<PollResult> PollRetrieveAsync(string emailId, CancellationToken cancellationToken = default);
}
