using System.Text.Json;
using System.Text.Json.Serialization;

namespace Mailblastr;

/// <summary>A step of an automation as returned by the API.</summary>
public class AutomationStep
{
    /// <summary>Absent on the synthesized trigger step.</summary>
    [JsonPropertyName("id")]
    public string? Id { get; set; }

    [JsonPropertyName("key")]
    public string Key { get; set; } = null!;

    [JsonPropertyName("type")]
    public string Type { get; set; } = null!;

    [JsonPropertyName("position")]
    public int? Position { get; set; }

    [JsonPropertyName("config")]
    public Dictionary<string, JsonElement>? Config { get; set; }
}

/// <summary>A typed edge between two step keys.</summary>
public class AutomationConnection
{
    [JsonPropertyName("from")]
    public string From { get; set; } = null!;

    [JsonPropertyName("to")]
    public string To { get; set; } = null!;

    [JsonPropertyName("type")]
    public string? Type { get; set; }
}

/// <summary>An inline step supplied on automation create.</summary>
public class AutomationStepInput
{
    /// <summary>Optional key referenced by connections.</summary>
    [JsonPropertyName("key")]
    public string? Key { get; set; }

    [JsonPropertyName("type")]
    public string Type { get; set; } = null!;

    [JsonPropertyName("config")]
    public Dictionary<string, object?>? Config { get; set; }
}

/// <summary>Enrollment counts (retrieve only).</summary>
public class AutomationEnrollments
{
    [JsonPropertyName("active")]
    public int Active { get; set; }

    [JsonPropertyName("completed")]
    public int Completed { get; set; }
}

/// <summary>An automation (multi-step workflow triggered by an event).</summary>
public class Automation
{
    [JsonPropertyName("object")]
    public string Object { get; set; } = "automation";

    [JsonPropertyName("id")]
    public string Id { get; set; } = null!;

    [JsonPropertyName("audience_id")]
    public string AudienceId { get; set; } = null!;

    [JsonPropertyName("name")]
    public string Name { get; set; } = null!;

    [JsonPropertyName("trigger")]
    public string? Trigger { get; set; }

    /// <summary>
    /// The sending domain this automation belongs to. Only EventSendAsync calls
    /// with the same domain trigger it. Null on pre-domain rows (treated as the
    /// account's single domain when exactly one exists).
    /// </summary>
    [JsonPropertyName("domain")]
    public string? Domain { get; set; }

    /// <summary><c>enabled</c> | <c>disabled</c>.</summary>
    [JsonPropertyName("status")]
    public string Status { get; set; } = null!;

    [JsonPropertyName("steps")]
    public List<AutomationStep>? Steps { get; set; }

    [JsonPropertyName("connections")]
    public List<AutomationConnection>? Connections { get; set; }

    /// <summary>Enrollment counts — included only on retrieve (GET /automations/:id).</summary>
    [JsonPropertyName("enrollments")]
    public AutomationEnrollments? Enrollments { get; set; }

    [JsonPropertyName("created_at")]
    public string CreatedAt { get; set; } = null!;

    [JsonPropertyName("updated_at")]
    public string? UpdatedAt { get; set; }
}

/// <summary>
/// Config for the <c>mailblastr:schedule</c> trigger: the automation fires ONCE
/// at <see cref="At"/>, enrolling every contact of its domain's pool. Required
/// with that trigger; not accepted on any other.
/// </summary>
public class AutomationTriggerConfig
{
    /// <summary>ISO 8601 instant the automation fires (future, at most 366 days ahead).</summary>
    [JsonPropertyName("at")]
    public string At { get; set; } = null!;

    /// <summary>IANA timezone the schedule was picked in (e.g. <c>America/New_York</c>).</summary>
    [JsonPropertyName("timezone")]
    public string Timezone { get; set; } = null!;
}

/// <summary>Payload for creating an automation (POST /automations). <see cref="Domain"/> is REQUIRED.</summary>
public class AutomationCreateOptions
{
    [JsonPropertyName("name")]
    public string Name { get; set; } = null!;

    /// <summary>
    /// REQUIRED. The sending domain this automation belongs to (e.g.
    /// <c>yourdomain.com</c> — one of your domains). Only EventSendAsync calls
    /// with the same domain trigger it.
    /// </summary>
    [JsonPropertyName("domain")]
    public string Domain { get; set; } = null!;

    /// <summary>
    /// The event that starts a run: <c>contact.created</c>, the built-in
    /// scheduled trigger <c>mailblastr:schedule</c> (requires
    /// <see cref="TriggerConfig"/>), an engagement event (<c>email.opened</c>,
    /// <c>email.clicked</c>, <c>email.replied</c>, <c>email.bounced</c>,
    /// <c>email.delivered</c>), or any custom event name you send via
    /// EventSendAsync. Usually supplied as a Steps[0] trigger step instead.
    /// </summary>
    [JsonPropertyName("trigger")]
    public string? Trigger { get; set; }

