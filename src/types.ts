// Public types for the MailBlastr SDK. These mirror the REST API shapes.

export interface MailBlastrError {
  statusCode: number;
  name: string;
  message: string;
}

/** Every method returns either { data, error: null } or { data: null, error }. */
export type Result<T> = { data: T; error: null } | { data: null; error: MailBlastrError };

/** Slim acknowledgement shape returned by several create/update routes. */
export interface ObjectRef<O extends string> { object: O; id: string }

export interface RequestOptions {
  /** Optional Idempotency-Key for safely retrying a create (24h window). */
  idempotencyKey?: string;
}

// ---- Emails ----
export interface Attachment {
  filename: string;
  /** Base64-encoded file content. Provide `content` OR `path`. */
  content?: string;
  /** A hosted URL to fetch the file from. Provide `content` OR `path`. */
  path?: string;
  content_type?: string;
  /** Content-ID for inline/related parts (renders as `cid:` references). */
  content_id?: string;
}
export interface Tag { name: string; value: string }

export interface SendEmailOptions {
  from: string;
  to: string | string[];
  subject: string;
  bcc?: string | string[];
  cc?: string | string[];
  reply_to?: string | string[];
  html?: string;
  text?: string;
  headers?: Record<string, string>;
  attachments?: Attachment[];
  tags?: Tag[];
  /** ISO 8601 timestamp to schedule the send. */
  scheduled_at?: string;
  /** Drop recipients unsubscribed from this topic (topic gating). */
  topic_id?: string;
  /** Send using a saved template; its subject/html/text fill any omitted field. */
  template_id?: string;
  /**
   * Nested template reference. Provide `template` OR
   * `html`/`text`, not both. `id` OR `alias` selects the template.
   */
  template?: { id?: string; alias?: string; variables?: Record<string, unknown> };
  /** Values for the template's {{ placeholder }} variables. */
  variables?: Record<string, string | number | boolean | null>;
}
export interface CreateEmailResponse { id: string }
export interface EmailEvent { type: string; created_at: string }
export interface Email {
  object: 'email';
  id: string;
  /** Provider message id; null until the send is accepted. */
  message_id?: string | null;
  from: string;
  to: string[];
  cc?: string[];
  bcc?: string[];
  reply_to?: string[];
  subject: string | null;
  html?: string;
  text?: string;
  status: string;
  last_event?: string;
  scheduled_at?: string | null;
  created_at: string;
  events?: EmailEvent[];
}

// ---- Domains ----
export interface DomainRecord {
  record: 'DKIM' | 'SPF' | 'DMARC' | 'Tracking' | 'Receiving MX';
  name: string;
  type: 'CNAME' | 'MX' | 'TXT' | 'CAA';
  ttl: string;
  status: string;
  value: string;
  priority?: number;
}
export interface DomainCapabilities {
  sending: 'enabled' | 'disabled';
  receiving: 'enabled' | 'disabled';
}
export interface Domain {
  object: 'domain';
  id: string;
  name: string;
  status: string;
  region: string;
  created_at: string;
  records: DomainRecord[];
  custom_return_path?: string;
  open_tracking?: boolean;
  click_tracking?: boolean;
  /** Open/click tracking subdomain label (null ⇒ tracking on the shared host). */
  tracking_subdomain?: string | null;
  /** Outbound TLS policy. */
  tls?: 'opportunistic' | 'enforced';
  /** Sending/receiving capabilities. */
  capabilities?: DomainCapabilities;
  tracking_domain?: string | null;
  tracking_verified?: boolean;
}
export interface CreateDomainOptions {
  name: string;
  region?: string;
  /** MAIL FROM subdomain (Return-Path); defaults to 'send'. */
  custom_return_path?: string;
  open_tracking?: boolean;
  click_tracking?: boolean;
  /** Custom tracking host label, e.g. 'email' ⇒ email.<domain>. */
  tracking_subdomain?: string;
  /** Outbound TLS policy. */
  tls?: 'opportunistic' | 'enforced';
  /** Capabilities to enable, e.g. { receiving: 'enabled' }. */
  capabilities?: { receiving?: 'enabled' | 'disabled' };
}
export interface UpdateDomainOptions {
  open_tracking?: boolean;
  click_tracking?: boolean;
  tracking_subdomain?: string;
  tls?: 'opportunistic' | 'enforced';
  capabilities?: { receiving?: 'enabled' | 'disabled' };
}

