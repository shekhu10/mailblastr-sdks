import { createHmac, timingSafeEqual } from 'node:crypto';
import { HttpClient, type ClientConfig, DEFAULT_BASE_URL, VERSION, USER_AGENT } from './client';
import type {
  Result, RequestOptions, PaginationParams, ObjectRef,
  SendEmailOptions, CreateEmailResponse, Email, SentEmailListItem,
  AttachmentMeta, ReceivedAttachment, ReceivedEmail, ForwardReceivedEmailOptions,
  CreateDomainOptions, UpdateDomainOptions, Domain,
  DomainClaim, ClaimDomainOptions,
  Audience, Contact, CreateContactOptions, UpdateContactOptions,
  ContactInput, ImportContactsResponse, ListContactsParams,
  ContactTopics, UpdateContactTopicsOptions,
  ContactProperty, CreateContactPropertyOptions, UpdateContactPropertyOptions,
  Poll, PollResult,
  Campaign, CreateCampaignOptions, UpdateCampaignOptions, CampaignStats, CampaignAbResult,
  Segment, CreateSegmentOptions, UpdateSegmentOptions,
  Topic, CreateTopicOptions, UpdateTopicOptions,
  Template, CreateTemplateOptions, UpdateTemplateOptions, DuplicateTemplateOptions,
  Automation, CreateAutomationOptions, UpdateAutomationOptions,
  AddAutomationStepOptions, AutomationStep, AutomationRun,
  Webhook, CreateWebhookOptions, UpdateWebhookOptions,
  WebhookHeaders, VerifyWebhookResult, VerifyWebhookOptions,
  LogEntry, SendEventOptions, SendEventResponse, CreateEventOptions, EventDefinition,
  ApiKey, CreateApiKeyOptions, CreateApiKeyResponse,
  ListResponse, RemovedResponse,
} from './types';

export * from './types';
export { DEFAULT_BASE_URL, VERSION, USER_AGENT };

/** Build a `?limit=&after=&before=` query string from pagination params. */
function paginate(params?: PaginationParams): string {
  if (!params) return '';
  const q = new URLSearchParams();
  if (params.limit != null) q.set('limit', String(params.limit));
  if (params.after != null) q.set('after', params.after);
  if (params.before != null) q.set('before', params.before);
  const qs = q.toString();
  return qs ? `?${qs}` : '';
}

/** Inbound (received) email — accessed as `mb.emails.receiving`. */
class ReceivingEmails {
  constructor(private readonly http: HttpClient) {}
  /** List received emails. GET /emails/receiving */
  list(params?: PaginationParams): Promise<Result<ListResponse<ReceivedEmail>>> {
    return this.http.request('GET', `/emails/receiving${paginate(params)}`);
  }
  /** Retrieve a received email. GET /emails/receiving/:id */
  get(id: string): Promise<Result<ReceivedEmail>> {
    return this.http.request('GET', `/emails/receiving/${id}`);
  }
  /** List a received email's attachments. GET /emails/receiving/:id/attachments */
  listAttachments(id: string): Promise<Result<ListResponse<ReceivedAttachment>>> {
    return this.http.request('GET', `/emails/receiving/${id}/attachments`);
  }
  /**
   * Download one attachment of a received email as raw bytes.
   * GET /emails/receiving/:id/attachments/:attachmentId — this route streams the
   * binary file (not JSON), so the result `data` is an ArrayBuffer.
   */
  getAttachment(id: string, attachmentId: string): Promise<Result<ArrayBuffer>> {
    return this.http.requestRaw('GET', `/emails/receiving/${id}/attachments/${attachmentId}`);
  }
  /**
   * Download the original RFC822/MIME message as raw bytes.
   * GET /emails/receiving/:id/raw — streams `message/rfc822`, so `data` is an
   * ArrayBuffer.
   */
  getRaw(id: string): Promise<Result<ArrayBuffer>> {
    return this.http.requestRaw('GET', `/emails/receiving/${id}/raw`);
  }
  /** Forward a received email. POST /emails/receiving/:id/forward */
  forward(id: string, payload: ForwardReceivedEmailOptions): Promise<Result<CreateEmailResponse>> {
    return this.http.request('POST', `/emails/receiving/${id}/forward`, payload);
  }
  /**
   * Reply to a received email's sender, threaded into the same conversation
   * (In-Reply-To the received message; subject defaults to `Re: …`).
   * POST /emails/receiving/:id/reply
   */
  reply(id: string, payload: { from: string; html?: string; text?: string; subject?: string }): Promise<Result<CreateEmailResponse>> {
    return this.http.request('POST', `/emails/receiving/${id}/reply`, payload);
  }
  /** Delete a received email. DELETE /emails/receiving/:id */
  remove(id: string): Promise<Result<RemovedResponse>> {
    return this.http.request('DELETE', `/emails/receiving/${id}`);
  }
}

