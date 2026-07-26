package mailblastr

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

// Attachment is a file attached to an outgoing email. Provide Content
// (base64) OR Path (a hosted URL fetched at send time).
type Attachment struct {
	Filename string `json:"filename"`
	// Content is the base64-encoded file content.
	Content string `json:"content,omitempty"`
	// Path is a hosted URL to fetch the file from.
	Path        string `json:"path,omitempty"`
	ContentType string `json:"content_type,omitempty"`
	// ContentId is the Content-ID for inline/related parts (renders as cid: references).
	ContentId string `json:"content_id,omitempty"`
}

// TemplateRef is a nested template reference on SendEmailRequest. Provide Id
// OR Alias to select the template.
type TemplateRef struct {
	Id        string         `json:"id,omitempty"`
	Alias     string         `json:"alias,omitempty"`
	Variables map[string]any `json:"variables,omitempty"`
}

// SendEmailRequest is the payload for POST /emails.
type SendEmailRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Bcc     []string `json:"bcc,omitempty"`
	Cc      []string `json:"cc,omitempty"`
	ReplyTo []string `json:"reply_to,omitempty"`
	// Html body. Markdown-style [text](url) links and bare URLs are converted
	// to tracked hyperlinks automatically at send time.
	Html string `json:"html,omitempty"`
	Text string `json:"text,omitempty"`
	// PreviewText is the inbox preview (preheader) shown next to the subject.
	// Max 150 characters.
	PreviewText string            `json:"preview_text,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
	Attachments []Attachment      `json:"attachments,omitempty"`
	// ScheduledAt is an ISO 8601 timestamp to schedule the send.
	ScheduledAt string `json:"scheduled_at,omitempty"`
	// TopicId drops recipients unsubscribed from this topic (topic gating).
	TopicId string `json:"topic_id,omitempty"`
	// TemplateId sends using a saved template; its subject/html/text fill any
	// omitted field.
	TemplateId string `json:"template_id,omitempty"`
	// Template is a nested template reference. Provide Template OR Html/Text.
	Template *TemplateRef `json:"template,omitempty"`
	// Variables fills the template's {{ placeholder }} variables.
	Variables map[string]any `json:"variables,omitempty"`
}

// BatchEmailRequest is a single email in a batch send (POST /emails/batch).
// Identical to SendEmailRequest minus Attachments and ScheduledAt — the batch
// endpoint rejects both per item; send those individually via Emails.Send.
type BatchEmailRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Bcc     []string `json:"bcc,omitempty"`
	Cc      []string `json:"cc,omitempty"`
	ReplyTo []string `json:"reply_to,omitempty"`
	// Html body. Markdown-style [text](url) links and bare URLs are converted
	// to tracked hyperlinks automatically at send time.
	Html string `json:"html,omitempty"`
	Text string `json:"text,omitempty"`
	// PreviewText is the inbox preview (preheader) shown next to the subject.
	// Max 150 characters.
	PreviewText string            `json:"preview_text,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
	// TopicId drops recipients unsubscribed from this topic (topic gating).
	TopicId string `json:"topic_id,omitempty"`
	// TemplateId sends using a saved template; its subject/html/text fill any
	// omitted field.
	TemplateId string `json:"template_id,omitempty"`
	// Template is a nested template reference. Provide Template OR Html/Text.
	Template *TemplateRef `json:"template,omitempty"`
	// Variables fills the template's {{ placeholder }} variables.
	Variables map[string]any `json:"variables,omitempty"`
}

// CreateEmailResponse is the { id } acknowledgement of a send.
type CreateEmailResponse struct {
	Id string `json:"id"`
}

// BatchSendResponse wraps the ids created by a batch send.
type BatchSendResponse struct {
	Data []CreateEmailResponse `json:"data"`
}

// EmailEvent is one entry of a sent email's event timeline.
type EmailEvent struct {
	Type      string `json:"type"`
	CreatedAt string `json:"created_at"`
}