// ---- Audiences & Contacts ----
export interface Audience { object: 'audience'; id: string; name: string; created_at: string }
export interface Contact {
  object: 'contact';
  id: string;
  email: string;
  first_name: string | null;
  last_name: string | null;
  unsubscribed: boolean;
  /** Custom contact properties (own values merged over registered fallbacks). */
  properties?: Record<string, unknown>;
  created_at: string;
}
export interface CreateContactOptions {
  audienceId: string;
  email: string;
  first_name?: string;
  last_name?: string;
  unsubscribed?: boolean;
  /** Custom contact property values keyed by the property key. */
  properties?: Record<string, string | number>;
}
export interface UpdateContactOptions {
  audienceId: string;
  /** Contact id OR email. */
  id: string;
  first_name?: string;
  last_name?: string;
  unsubscribed?: boolean;
  properties?: Record<string, string | number>;
}
/** A single contact in a bulk import (no audienceId — it's on the call). */
export interface ContactInput {
  email: string;
  first_name?: string;
  last_name?: string;
  unsubscribed?: boolean;
  properties?: Record<string, string | number>;
}
/** Result of a batch / CSV contact import. */
export interface ImportContactsResponse {
  object: 'list';
  imported: number; // newly inserted
  updated: number;  // existing contacts updated
  skipped: number;  // rows dropped for a missing/invalid email
  total: number;    // distinct contacts processed
}
/** Cursor paging for listing contacts (limit ≤ 100). */
export interface ListContactsParams {
  audienceId: string;
  limit?: number;
  /** Cursor: id of the last item on the previous page. */
  after?: string;
  /** Cursor: id of the first item on the next page. */
  before?: string;
}

// ---- Broadcasts ----
export interface Broadcast {
  object: 'broadcast';
  id: string;
  name: string | null;
  audience_id: string;
  from: string | null;
  subject: string | null;
  html?: string | null;
  text?: string | null;
  reply_to?: string | null;
  preview_text?: string | null;
  status: string;
  scheduled_at: string | null;
  sent_at: string | null;
  created_at: string;
}
/**
 * A/B-test config accepted on broadcast create. When `enabled`, supply at least
 * one variant-B field (`subject_b`, `html_b`, or `text_b`). `test_pct` is the
 * percentage of the audience used for the test split; `metric` decides the winner.
 */
