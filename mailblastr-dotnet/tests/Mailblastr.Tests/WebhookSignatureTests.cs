using System.Security.Cryptography;
using System.Text;
using Xunit;

namespace Mailblastr.Tests;

public class WebhookSignatureTests
{
    private static string Sign(string id, string timestamp, string payload, byte[] key)
    {
        using var hmac = new HMACSHA256(key);
        return Convert.ToBase64String(hmac.ComputeHash(Encoding.UTF8.GetBytes($"{id}.{timestamp}.{payload}")));
    }

    private static string Now() => DateTimeOffset.UtcNow.ToUnixTimeSeconds().ToString();

    [Fact]
    public void Verify_ValidSignature_WithWhsecSecret()
    {
        var rawKey = new byte[32];
        RandomNumberGenerator.Fill(rawKey);
        var secret = "whsec_" + Convert.ToBase64String(rawKey);

        var payload = """{"type":"email.delivered","data":{"id":"em_1"}}""";
        var id = "msg_abc";
        var ts = Now();
        var signature = "v1," + Sign(id, ts, payload, rawKey);

        var result = WebhookSignature.Verify(payload, new Dictionary<string, string>
        {
            // Mixed-case header names must be read case-insensitively.
            ["Svix-Id"] = id,
            ["SVIX-TIMESTAMP"] = ts,
            ["svix-signature"] = signature,
        }, secret);

        Assert.True(result.Valid);
        Assert.Null(result.Reason);
    }

    [Fact]
    public void Verify_ValidSignature_WithRawUtf8Secret()
    {
        var secret = "plain-secret";
        var payload = "{}";
        var id = "msg_1";
        var ts = Now();
        var signature = "v1," + Sign(id, ts, payload, Encoding.UTF8.GetBytes(secret));

        var result = WebhookSignature.Verify(payload, id, ts, signature, secret);

        Assert.True(result.Valid);
    }

    [Fact]
    public void Verify_MultipleSpaceSeparatedSignatures_AnyMatchWins()
    {
        var secret = "plain-secret";
        var payload = "{}";
        var id = "msg_1";
        var ts = Now();
        var good = "v1," + Sign(id, ts, payload, Encoding.UTF8.GetBytes(secret));
        var header = "v1,AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA= " + good;

        var result = WebhookSignature.Verify(payload, id, ts, header, secret);

        Assert.True(result.Valid);
    }

    [Fact]
    public void Verify_TamperedPayload_ReturnsNoMatch()
    {
        var secret = "plain-secret";
        var id = "msg_1";
        var ts = Now();
        var signature = "v1," + Sign(id, ts, "{}", Encoding.UTF8.GetBytes(secret));

        var result = WebhookSignature.Verify("""{"tampered":true}""", id, ts, signature, secret);

        Assert.False(result.Valid);
        Assert.Equal("no_match", result.Reason);
    }

    [Fact]
    public void Verify_MissingHeaders_ReturnsMissingHeaders()
    {
        var result = WebhookSignature.Verify("{}", new Dictionary<string, string>(), "secret");

        Assert.False(result.Valid);
        Assert.Equal("missing_headers", result.Reason);
    }

    [Fact]
    public void Verify_StaleTimestamp_ReturnsOutOfTolerance_UnlessDisabled()
    {
        var secret = "plain-secret";
        var payload = "{}";
        var id = "msg_1";
        var stale = DateTimeOffset.UtcNow.AddHours(-1).ToUnixTimeSeconds().ToString();
        var signature = "v1," + Sign(id, stale, payload, Encoding.UTF8.GetBytes(secret));

        var strict = WebhookSignature.Verify(payload, id, stale, signature, secret);
        Assert.False(strict.Valid);
        Assert.Equal("timestamp_out_of_tolerance", strict.Reason);

        var relaxed = WebhookSignature.Verify(payload, id, stale, signature, secret, toleranceSeconds: 0);
        Assert.True(relaxed.Valid);
    }

    [Fact]
    public void Verify_NonNumericTimestamp_ReturnsInvalidTimestamp()
    {
        var result = WebhookSignature.Verify("{}", "msg_1", "not-a-number", "v1,abc", "secret");

        Assert.False(result.Valid);
        Assert.Equal("invalid_timestamp", result.Reason);
    }
}
