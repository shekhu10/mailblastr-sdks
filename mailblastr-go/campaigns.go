package mailblastr

import (
	"context"
	"net/http"
)

// CampaignAbTest is the A/B-test config accepted on campaign create/update.
// When Enabled, supply at least one variant-B field (SubjectB, HtmlB, or
// TextB).
type CampaignAbTest struct {
	Enabled  bool   `json:"enabled"`
	SubjectB string `json:"subject_b,omitempty"`
	HtmlB    string `json:"html_b,omitempty"`
	TextB    string `json:"text_b,omitempty"`
	// TestPct is the percentage (1-100) of the audience used to pick the
	// winner. Defaults to 20.
	TestPct int `json:"test_pct,omitempty"`
	// Metric decides the winner: "open" | "click" | "reply". Defaults to "open".
	Metric string `json:"metric,omitempty"`
	// EvalHours is the hours (1-168) to run the test before evaluating and
	// sending the winner.
	EvalHours int `json:"eval_hours,omitempty"`
}

// CampaignAbState is the A/B config + decision echoed back on retrieve.
type CampaignAbState struct {
	Enabled  bool   `json:"enabled"`
	SubjectB string `json:"subject_b,omitempty"`
	TestPct  int    `json:"test_pct,omitempty"`
	Metric   string `json:"metric,omitempty"`
	Status   string `json:"status,omitempty"`
	Winner   string `json:"winner,omitempty"`
}

// CampaignFollowupInput is an engagement follow-up accepted on create (max 5):
// sent Delay after the campaign finishes to recipients matching Condition,
// threaded as replies to the original email.
type CampaignFollowupInput struct {
	// Condition is one of: opened, clicked, not_opened, not_clicked, replied,
	// not_replied.
	Condition string `json:"condition"`
	// Delay is a natural-language duration, e.g. "5 hours" or "4 days" (max 30 days).
	Delay string `json:"delay"`
	// Subject defaults to "Re: <campaign subject>" (keeps the thread).
	Subject string `json:"subject,omitempty"`
	Html    string `json:"html"`
}

// CampaignFollowup is a follow-up echoed back on retrieve.
type CampaignFollowup struct {
	Id        string `json:"id"`
	Condition string `json:"condition"`
	Delay     string `json:"delay"`
	Subject   string `json:"subject"`
	Status    string `json:"status"`
	RunAt     string `json:"run_at"`
	SentCount int    `json:"sent_count"`
}

// CampaignRecurrenceState is the recurrence cadence echoed back on retrieve.
type CampaignRecurrenceState struct {
	// Interval is "daily" | "weekly" | "monthly".
	Interval string `json:"interval"`
	Every    int    `json:"every"`
}

// Campaign is a bulk send to an audience or segment.
type Campaign struct {
	Object     string `json:"object"`
	Id         string `json:"id"`
	Name       string `json:"name"`
	AudienceId string `json:"audience_id"`
	// SegmentId is the segment target (empty => whole audience).
	SegmentId string `json:"segment_id"`
	// TopicId is the topic gate (empty => no gating).
	TopicId     string `json:"topic_id"`
	From        string `json:"from"`
	Subject     string `json:"subject"`
	Html        string `json:"html,omitempty"`
	Text        string `json:"text,omitempty"`
	ReplyTo     string `json:"reply_to,omitempty"`
	PreviewText string `json:"preview_text,omitempty"`
	Status      string `json:"status"`
	ScheduledAt string `json:"scheduled_at"`
	SentAt      string `json:"sent_at"`
	CreatedAt   string `json:"created_at"`
	// AbTest is the A/B config + decision ({ enabled: false } when not an A/B
	// campaign).
	AbTest *CampaignAbState `json:"ab_test,omitempty"`
	// Followups are the engagement follow-ups (retrieve only).
	Followups []CampaignFollowup `json:"followups,omitempty"`
	// ListAddress is the generated mailing-list To address (retrieve only;
	// empty unless list_to was set).
	ListAddress string `json:"list_address,omitempty"`
	// UnsubscribePolicy: "account" | "domain" | "ignore" (retrieve only).
	UnsubscribePolicy string `json:"unsubscribe_policy,omitempty"`
	// Recurrence is the cadence (nil => one-off).
	Recurrence *CampaignRecurrenceState `json:"recurrence,omitempty"`
	// ParentCampaignId is set on auto-generated occurrences of a recurring
	// campaign.
	ParentCampaignId string `json:"parent_campaign_id,omitempty"`
	// Statistics holds engagement counts — included only on retrieve.
	Statistics map[string]any `json:"statistics,omitempty"`
}

