using System.Text.Json;
using System.Text.Json.Serialization;

namespace Mailblastr;

/// <summary>A template variable definition as returned by the API.</summary>
public class TemplateVariable
{
    [JsonPropertyName("id")]
    public string Id { get; set; } = null!;

    [JsonPropertyName("key")]
    public string Key { get; set; } = null!;

    /// <summary><c>string</c> | <c>number</c>.</summary>
    [JsonPropertyName("type")]
    public string Type { get; set; } = null!;

    /// <summary>String, number or null.</summary>
    [JsonPropertyName("fallback_value")]
    public JsonElement? FallbackValue { get; set; }

    [JsonPropertyName("created_at")]
    public string? CreatedAt { get; set; }

    [JsonPropertyName("updated_at")]
    public string? UpdatedAt { get; set; }
}

/// <summary>A template variable definition accepted on create/update.</summary>
public class TemplateVariableInput
{
    [JsonPropertyName("key")]
    public string Key { get; set; } = null!;

    /// <summary><c>string</c> | <c>number</c>.</summary>
    [JsonPropertyName("type")]
    public string? Type { get; set; }

    /// <summary>String or number.</summary>
    [JsonPropertyName("fallback_value")]
    public object? FallbackValue { get; set; }
}

/// <summary>A saved email template.</summary>
public class Template
{
    [JsonPropertyName("object")]
    public string Object { get; set; } = "template";

    [JsonPropertyName("id")]
    public string Id { get; set; } = null!;

    [JsonPropertyName("name")]
    public string Name { get; set; } = null!;

    [JsonPropertyName("subject")]
    public string? Subject { get; set; }

    [JsonPropertyName("from")]
    public string? From { get; set; }

    [JsonPropertyName("reply_to")]
    public string? ReplyTo { get; set; }

    [JsonPropertyName("html")]
    public string? HtmlBody { get; set; }

    [JsonPropertyName("text")]
    public string? TextBody { get; set; }

    [JsonPropertyName("status")]
    public string? Status { get; set; }

    /// <summary>Stable handle usable anywhere an id is accepted.</summary>
    [JsonPropertyName("alias")]
    public string? Alias { get; set; }

    /// <summary>When the template was last published (null while only a draft exists).</summary>
    [JsonPropertyName("published_at")]
    public string? PublishedAt { get; set; }

    /// <summary>True when the draft has edits not yet published (retrieve only).</summary>
    [JsonPropertyName("has_unpublished_versions")]
    public bool? HasUnpublishedVersions { get; set; }

    [JsonPropertyName("current_version_id")]
    public string? CurrentVersionId { get; set; }

    [JsonPropertyName("variables")]
    public List<TemplateVariable>? Variables { get; set; }

    [JsonPropertyName("created_at")]
    public string CreatedAt { get; set; } = null!;

    [JsonPropertyName("updated_at")]
    public string? UpdatedAt { get; set; }
}

/// <summary>Payload for creating a template (POST /templates).</summary>
public class TemplateCreateOptions
{
    [JsonPropertyName("name")]
    public string Name { get; set; } = null!;

    /// <summary>Optional stable handle for sending by alias.</summary>
    [JsonPropertyName("alias")]
    public string? Alias { get; set; }

    [JsonPropertyName("subject")]
    public string? Subject { get; set; }

    [JsonPropertyName("from")]
    public string? From { get; set; }

    [JsonPropertyName("reply_to")]
    public EmailAddressList? ReplyTo { get; set; }

    [JsonPropertyName("html")]
    public string? HtmlBody { get; set; }

    [JsonPropertyName("text")]
    public string? TextBody { get; set; }

    [JsonPropertyName("variables")]
    public List<TemplateVariableInput>? Variables { get; set; }
}

/// <summary>Payload for updating a template (PATCH /templates/:id).</summary>
public class TemplateUpdateOptions
{
    [JsonPropertyName("name")]
    public string? Name { get; set; }

    /// <summary>Optional stable handle for sending by alias.</summary>
    [JsonPropertyName("alias")]
    public string? Alias { get; set; }

    [JsonPropertyName("subject")]
    public string? Subject { get; set; }

    [JsonPropertyName("from")]
    public string? From { get; set; }

    [JsonPropertyName("reply_to")]
    public EmailAddressList? ReplyTo { get; set; }

    [JsonPropertyName("html")]
    public string? HtmlBody { get; set; }

    [JsonPropertyName("text")]
    public string? TextBody { get; set; }

    [JsonPropertyName("variables")]
    public List<TemplateVariableInput>? Variables { get; set; }
}

/// <summary>Payload for TemplateDuplicateAsync (POST /templates/:id/duplicate).</summary>
public class TemplateDuplicateOptions
{
    [JsonPropertyName("name")]
    public string? Name { get; set; }

    [JsonPropertyName("alias")]
    public string? Alias { get; set; }
}
