package mailblastr

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

// PropertyFilter is a single custom-property predicate. Value is required
// for the "eq" and "contains" operators.
type PropertyFilter struct {
	Key string `json:"key"`
	// Operator is "eq" | "contains" | "exists".
	Operator string `json:"operator"`
	// Value is a string or number.
	Value any `json:"value,omitempty"`
}

// SegmentFilter describes which contacts a segment matches.
type SegmentFilter struct {
	// Status is "all" | "subscribed" | "unsubscribed".
	Status          string           `json:"status,omitempty"`
	EmailContains   string           `json:"email_contains,omitempty"`
	PropertyFilters []PropertyFilter `json:"property_filters,omitempty"`
}

// Segment is a saved contact filter within a domain's contact pool.
type Segment struct {
	Object     string        `json:"object"`
	Id         string        `json:"id"`
	AudienceId string        `json:"audience_id"`
	Name       string        `json:"name"`
	Filter     SegmentFilter `json:"filter"`
	CreatedAt  string        `json:"created_at"`
	UpdatedAt  string        `json:"updated_at"`
}

// CreateSegmentRequest is the payload for POST /segments. Domain is REQUIRED
// (domain-first): segment names are unique WITHIN a domain but freely
// reusable across domains; every domain also carries an auto-created
// "General" (all contacts) segment.
type CreateSegmentRequest struct {
	// Domain is REQUIRED — the sending domain this segment belongs to (e.g.
	// "yourdomain.com" — one of your domains).
	Domain string         `json:"domain"`
	Name   string         `json:"name"`
	Filter *SegmentFilter `json:"filter,omitempty"`
}

// ListSegmentsRequest lists a domain's segments. Domain is REQUIRED.
type ListSegmentsRequest struct {
	// Domain is REQUIRED — the sending domain whose segments to list.
	Domain string
	Limit  int
	After  string
	Before string
}

// UpdateSegmentRequest is the payload for PATCH /segments/:id.
type UpdateSegmentRequest struct {
	Name   string         `json:"name,omitempty"`
	Filter *SegmentFilter `json:"filter,omitempty"`
}

// SegmentsService handles the /segments endpoints.
type SegmentsService struct {
	client *Client
}

// Create creates a segment on a sending domain (Domain is required).
// POST /segments
func (s *SegmentsService) Create(params *CreateSegmentRequest) (*Segment, error) {
	return s.CreateWithContext(context.Background(), params)
}

// CreateWithContext creates a segment. POST /segments
func (s *SegmentsService) CreateWithContext(ctx context.Context, params *CreateSegmentRequest) (*Segment, error) {
	return request[Segment](ctx, s.client, http.MethodPost, "/segments", params, nil)
}

// Get retrieves a segment. GET /segments/:id
func (s *SegmentsService) Get(id string) (*Segment, error) {
	return s.GetWithContext(context.Background(), id)
}

// GetWithContext retrieves a segment. GET /segments/:id
func (s *SegmentsService) GetWithContext(ctx context.Context, id string) (*Segment, error) {
	return request[Segment](ctx, s.client, http.MethodGet, "/segments/"+esc(id), nil, nil)
}

// List lists a domain's segments (Domain is required; includes its
// auto-created "General" segment). GET /segments?domain=...
func (s *SegmentsService) List(params *ListSegmentsRequest) (*ListResponse[Segment], error) {
	return s.ListWithContext(context.Background(), params)
}

// ListWithContext lists a domain's segments. GET /segments?domain=...
func (s *SegmentsService) ListWithContext(ctx context.Context, params *ListSegmentsRequest) (*ListResponse[Segment], error) {
	q := url.Values{}
	q.Set("domain", params.Domain)
	if params.Limit > 0 {
		q.Set("limit", strconv.Itoa(params.Limit))
	}
	if params.After != "" {
		q.Set("after", params.After)
	}
	if params.Before != "" {
		q.Set("before", params.Before)
	}
	return request[ListResponse[Segment]](ctx, s.client, http.MethodGet, "/segments?"+q.Encode(), nil, nil)
}

// Contacts previews the contacts a segment currently resolves to.
// GET /segments/:id/contacts
func (s *SegmentsService) Contacts(id string) (*ListResponse[Contact], error) {
	return s.ContactsWithContext(context.Background(), id)
}

// ContactsWithContext previews the contacts a segment resolves to.
func (s *SegmentsService) ContactsWithContext(ctx context.Context, id string) (*ListResponse[Contact], error) {
	return request[ListResponse[Contact]](ctx, s.client, http.MethodGet, "/segments/"+esc(id)+"/contacts", nil, nil)
}

// Update updates a segment. PATCH /segments/:id
func (s *SegmentsService) Update(id string, params *UpdateSegmentRequest) (*Segment, error) {
	return s.UpdateWithContext(context.Background(), id, params)
}

// UpdateWithContext updates a segment. PATCH /segments/:id
func (s *SegmentsService) UpdateWithContext(ctx context.Context, id string, params *UpdateSegmentRequest) (*Segment, error) {
	return request[Segment](ctx, s.client, http.MethodPatch, "/segments/"+esc(id), params, nil)
}

// Remove deletes a segment. DELETE /segments/:id
func (s *SegmentsService) Remove(id string) (*RemovedResponse, error) {
	return s.RemoveWithContext(context.Background(), id)
}

// RemoveWithContext deletes a segment. DELETE /segments/:id
func (s *SegmentsService) RemoveWithContext(ctx context.Context, id string) (*RemovedResponse, error) {
	return request[RemovedResponse](ctx, s.client, http.MethodDelete, "/segments/"+esc(id), nil, nil)
}