// CreateCampaignRequest is the payload for POST /campaigns. Domain is
// REQUIRED (domain-first): it names the sending domain whose contact pool
// this campaign targets; it is orthogonal to From, which may be a different
// verified domain.
type CreateCampaignRequest struct {
	// Domain is REQUIRED — the sending domain whose contact pool this
	// campaign targets (e.g. "yourdomain.com" — one of your domains).
	Domain      string   `json:"domain"`
	From        string   `json:"from"`
	Subject     string   `json:"subject"`
	Html        string   `json:"html,omitempty"`
	Text        string   `json:"text,omitempty"`
	ReplyTo     []string `json:"reply_to,omitempty"`
	PreviewText string   `json:"preview_text,omitempty"`
	Name        string   `json:"name,omitempty"`
	// SegmentId targets a segment (subset of the audience) instead of everyone.
	SegmentId string `json:"segment_id,omitempty"`
	// TopicId gates recipients by a topic subscription.
	TopicId string `json:"topic_id,omitempty"`
	// Recurrence makes this a recurring campaign: "daily" | "weekly" | "monthly".
	Recurrence string `json:"recurrence,omitempty"`
	// RecurrenceEvery is the number of periods between recurring sends
	// (1-365). Defaults to 1.
	RecurrenceEvery int             `json:"recurrence_every,omitempty"`
	AbTest          *CampaignAbTest `json:"ab_test,omitempty"`
	// Followups are engagement follow-ups (max 5).
	Followups []CampaignFollowupInput `json:"followups,omitempty"`
	// ListTo shows a generated mailing-list address as the visible To.
	// Delivery stays individual.
	ListTo bool `json:"list_to,omitempty"`
	// UnsubscribePolicy: "account" (default), "domain", or "ignore"
	// (bounced/complained addresses are ALWAYS excluded).
	UnsubscribePolicy string `json:"unsubscribe_policy,omitempty"`
	// Send sends immediately on create (or schedules when ScheduledAt is given).
	Send bool `json:"send,omitempty"`
	// ScheduledAt is an ISO 8601 (or natural-language) schedule used when
	// Send is true.
	ScheduledAt string `json:"scheduled_at,omitempty"`
}

// UpdateCampaignRequest is the payload for PATCH /campaigns/:id.
type UpdateCampaignRequest struct {
	Name        string   `json:"name,omitempty"`
	From        string   `json:"from,omitempty"`
	Subject     string   `json:"subject,omitempty"`
	Html        string   `json:"html,omitempty"`
	Text        string   `json:"text,omitempty"`
	ReplyTo     []string `json:"reply_to,omitempty"`
	PreviewText string   `json:"preview_text,omitempty"`
	// Domain re-points the campaign at another of your domains' contact pools
	// (draft campaigns only).
	Domain string `json:"domain,omitempty"`
	// SegmentId re-targets a segment.
	SegmentId string `json:"segment_id,omitempty"`
	// TopicId re-targets a topic gate.
	TopicId         string          `json:"topic_id,omitempty"`
	Recurrence      string          `json:"recurrence,omitempty"`
	RecurrenceEvery int             `json:"recurrence_every,omitempty"`
	AbTest          *CampaignAbTest `json:"ab_test,omitempty"`
}

// SendCampaignRequest sends a campaign now, or schedules it with ScheduledAt.
type SendCampaignRequest struct {
	ScheduledAt string `json:"scheduled_at,omitempty"`
}

// CampaignStats is the per-campaign analytics from GET /campaigns/:id/stats.
type CampaignStats struct {
	Object     string           `json:"object"`
	CampaignId string           `json:"campaign_id"`
	Links      []map[string]any `json:"links,omitempty"`
}

// CampaignAbResult is the A/B winner evaluation from GET /campaigns/:id/ab.
type CampaignAbResult struct {
	Object     string `json:"object"`
	CampaignId string `json:"campaign_id"`
	// Metric is "open" | "click" | "reply".
	Metric string `json:"metric"`
	Status string `json:"status,omitempty"`
	Winner string `json:"winner,omitempty"`
}

// CampaignsService handles the /campaigns endpoints.
type CampaignsService struct {
	client *Client
}

// Create creates a campaign; returns { id }. POST /campaigns
func (s *CampaignsService) Create(params *CreateCampaignRequest) (*IdResponse, error) {
	return s.CreateWithContext(context.Background(), params)
}

// CreateWithContext creates a campaign. POST /campaigns
func (s *CampaignsService) CreateWithContext(ctx context.Context, params *CreateCampaignRequest) (*IdResponse, error) {
	return request[IdResponse](ctx, s.client, http.MethodPost, "/campaigns", params, nil)
}