// Email is a sent email as returned by GET /emails/:id.
type Email struct {
	Object string `json:"object"`
	Id     string `json:"id"`
	// MessageId is the provider message id; empty until the send is accepted.
	MessageId string   `json:"message_id"`
	From      string   `json:"from"`
	To        []string `json:"to"`
	Cc        []string `json:"cc,omitempty"`
	Bcc       []string `json:"bcc,omitempty"`
	ReplyTo   []string `json:"reply_to,omitempty"`
	Subject   string   `json:"subject"`
	Html      string   `json:"html,omitempty"`
	Text      string   `json:"text,omitempty"`
	Status    string   `json:"status"`
	LastEvent string   `json:"last_event,omitempty"`
	// Error is a plain-language failure reason; set only when the send failed.
	Error       string       `json:"error,omitempty"`
	ScheduledAt string       `json:"scheduled_at,omitempty"`
	CreatedAt   string       `json:"created_at"`
	Events      []EmailEvent `json:"events,omitempty"`
}

// SentEmailListItem is the trimmed shape returned by Emails.List (GET
// /emails): no Status/Html/Text/Events. Use Emails.Get for the full
// email with its event timeline.
type SentEmailListItem struct {
	Object      string   `json:"object"`
	Id          string   `json:"id"`
	MessageId   string   `json:"message_id"`
	From        string   `json:"from"`
	To          []string `json:"to"`
	Cc          []string `json:"cc"`
	Bcc         []string `json:"bcc"`
	ReplyTo     []string `json:"reply_to"`
	Subject     string   `json:"subject"`
	LastEvent   string   `json:"last_event"`
	ScheduledAt string   `json:"scheduled_at"`
	CreatedAt   string   `json:"created_at"`
}

// AttachmentMeta describes an attachment of a sent email.
type AttachmentMeta struct {
	Object             string `json:"object"`
	Id                 string `json:"id"`
	Filename           string `json:"filename"`
	ContentType        string `json:"content_type"`
	ContentDisposition string `json:"content_disposition"`
	ContentId          string `json:"content_id"`
	Size               int64  `json:"size"`
	DownloadUrl        string `json:"download_url"`
	ExpiresAt          string `json:"expires_at"`
}

// ListEmailsRequest lists sent emails: cursor pagination plus optional
// server-side source filters.
type ListEmailsRequest struct {
	Limit  int
	After  string
	Before string
	// CampaignId restricts to emails sent by this campaign. Takes precedence
	// over AutomationId/Source.
	CampaignId string
	// AutomationId restricts to emails sent by this automation.
	AutomationId string
	// Source: "individual" restricts to one-off API sends (no
	// campaign/automation).
	Source string
	// DomainId restricts to emails sent from this sending domain (domain id).
	// Composes with the source filters.
	DomainId string
}

// UpdateEmailRequest reschedules a scheduled email.
type UpdateEmailRequest struct {
	// ScheduledAt is the new ISO 8601 send time.
	ScheduledAt string `json:"scheduled_at"`
}

// EmailsService handles the /emails endpoints. Inbound email lives on the
// nested Receiving sub-service.
type EmailsService struct {
	client *Client
	// Receiving is the inbound (received) email sub-resource.
	Receiving *ReceivingService
}

// Send sends a single email. POST /emails
func (s *EmailsService) Send(params *SendEmailRequest) (*CreateEmailResponse, error) {
	return s.SendWithContext(context.Background(), params)
}

// SendWithContext sends a single email. POST /emails
func (s *EmailsService) SendWithContext(ctx context.Context, params *SendEmailRequest) (*CreateEmailResponse, error) {
	return request[CreateEmailResponse](ctx, s.client, http.MethodPost, "/emails", params, nil)
}

// SendWithOptions sends a single email with per-request options (e.g. an
// Idempotency-Key). POST /emails
func (s *EmailsService) SendWithOptions(ctx context.Context, params *SendEmailRequest, opts *RequestOptions) (*CreateEmailResponse, error) {
	return request[CreateEmailResponse](ctx, s.client, http.MethodPost, "/emails", params, opts)
}