class Emails {
  /** Inbound email sub-resource. */
  readonly receiving: ReceivingEmails;
  constructor(private readonly http: HttpClient) {
    this.receiving = new ReceivingEmails(http);
  }
  /** Send a single email. POST /emails */
  send(payload: SendEmailOptions, options?: RequestOptions): Promise<Result<CreateEmailResponse>> {
    return this.http.request('POST', '/emails', payload, options);
  }
  /** Send up to 100 emails in one request. POST /emails/batch (alias of `mb.batch.send`). */
  batch(payloads: SendEmailOptions[], options?: RequestOptions): Promise<Result<{ data: CreateEmailResponse[] }>> {
    return this.http.request('POST', '/emails/batch', payloads, options);
  }
  /** List sent emails. GET /emails — returns trimmed {@link SentEmailListItem}s (no status/html/text/events). */
  list(params?: PaginationParams): Promise<Result<ListResponse<SentEmailListItem>>> {
    return this.http.request('GET', `/emails${paginate(params)}`);
  }
  /** Retrieve a sent email and its events. GET /emails/:id */
  get(id: string): Promise<Result<Email>> {
    return this.http.request('GET', `/emails/${id}`);
  }
  /** List a sent email's attachments. GET /emails/:id/attachments */
  listAttachments(id: string): Promise<Result<ListResponse<AttachmentMeta>>> {
    return this.http.request('GET', `/emails/${id}/attachments`);
  }
  /** Retrieve one attachment of a sent email. GET /emails/:id/attachments/:attachmentId */
  getAttachment(id: string, attachmentId: string): Promise<Result<AttachmentMeta>> {
    return this.http.request('GET', `/emails/${id}/attachments/${attachmentId}`);
  }
  /** Reschedule a scheduled email. PATCH /emails/:id */
  update(id: string, payload: { scheduled_at: string }): Promise<Result<{ id: string; object: 'email' }>> {
    return this.http.request('PATCH', `/emails/${id}`, payload);
  }
  /** Cancel a scheduled email. POST /emails/:id/cancel */
  cancel(id: string): Promise<Result<{ id: string; object: 'email' }>> {
    return this.http.request('POST', `/emails/${id}/cancel`);
  }
}

/** Batch send — `mb.batch.send([...])`. The `batch` resource. */
class Batch {
  constructor(private readonly http: HttpClient) {}
  /** Send up to 100 emails in one request. POST /emails/batch */
  send(payloads: SendEmailOptions[], options?: RequestOptions): Promise<Result<{ data: CreateEmailResponse[] }>> {
    return this.http.request('POST', '/emails/batch', payloads, options);
  }
}

class Domains {
  constructor(private readonly http: HttpClient) {}
  create(payload: CreateDomainOptions): Promise<Result<Domain>> {
    return this.http.request('POST', '/domains', payload);
  }
  get(id: string): Promise<Result<Domain>> {
    return this.http.request('GET', `/domains/${id}`);
  }
  list(params?: PaginationParams): Promise<Result<ListResponse<Domain>>> {
    return this.http.request('GET', `/domains${paginate(params)}`);
  }
  /** Returns the slim ack { object: 'domain', id }. */
  update(id: string, payload: UpdateDomainOptions): Promise<Result<ObjectRef<'domain'>>> {
    return this.http.request('PATCH', `/domains/${id}`, payload);
  }
  /** Returns the slim ack { object: 'domain', id }. */
  verify(id: string): Promise<Result<ObjectRef<'domain'>>> {
    return this.http.request('POST', `/domains/${id}/verify`);
  }
  /** Claim a domain already verified elsewhere. POST /domains/claim */
  claim(payload: ClaimDomainOptions): Promise<Result<DomainClaim>> {
    return this.http.request('POST', '/domains/claim', payload);
  }
  /** Retrieve a domain's claim record. GET /domains/:id/claim */
  getClaim(id: string): Promise<Result<DomainClaim>> {
    return this.http.request('GET', `/domains/${id}/claim`);
  }
  /** Verify a domain claim. POST /domains/:id/claim/verify */
  verifyClaim(id: string): Promise<Result<DomainClaim>> {
    return this.http.request('POST', `/domains/${id}/claim/verify`);
  }
  /**
   * Detect a domain's DNS provider and the one-click apply methods available
   * (Cloudflare token, GoDaddy key/secret, hosted Domain Connect, panel deep-link).
   * GET /domains/:id/dns/detect
   */
  detectDns(id: string): Promise<Result<{ provider: unknown; nameservers: string[]; methods: Array<Record<string, unknown>> }>> {
    return this.http.request('GET', `/domains/${id}/dns/detect`);
  }
  /** Apply this domain's DNS records via the Cloudflare API, then auto-verify. POST /domains/:id/dns/cloudflare */
  applyCloudflareDns(id: string, payload: { token: string }): Promise<Result<Record<string, unknown>>> {
    return this.http.request('POST', `/domains/${id}/dns/cloudflare`, payload);
  }
  /** Apply this domain's DNS records via the GoDaddy API, then auto-verify. POST /domains/:id/dns/godaddy */
  applyGoDaddyDns(id: string, payload: { key: string; secret: string }): Promise<Result<Record<string, unknown>>> {
    return this.http.request('POST', `/domains/${id}/dns/godaddy`, payload);
  }
  /**
   * Apply this domain's DNS records via the Namecheap API (existing records are
   * preserved), then auto-verify. Namecheap must have the calling server's IP
   * whitelisted in the account's API settings. POST /domains/:id/dns/namecheap
   */
  applyNamecheapDns(id: string, payload: { apiUser: string; apiKey: string; userName?: string }): Promise<Result<Record<string, unknown>>> {
    return this.http.request('POST', `/domains/${id}/dns/namecheap`, payload);
  }
  remove(id: string): Promise<Result<RemovedResponse>> {
    return this.http.request('DELETE', `/domains/${id}`);
  }
}

