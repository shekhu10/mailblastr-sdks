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