export interface BroadcastAbTest {
  enabled: boolean;
  subject_b?: string | null;
  html_b?: string | null;
  text_b?: string | null;
  /** Percentage (1-100) of the audience used to pick the winner. Defaults to 20. */
  test_pct?: number;
  /** Winner metric. Defaults to 'open'. */
  metric?: 'open' | 'click';
}
/** A recurring broadcast re-sends every `recurrence_every` periods. */
export type BroadcastRecurrence = 'daily' | 'weekly' | 'monthly';
export interface CreateBroadcastOptions {
  audience_id: string;
  from: string;
  subject: string;
  html?: string;
  text?: string;
  reply_to?: string | string[];
  preview_text?: string;
  name?: string;
  /** Target a segment (subset of the audience) instead of everyone. */
  segment_id?: string;
  /** Gate recipients by a topic subscription. */
  topic_id?: string;
  /** Make this a recurring broadcast (re-sends every `recurrence_every` periods). */
  recurrence?: BroadcastRecurrence;
  /** Number of periods between recurring sends (1-365). Defaults to 1. */
  recurrence_every?: number;
  /** A/B-test configuration. */
  ab_test?: BroadcastAbTest;
  /** Send immediately on create (or schedule it when `scheduled_at` is given). */
  send?: boolean;
  /** ISO 8601 (or natural-language) schedule used when `send` is true. */
  scheduled_at?: string;
}
export interface UpdateBroadcastOptions {
  name?: string;
  from?: string;
  subject?: string;
  html?: string;
  text?: string;
  reply_to?: string | string[];
  preview_text?: string;
  audience_id?: string;
  /** Re-target a segment (pass null to clear). */
  segment_id?: string | null;
  /** Re-target a topic gate (pass null to clear). */
  topic_id?: string | null;
  /** Update the recurrence cadence. */
  recurrence?: BroadcastRecurrence;
  /** Update the number of periods between recurring sends (1-365). */
  recurrence_every?: number;
  /** Update the A/B-test configuration. */
  ab_test?: BroadcastAbTest;
}
/** Per-broadcast analytics returned by GET /broadcasts/:id/stats. */
export interface BroadcastStats {
  object: 'broadcast_stats';
  broadcast_id: string;
  links?: Array<Record<string, unknown>>;
  [k: string]: unknown;
}
/** A/B winner evaluation returned by GET /broadcasts/:id/ab. */
export interface BroadcastAbResult {
  object: 'broadcast_ab';
  broadcast_id: string;
  metric: 'open' | 'click';
  [k: string]: unknown;
}

// ---- Segments ----
export type SegmentStatus = 'all' | 'subscribed' | 'unsubscribed';
/** Operators for a custom-property segment predicate. */
export type PropertyOperator = 'eq' | 'contains' | 'exists';
/** A single custom-property predicate. `value` is required for eq/contains. */
export interface PropertyFilter {
  key: string;
  operator: PropertyOperator;
  value?: string | number | null;
}
export interface Segment {
  object: 'segment';
  id: string;
  audience_id: string;
  name: string;
  filter: { status: SegmentStatus; email_contains: string | null; property_filters?: PropertyFilter[] };
  created_at: string;
  updated_at: string;
}
export interface CreateSegmentOptions {
  audience_id: string;
  name: string;
  filter?: { status?: SegmentStatus; email_contains?: string | null; property_filters?: PropertyFilter[] };
}
export interface UpdateSegmentOptions {
  name?: string;
  filter?: { status?: SegmentStatus; email_contains?: string | null; property_filters?: PropertyFilter[] };
}

// ---- Templates ----
export interface TemplateVariable {
  id: string;
  key: string;
  type: 'string' | 'number';
  fallback_value: string | number | null;
  created_at: string;
  updated_at: string;
}
export interface Template {
  object: 'template';
  id: string;
  name: string;
  subject: string | null;
  from?: string | null;
  reply_to?: string | null;
  html: string | null;
  text: string | null;
  status?: string;
  current_version_id?: string | null;
  variables?: TemplateVariable[];
  created_at: string;
  updated_at: string;
}
/** A template variable definition accepted on create/update. */
export interface TemplateVariableInput {
  key: string;
  type?: 'string' | 'number';
  fallback_value?: string;
}
export interface CreateTemplateOptions {
  name: string;
  /** Optional stable handle for sending by alias. */
  alias?: string;
  subject?: string | null;
  from?: string | null;
  reply_to?: string | string[];
  html?: string | null;
  text?: string | null;
  variables?: TemplateVariableInput[];
}
export interface UpdateTemplateOptions {
  name?: string;
  /** Optional stable handle for sending by alias. */
  alias?: string;
  subject?: string | null;
  from?: string | null;
  reply_to?: string | string[];
  html?: string | null;
  text?: string | null;
  variables?: TemplateVariableInput[];
}
/** Body for templates.duplicate. */
export interface DuplicateTemplateOptions {
  name?: string;
  alias?: string;
}

// ---- API keys ----
export interface ApiKey { object: 'api_key'; id: string; name: string; created_at: string }
export interface CreateApiKeyResponse { id: string; token: string }
export interface CreateApiKeyOptions {
  name: string;
  permission?: 'full_access' | 'sending_access';
  /** Scope a `sending_access` key to one domain (ignored otherwise). */
  domain_id?: string;
}