class Audiences {
  constructor(private readonly http: HttpClient) {}
  create(payload: { name: string }): Promise<Result<Audience>> {
    return this.http.request('POST', '/audiences', payload);
  }
  get(id: string): Promise<Result<Audience>> {
    return this.http.request('GET', `/audiences/${id}`);
  }
  list(params?: PaginationParams): Promise<Result<ListResponse<Audience>>> {
    return this.http.request('GET', `/audiences${paginate(params)}`);
  }
  /**
   * Import contacts from a link-shared Google Sheet. Header columns become
   * contact properties (usable as {{merge_tags}}); rows land in a fresh
   * segment. POST /audiences/:id/contacts/import-sheet
   */
  importSheet(audienceId: string, payload: { url: string; segment_name?: string }): Promise<Result<{
    object: 'sheet_import'; imported: number; updated: number; skipped: number; total: number;
    segment_id: string; segment_name: string; segment_added: number; variables: string[];
  }>> {
    return this.http.request('POST', `/audiences/${encodeURIComponent(audienceId)}/contacts/import-sheet`, payload);
  }
  /** Rename an audience. PATCH /audiences/:id */
  update(id: string, payload: { name: string }): Promise<Result<Audience>> {
    return this.http.request('PATCH', `/audiences/${id}`, payload);
  }
  remove(id: string): Promise<Result<RemovedResponse>> {
    return this.http.request('DELETE', `/audiences/${id}`);
  }
}

