using System.Text.Json;
using System.Text.Json.Serialization;

namespace Mailblastr;

/// <summary>Shared serializer settings for every API request/response.</summary>
internal static class MailblastrJson
{
    public static readonly JsonSerializerOptions Options = new(JsonSerializerDefaults.Web)
    {
        // All model properties carry explicit [JsonPropertyName] attributes;
        // snake_case is the fallback policy so any unannotated member still
        // matches the wire format.
        PropertyNamingPolicy = JsonNamingPolicy.SnakeCaseLower,
        DefaultIgnoreCondition = JsonIgnoreCondition.WhenWritingNull,
    };
}

/// <summary>The API error body shape: <c>{ statusCode, name, message }</c>.</summary>
internal sealed class ApiErrorBody
{
    [JsonPropertyName("statusCode")]
    public int? StatusCode { get; set; }

    [JsonPropertyName("name")]
    public string? Name { get; set; }

    [JsonPropertyName("message")]
    public string? Message { get; set; }

    /// <summary>
    /// The additive fields some errors carry on top of the envelope —
    /// <c>limit</c>, <c>reputation</c>, <c>sent</c>, <c>sent_count</c>. Captured
    /// unparsed so a shape this SDK version does not know still reaches the
    /// caller via <see cref="MailblastrException.Extra"/>.
    /// </summary>
    [JsonExtensionData]
    public Dictionary<string, JsonElement>? Extra { get; set; }
}
