using System.Net;
using System.Text.Json;
using Xunit;

namespace Mailblastr.Tests;

public class MailblastrClientTests
{
    private static (IMailblastr Client, StubHttpMessageHandler Stub) CreateClient(string responseBody = "{}", HttpStatusCode status = HttpStatusCode.OK)
    {
        var stub = new StubHttpMessageHandler { ResponseBody = responseBody, StatusCode = status };
        var client = MailblastrClient.Create("mb_test_key", new MailblastrClientOptions { HttpMessageHandler = stub });
        return (client, stub);
    }

    [Fact]
    public void Create_RequiresApiKey()
    {
        Assert.Throws<ArgumentException>(() => MailblastrClient.Create(""));
        Assert.Throws<ArgumentException>(() => MailblastrClient.Create("  "));
    }

    [Fact]
    public async Task EmailSend_PostsToEmailsWithBearerAndSnakeCaseBody()
    {
        var (client, stub) = CreateClient("""{"id":"em_123"}""");

        var created = await client.EmailSendAsync(new EmailMessage
        {
            From = "Acme <hello@acme.com>",
            To = "user@example.com",
            Subject = "Hi",
            HtmlBody = "<p>Hi</p>",
            PreviewText = "peek",
        }, idempotencyKey: "order-1");

        Assert.Equal("em_123", created.Id);
        Assert.Equal(HttpMethod.Post, stub.LastRequest.Method);
        Assert.Equal("https://api.mailblastr.com/emails", stub.LastRequest.RequestUri!.AbsoluteUri);
        Assert.Equal("Bearer", stub.LastRequest.Headers.Authorization!.Scheme);
        Assert.Equal("mb_test_key", stub.LastRequest.Headers.Authorization!.Parameter);
        Assert.Contains("order-1", stub.LastRequest.Headers.GetValues("Idempotency-Key"));

        using var body = JsonDocument.Parse(stub.LastRequestBody!);
        Assert.Equal("Acme <hello@acme.com>", body.RootElement.GetProperty("from").GetString());
        Assert.Equal("user@example.com", body.RootElement.GetProperty("to")[0].GetString());
        Assert.Equal("<p>Hi</p>", body.RootElement.GetProperty("html").GetString());
        Assert.Equal("peek", body.RootElement.GetProperty("preview_text").GetString());
        // Null members are omitted.
        Assert.False(body.RootElement.TryGetProperty("text", out _));
    }

    [Fact]
    public async Task EmailBatch_UnwrapsDataEnvelope()
    {
        var (client, stub) = CreateClient("""{"data":[{"id":"em_1"},{"id":"em_2"}]}""");

        var results = await client.EmailBatchAsync(new[]
        {
            new EmailMessage { From = "a@b.com", To = "x@y.com", Subject = "1" },
            new EmailMessage { From = "a@b.com", To = "z@y.com", Subject = "2" },
        });

        Assert.Equal("https://api.mailblastr.com/emails/batch", stub.LastRequest.RequestUri!.AbsoluteUri);
        Assert.Equal(2, results.Count);
        Assert.Equal("em_2", results[1].Id);
        // The batch body is a bare JSON array.
        Assert.StartsWith("[", stub.LastRequestBody!.TrimStart());
    }

    [Fact]
    public async Task NonSuccess_ThrowsMailblastrExceptionWithParsedErrorShape()
    {
        var (client, _) = CreateClient(
            """{"statusCode":422,"name":"validation_error","message":"domain is required"}""",
            HttpStatusCode.UnprocessableEntity);

        var ex = await Assert.ThrowsAsync<MailblastrException>(
            () => client.SegmentCreateAsync(new SegmentCreateOptions { Domain = null!, Name = "x" }));

        Assert.Equal(422, ex.StatusCode);
        Assert.Equal("validation_error", ex.Name);
        Assert.Equal("domain is required", ex.Message);
    }

    [Fact]
    public async Task NonSuccess_WithNonJsonBody_FallsBackToHttpStatus()
    {
        var (client, _) = CreateClient("gateway timeout", HttpStatusCode.BadGateway);

        var ex = await Assert.ThrowsAsync<MailblastrException>(() => client.EmailRetrieveAsync("em_1"));

        Assert.Equal(502, ex.StatusCode);
        Assert.Equal("application_error", ex.Name);
    }

