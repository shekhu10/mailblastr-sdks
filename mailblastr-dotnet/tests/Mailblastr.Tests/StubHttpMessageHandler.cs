using System.Net;
using System.Text;

namespace Mailblastr.Tests;

/// <summary>
/// Injectable HttpMessageHandler stub: records every request (method, URL,
/// headers, body) and replies with a canned status + JSON body.
/// </summary>
public sealed class StubHttpMessageHandler : HttpMessageHandler
{
    public List<HttpRequestMessage> Requests { get; } = new();

    /// <summary>Request bodies read eagerly (index-aligned with <see cref="Requests"/>).</summary>
    public List<string?> RequestBodies { get; } = new();

    public HttpStatusCode StatusCode { get; set; } = HttpStatusCode.OK;

    public string ResponseBody { get; set; } = "{}";

    public HttpRequestMessage LastRequest => Requests[^1];

    public string? LastRequestBody => RequestBodies[^1];

    protected override async Task<HttpResponseMessage> SendAsync(HttpRequestMessage request, CancellationToken cancellationToken)
    {
        Requests.Add(request);
        RequestBodies.Add(request.Content is null
            ? null
            : await request.Content.ReadAsStringAsync(cancellationToken));

        return new HttpResponseMessage(StatusCode)
        {
            Content = new StringContent(ResponseBody, Encoding.UTF8, "application/json"),
            RequestMessage = request,
        };
    }
}
