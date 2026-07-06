namespace Mailblastr;

public partial interface IMailblastr
{
    // ---- Campaigns (domain-first) ----

    /// <summary>Create a campaign. <c>options.Domain</c> is REQUIRED (names the contact pool it targets). POST /campaigns</summary>
    Task<IdResponse> CampaignCreateAsync(CampaignCreateOptions options, CancellationToken cancellationToken = default);

    /// <summary>Retrieve a campaign (includes statistics/followups/list_address). GET /campaigns/:id</summary>
    Task<Campaign> CampaignRetrieveAsync(string campaignId, CancellationToken cancellationToken = default);

    /// <summary>List campaigns. GET /campaigns</summary>
    Task<ListResponse<Campaign>> CampaignListAsync(PaginationOptions? pagination = null, CancellationToken cancellationToken = default);

    /// <summary>Update a draft campaign. PATCH /campaigns/:id</summary>
    Task<IdResponse> CampaignUpdateAsync(string campaignId, CampaignUpdateOptions options, CancellationToken cancellationToken = default);

    /// <summary>Send a campaign now, or schedule it with <paramref name="scheduledAt"/>. POST /campaigns/:id/send</summary>
    /// <param name="campaignId">Id of the campaign to send.</param>
    /// <param name="scheduledAt">ISO 8601 (or natural-language) schedule; null sends immediately.</param>
    /// <param name="cancellationToken">Token to cancel the request.</param>
    Task<IdResponse> CampaignSendAsync(string campaignId, string? scheduledAt = null, CancellationToken cancellationToken = default);

    /// <summary>Cancel a scheduled campaign (returns it to draft). POST /campaigns/:id/cancel</summary>
    Task<Campaign> CampaignCancelAsync(string campaignId, CancellationToken cancellationToken = default);

    /// <summary>Per-campaign analytics (counts, engagement rates, top links). GET /campaigns/:id/stats</summary>
    Task<CampaignStats> CampaignRetrieveStatsAsync(string campaignId, CancellationToken cancellationToken = default);

    /// <summary>A/B winner evaluation for an A/B campaign. GET /campaigns/:id/ab</summary>
    Task<CampaignAbResult> CampaignRetrieveAbResultAsync(string campaignId, CancellationToken cancellationToken = default);

    /// <summary>Delete a campaign. DELETE /campaigns/:id</summary>
    Task<RemovedResponse> CampaignDeleteAsync(string campaignId, CancellationToken cancellationToken = default);

    // ---- Segments (domain-first) ----

    /// <summary>
    /// Create a segment on a sending domain. <c>options.Domain</c> is REQUIRED;
    /// segment names are unique within a domain. POST /segments
    /// </summary>
    Task<Segment> SegmentCreateAsync(SegmentCreateOptions options, CancellationToken cancellationToken = default);

    /// <summary>Retrieve a segment. GET /segments/:id</summary>
    Task<Segment> SegmentRetrieveAsync(string segmentId, CancellationToken cancellationToken = default);

    /// <summary>
    /// List a domain's segments (<paramref name="domain"/> is REQUIRED; includes
    /// its auto-created "General" segment). GET /segments?domain=…
    /// </summary>
    Task<ListResponse<Segment>> SegmentListAsync(string domain, PaginationOptions? pagination = null, CancellationToken cancellationToken = default);

    /// <summary>Preview the contacts a segment currently resolves to. GET /segments/:id/contacts</summary>
    Task<ListResponse<Contact>> SegmentListContactsAsync(string segmentId, CancellationToken cancellationToken = default);

    /// <summary>Update a segment. PATCH /segments/:id</summary>
    Task<Segment> SegmentUpdateAsync(string segmentId, SegmentUpdateOptions options, CancellationToken cancellationToken = default);

    /// <summary>Delete a segment. DELETE /segments/:id</summary>
    Task<RemovedResponse> SegmentDeleteAsync(string segmentId, CancellationToken cancellationToken = default);

    // ---- Topics (domain-first) ----

    /// <summary>Create a topic on a sending domain. <c>options.Domain</c> is REQUIRED. POST /topics</summary>
    Task<Topic> TopicCreateAsync(TopicCreateOptions options, CancellationToken cancellationToken = default);

    /// <summary>Retrieve a topic. GET /topics/:id</summary>
    Task<Topic> TopicRetrieveAsync(string topicId, CancellationToken cancellationToken = default);

    /// <summary>List a domain's topics (<paramref name="domain"/> is REQUIRED). GET /topics?domain=…</summary>
    Task<ListResponse<Topic>> TopicListAsync(string domain, PaginationOptions? pagination = null, CancellationToken cancellationToken = default);

    /// <summary>Update a topic. PATCH /topics/:id</summary>
    Task<Topic> TopicUpdateAsync(string topicId, TopicUpdateOptions options, CancellationToken cancellationToken = default);

    /// <summary>Delete a topic. DELETE /topics/:id</summary>
    Task<RemovedResponse> TopicDeleteAsync(string topicId, CancellationToken cancellationToken = default);

    // ---- Templates ----

    /// <summary>Create a template. POST /templates</summary>
    Task<ObjectRef> TemplateCreateAsync(TemplateCreateOptions options, CancellationToken cancellationToken = default);

    /// <summary>Retrieve a template. GET /templates/:id</summary>
    Task<Template> TemplateRetrieveAsync(string templateId, CancellationToken cancellationToken = default);

    /// <summary>List templates. GET /templates</summary>
    Task<ListResponse<Template>> TemplateListAsync(PaginationOptions? pagination = null, CancellationToken cancellationToken = default);

    /// <summary>Update a template (creates a new draft version). PATCH /templates/:id</summary>
    Task<ObjectRef> TemplateUpdateAsync(string templateId, TemplateUpdateOptions options, CancellationToken cancellationToken = default);

    /// <summary>Duplicate a template. POST /templates/:id/duplicate</summary>
    Task<ObjectRef> TemplateDuplicateAsync(string templateId, TemplateDuplicateOptions? options = null, CancellationToken cancellationToken = default);

    /// <summary>Publish a template (make its latest draft live). POST /templates/:id/publish</summary>
    Task<ObjectRef> TemplatePublishAsync(string templateId, CancellationToken cancellationToken = default);

    /// <summary>Delete a template. DELETE /templates/:id</summary>
    Task<RemovedResponse> TemplateDeleteAsync(string templateId, CancellationToken cancellationToken = default);
}