// Batch sends up to 100 emails in one request (alias of client.Batch.Send).
// POST /emails/batch
//
// Deprecated: use client.Batch.SendEmails — batch items reject Attachments
// and ScheduledAt (send those individually via Emails.Send), which
// BatchEmailRequest enforces at compile time.
func (s *EmailsService) Batch(params []*SendEmailRequest) (*BatchSendResponse, error) {
	return s.BatchWithContext(context.Background(), params)
}

// BatchWithContext sends up to 100 emails in one request. POST /emails/batch
//
// Deprecated: use client.Batch.SendEmailsWithContext — batch items reject
// Attachments and ScheduledAt, which BatchEmailRequest enforces at compile
// time.
func (s *EmailsService) BatchWithContext(ctx context.Context, params []*SendEmailRequest) (*BatchSendResponse, error) {
	return request[BatchSendResponse](ctx, s.client, http.MethodPost, "/emails/batch", params, nil)
}

// List lists sent emails (trimmed SentEmailListItem rows). GET /emails
// For the campaign_id/automation_id/source/domain_id filters use ListFiltered.
func (s *EmailsService) List(params *ListParams) (*ListResponse[SentEmailListItem], error) {
	return s.ListWithContext(context.Background(), params)
}

// ListWithContext lists sent emails. GET /emails
func (s *EmailsService) ListWithContext(ctx context.Context, params *ListParams) (*ListResponse[SentEmailListItem], error) {
	return request[ListResponse[SentEmailListItem]](ctx, s.client, http.MethodGet, listPath("/emails", params), nil, nil)
}

// ListFiltered lists sent emails with optional campaign_id/automation_id/
// source/domain_id filters. GET /emails
func (s *EmailsService) ListFiltered(params *ListEmailsRequest) (*ListResponse[SentEmailListItem], error) {
	return s.ListFilteredWithContext(context.Background(), params)
}

// ListFilteredWithContext lists sent emails with optional filters. GET /emails
func (s *EmailsService) ListFilteredWithContext(ctx context.Context, params *ListEmailsRequest) (*ListResponse[SentEmailListItem], error) {
	q := url.Values{}
	if params != nil {
		if params.Limit > 0 {
			q.Set("limit", strconv.Itoa(params.Limit))
		}
		if params.After != "" {
			q.Set("after", params.After)
		}
		if params.Before != "" {
			q.Set("before", params.Before)
		}
		if params.CampaignId != "" {
			q.Set("campaign_id", params.CampaignId)
		}
		if params.AutomationId != "" {
			q.Set("automation_id", params.AutomationId)
		}
		if params.Source != "" {
			q.Set("source", params.Source)
		}
		if params.DomainId != "" {
			q.Set("domain_id", params.DomainId)
		}
	}
	path := "/emails"
	if enc := q.Encode(); enc != "" {
		path += "?" + enc
	}
	return request[ListResponse[SentEmailListItem]](ctx, s.client, http.MethodGet, path, nil, nil)
}

// Get retrieves a sent email and its events. GET /emails/:id
func (s *EmailsService) Get(id string) (*Email, error) {
	return s.GetWithContext(context.Background(), id)
}

// GetWithContext retrieves a sent email and its events. GET /emails/:id
func (s *EmailsService) GetWithContext(ctx context.Context, id string) (*Email, error) {
	return request[Email](ctx, s.client, http.MethodGet, "/emails/"+esc(id), nil, nil)
}

// ListAttachments lists a sent email's attachments. GET /emails/:id/attachments
func (s *EmailsService) ListAttachments(id string) (*ListResponse[AttachmentMeta], error) {
	return s.ListAttachmentsWithContext(context.Background(), id)
}

// ListAttachmentsWithContext lists a sent email's attachments.
func (s *EmailsService) ListAttachmentsWithContext(ctx context.Context, id string) (*ListResponse[AttachmentMeta], error) {
	return request[ListResponse[AttachmentMeta]](ctx, s.client, http.MethodGet, "/emails/"+esc(id)+"/attachments", nil, nil)
}

// GetAttachment retrieves one attachment of a sent email.
// GET /emails/:id/attachments/:attachmentId
func (s *EmailsService) GetAttachment(id, attachmentId string) (*AttachmentMeta, error) {
	return s.GetAttachmentWithContext(context.Background(), id, attachmentId)
}