export interface ListResponse<T> { object: 'list'; has_more: boolean; data: T[] }
export interface RemovedResponse { object: string; id: string; deleted: true }

/** Cursor pagination params accepted by most list() methods. */
export interface PaginationParams {
  limit?: number;
  after?: string;
  before?: string;
}

// ---- Email attachments (sent) ----
// Matches the backend `toSentAttachment` shape (src/routes/emails_api.ts).
export interface AttachmentMeta {
  object: 'attachment';
  id: string;
  filename: string | null;
  content_type: string | null;
  content_disposition: string;
  content_id: string | null;
  size: number | null;
  download_url: string | null;
  expires_at: string | null;
}

// ---- Inbound / received emails ----
// Matches the backend received-email output (src/services/inbound.ts).
export interface ReceivedAttachment {
  id?: string;
  filename: string | null;
  content_type: string | null;
  size: number;
  content_id?: string;
  content_disposition?: string;
  downloadable: boolean;
  download_url?: string;
  expires_at?: string;
  /** Backward-compatible relative-path alias. */
  url?: string;
}
export interface ReceivedEmail {
  object: 'received_email';
  id: string;
  from: string;
  to: string[];
  cc?: string[];
  bcc?: string[];
  received_for?: string[];
  subject: string | null;
  html?: string | null;
  text?: string | null;
  /** Parsed message headers as a name → value map. */
  headers?: Record<string, string>;
  message_id?: string;
  spf?: string;
  verdicts?: Record<string, unknown>;
  attachments?: ReceivedAttachment[];
  raw_available: boolean;
  raw?: { download_url: string; expires_at: string };
  raw_url?: string;
  created_at: string;
}
export interface ForwardReceivedEmailOptions {
  /** A verified sending address to forward from (required by the backend). */
  from: string;
  to: string | string[];
  subject?: string;
}

// ---- Domain claims ----
export interface DomainClaimRecord {
  type: 'TXT';
  name: string;
  value: string | null;
  ttl: string;
}
export interface DomainClaim {
  object: 'domain_claim';
  id: string;
  name: string;
  domain_id: string;
  region: string;
  status: string;
  record: DomainClaimRecord;
  blocked_reason: string | null;
  failure_reason: string | null;
  created_at: string;
  expires_at: string;
}
export interface ClaimDomainOptions { name: string; region?: string }

// ---- Contact properties ----
// Backend only supports string | number (src/services/contact_properties.ts).
export type ContactPropertyType = 'string' | 'number';
export interface ContactProperty {
  object: 'contact_property';
  id: string;
  key: string;
  type: ContactPropertyType;
  fallback_value: string | number | null;
  created_at: string;
}
export interface CreateContactPropertyOptions {
  /** Canonical merge-tag key. `name` is accepted as an alias. */
  key?: string;
  name?: string;
  type: ContactPropertyType;
  fallback_value?: string | number;
}
/** Only fallback_value is mutable; key/type are immutable. */
export interface UpdateContactPropertyOptions {
  fallback_value?: string | number | null;
}

// ---- Contact <-> topics ----
export interface ContactTopicSubscription {
  id: string;
  name: string;
  description: string | null;
  subscription: 'opt_in' | 'opt_out';
}
export interface ContactTopics {
  object: 'list';
  data: ContactTopicSubscription[];
}
export interface UpdateContactTopicsOptions {
  topics: { id: string; subscription: 'opt_in' | 'opt_out' }[];
}

// ---- Topics ----
export interface Topic {
  object: 'topic';
  id: string;
  audience_id: string;
  name: string;
  description: string | null;
  default_subscription: 'opt_in' | 'opt_out';
  visibility: 'public' | 'private';
  created_at: string;
}
export interface CreateTopicOptions {
  name: string;
  default_subscription: 'opt_in' | 'opt_out';
  visibility?: 'public' | 'private';
  description?: string | null;
  audience_id?: string;
}
export interface UpdateTopicOptions {
  name?: string;
  description?: string | null;
  visibility?: 'public' | 'private';
}

