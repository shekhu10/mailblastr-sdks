using System.Security.Cryptography;
using System.Text;

namespace Mailblastr;

/// <summary>
/// Verifies MailBlastr webhook delivery signatures (Svix-style scheme:
/// base64 HMAC-SHA256 over <c>{svix-id}.{svix-timestamp}.{raw body}</c>,
/// tagged <c>v1,</c> in the <c>svix-signature</c> header).
/// Pure local computation — makes no HTTP request.
/// </summary>
public static class WebhookSignature
{
    /// <summary>
    /// Verify a delivery given its raw body and headers.
    /// <paramref name="payload"/> MUST be the exact raw request body string the
    /// server sent — do not re-serialize the parsed JSON, whitespace differences
    /// break the signature. <paramref name="headers"/> is read case-insensitively
    /// for <c>svix-id</c> / <c>svix-timestamp</c> / <c>svix-signature</c>.
    /// </summary>
    /// <param name="payload">The exact raw request body string as received.</param>
    /// <param name="headers">The delivery's headers; svix-id / svix-timestamp / svix-signature are read case-insensitively.</param>
    /// <param name="secret">Your endpoint's signing secret (typically <c>whsec_…</c>).</param>
    /// <param name="toleranceSeconds">Max allowed clock skew in seconds (default 300). Pass 0 to skip the check.</param>
    public static VerifyWebhookResult Verify(string payload, IReadOnlyDictionary<string, string> headers, string secret, int toleranceSeconds = 300)
    {
        ArgumentNullException.ThrowIfNull(headers);
        return Verify(
            payload,
            ReadHeader(headers, "svix-id"),
            ReadHeader(headers, "svix-timestamp"),
            ReadHeader(headers, "svix-signature"),
            secret,
            toleranceSeconds);
    }

    /// <summary>Verify a delivery given the three svix header values directly.</summary>
    public static VerifyWebhookResult Verify(string payload, string? svixId, string? svixTimestamp, string? svixSignature, string secret, int toleranceSeconds = 300)
    {
        ArgumentNullException.ThrowIfNull(payload);

        if (string.IsNullOrEmpty(svixId) || string.IsNullOrEmpty(svixTimestamp) || string.IsNullOrEmpty(svixSignature))
        {
            return Invalid("missing_headers");
        }
        if (string.IsNullOrEmpty(secret))
        {
            return Invalid("missing_secret");
        }

        // Optional timestamp freshness check (default 5-minute tolerance; 0 disables).
        if (toleranceSeconds > 0)
        {
            if (!long.TryParse(svixTimestamp, out var timestamp))
            {
                return Invalid("invalid_timestamp");
            }
            var skew = Math.Abs(DateTimeOffset.UtcNow.ToUnixTimeSeconds() - timestamp);
            if (skew > toleranceSeconds)
            {
                return Invalid("timestamp_out_of_tolerance");
            }
        }

        var signedContent = $"{svixId}.{svixTimestamp}.{payload}";
        string expected;
        using (var hmac = new HMACSHA256(SecretToKey(secret)))
        {
            expected = Convert.ToBase64String(hmac.ComputeHash(Encoding.UTF8.GetBytes(signedContent)));
        }

        // The header may contain multiple space-separated `v1,<sig>` entries; any match wins.
        foreach (var part in svixSignature.Split(' ', StringSplitOptions.RemoveEmptyEntries | StringSplitOptions.TrimEntries))
        {
            var candidate = part.StartsWith("v1,", StringComparison.Ordinal) ? part["v1,".Length..] : part;
            if (FixedTimeEquals(candidate, expected))
            {
                return new VerifyWebhookResult { Valid = true };
            }
        }
        return Invalid("no_match");
    }

    /// <summary>
    /// Derive the HMAC key from a <c>whsec_</c>-prefixed secret (base64-decode
    /// the suffix); a secret without the prefix is used as raw UTF-8 bytes.
    /// Mirrors the backend's key derivation.
    /// </summary>
    private static byte[] SecretToKey(string secret)
    {
        const string prefix = "whsec_";
        if (secret.StartsWith(prefix, StringComparison.Ordinal))
        {
            try
            {
                var decoded = Convert.FromBase64String(secret[prefix.Length..]);
                if (decoded.Length > 0)
                {
                    return decoded;
                }
            }
            catch (FormatException)
            {
                // Fall through to raw bytes.
            }
        }
        return Encoding.UTF8.GetBytes(secret);
    }

    /// <summary>Constant-time compare of two base64 signature strings.</summary>
    private static bool FixedTimeEquals(string a, string b)
        => CryptographicOperations.FixedTimeEquals(Encoding.UTF8.GetBytes(a), Encoding.UTF8.GetBytes(b));

    /// <summary>Case-insensitively read a single header value.</summary>
    private static string? ReadHeader(IReadOnlyDictionary<string, string> headers, string name)
    {
        if (headers.TryGetValue(name, out var direct))
        {
            return direct;
        }
        foreach (var pair in headers)
        {
            if (string.Equals(pair.Key, name, StringComparison.OrdinalIgnoreCase))
            {
                return pair.Value;
            }
        }
        return null;
    }

    private static VerifyWebhookResult Invalid(string reason) => new() { Valid = false, Reason = reason };
}