// GetAttachmentWithContext retrieves one attachment of a sent email.
func (s *EmailsService) GetAttachmentWithContext(ctx context.Context, id, attachmentId string) (*AttachmentMeta, error) {
	return request[AttachmentMeta](ctx, s.client, http.MethodGet, "/emails/"+esc(id)+"/attachments/"+esc(attachmentId), nil, nil)
}

// Update reschedules a scheduled email. PATCH /emails/:id
func (s *EmailsService) Update(id string, params *UpdateEmailRequest) (*ObjectRef, error) {
	return s.UpdateWithContext(context.Background(), id, params)
}

// UpdateWithContext reschedules a scheduled email. PATCH /emails/:id
func (s *EmailsService) UpdateWithContext(ctx context.Context, id string, params *UpdateEmailRequest) (*ObjectRef, error) {
	return request[ObjectRef](ctx, s.client, http.MethodPatch, "/emails/"+esc(id), params, nil)
}

// Cancel cancels a scheduled email. POST /emails/:id/cancel
func (s *EmailsService) Cancel(id string) (*ObjectRef, error) {
	return s.CancelWithContext(context.Background(), id)
}

// CancelWithContext cancels a scheduled email. POST /emails/:id/cancel
func (s *EmailsService) CancelWithContext(ctx context.Context, id string) (*ObjectRef, error) {
	return request[ObjectRef](ctx, s.client, http.MethodPost, "/emails/"+esc(id)+"/cancel", nil, nil)
}

// BatchService sends up to 100 emails in one request: client.Batch.Send([...]).
type BatchService struct {
	client *Client
}

// Send sends up to 100 emails in one request. POST /emails/batch
//
// Deprecated: use SendEmails — batch items reject Attachments and ScheduledAt
// (send those individually via Emails.Send), which BatchEmailRequest enforces
// at compile time.
func (s *BatchService) Send(params []*SendEmailRequest) (*BatchSendResponse, error) {
	return s.SendWithContext(context.Background(), params)
}

// SendWithContext sends up to 100 emails in one request. POST /emails/batch
//
// Deprecated: use SendEmailsWithContext — batch items reject Attachments and
// ScheduledAt, which BatchEmailRequest enforces at compile time.
func (s *BatchService) SendWithContext(ctx context.Context, params []*SendEmailRequest) (*BatchSendResponse, error) {
	return request[BatchSendResponse](ctx, s.client, http.MethodPost, "/emails/batch", params, nil)
}

// SendWithOptions sends a batch with per-request options (e.g. an
// Idempotency-Key). POST /emails/batch
//
// Deprecated: use SendEmailsWithOptions — batch items reject Attachments and
// ScheduledAt, which BatchEmailRequest enforces at compile time.
func (s *BatchService) SendWithOptions(ctx context.Context, params []*SendEmailRequest, opts *RequestOptions) (*BatchSendResponse, error) {
	return request[BatchSendResponse](ctx, s.client, http.MethodPost, "/emails/batch", params, opts)
}

// SendEmails sends up to 100 emails in one request. Batch items reject
// Attachments and ScheduledAt — send those individually via Emails.Send.
// POST /emails/batch
func (s *BatchService) SendEmails(params []*BatchEmailRequest) (*BatchSendResponse, error) {
	return s.SendEmailsWithContext(context.Background(), params)
}

// SendEmailsWithContext sends up to 100 emails in one request.
// POST /emails/batch
func (s *BatchService) SendEmailsWithContext(ctx context.Context, params []*BatchEmailRequest) (*BatchSendResponse, error) {
	return request[BatchSendResponse](ctx, s.client, http.MethodPost, "/emails/batch", params, nil)
}

// SendEmailsWithOptions sends a batch with per-request options (e.g. an
// Idempotency-Key). POST /emails/batch
func (s *BatchService) SendEmailsWithOptions(ctx context.Context, params []*BatchEmailRequest, opts *RequestOptions) (*BatchSendResponse, error) {
	return request[BatchSendResponse](ctx, s.client, http.MethodPost, "/emails/batch", params, opts)
}