class Contacts {
  constructor(private readonly http: HttpClient) {}
  /**
   * Create a contact. Returns the slim ack { object: 'contact', id }.
   * Pass `audienceId` to target an audience, or OMIT it to use the flat
   * top-level `/contacts` API (creates in your default/oldest audience).
   */
  create(payload: CreateContactOptions): Promise<Result<ObjectRef<'contact'>>> {
    const { audienceId, ...body } = payload;
    return this.http.request('POST', audienceId ? `/audiences/${audienceId}/contacts` : '/contacts', body);
  }
  /** Retrieve a contact by id or email. Omit `audienceId` to resolve across all your audiences. */
  get(params: { audienceId?: string; id: string }): Promise<Result<Contact>> {
    const id = encodeURIComponent(params.id);
    return this.http.request('GET', params.audienceId ? `/audiences/${params.audienceId}/contacts/${id}` : `/contacts/${id}`);
  }
  list(params: ListContactsParams = {}): Promise<Result<ListResponse<Contact>>> {
    const q = new URLSearchParams();
    if (params.limit != null) q.set('limit', String(params.limit));
    if (params.after != null) q.set('after', params.after);
    if (params.before != null) q.set('before', params.before);
    // segment_id filter is honored by both the flat and audience-scoped list.
    if (params.segment_id != null) q.set('segment_id', params.segment_id);
    const qs = q.toString();
    const base = params.audienceId ? `/audiences/${params.audienceId}/contacts` : '/contacts';
    return this.http.request('GET', `${base}${qs ? `?${qs}` : ''}`);
  }
  /**
   * Bulk-import contacts from an array. Upserts by email; max 10,000 per call.
   * `on_conflict: 'skip'` leaves existing contacts untouched (default 'upsert').
   */
  batch(params: { audienceId: string; contacts: ContactInput[]; on_conflict?: 'upsert' | 'skip' }): Promise<Result<ImportContactsResponse>> {
    const qs = params.on_conflict ? `?on_conflict=${params.on_conflict}` : '';
    return this.http.request('POST', `/audiences/${params.audienceId}/contacts/batch${qs}`, { contacts: params.contacts });
  }
  /**
   * Bulk-import contacts from CSV text (header row optional). Upserts by email.
   * `on_conflict: 'skip'` leaves existing contacts untouched (default 'upsert').
   * By default every non-builtin CSV column is auto-registered as a custom property
   * (so `company`, `plan`, … survive and become `{{merge}}` tags) — same as the
   * dashboard and Google-Sheet import. Pass `create_properties: false` for strict
   * mode, where only already-registered columns are kept and the rest are returned
   * in `ignored_columns`.
   */
  import(params: { audienceId: string; csv: string; on_conflict?: 'upsert' | 'skip'; create_properties?: boolean }): Promise<Result<ImportContactsResponse>> {
    const q = new URLSearchParams();
    if (params.on_conflict) q.set('on_conflict', params.on_conflict);
    if (params.create_properties === false) q.set('create_properties', 'false');
    const qs = q.toString();
    return this.http.request('POST', `/audiences/${params.audienceId}/contacts/import${qs ? `?${qs}` : ''}`, { csv: params.csv });
  }
  /** Update a contact. Returns the slim ack { object: 'contact', id }. Omit `audienceId` for the flat API. */
  update(payload: UpdateContactOptions): Promise<Result<ObjectRef<'contact'>>> {
    const { audienceId, id, ...body } = payload;
    const eid = encodeURIComponent(id);
    return this.http.request('PATCH', audienceId ? `/audiences/${audienceId}/contacts/${eid}` : `/contacts/${eid}`, body);
  }
  /** Delete a contact. The id is returned under `contact`. Omit `audienceId` for the flat API. */
  remove(params: { audienceId?: string; id: string }): Promise<Result<{ object: 'contact'; contact: string; deleted: true }>> {
    const id = encodeURIComponent(params.id);
    return this.http.request('DELETE', params.audienceId ? `/audiences/${params.audienceId}/contacts/${id}` : `/contacts/${id}`);
  }
  /** Add a contact to a segment. POST /contacts/:id/segments/:segmentId → { id } (the segment id). */
  addToSegment(id: string, segmentId: string): Promise<Result<{ id: string }>> {
    return this.http.request('POST', `/contacts/${encodeURIComponent(id)}/segments/${segmentId}`);
  }
  /** Remove a contact from a segment. DELETE /contacts/:id/segments/:segmentId → { id, audienceId, deleted }. */
  removeFromSegment(id: string, segmentId: string): Promise<Result<{ id: string; audienceId: string; deleted: boolean }>> {
    return this.http.request('DELETE', `/contacts/${encodeURIComponent(id)}/segments/${segmentId}`);
  }
  /** List the segments a contact belongs to. GET /contacts/:id/segments */
  listSegments(id: string): Promise<Result<ListResponse<Segment>>> {
    return this.http.request('GET', `/contacts/${encodeURIComponent(id)}/segments`);
  }
  /** Get a contact's topic subscriptions. GET /contacts/:id/topics */
  getTopics(id: string): Promise<Result<ContactTopics>> {
    return this.http.request('GET', `/contacts/${encodeURIComponent(id)}/topics`);
  }
  /** Update a contact's topic subscriptions. PATCH /contacts/:id/topics → { id } (the contact id). */
  updateTopics(id: string, payload: UpdateContactTopicsOptions): Promise<Result<{ id: string }>> {
    return this.http.request('PATCH', `/contacts/${encodeURIComponent(id)}/topics`, payload);
  }
}

