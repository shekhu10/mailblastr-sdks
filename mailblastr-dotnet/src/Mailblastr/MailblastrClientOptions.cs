namespace Mailblastr;

/// <summary>Optional configuration for <see cref="MailblastrClient"/>.</summary>
public class MailblastrClientOptions
{
    /// <summary>API host override. Defaults to <c>https://api.mailblastr.com</c>.</summary>
    public string BaseUrl { get; set; } = MailblastrClient.DefaultBaseUrl;

    /// <summary>
    /// Supply your own <see cref="System.Net.Http.HttpClient"/> (e.g. from
    /// IHttpClientFactory). When set, the client is NOT disposed by the SDK.
    /// </summary>
    public HttpClient? HttpClient { get; set; }

    /// <summary>
    /// Supply a custom <see cref="HttpMessageHandler"/> (e.g. a test stub).
    /// Ignored when <see cref="HttpClient"/> is set.
    /// </summary>
    public HttpMessageHandler? HttpMessageHandler { get; set; }
}