    /// <summary>
    /// Schedule for the <c>mailblastr:schedule</c> trigger (<c>{at, timezone}</c>).
    /// Required with that trigger; not accepted on any other.
    /// </summary>
    [JsonPropertyName("trigger_config")]
    public AutomationTriggerConfig? TriggerConfig { get; set; }

    /// <summary>Initial status: <c>enabled</c> | <c>disabled</c> (default <c>disabled</c>).</summary>
    [JsonPropertyName("status")]
    public string? Status { get; set; }

    /// <summary>Optional inline step graph; each step may carry a Key for connections.</summary>
    [JsonPropertyName("steps")]
    public List<AutomationStepInput>? Steps { get; set; }

    /// <summary>Optional typed edges between step keys.</summary>
    [JsonPropertyName("connections")]
    public List<AutomationConnection>? Connections { get; set; }
}

/// <summary>Payload for updating an automation (PATCH /automations/:id).</summary>
public class AutomationUpdateOptions
{
    [JsonPropertyName("name")]
    public string? Name { get; set; }

    /// <summary><c>enabled</c> | <c>disabled</c>.</summary>
    [JsonPropertyName("status")]
    public string? Status { get; set; }

    /// <summary>Re-point the automation at another of your domains (disabled automations only).</summary>
    [JsonPropertyName("domain")]
    public string? Domain { get; set; }

    /// <summary>
    /// Update the <c>mailblastr:schedule</c> trigger's schedule
    /// (<c>{at, timezone}</c>). Only valid on automations with that trigger.
    /// </summary>
    [JsonPropertyName("trigger_config")]
    public AutomationTriggerConfig? TriggerConfig { get; set; }

    [JsonPropertyName("connections")]
    public List<AutomationConnection>? Connections { get; set; }
}

/// <summary>Payload for appending a step (POST /automations/:id/steps).</summary>
public class AutomationAddStepOptions
{
    [JsonPropertyName("type")]
    public string Type { get; set; } = null!;

    [JsonPropertyName("config")]
    public Dictionary<string, object?>? Config { get; set; }

    [JsonPropertyName("key")]
    public string? Key { get; set; }
}

/// <summary>Acknowledgement of deleting an automation step: <c>{ id, deleted }</c>.</summary>
public class AutomationStepDeleted
{
    [JsonPropertyName("id")]
    public string Id { get; set; } = null!;

    [JsonPropertyName("deleted")]
    public bool Deleted { get; set; }
}

/// <summary>One executed step within an automation run.</summary>
public class AutomationRunStep
{
    [JsonPropertyName("key")]
    public string Key { get; set; } = null!;

    [JsonPropertyName("type")]
    public string Type { get; set; } = null!;

    /// <summary><c>completed</c> | <c>failed</c> | <c>skipped</c>.</summary>
    [JsonPropertyName("status")]
    public string Status { get; set; } = null!;

    [JsonPropertyName("started_at")]
    public string? StartedAt { get; set; }

    [JsonPropertyName("completed_at")]
    public string? CompletedAt { get; set; }

    [JsonPropertyName("output")]
    public Dictionary<string, JsonElement>? Output { get; set; }

    [JsonPropertyName("error")]
    public string? Error { get; set; }
}

/// <summary>One enrollment (run) of a contact through an automation.</summary>
public class AutomationRun
{
    [JsonPropertyName("object")]
    public string Object { get; set; } = "automation_run";

    [JsonPropertyName("id")]
    public string Id { get; set; } = null!;

    [JsonPropertyName("contact_id")]
    public string ContactId { get; set; } = null!;

    /// <summary>Email of the contact the run is for; null if that contact was deleted.</summary>
    [JsonPropertyName("contact_email")]
    public string? ContactEmail { get; set; }

    /// <summary><c>running</c> | <c>completed</c> | <c>failed</c> | <c>cancelled</c> | <c>skipped</c>.</summary>
    [JsonPropertyName("status")]
    public string Status { get; set; } = null!;

    [JsonPropertyName("started_at")]
    public string? StartedAt { get; set; }

    [JsonPropertyName("completed_at")]
    public string? CompletedAt { get; set; }

    [JsonPropertyName("created_at")]
    public string CreatedAt { get; set; } = null!;

    /// <summary>Present on retrieve (GET /automations/:id/runs/:runId) only.</summary>
    [JsonPropertyName("automation_id")]
    public string? AutomationId { get; set; }

    [JsonPropertyName("steps")]
    public List<AutomationRunStep>? Steps { get; set; }

    [JsonPropertyName("error")]
    public string? Error { get; set; }
}