    [Fact]
    public async Task ContactCreate_FlatRoute_IncludesDomainInBody()
    {
        var (client, stub) = CreateClient("""{"object":"contact","id":"con_1"}""");

        var created = await client.ContactCreateAsync(new ContactCreateOptions
        {
            Domain = "acme.com",
            Email = "user@example.com",
            FirstName = "Ada",
        });

        Assert.Equal("con_1", created.Id);
        Assert.Equal("https://api.mailblastr.com/contacts", stub.LastRequest.RequestUri!.AbsoluteUri);

        using var body = JsonDocument.Parse(stub.LastRequestBody!);
        Assert.Equal("acme.com", body.RootElement.GetProperty("domain").GetString());
        Assert.Equal("Ada", body.RootElement.GetProperty("first_name").GetString());
    }

    [Fact]
    public async Task ContactCreate_NestedRoute_TargetsAudienceAndOmitsDomain()
    {
        var (client, stub) = CreateClient("""{"object":"contact","id":"con_1"}""");

        await client.ContactCreateAsync(new ContactCreateOptions
        {
            AudienceId = "aud_9",
            Email = "user@example.com",
        });

        Assert.Equal("https://api.mailblastr.com/audiences/aud_9/contacts", stub.LastRequest.RequestUri!.AbsoluteUri);
        using var body = JsonDocument.Parse(stub.LastRequestBody!);
        Assert.False(body.RootElement.TryGetProperty("domain", out _));
        Assert.False(body.RootElement.TryGetProperty("audience_id", out _));
    }

    [Fact]
    public async Task ContactRetrieve_ByEmailWithDomain_AppendsQueryAndEscapes()
    {
        var (client, stub) = CreateClient("""{"object":"contact","id":"con_1","email":"a@b.com","unsubscribed":false,"created_at":"2026-01-01T00:00:00Z"}""");

        await client.ContactRetrieveAsync("user@example.com", domain: "acme.com");

        Assert.Equal(
            "https://api.mailblastr.com/contacts/user%40example.com?domain=acme.com",
            stub.LastRequest.RequestUri!.AbsoluteUri);
    }

    [Fact]
    public async Task SegmentList_RequiresDomainQuery_AndPaginates()
    {
        var (client, stub) = CreateClient("""{"object":"list","has_more":false,"data":[]}""");

        await client.SegmentListAsync("acme.com", new PaginationOptions { Limit = 5, After = "seg_1" });

        Assert.Equal(
            "https://api.mailblastr.com/segments?domain=acme.com&limit=5&after=seg_1",
            stub.LastRequest.RequestUri!.AbsoluteUri);
    }

    [Fact]
    public async Task IdBearingPaths_AreEscapedAgainstTraversal()
    {
        var (client, stub) = CreateClient("""{"object":"email","id":"x","from":"a@b.com","to":[],"created_at":"2026-01-01T00:00:00Z"}""");

        await client.EmailRetrieveAsync("../api-keys");

        Assert.Equal("https://api.mailblastr.com/emails/..%2Fapi-keys", stub.LastRequest.RequestUri!.AbsoluteUri);
    }

    [Fact]
    public async Task ContactPropertyUpdate_SendsExplicitNullToClearFallback()
    {
        var (client, stub) = CreateClient("""{"object":"contact_property","id":"prop_1"}""");

        await client.ContactPropertyUpdateAsync("prop_1", null);

        using var body = JsonDocument.Parse(stub.LastRequestBody!);
        Assert.Equal(JsonValueKind.Null, body.RootElement.GetProperty("fallback_value").ValueKind);
    }

    [Fact]
    public async Task EventSend_PostsDomainRequiredPayload()
    {
        var (client, stub) = CreateClient("""{"object":"event","id":"evt_1","enrolled":1}""");

        var result = await client.EventSendAsync(new EventSendOptions
        {
            Name = "signup.completed",
            Domain = "acme.com",
            Email = "user@example.com",
            Data = new Dictionary<string, object?> { ["plan"] = "pro" },
        });

        Assert.Equal("https://api.mailblastr.com/events/send", stub.LastRequest.RequestUri!.AbsoluteUri);
        Assert.Equal(1, result.Enrolled);
        using var body = JsonDocument.Parse(stub.LastRequestBody!);
        Assert.Equal("acme.com", body.RootElement.GetProperty("domain").GetString());
        Assert.Equal("pro", body.RootElement.GetProperty("data").GetProperty("plan").GetString());
    }