class ContactProperties {
  constructor(private readonly http: HttpClient) {}
  /** Returns the slim ack { object: 'contact_property', id }. */
  create(payload: CreateContactPropertyOptions): Promise<Result<ObjectRef<'contact_property'>>> {
    return this.http.request('POST', '/contact-properties', payload);
  }
  get(id: string): Promise<Result<ContactProperty>> {
    return this.http.request('GET', `/contact-properties/${id}`);
  }
  list(params?: PaginationParams): Promise<Result<ListResponse<ContactProperty>>> {
    return this.http.request('GET', `/contact-properties${paginate(params)}`);
  }
  /** Returns the slim ack { object: 'contact_property', id }. */
  update(id: string, payload: UpdateContactPropertyOptions): Promise<Result<ObjectRef<'contact_property'>>> {
    return this.http.request('PATCH', `/contact-properties/${id}`, payload);
  }
  remove(id: string): Promise<Result<RemovedResponse>> {
    return this.http.request('DELETE', `/contact-properties/${id}`);
  }
}

/** Read-only results of the in-email poll widget. `mb.polls`. */
class Polls {
  constructor(private readonly http: HttpClient) {}
  /** One summary row per email that has poll responses. GET /polls */
  list(params?: PaginationParams): Promise<Result<ListResponse<Poll>>> {
    return this.http.request('GET', `/polls${paginate(params)}`);
  }
  /** The aggregated answer breakdown for one email. GET /polls/:emailId */
  get(emailId: string): Promise<Result<PollResult>> {
    return this.http.request('GET', `/polls/${encodeURIComponent(emailId)}`);
  }
}

/** Campaigns — bulk sends to an audience or segment. Calls the /campaigns endpoints. */
class Campaigns {
  constructor(private readonly http: HttpClient) {}
  create(payload: CreateCampaignOptions): Promise<Result<{ id: string }>> {
    return this.http.request('POST', '/campaigns', payload);
  }
  get(id: string): Promise<Result<Campaign>> {
    return this.http.request('GET', `/campaigns/${id}`);
  }
  list(params?: PaginationParams): Promise<Result<ListResponse<Campaign>>> {
    return this.http.request('GET', `/campaigns${paginate(params)}`);
  }
  update(id: string, payload: UpdateCampaignOptions): Promise<Result<{ id: string }>> {
    return this.http.request('PATCH', `/campaigns/${id}`, payload);
  }
  /** Send now, or schedule with { scheduled_at }. */
  send(id: string, payload?: { scheduled_at?: string }): Promise<Result<{ id: string }>> {
    return this.http.request('POST', `/campaigns/${id}/send`, payload ?? {});
  }
  /** Cancel a scheduled campaign (returns it to draft). POST /campaigns/:id/cancel */
  cancel(id: string): Promise<Result<Campaign>> {
    return this.http.request('POST', `/campaigns/${id}/cancel`);
  }
  /** Per-campaign analytics (counts, engagement rates, top links). GET /campaigns/:id/stats */
  stats(id: string): Promise<Result<CampaignStats>> {
    return this.http.request('GET', `/campaigns/${id}/stats`);
  }
  /** A/B winner evaluation for an A/B campaign. GET /campaigns/:id/ab */
  ab(id: string): Promise<Result<CampaignAbResult>> {
    return this.http.request('GET', `/campaigns/${id}/ab`);
  }
  remove(id: string): Promise<Result<RemovedResponse>> {
    return this.http.request('DELETE', `/campaigns/${id}`);
  }
}

class Segments {
  constructor(private readonly http: HttpClient) {}
  create(payload: CreateSegmentOptions): Promise<Result<Segment>> {
    return this.http.request('POST', '/segments', payload);
  }
  get(id: string): Promise<Result<Segment>> {
    return this.http.request('GET', `/segments/${id}`);
  }
  list(params?: PaginationParams): Promise<Result<ListResponse<Segment>>> {
    return this.http.request('GET', `/segments${paginate(params)}`);
  }
  /** Preview the contacts a segment currently resolves to. */
  contacts(id: string): Promise<Result<ListResponse<Contact>>> {
    return this.http.request('GET', `/segments/${id}/contacts`);
  }
  update(id: string, payload: UpdateSegmentOptions): Promise<Result<Segment>> {
    return this.http.request('PATCH', `/segments/${id}`, payload);
  }
  remove(id: string): Promise<Result<RemovedResponse>> {
    return this.http.request('DELETE', `/segments/${id}`);
  }
}