// Get retrieves a campaign. GET /campaigns/:id
func (s *CampaignsService) Get(id string) (*Campaign, error) {
	return s.GetWithContext(context.Background(), id)
}

// GetWithContext retrieves a campaign. GET /campaigns/:id
func (s *CampaignsService) GetWithContext(ctx context.Context, id string) (*Campaign, error) {
	return request[Campaign](ctx, s.client, http.MethodGet, "/campaigns/"+esc(id), nil, nil)
}

// List lists campaigns. GET /campaigns
func (s *CampaignsService) List(params *ListParams) (*ListResponse[Campaign], error) {
	return s.ListWithContext(context.Background(), params)
}

// ListWithContext lists campaigns. GET /campaigns
func (s *CampaignsService) ListWithContext(ctx context.Context, params *ListParams) (*ListResponse[Campaign], error) {
	return request[ListResponse[Campaign]](ctx, s.client, http.MethodGet, listPath("/campaigns", params), nil, nil)
}

// Update updates a campaign; returns { id }. PATCH /campaigns/:id
func (s *CampaignsService) Update(id string, params *UpdateCampaignRequest) (*IdResponse, error) {
	return s.UpdateWithContext(context.Background(), id, params)
}

// UpdateWithContext updates a campaign. PATCH /campaigns/:id
func (s *CampaignsService) UpdateWithContext(ctx context.Context, id string, params *UpdateCampaignRequest) (*IdResponse, error) {
	return request[IdResponse](ctx, s.client, http.MethodPatch, "/campaigns/"+esc(id), params, nil)
}

// Send sends a campaign now, or schedules it with ScheduledAt (nil params =
// send now). POST /campaigns/:id/send
func (s *CampaignsService) Send(id string, params *SendCampaignRequest) (*IdResponse, error) {
	return s.SendWithContext(context.Background(), id, params)
}

// SendWithContext sends a campaign. POST /campaigns/:id/send
func (s *CampaignsService) SendWithContext(ctx context.Context, id string, params *SendCampaignRequest) (*IdResponse, error) {
	body := params
	if body == nil {
		body = &SendCampaignRequest{}
	}
	return request[IdResponse](ctx, s.client, http.MethodPost, "/campaigns/"+esc(id)+"/send", body, nil)
}

// Cancel cancels a scheduled campaign (returns it to draft).
// POST /campaigns/:id/cancel
func (s *CampaignsService) Cancel(id string) (*Campaign, error) {
	return s.CancelWithContext(context.Background(), id)
}

// CancelWithContext cancels a scheduled campaign. POST /campaigns/:id/cancel
func (s *CampaignsService) CancelWithContext(ctx context.Context, id string) (*Campaign, error) {
	return request[Campaign](ctx, s.client, http.MethodPost, "/campaigns/"+esc(id)+"/cancel", nil, nil)
}

// Stats returns per-campaign analytics (counts, engagement rates, top links).
// GET /campaigns/:id/stats
func (s *CampaignsService) Stats(id string) (*CampaignStats, error) {
	return s.StatsWithContext(context.Background(), id)
}

// StatsWithContext returns per-campaign analytics. GET /campaigns/:id/stats
func (s *CampaignsService) StatsWithContext(ctx context.Context, id string) (*CampaignStats, error) {
	return request[CampaignStats](ctx, s.client, http.MethodGet, "/campaigns/"+esc(id)+"/stats", nil, nil)
}

// Ab returns the A/B winner evaluation for an A/B campaign.
// GET /campaigns/:id/ab
func (s *CampaignsService) Ab(id string) (*CampaignAbResult, error) {
	return s.AbWithContext(context.Background(), id)
}

// AbWithContext returns the A/B winner evaluation. GET /campaigns/:id/ab
func (s *CampaignsService) AbWithContext(ctx context.Context, id string) (*CampaignAbResult, error) {
	return request[CampaignAbResult](ctx, s.client, http.MethodGet, "/campaigns/"+esc(id)+"/ab", nil, nil)
}

// Remove deletes a campaign. DELETE /campaigns/:id
func (s *CampaignsService) Remove(id string) (*RemovedResponse, error) {
	return s.RemoveWithContext(context.Background(), id)
}

// RemoveWithContext deletes a campaign. DELETE /campaigns/:id
func (s *CampaignsService) RemoveWithContext(ctx context.Context, id string) (*RemovedResponse, error) {
	return request[RemovedResponse](ctx, s.client, http.MethodDelete, "/campaigns/"+esc(id), nil, nil)
}
