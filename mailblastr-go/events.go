package mailblastr

import (
	"context"
	"net/http"
)

// SendEventRequest is the payload for POST /events/send. Domain is REQUIRED
// (domain-first): only automations belonging to that domain are triggered, so
// the same event name across several products can never double-fire.
type SendEventRequest struct {
	// Event is the custom event name automations can trigger on. (Name is
	// accepted as an alias.)
	Event string `json:"event,omitempty"`
	// Name is an alias for Event.
	Name string `json:"name,omitempty"`
	// Domain is REQUIRED — the sending domain this event belongs to (e.g.
	// "yourdomain.com" — one of your verified domains). Contacts auto-created
	// by this event land in the domain's own contact pool, with unsubscribe
	// state separate per domain.
	Domain string `json:"domain"`
	// ContactId identifies the contact by id. Provide ContactId OR Email.
	ContactId string `json:"contact_id,omitempty"`
	// Email identifies the contact by email. Provide ContactId OR Email.
	Email string `json:"email,omitempty"`
	// Payload is the arbitrary event payload. (Data is accepted as an alias.)
	Payload map[string]any `json:"payload,omitempty"`
	// Data is an alias for Payload.
	Data map[string]any `json:"data,omitempty"`
}

// SendEventResponse acknowledges an ingested event.
type SendEventResponse struct {
	Object string `json:"object"`
	Id     string `json:"id"`
	// Event is the event name that was ingested.
	Event string `json:"event,omitempty"`
	// ContactId is the resolved contact id the event was attributed to.
	ContactId string `json:"contact_id,omitempty"`
	// Enrolled is the number of automations the event enrolled the contact into.
	Enrolled int `json:"enrolled,omitempty"`
}

// CreateEventRequest creates a custom-event definition.
type CreateEventRequest struct {
	// Name is the custom event name (cannot start with the reserved
	// "mailblastr:" prefix).
	Name string `json:"name"`
	// Schema is an optional flat key -> type map; types: "string" | "number"
	// | "boolean" | "date".
	Schema map[string]string `json:"schema,omitempty"`
}

// EventDefinition is a registered custom-event definition.
type EventDefinition struct {
	Object    string            `json:"object"`
	Id        string            `json:"id"`
	Name      string            `json:"name"`
	Schema    map[string]string `json:"schema"`
	CreatedAt string            `json:"created_at"`
	UpdatedAt string            `json:"updated_at"`
}

// EventsService handles the /events endpoints (automation custom events).
type EventsService struct {
	client *Client
}

// Send sends a custom event that automations can trigger on (Domain is
// required). POST /events/send
func (s *EventsService) Send(params *SendEventRequest) (*SendEventResponse, error) {
	return s.SendWithContext(context.Background(), params)
}

// SendWithContext sends a custom event. POST /events/send
func (s *EventsService) SendWithContext(ctx context.Context, params *SendEventRequest) (*SendEventResponse, error) {
	return request[SendEventResponse](ctx, s.client, http.MethodPost, "/events/send", params, nil)
}

// SendWithOptions sends a custom event with per-request options (e.g. an
// Idempotency-Key). POST /events/send
func (s *EventsService) SendWithOptions(ctx context.Context, params *SendEventRequest, opts *RequestOptions) (*SendEventResponse, error) {
	return request[SendEventResponse](ctx, s.client, http.MethodPost, "/events/send", params, opts)
}

// Create creates a custom-event definition (name + optional payload schema).
// POST /events
func (s *EventsService) Create(params *CreateEventRequest) (*EventDefinition, error) {
	return s.CreateWithContext(context.Background(), params)
}

// CreateWithContext creates a custom-event definition. POST /events
func (s *EventsService) CreateWithContext(ctx context.Context, params *CreateEventRequest) (*EventDefinition, error) {
	return request[EventDefinition](ctx, s.client, http.MethodPost, "/events", params, nil)
}

// List lists custom-event definitions. GET /events
func (s *EventsService) List(params *ListParams) (*ListResponse[EventDefinition], error) {
	return s.ListWithContext(context.Background(), params)
}

// ListWithContext lists custom-event definitions. GET /events
func (s *EventsService) ListWithContext(ctx context.Context, params *ListParams) (*ListResponse[EventDefinition], error) {
	return request[ListResponse[EventDefinition]](ctx, s.client, http.MethodGet, listPath("/events", params), nil, nil)
}

// Remove deletes a custom-event definition. DELETE /events/:id
func (s *EventsService) Remove(id string) (*RemovedResponse, error) {
	return s.RemoveWithContext(context.Background(), id)
}

// RemoveWithContext deletes a custom-event definition. DELETE /events/:id
func (s *EventsService) RemoveWithContext(ctx context.Context, id string) (*RemovedResponse, error) {
	return request[RemovedResponse](ctx, s.client, http.MethodDelete, "/events/"+esc(id), nil, nil)
}
