using System.Globalization;

namespace Mailblastr;

public partial class MailblastrClient
{
    // ---- API keys ----

    public Task<ApiKeyCreated> ApiKeyCreateAsync(ApiKeyCreateOptions options, CancellationToken cancellationToken = default)
    {
        ArgumentNullException.ThrowIfNull(options);
        return RequestAsync<ApiKeyCreated>(HttpMethod.Post, "/api-keys", options, null, cancellationToken);
    }

    public Task<ListResponse<ApiKey>> ApiKeyListAsync(CancellationToken cancellationToken = default)
        => RequestAsync<ListResponse<ApiKey>>(HttpMethod.Get, "/api-keys", null, null, cancellationToken);

    public Task<RemovedResponse> ApiKeyDeleteAsync(string apiKeyId, CancellationToken cancellationToken = default)
        => RequestAsync<RemovedResponse>(HttpMethod.Delete, $"/api-keys/{E(apiKeyId)}", null, null, cancellationToken);

    // ---- Logs ----

    public Task<ListResponse<LogEntry>> LogListAsync(LogListOptions? options = null, CancellationToken cancellationToken = default)
    {
        var query = Query(
            ("limit", options?.Limit?.ToString(CultureInfo.InvariantCulture)),
            ("after", options?.After),
            ("before", options?.Before),
            ("method", options?.Method),
            ("status", options?.Status?.ToString(CultureInfo.InvariantCulture)));
        return RequestAsync<ListResponse<LogEntry>>(HttpMethod.Get, "/logs" + query, null, null, cancellationToken);
    }

    public Task<LogEntry> LogRetrieveAsync(string logId, CancellationToken cancellationToken = default)
        => RequestAsync<LogEntry>(HttpMethod.Get, $"/logs/{E(logId)}", null, null, cancellationToken);

    // ---- Polls ----

    public Task<ListResponse<Poll>> PollListAsync(PaginationOptions? pagination = null, CancellationToken cancellationToken = default)
        => RequestAsync<ListResponse<Poll>>(HttpMethod.Get, "/polls" + Paginate(pagination), null, null, cancellationToken);

    public Task<PollResult> PollRetrieveAsync(string emailId, CancellationToken cancellationToken = default)
        => RequestAsync<PollResult>(HttpMethod.Get, $"/polls/{E(emailId)}", null, null, cancellationToken);
}