    [Fact]
    public async Task LogList_BuildsMethodAndStatusFilters()
    {
        var (client, stub) = CreateClient("""{"object":"list","has_more":false,"data":[]}""");

        await client.LogListAsync(new LogListOptions { Limit = 100, Method = "POST", Status = 429 });

        Assert.Equal(
            "https://api.mailblastr.com/logs?limit=100&method=POST&status=429",
            stub.LastRequest.RequestUri!.AbsoluteUri);
    }

    [Fact]
    public async Task Retries_On429_ThenSucceeds()
    {
        var stub = new StubHttpMessageHandler();
        // First attempt rate-limited (Retry-After: 0 keeps the test fast), then 200.
        stub.ScriptedResponses.Enqueue((HttpStatusCode.TooManyRequests, """{"name":"rate_limited","message":"slow down"}""", "0"));
        stub.ScriptedResponses.Enqueue((HttpStatusCode.OK, """{"id":"em_123"}""", null));
        var client = MailblastrClient.Create("mb_test_key", new MailblastrClientOptions { HttpMessageHandler = stub });

        var created = await client.EmailSendAsync(new EmailMessage
        {
            From = "a@b.com",
            To = "user@example.com",
            Subject = "Hi",
        });

        Assert.Equal("em_123", created.Id);
        // Two total attempts: the 429 then the 200.
        Assert.Equal(2, stub.Requests.Count);
        Assert.Equal("Bearer", stub.Requests[0].Headers.Authorization!.Scheme);
        Assert.Equal("Bearer", stub.Requests[1].Headers.Authorization!.Scheme);
    }

    [Fact]
    public async Task Retries_ExhaustMaxRetries_ThenThrowsLastError()
    {
        var stub = new StubHttpMessageHandler
        {
            StatusCode = HttpStatusCode.ServiceUnavailable,
            ResponseBody = """{"name":"unavailable","message":"try later"}""",
        };
        // maxRetries=1 => 2 total attempts, both 503.
        var client = MailblastrClient.Create("mb_test_key", new MailblastrClientOptions
        {
            HttpMessageHandler = stub,
            MaxRetries = 1,
        });

        var ex = await Assert.ThrowsAsync<MailblastrException>(() => client.EmailRetrieveAsync("em_1"));

        Assert.Equal(503, ex.StatusCode);
        Assert.Equal(2, stub.Requests.Count);
    }

    [Fact]
    public async Task DoesNotRetry_On500_OrWhenMaxRetriesZero()
    {
        // 500 is never retried, even with retries enabled.
        var stub500 = new StubHttpMessageHandler
        {
            StatusCode = HttpStatusCode.InternalServerError,
            ResponseBody = """{"name":"server_error","message":"boom"}""",
        };
        var client500 = MailblastrClient.Create("mb_test_key", new MailblastrClientOptions { HttpMessageHandler = stub500, MaxRetries = 3 });
        await Assert.ThrowsAsync<MailblastrException>(() => client500.EmailRetrieveAsync("em_1"));
        Assert.Single(stub500.Requests);

        // MaxRetries=0 disables retry even for 429.
        var stub429 = new StubHttpMessageHandler
        {
            StatusCode = HttpStatusCode.TooManyRequests,
            ResponseBody = """{"name":"rate_limited","message":"slow down"}""",
        };
        var client429 = MailblastrClient.Create("mb_test_key", new MailblastrClientOptions { HttpMessageHandler = stub429, MaxRetries = 0 });
        await Assert.ThrowsAsync<MailblastrException>(() => client429.EmailRetrieveAsync("em_1"));
        Assert.Single(stub429.Requests);
    }

    [Fact]
    public async Task BaseUrl_Override_IsRespectedAndTrailingSlashTrimmed()
    {
        var stub = new StubHttpMessageHandler { ResponseBody = """{"object":"list","has_more":false,"data":[]}""" };
        var client = MailblastrClient.Create("mb_test_key", new MailblastrClientOptions
        {
            BaseUrl = "http://localhost:3000/",
            HttpMessageHandler = stub,
        });

        await client.DomainListAsync();

        Assert.Equal("http://localhost:3000/domains", stub.LastRequest.RequestUri!.AbsoluteUri);
    }
}