class Templates {
  constructor(private readonly http: HttpClient) {}
  /** Returns the slim ack { object: 'template', id }. */
  create(payload: CreateTemplateOptions): Promise<Result<ObjectRef<'template'>>> {
    return this.http.request('POST', '/templates', payload);
  }
  get(id: string): Promise<Result<Template>> {
    return this.http.request('GET', `/templates/${id}`);
  }
  list(params?: PaginationParams): Promise<Result<ListResponse<Template>>> {
    return this.http.request('GET', `/templates${paginate(params)}`);
  }
  /** Returns the slim ack { object: 'template', id }. */
  update(id: string, payload: UpdateTemplateOptions): Promise<Result<ObjectRef<'template'>>> {
    return this.http.request('PATCH', `/templates/${id}`, payload);
  }
  /** Duplicate a template. POST /templates/:id/duplicate → { object: 'template', id }. */
  duplicate(id: string, payload?: DuplicateTemplateOptions): Promise<Result<ObjectRef<'template'>>> {
    return this.http.request('POST', `/templates/${id}/duplicate`, payload ?? {});
  }
  /** Publish a template (make its latest draft live). POST /templates/:id/publish → { object: 'template', id }. */
  publish(id: string): Promise<Result<ObjectRef<'template'>>> {
    return this.http.request('POST', `/templates/${id}/publish`);
  }
  remove(id: string): Promise<Result<RemovedResponse>> {
    return this.http.request('DELETE', `/templates/${id}`);
  }
}

class Topics {
  constructor(private readonly http: HttpClient) {}
  create(payload: CreateTopicOptions): Promise<Result<Topic>> {
    return this.http.request('POST', '/topics', payload);
  }
  get(id: string): Promise<Result<Topic>> {
    return this.http.request('GET', `/topics/${id}`);
  }
  list(params?: PaginationParams): Promise<Result<ListResponse<Topic>>> {
    return this.http.request('GET', `/topics${paginate(params)}`);
  }
  update(id: string, payload: UpdateTopicOptions): Promise<Result<Topic>> {
    return this.http.request('PATCH', `/topics/${id}`, payload);
  }
  remove(id: string): Promise<Result<RemovedResponse>> {
    return this.http.request('DELETE', `/topics/${id}`);
  }
}

class Automations {
  constructor(private readonly http: HttpClient) {}
  create(payload: CreateAutomationOptions): Promise<Result<Automation>> {
    return this.http.request('POST', '/automations', payload);
  }
  get(id: string): Promise<Result<Automation>> {
    return this.http.request('GET', `/automations/${id}`);
  }
  list(params?: PaginationParams): Promise<Result<ListResponse<Automation>>> {
    return this.http.request('GET', `/automations${paginate(params)}`);
  }
  update(id: string, payload: UpdateAutomationOptions): Promise<Result<Automation>> {
    return this.http.request('PATCH', `/automations/${id}`, payload);
  }
  /** Append a step to an automation. POST /automations/:id/steps → the created step. */
  addStep(id: string, payload: AddAutomationStepOptions): Promise<Result<AutomationStep>> {
    return this.http.request('POST', `/automations/${id}/steps`, payload);
  }
  /** Delete a step from an automation. DELETE /automations/:id/steps/:stepId → { id, deleted }. */
  deleteStep(id: string, stepId: string): Promise<Result<{ id: string; deleted: boolean }>> {
    return this.http.request('DELETE', `/automations/${id}/steps/${stepId}`);
  }
  /** List an automation's runs. GET /automations/:id/runs */
  runs(id: string, params?: PaginationParams): Promise<Result<ListResponse<AutomationRun>>> {
    return this.http.request('GET', `/automations/${id}/runs${paginate(params)}`);
  }
  /** Retrieve a single automation run. GET /automations/:id/runs/:runId */
  getRun(id: string, runId: string): Promise<Result<AutomationRun>> {
    return this.http.request('GET', `/automations/${id}/runs/${runId}`);
  }
  /** Stop an automation — prevents new runs; in-progress runs finish. POST /automations/:id/stop */
  stop(id: string): Promise<Result<Automation>> {
    return this.http.request('POST', `/automations/${id}/stop`);
  }
  remove(id: string): Promise<Result<RemovedResponse>> {
    return this.http.request('DELETE', `/automations/${id}`);
  }
}

