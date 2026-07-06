using System.Text.Json.Serialization;

namespace Mailblastr;

/// <summary>A single custom-property predicate. <see cref="Value"/> is required for eq/contains.</summary>
public class PropertyFilter
{
    [JsonPropertyName("key")]
    public string Key { get; set; } = null!;

    /// <summary><c>eq</c> | <c>contains</c> | <c>exists</c>.</summary>
    [JsonPropertyName("operator")]
    public string Operator { get; set; } = null!;

    /// <summary>String, number or null.</summary>
    [JsonPropertyName("value")]
    public object? Value { get; set; }
}

/// <summary>A segment's filter: subscription status + email substring + property predicates.</summary>
public class SegmentFilter
{
    /// <summary><c>all</c> | <c>subscribed</c> | <c>unsubscribed</c>.</summary>
    [JsonPropertyName("status")]
    public string? Status { get; set; }

    [JsonPropertyName("email_contains")]
    public string? EmailContains { get; set; }

    [JsonPropertyName("property_filters")]
    public List<PropertyFilter>? PropertyFilters { get; set; }
}

/// <summary>A saved segment of a domain's contact pool.</summary>
public class Segment
{
    [JsonPropertyName("object")]
    public string Object { get; set; } = "segment";

    [JsonPropertyName("id")]
    public string Id { get; set; } = null!;

    [JsonPropertyName("audience_id")]
    public string AudienceId { get; set; } = null!;

    [JsonPropertyName("name")]
    public string Name { get; set; } = null!;

    [JsonPropertyName("filter")]
    public SegmentFilter Filter { get; set; } = new();

    [JsonPropertyName("created_at")]
    public string CreatedAt { get; set; } = null!;

    [JsonPropertyName("updated_at")]
    public string? UpdatedAt { get; set; }
}

/// <summary>
/// Payload for creating a segment (POST /segments). <see cref="Domain"/> is
/// REQUIRED: segment names are unique WITHIN a domain but freely reusable
/// across domains; every domain also carries an auto-created "General"
/// (all contacts) segment.
/// </summary>
public class SegmentCreateOptions
{
    /// <summary>REQUIRED. The sending domain this segment belongs to (e.g. <c>yourdomain.com</c>).</summary>
    [JsonPropertyName("domain")]
    public string Domain { get; set; } = null!;

    [JsonPropertyName("name")]
    public string Name { get; set; } = null!;

    [JsonPropertyName("filter")]
    public SegmentFilter? Filter { get; set; }
}

/// <summary>Payload for updating a segment (PATCH /segments/:id).</summary>
public class SegmentUpdateOptions
{
    [JsonPropertyName("name")]
    public string? Name { get; set; }

    [JsonPropertyName("filter")]
    public SegmentFilter? Filter { get; set; }
}