// ---- Automations ----
export interface AutomationStep {
  id: string;
  key: string;
  type: string;
  position: number;
  config: Record<string, unknown>;
}
export interface AutomationConnection {
  from: string;
  to: string;
  type: string;
}
export interface Automation {
  object: 'automation';
  id: string;
  audience_id: string;
  name: string;
  trigger: string;
  status: string;
  steps?: AutomationStep[];
  connections?: AutomationConnection[];
  created_at: string;
  updated_at: string;
}
export interface CreateAutomationOptions {
  name: string;
  audience_id: string;
  /** Built-in 'contact.created' or any custom event name. Defaults server-side. */
  trigger?: string;
  /** Initial status, e.g. 'draft' | 'active' | 'paused'. */
  status?: string;
  /** Optional inline step graph; each step may carry a `key` for connections. */
  steps?: Array<{ key?: string; type: string; config?: Record<string, unknown>; [k: string]: unknown }>;
  /** Optional typed edges between step keys. */
  connections?: Array<{ from: string; to: string; type?: string }>;
}
export interface UpdateAutomationOptions {
  name?: string;
  status?: 'draft' | 'active' | 'paused';
  connections?: Array<{ from: string; to: string; type?: string }>;
}
export interface AddAutomationStepOptions {
  type: string;
  config?: Record<string, unknown>;
  key?: string;
  [k: string]: unknown;
}
export interface AutomationRun {
  object: 'automation_run';
  id: string;
  automation_id: string;
  contact_id: string;
  status: string;
  steps: Array<Record<string, unknown>>;
  error: string | null;
  started_at: string;
  completed_at: string | null;
  created_at: string;
}

// ---- Webhooks ----
export interface Webhook {
  object: 'webhook';
  id: string;
  endpoint: string;
  events: string[];
  status: string;
  /** Signing secret — returned on create + retrieve, omitted from list. */
  signing_secret?: string;
  created_at: string;
}
export interface CreateWebhookOptions {
  endpoint: string;
  events: string[];
}
export interface UpdateWebhookOptions {
  endpoint?: string;
  events?: string[];
  status?: 'enabled' | 'disabled';
}
/**
 * The Svix-style delivery headers MailBlastr sends with each webhook. Either the
 * lowercase header names or a Headers-like object can be passed to
 * `webhooks.verify` — it reads them case-insensitively.
 */
export interface WebhookHeaders {
  'svix-id'?: string;
  'svix-timestamp'?: string;
  'svix-signature'?: string;
  [k: string]: string | string[] | undefined;
}
/** Outcome of verifying a webhook delivery signature. */
export interface VerifyWebhookResult {
  /** True when the signature matches and (when checked) the timestamp is fresh. */
  valid: boolean;
  /** A machine reason when `valid` is false (e.g. 'missing_headers', 'no_match'). */
  reason?: string;
}
/** Options for webhooks.verify (timestamp tolerance). */
export interface VerifyWebhookOptions {
  /** Max allowed clock skew in seconds (default 300). Pass 0 to skip the check. */
  toleranceSec?: number;
}

// ---- Logs ----
export interface LogEntry {
  object: 'log';
  id: string;
  endpoint: string | null;
  method: string;
  response_status: number;
  user_agent: string | null;
  created_at: string;
  /** Present only on retrieve (GET /logs/:id). */
  request_body?: unknown;
  response_body?: unknown;
}

// ---- Events (automation custom events) ----
export interface SendEventOptions {
  /** The custom event name automations can trigger on. */
  name: string;
  /** Identify the contact by id. Provide `contact_id` OR `email`. */
  contact_id?: string;
  /** Identify the contact by email. Provide `contact_id` OR `email`. */
  email?: string;
  /** Arbitrary event payload. */
  data?: Record<string, unknown>;
}
export interface SendEventResponse { object: 'event'; id: string }