class Webhooks {
  constructor(private readonly http: HttpClient) {}
  /** Create a webhook. The signing secret is shown ONCE, only here. */
  create(payload: CreateWebhookOptions): Promise<Result<ObjectRef<'webhook'> & { signing_secret: string }>> {
    return this.http.request('POST', '/webhooks', payload);
  }
  get(id: string): Promise<Result<Webhook>> {
    return this.http.request('GET', `/webhooks/${id}`);
  }
  list(params?: PaginationParams): Promise<Result<ListResponse<Webhook>>> {
    return this.http.request('GET', `/webhooks${paginate(params)}`);
  }
  /** Returns the slim ack { object: 'webhook', id }. */
  update(id: string, payload: UpdateWebhookOptions): Promise<Result<ObjectRef<'webhook'>>> {
    return this.http.request('PATCH', `/webhooks/${id}`, payload);
  }
  /**
   * Rotate the signing secret. The new plaintext `signing_secret` is returned
   * ONCE (reveal-once); the old secret stops verifying immediately.
   * POST /webhooks/:id/rotate
   */
  rotate(id: string): Promise<Result<ObjectRef<'webhook'> & { signing_secret: string }>> {
    return this.http.request('POST', `/webhooks/${id}/rotate`);
  }
  /** Send a synchronous test delivery and return the endpoint's live result. POST /webhooks/:id/test */
  test(id: string): Promise<Result<{ object: 'webhook_test'; id: string; [k: string]: unknown }>> {
    return this.http.request('POST', `/webhooks/${id}/test`);
  }
  remove(id: string): Promise<Result<RemovedResponse>> {
    return this.http.request('DELETE', `/webhooks/${id}`);
  }
  /**
   * Verify a webhook delivery's Svix-style signature against your endpoint's
   * signing secret. Mirrors the backend signing scheme
   * (`${id}.${timestamp}.${body}` → base64 HMAC-SHA256, tagged `v1,`).
   *
   * `payload` MUST be the exact raw request body string the server sent (do not
   * re-serialize the parsed JSON — whitespace differences break the signature).
   * `headers` accepts the svix-id/svix-timestamp/svix-signature headers (read
   * case-insensitively). A header may carry multiple space-separated signatures;
   * any one matching makes the delivery valid.
   *
   * This is a pure local computation — it makes no HTTP request.
   */
  verify(payload: string, headers: WebhookHeaders, secret: string, options: VerifyWebhookOptions = {}): VerifyWebhookResult {
    return verifyWebhookSignature(payload, headers, secret, options);
  }
}

/** Case-insensitively read a single header value (first if it's an array). */
function readHeader(headers: WebhookHeaders, name: string): string | undefined {
  const lower = name.toLowerCase();
  for (const k of Object.keys(headers)) {
    if (k.toLowerCase() === lower) {
      const v = headers[k];
      return Array.isArray(v) ? v[0] : v;
    }
  }
  return undefined;
}

/**
 * Derive the HMAC key from a `whsec_`-prefixed secret (base64-decode the suffix);
 * a secret without the prefix is used as raw UTF-8 bytes. Mirrors the backend's
 * `secretToKey` (src/services/svix_signature.ts).
 */
function secretToKey(secret: string): Buffer {
  const s = String(secret || '');
  if (s.startsWith('whsec_')) {
    const b64 = s.slice('whsec_'.length);
    try {
      const buf = Buffer.from(b64, 'base64');
      if (buf.length) return buf;
    } catch { /* fall through to raw */ }
  }
  return Buffer.from(s, 'utf8');
}

/** Constant-time compare of two base64 signature strings. */
function safeEqual(a: string, b: string): boolean {
  const ab = Buffer.from(a, 'utf8');
  const bb = Buffer.from(b, 'utf8');
  if (ab.length !== bb.length) return false;
  return timingSafeEqual(ab, bb);
}

/**
 * Pure Svix-style signature verification. Exported so it can be used without
 * constructing a client (e.g. in a request handler that only has the raw body).
 */
export function verifyWebhookSignature(
  payload: string,
  headers: WebhookHeaders,
  secret: string,
  options: VerifyWebhookOptions = {},
): VerifyWebhookResult {
  const id = readHeader(headers, 'svix-id');
  const timestamp = readHeader(headers, 'svix-timestamp');
  const sigHeader = readHeader(headers, 'svix-signature');
  if (!id || !timestamp || !sigHeader) return { valid: false, reason: 'missing_headers' };
  if (!secret) return { valid: false, reason: 'missing_secret' };

  // Optional timestamp freshness check (default 5-minute tolerance; 0 disables).
  const toleranceSec = options.toleranceSec ?? 300;
  if (toleranceSec > 0) {
    const ts = Number(timestamp);
    if (!Number.isFinite(ts)) return { valid: false, reason: 'invalid_timestamp' };
    const skew = Math.abs(Math.floor(Date.now() / 1000) - ts);
    if (skew > toleranceSec) return { valid: false, reason: 'timestamp_out_of_tolerance' };
  }

  const signed = `${id}.${timestamp}.${payload}`;
  const expected = createHmac('sha256', secretToKey(secret)).update(signed).digest('base64');

  // The header may contain multiple space-separated `v1,<sig>` entries; any match wins.
  for (const part of sigHeader.split(' ')) {
    const trimmed = part.trim();
    if (!trimmed) continue;
    const sig = trimmed.startsWith('v1,') ? trimmed.slice('v1,'.length) : trimmed;
    if (safeEqual(sig, expected)) return { valid: true };
  }
  return { valid: false, reason: 'no_match' };
}

