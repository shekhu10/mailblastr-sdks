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
}

/// <summary>Envelope for endpoints that return a bare <c>{ data: [...] }</c> (e.g. POST /emails/batch).</summary>
internal sealed class DataEnvelope<T>
{
    [JsonPropertyName("data")]
    public List<T> Data { get; set; } = new();
}
