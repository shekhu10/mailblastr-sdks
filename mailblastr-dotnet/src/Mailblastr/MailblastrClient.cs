using System.Globalization;
using System.Net.Http.Headers;
using System.Text;
using System.Text.Json;

namespace Mailblastr;

/// <summary>
/// The MailBlastr API client. Create one via <see cref="Create(string)"/> and
/// reuse it for the lifetime of your application:
/// <code>
/// IMailblastr mailblastr = MailblastrClient.Create("mb_xxxxxxxxx");
/// </code>
/// </summary>
public partial class MailblastrClient : IMailblastr, IDisposable
{
    /// <summary>Default API host.</summary>
    public const string DefaultBaseUrl = "https://api.mailblastr.com";

    /// <summary>SDK version. Keep in sync with the csproj &lt;Version&gt;.</summary>
    public const string Version = "0.1.0";

    private const string UserAgent = "mailblastr-dotnet/" + Version;

    private readonly HttpClient _http;
    private readonly bool _ownsHttpClient;
    private readonly string _apiKey;
    private readonly string _baseUrl;

    /// <summary>Create a client with the default configuration.</summary>
    /// <param name="apiKey">Your MailBlastr API key, e.g. <c>mb_xxxxxxxxx</c>.</param>
    public static IMailblastr Create(string apiKey) => new MailblastrClient(apiKey);

    /// <summary>Create a client with custom options (base URL, HttpClient/handler).</summary>
    public static IMailblastr Create(string apiKey, MailblastrClientOptions options) => new MailblastrClient(apiKey, options);

    public MailblastrClient(string apiKey, MailblastrClientOptions? options = null)
    {
        if (string.IsNullOrWhiteSpace(apiKey))
        {
            throw new ArgumentException("MailBlastr: an API key is required, e.g. MailblastrClient.Create(\"mb_...\").", nameof(apiKey));
        }

        _apiKey = apiKey;
        options ??= new MailblastrClientOptions();
        _baseUrl = string.IsNullOrEmpty(options.BaseUrl) ? DefaultBaseUrl : options.BaseUrl.TrimEnd('/');

        if (options.HttpClient is not null)
        {
            _http = options.HttpClient;
            _ownsHttpClient = false;
        }
        else
        {
            _http = options.HttpMessageHandler is null ? new HttpClient() : new HttpClient(options.HttpMessageHandler);
            _ownsHttpClient = true;
        }
    }

    public void Dispose()
    {
        if (_ownsHttpClient)
        {
            _http.Dispose();
        }
        GC.SuppressFinalize(this);
    }

    // ---- HTTP core ----

    private HttpRequestMessage BuildRequest(HttpMethod method, string path, object? body, string? idempotencyKey)
    {
        var request = new HttpRequestMessage(method, _baseUrl + path);
        request.Headers.Authorization = new AuthenticationHeaderValue("Bearer", _apiKey);
        request.Headers.TryAddWithoutValidation("User-Agent", UserAgent);
        if (idempotencyKey is not null)
        {
            request.Headers.TryAddWithoutValidation("Idempotency-Key", idempotencyKey);
        }
        if (body is not null)
        {
            var json = JsonSerializer.Serialize(body, MailblastrJson.Options);
            request.Content = new StringContent(json, Encoding.UTF8, "application/json");
        }
        return request;
    }

    private async Task<T> RequestAsync<T>(HttpMethod method, string path, object? body, string? idempotencyKey, CancellationToken cancellationToken)
    {
        using var request = BuildRequest(method, path, body, idempotencyKey);
        HttpResponseMessage response;
        try
        {
            response = await _http.SendAsync(request, cancellationToken).ConfigureAwait(false);
        }
        catch (HttpRequestException ex)
        {
            throw new MailblastrException(0, "network_error", ex.Message, ex);
        }

        using (response)
        {
            var text = await response.Content.ReadAsStringAsync(cancellationToken).ConfigureAwait(false);
            if (!response.IsSuccessStatusCode)
            {
                throw CreateError((int)response.StatusCode, text);
            }
            try
            {
                var value = JsonSerializer.Deserialize<T>(text, MailblastrJson.Options);
                if (value is null)
                {
                    throw new MailblastrException((int)response.StatusCode, "invalid_response", "The API returned an empty response body.");
                }
                return value;
            }
            catch (JsonException ex)
            {
                throw new MailblastrException((int)response.StatusCode, "invalid_response", $"Failed to parse the API response as JSON: {ex.Message}", ex);
            }
        }
    }

    /// <summary>For endpoints that stream raw binary bytes (attachment / raw MIME downloads).</summary>
    private async Task<byte[]> RequestBytesAsync(HttpMethod method, string path, CancellationToken cancellationToken)
    {
        using var request = BuildRequest(method, path, body: null, idempotencyKey: null);
        HttpResponseMessage response;
        try
        {
            response = await _http.SendAsync(request, cancellationToken).ConfigureAwait(false);
        }
        catch (HttpRequestException ex)
        {
            throw new MailblastrException(0, "network_error", ex.Message, ex);
        }

        using (response)
        {
            if (!response.IsSuccessStatusCode)
            {
                var text = await response.Content.ReadAsStringAsync(cancellationToken).ConfigureAwait(false);
                throw CreateError((int)response.StatusCode, text);
            }
            return await response.Content.ReadAsByteArrayAsync(cancellationToken).ConfigureAwait(false);
        }
    }

    private static MailblastrException CreateError(int httpStatus, string body)
    {
        var status = httpStatus;
        var name = "application_error";
        var message = $"Request failed with status {httpStatus}.";
        if (!string.IsNullOrEmpty(body))
        {
            try
            {
                var parsed = JsonSerializer.Deserialize<ApiErrorBody>(body, MailblastrJson.Options);
                if (parsed is not null)
                {
                    if (parsed.StatusCode is int sc && sc > 0) status = sc;
                    if (!string.IsNullOrEmpty(parsed.Name)) name = parsed.Name;
                    if (!string.IsNullOrEmpty(parsed.Message)) message = parsed.Message;
                }
            }
            catch (JsonException)
            {
                // Non-JSON error body; keep the defaults.
            }
        }
        return new MailblastrException(status, name, message);
    }

    // ---- URL helpers ----

    /// <summary>Percent-encode an id-bearing path segment so it cannot traverse the URL path.</summary>
    private static string E(string value)
    {
        ArgumentNullException.ThrowIfNull(value);
        return Uri.EscapeDataString(value);
    }

    /// <summary>Build a query string from key/value pairs, skipping null values.</summary>
    private static string Query(params (string Key, string? Value)[] pairs)
    {
        var sb = new StringBuilder();
        foreach (var (key, value) in pairs)
        {
            if (value is null) continue;
            sb.Append(sb.Length == 0 ? '?' : '&');
            sb.Append(key).Append('=').Append(Uri.EscapeDataString(value));
        }
        return sb.ToString();
    }

    /// <summary>Build a <c>?limit=&amp;after=&amp;before=</c> query string from pagination options.</summary>
    private static string Paginate(PaginationOptions? pagination) => Query(
        ("limit", pagination?.Limit?.ToString(CultureInfo.InvariantCulture)),
        ("after", pagination?.After),
        ("before", pagination?.Before));
}