/** List-logs params: cursor pagination plus optional method/status filters. */
export interface ListLogsParams extends PaginationParams {
  /** Filter to an exact HTTP method, e.g. 'POST'. */
  method?: string;
  /** Filter to an exact response status, e.g. 200 or 429. */
  status?: number;
}

class Logs {
  constructor(private readonly http: HttpClient) {}
  /**
   * List API request logs. Cursor-paginated (`limit`/`after`/`before`) with
   * optional `method` and `status` filters (server-side).
   */
  list(params?: ListLogsParams): Promise<Result<ListResponse<LogEntry>>> {
    const qs = new URLSearchParams();
    if (params?.limit != null) qs.set('limit', String(params.limit));
    if (params?.after) qs.set('after', params.after);
    if (params?.before) qs.set('before', params.before);
    if (params?.method) qs.set('method', params.method);
    if (params?.status != null) qs.set('status', String(params.status));
    const q = qs.toString();
    return this.http.request('GET', `/logs${q ? `?${q}` : ''}`);
  }
  get(id: string): Promise<Result<LogEntry>> {
    return this.http.request('GET', `/logs/${id}`);
  }
}

class Events {
  constructor(private readonly http: HttpClient) {}
  /** Send a custom event that automations can trigger on. POST /events/send */
  send(payload: SendEventOptions, options?: RequestOptions): Promise<Result<SendEventResponse>> {
    return this.http.request('POST', '/events/send', payload, options);
  }
  /** Create a custom-event definition (name + optional payload schema). POST /events */
  create(payload: CreateEventOptions, options?: RequestOptions): Promise<Result<EventDefinition>> {
    return this.http.request('POST', '/events', payload, options);
  }
  /** List custom-event definitions. GET /events */
  list(params?: PaginationParams): Promise<Result<ListResponse<EventDefinition>>> {
    return this.http.request('GET', `/events${paginate(params)}`);
  }
  /** Delete a custom-event definition. DELETE /events/:id */
  remove(id: string): Promise<Result<RemovedResponse>> {
    return this.http.request('DELETE', `/events/${id}`);
  }
}

class ApiKeys {
  constructor(private readonly http: HttpClient) {}
  create(payload: CreateApiKeyOptions): Promise<Result<CreateApiKeyResponse>> {
    return this.http.request('POST', '/api-keys', payload);
  }
  list(): Promise<Result<ListResponse<ApiKey>>> {
    return this.http.request('GET', '/api-keys');
  }
  remove(id: string): Promise<Result<RemovedResponse>> {
    return this.http.request('DELETE', `/api-keys/${id}`);
  }
}

export interface MailBlastrOptions extends ClientConfig {}

/**
 * The MailBlastr API client.
 *
 * ```ts
 * import { MailBlastr } from 'mailblastr';
 * const mb = new MailBlastr('mb_xxxxxxxxx');
 * const { data, error } = await mb.emails.send({
 *   from: 'you@yourdomain.com',
 *   to: ['user@example.com'],
 *   subject: 'Hello',
 *   html: '<p>Hi 👋</p>',
 * });
 * ```
 */
export class MailBlastr {
  readonly emails: Emails;
  readonly batch: Batch;
  readonly domains: Domains;
  readonly audiences: Audiences;
  readonly contacts: Contacts;
  readonly contactProperties: ContactProperties;
  readonly campaigns: Campaigns;
  readonly segments: Segments;
  readonly topics: Topics;
  readonly templates: Templates;
  readonly automations: Automations;
  readonly webhooks: Webhooks;
  readonly logs: Logs;
  readonly events: Events;
  readonly apiKeys: ApiKeys;
  readonly polls: Polls;

  constructor(apiKey: string, options: MailBlastrOptions = {}) {
    const http = new HttpClient(apiKey, options);
    this.emails = new Emails(http);
    this.batch = new Batch(http);
    this.domains = new Domains(http);
    this.audiences = new Audiences(http);
    this.contacts = new Contacts(http);
    this.contactProperties = new ContactProperties(http);
    this.campaigns = new Campaigns(http);
    this.segments = new Segments(http);
    this.topics = new Topics(http);
    this.templates = new Templates(http);
    this.automations = new Automations(http);
    this.webhooks = new Webhooks(http);
    this.logs = new Logs(http);
    this.events = new Events(http);
    this.apiKeys = new ApiKeys(http);
    this.polls = new Polls(http);
  }
}

export default MailBlastr;
