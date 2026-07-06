package mailblastr

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

// Contact is a single contact record. Contacts are DOMAIN-FIRST: each sending
// domain has its own contact pool, so the same email address on two domains
// is two records with separate consent.
type Contact struct {
	Object       string `json:"object"`
	Id           string `json:"id"`
	Email        string `json:"email"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Unsubscribed bool   `json:"unsubscribed"`
	// Properties are custom contact properties (own values merged over
	// registered fallbacks).
	Properties map[string]any `json:"properties,omitempty"`
	CreatedAt  string         `json:"created_at"`
}

// CreateContactRequest creates a contact. Domain-first: set Domain to create
// in that domain's contact pool via the flat /contacts API (REQUIRED there),
// or AudienceId to target a specific audience via the nested API.
type CreateContactRequest struct {
	// Domain is the sending domain whose contact pool the contact lands in
	// (e.g. "yourdomain.com" — one of your domains). REQUIRED whenever
	// AudienceId is omitted.
	Domain string `json:"domain,omitempty"`
	// AudienceId targets a specific audience via the nested
	// /audiences/:id/contacts API instead of Domain. Not sent in the body.
	AudienceId   string         `json:"-"`
	Email        string         `json:"email"`
	FirstName    string         `json:"first_name,omitempty"`
	LastName     string         `json:"last_name,omitempty"`
	Unsubscribed bool           `json:"unsubscribed,omitempty"`
	Properties   map[string]any `json:"properties,omitempty"`
}

// GetContactRequest identifies a contact for retrieval. Id may be a contact
// id (exact) or an EMAIL; an email can exist in several domains' pools, so
// set Domain to pick the pool (omitted => the oldest match anywhere).
type GetContactRequest struct {
	AudienceId string
	Domain     string
	// Id is a contact id OR email.
	Id string
}

// ListContactsRequest lists contacts. Domain-first: Domain is REQUIRED on the
// flat /contacts list (names the pool) whenever AudienceId is omitted.
type ListContactsRequest struct {
	Domain     string
	AudienceId string
	Limit      int
	After      string
	Before     string
	// SegmentId restricts to members of that segment (works with or without
	// AudienceId).
	SegmentId string
}

// ContactInput is a single contact in a bulk import (no audience — it's on
// the call).
type ContactInput struct {
	Email        string         `json:"email"`
	FirstName    string         `json:"first_name,omitempty"`
	LastName     string         `json:"last_name,omitempty"`
	Unsubscribed bool           `json:"unsubscribed,omitempty"`
	Properties   map[string]any `json:"properties,omitempty"`
}

// BatchContactsRequest bulk-imports contacts from a slice. Upserts by email;
// max 10,000 per call.
type BatchContactsRequest struct {
	AudienceId string
	Contacts   []ContactInput
	// OnConflict "skip" leaves existing contacts untouched (default "upsert").
	OnConflict string
}

// ImportContactsRequest bulk-imports contacts from CSV text (header row
// optional). Upserts by email. By default every non-builtin CSV column is
// auto-registered as a custom property; set CreateProperties to
// mailblastr.Bool(false) for strict mode.
type ImportContactsRequest struct {
	AudienceId string
	Csv        string
	// OnConflict "skip" leaves existing contacts untouched (default "upsert").
	OnConflict       string
	CreateProperties *bool
}

// ImportContactsResponse is the result of a batch / CSV contact import.
type ImportContactsResponse struct {
	Object string `json:"object"`
	// Imported counts newly inserted contacts.
	Imported int `json:"imported"`
	// Updated counts existing contacts updated.
	Updated int `json:"updated"`
	// Skipped counts rows dropped: missing/invalid email OR left untouched
	// under OnConflict "skip".
	Skipped int `json:"skipped"`
	// Total counts distinct contacts processed.
	Total int `json:"total"`
}

// UpdateContactRequest updates a contact.
type UpdateContactRequest struct {
	// Domain disambiguates an EMAIL Id across domains (a contact id is exact
	// and needs no domain). Flat API only.
	Domain string `json:"domain,omitempty"`
	// AudienceId is the audience the contact belongs to. OMIT for the flat
	// /contacts/:id API. Not sent in the body.
	AudienceId string `json:"-"`
	// Id is a contact id OR email. Not sent in the body.
	Id           string         `json:"-"`
	FirstName    *string        `json:"first_name,omitempty"`
	LastName     *string        `json:"last_name,omitempty"`
	Unsubscribed *bool          `json:"unsubscribed,omitempty"`
	Properties   map[string]any `json:"properties,omitempty"`
}

// RemoveContactRequest identifies a contact for deletion. On the flat API,
// set Domain when Id is an EMAIL (disambiguates across pools).
type RemoveContactRequest struct {
	AudienceId string
	Domain     string
	// Id is a contact id OR email.
	Id string
}

// RemoveContactResponse is returned by contact deletion (the id is under
// Contact).
type RemoveContactResponse struct {
	Object  string `json:"object"`
	Contact string `json:"contact"`
	Deleted bool   `json:"deleted"`
}

// SegmentMembershipRemoved is returned when removing a contact from a segment.
type SegmentMembershipRemoved struct {
	Id         string `json:"id"`
	AudienceId string `json:"audienceId"`
	Deleted    bool   `json:"deleted"`
}

// ContactTopicSubscription is one topic subscription state of a contact.
type ContactTopicSubscription struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	// Subscription is "opt_in" | "opt_out".
	Subscription string `json:"subscription"`
}

// ContactTopics is a contact's topic subscription list.
type ContactTopics struct {
	Object string                     `json:"object"`
	Data   []ContactTopicSubscription `json:"data"`
}

// TopicSubscriptionUpdate sets one topic's subscription state.
type TopicSubscriptionUpdate struct {
	Id string `json:"id"`
	// Subscription is "opt_in" | "opt_out".
	Subscription string `json:"subscription"`
}

// UpdateContactTopicsRequest updates a contact's topic subscriptions.
type UpdateContactTopicsRequest struct {
	Topics []TopicSubscriptionUpdate `json:"topics"`
}

// ContactsService handles the flat /contacts and nested
// /audiences/:id/contacts endpoints.
type ContactsService struct {
	client *Client
}

// Create creates a contact; returns the slim ack { object, id }.
// POST /contacts (flat, Domain required) or POST /audiences/:id/contacts.
func (s *ContactsService) Create(params *CreateContactRequest) (*ObjectRef, error) {
	return s.CreateWithContext(context.Background(), params)
}

// CreateWithContext creates a contact.
func (s *ContactsService) CreateWithContext(ctx context.Context, params *CreateContactRequest) (*ObjectRef, error) {
	body := *params
	if params.AudienceId != "" {
		// The nested audience route derives its pool from the path; only the
		// flat route takes domain in the body.
		body.Domain = ""
		return request[ObjectRef](ctx, s.client, http.MethodPost, "/audiences/"+esc(params.AudienceId)+"/contacts", &body, nil)
	}
	return request[ObjectRef](ctx, s.client, http.MethodPost, "/contacts", &body, nil)
}

// Get retrieves a contact by id or email. An id is exact; an EMAIL can exist
// in several domains' pools, so set Domain to pick the pool.
func (s *ContactsService) Get(params *GetContactRequest) (*Contact, error) {
	return s.GetWithContext(context.Background(), params)
}

// GetWithContext retrieves a contact by id or email.
func (s *ContactsService) GetWithContext(ctx context.Context, params *GetContactRequest) (*Contact, error) {
	if params.AudienceId != "" {
		return request[Contact](ctx, s.client, http.MethodGet, "/audiences/"+esc(params.AudienceId)+"/contacts/"+esc(params.Id), nil, nil)
	}
	path := "/contacts/" + esc(params.Id)
	if params.Domain != "" {
		path += "?domain=" + url.QueryEscape(params.Domain)
	}
	return request[Contact](ctx, s.client, http.MethodGet, path, nil, nil)
}

// List lists contacts. Domain-first: Domain is required on the flat /contacts
// list (names the pool) whenever AudienceId is omitted.
func (s *ContactsService) List(params *ListContactsRequest) (*ListResponse[Contact], error) {
	return s.ListWithContext(context.Background(), params)
}

// ListWithContext lists contacts.
func (s *ContactsService) ListWithContext(ctx context.Context, params *ListContactsRequest) (*ListResponse[Contact], error) {
	base := "/contacts"
	q := url.Values{}
	if params != nil {
		if params.AudienceId != "" {
			base = "/audiences/" + esc(params.AudienceId) + "/contacts"
		} else if params.Domain != "" {
			q.Set("domain", params.Domain)
		}
		if params.Limit > 0 {
			q.Set("limit", strconv.Itoa(params.Limit))
		}
		if params.After != "" {
			q.Set("after", params.After)
		}
		if params.Before != "" {
			q.Set("before", params.Before)
		}
		// segment_id filter is honored by both the flat and audience-scoped list.
		if params.SegmentId != "" {
			q.Set("segment_id", params.SegmentId)
		}
	}
	path := base
	if enc := q.Encode(); enc != "" {
		path += "?" + enc
	}
	return request[ListResponse[Contact]](ctx, s.client, http.MethodGet, path, nil, nil)
}

// Batch bulk-imports contacts from a slice. Upserts by email; max 10,000 per
// call. POST /audiences/:id/contacts/batch
func (s *ContactsService) Batch(params *BatchContactsRequest) (*ImportContactsResponse, error) {
	return s.BatchWithContext(context.Background(), params)
}

// BatchWithContext bulk-imports contacts from a slice.
func (s *ContactsService) BatchWithContext(ctx context.Context, params *BatchContactsRequest) (*ImportContactsResponse, error) {
	path := "/audiences/" + esc(params.AudienceId) + "/contacts/batch"
	if params.OnConflict != "" {
		path += "?on_conflict=" + url.QueryEscape(params.OnConflict)
	}
	body := map[string]any{"contacts": params.Contacts}
	return request[ImportContactsResponse](ctx, s.client, http.MethodPost, path, body, nil)
}

// Import bulk-imports contacts from CSV text. Upserts by email.
// POST /audiences/:id/contacts/import
func (s *ContactsService) Import(params *ImportContactsRequest) (*ImportContactsResponse, error) {
	return s.ImportWithContext(context.Background(), params)
}

// ImportWithContext bulk-imports contacts from CSV text.
func (s *ContactsService) ImportWithContext(ctx context.Context, params *ImportContactsRequest) (*ImportContactsResponse, error) {
	q := url.Values{}
	if params.OnConflict != "" {
		q.Set("on_conflict", params.OnConflict)
	}
	if params.CreateProperties != nil && !*params.CreateProperties {
		q.Set("create_properties", "false")
	}
	path := "/audiences/" + esc(params.AudienceId) + "/contacts/import"
	if enc := q.Encode(); enc != "" {
		path += "?" + enc
	}
	body := map[string]any{"csv": params.Csv}
	return request[ImportContactsResponse](ctx, s.client, http.MethodPost, path, body, nil)
}

// Update updates a contact; returns the slim ack { object, id }. On the flat
// API, set Domain when Id is an EMAIL (disambiguates across pools).
func (s *ContactsService) Update(params *UpdateContactRequest) (*ObjectRef, error) {
	return s.UpdateWithContext(context.Background(), params)
}

// UpdateWithContext updates a contact.
func (s *ContactsService) UpdateWithContext(ctx context.Context, params *UpdateContactRequest) (*ObjectRef, error) {
	body := *params
	if params.AudienceId != "" {
		// The nested audience route derives its pool from the path.
		body.Domain = ""
		return request[ObjectRef](ctx, s.client, http.MethodPatch, "/audiences/"+esc(params.AudienceId)+"/contacts/"+esc(params.Id), &body, nil)
	}
	return request[ObjectRef](ctx, s.client, http.MethodPatch, "/contacts/"+esc(params.Id), &body, nil)
}

// Remove deletes a contact. On the flat API, set Domain when Id is an EMAIL.
func (s *ContactsService) Remove(params *RemoveContactRequest) (*RemoveContactResponse, error) {
	return s.RemoveWithContext(context.Background(), params)
}

// RemoveWithContext deletes a contact.
func (s *ContactsService) RemoveWithContext(ctx context.Context, params *RemoveContactRequest) (*RemoveContactResponse, error) {
	if params.AudienceId != "" {
		return request[RemoveContactResponse](ctx, s.client, http.MethodDelete, "/audiences/"+esc(params.AudienceId)+"/contacts/"+esc(params.Id), nil, nil)
	}
	path := "/contacts/" + esc(params.Id)
	if params.Domain != "" {
		path += "?domain=" + url.QueryEscape(params.Domain)
	}
	return request[RemoveContactResponse](ctx, s.client, http.MethodDelete, path, nil, nil)
}

// AddToSegment adds a contact to a segment; returns { id } (the segment id).
// POST /contacts/:id/segments/:segmentId
func (s *ContactsService) AddToSegment(id, segmentId string) (*IdResponse, error) {
	return s.AddToSegmentWithContext(context.Background(), id, segmentId)
}

// AddToSegmentWithContext adds a contact to a segment.
func (s *ContactsService) AddToSegmentWithContext(ctx context.Context, id, segmentId string) (*IdResponse, error) {
	return request[IdResponse](ctx, s.client, http.MethodPost, "/contacts/"+esc(id)+"/segments/"+esc(segmentId), nil, nil)
}

// RemoveFromSegment removes a contact from a segment.
// DELETE /contacts/:id/segments/:segmentId
func (s *ContactsService) RemoveFromSegment(id, segmentId string) (*SegmentMembershipRemoved, error) {
	return s.RemoveFromSegmentWithContext(context.Background(), id, segmentId)
}

// RemoveFromSegmentWithContext removes a contact from a segment.
func (s *ContactsService) RemoveFromSegmentWithContext(ctx context.Context, id, segmentId string) (*SegmentMembershipRemoved, error) {
	return request[SegmentMembershipRemoved](ctx, s.client, http.MethodDelete, "/contacts/"+esc(id)+"/segments/"+esc(segmentId), nil, nil)
}

// ListSegments lists the segments a contact belongs to.
// GET /contacts/:id/segments
func (s *ContactsService) ListSegments(id string) (*ListResponse[Segment], error) {
	return s.ListSegmentsWithContext(context.Background(), id)
}

// ListSegmentsWithContext lists the segments a contact belongs to.
func (s *ContactsService) ListSegmentsWithContext(ctx context.Context, id string) (*ListResponse[Segment], error) {
	return request[ListResponse[Segment]](ctx, s.client, http.MethodGet, "/contacts/"+esc(id)+"/segments", nil, nil)
}

// GetTopics gets a contact's topic subscriptions. GET /contacts/:id/topics
func (s *ContactsService) GetTopics(id string) (*ContactTopics, error) {
	return s.GetTopicsWithContext(context.Background(), id)
}

// GetTopicsWithContext gets a contact's topic subscriptions.
func (s *ContactsService) GetTopicsWithContext(ctx context.Context, id string) (*ContactTopics, error) {
	return request[ContactTopics](ctx, s.client, http.MethodGet, "/contacts/"+esc(id)+"/topics", nil, nil)
}

// UpdateTopics updates a contact's topic subscriptions; returns { id } (the
// contact id). PATCH /contacts/:id/topics
func (s *ContactsService) UpdateTopics(id string, params *UpdateContactTopicsRequest) (*IdResponse, error) {
	return s.UpdateTopicsWithContext(context.Background(), id, params)
}

// UpdateTopicsWithContext updates a contact's topic subscriptions.
func (s *ContactsService) UpdateTopicsWithContext(ctx context.Context, id string, params *UpdateContactTopicsRequest) (*IdResponse, error) {
	return request[IdResponse](ctx, s.client, http.MethodPatch, "/contacts/"+esc(id)+"/topics", params, nil)
}
