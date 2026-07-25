package mailblastr

import (
	"context"
	"net/http"
)

// AutomationStep is one step of an automation's graph.
type AutomationStep struct {
	// Id is absent on the synthesized trigger step.
	Id       string         `json:"id,omitempty"`
	Key      string         `json:"key"`
	Type     string         `json:"type"`
	Position int            `json:"position,omitempty"`
	Config   map[string]any `json:"config"`
}

// AutomationConnection is a typed edge between two step keys.
type AutomationConnection struct {
	From string `json:"from"`
	To   string `json:"to"`
	Type string `json:"type"`
}

// Automation is a multi-step workflow triggered by an event.
type Automation struct {
	Object     string `json:"object"`
	Id         string `json:"id"`
	AudienceId string `json:"audience_id"`
	Name       string `json:"name"`
	Trigger    string `json:"trigger"`
	// Domain is the sending domain this automation belongs to. Only
	// Events.Send calls with the same Domain trigger it.
	Domain string `json:"domain"`
	// Status is "enabled" | "disabled".
	Status      string                 `json:"status"`
	Steps       []AutomationStep       `json:"steps,omitempty"`
	Connections []AutomationConnection `json:"connections,omitempty"`
	// Enrollments holds enrollment counts — included only on retrieve.
	Enrollments *AutomationEnrollments `json:"enrollments,omitempty"`
	CreatedAt   string                 `json:"created_at"`
	UpdatedAt   string                 `json:"updated_at"`
}

// AutomationEnrollments are the counts on GET /automations/:id.
type AutomationEnrollments struct {
	Active    int `json:"active"`
	Completed int `json:"completed"`
}

// AutomationStepInput is an inline step accepted on automation create.
type AutomationStepInput struct {
	// Key optionally names the step for connections.
	Key    string         `json:"key,omitempty"`
	Type   string         `json:"type"`
	Config map[string]any `json:"config,omitempty"`
}

// AutomationConnectionInput is a typed edge accepted on create/update.
type AutomationConnectionInput struct {
	From string `json:"from"`
	To   string `json:"to"`
	Type string `json:"type,omitempty"`
}

// AutomationTriggerConfig is the config for the "mailblastr:schedule"
// trigger: the automation fires ONCE at At, enrolling every contact of its
// domain's pool. Required with that trigger; not accepted on any other.
type AutomationTriggerConfig struct {
	// At is the ISO 8601 instant the automation fires (future, at most 366
	// days ahead).
	At string `json:"at"`
	// Timezone is the IANA timezone the schedule was picked in (e.g.
	// "America/New_York").
	Timezone string `json:"timezone"`
}

// CreateAutomationRequest is the payload for POST /automations. Domain is
// REQUIRED (domain-first): only Events.Send calls with the same Domain
// trigger the automation.
type CreateAutomationRequest struct {
	Name string `json:"name"`
	// Domain is REQUIRED — the sending domain this automation belongs to
	// (e.g. "yourdomain.com" — one of your domains).
	Domain string `json:"domain"`
	// Trigger is the event that starts a run: "contact.created", the built-in
	// scheduled trigger "mailblastr:schedule" (requires TriggerConfig), an
	// engagement event ("email.opened", "email.clicked", "email.replied",
	// "email.bounced", "email.delivered"), or any custom event name you send
	// via Events.Send. Usually supplied as a Steps[0] trigger step instead.
	Trigger string `json:"trigger,omitempty"`
	// TriggerConfig is the schedule for the "mailblastr:schedule" trigger
	// ({at, timezone}). Required with that trigger; not accepted on any other.
	TriggerConfig *AutomationTriggerConfig `json:"trigger_config,omitempty"`
	// Status is the initial status: "enabled" | "disabled" (default "disabled").
	Status string `json:"status,omitempty"`
	// Steps is an optional inline step graph.
	Steps []AutomationStepInput `json:"steps,omitempty"`
	// Connections are optional typed edges between step keys.
	Connections []AutomationConnectionInput `json:"connections,omitempty"`
}

// UpdateAutomationRequest is the payload for PATCH /automations/:id.
type UpdateAutomationRequest struct {
	Name string `json:"name,omitempty"`
	// Status is "enabled" | "disabled".
	Status string `json:"status,omitempty"`
	// Domain re-points the automation at another of your domains (disabled
	// automations only).
	Domain string `json:"domain,omitempty"`
	// TriggerConfig updates the "mailblastr:schedule" trigger's schedule
	// ({at, timezone}). Only valid on automations with that trigger.
	TriggerConfig *AutomationTriggerConfig    `json:"trigger_config,omitempty"`
	Connections   []AutomationConnectionInput `json:"connections,omitempty"`
}

// AddAutomationStepRequest appends a step to an automation.
type AddAutomationStepRequest struct {
	Type   string         `json:"type"`
	Config map[string]any `json:"config,omitempty"`
	Key    string         `json:"key,omitempty"`
}

// DeleteAutomationStepResponse is returned by Automations.DeleteStep.
type DeleteAutomationStepResponse struct {
	Id      string `json:"id"`
	Deleted bool   `json:"deleted"`
}

// AutomationRunStep is one executed step of an automation run trace.
type AutomationRunStep struct {
	Key  string `json:"key"`
	Type string `json:"type"`
	// Status is "completed" | "failed" | "skipped".
	Status      string         `json:"status"`
	StartedAt   string         `json:"started_at"`
	CompletedAt string         `json:"completed_at"`
	Output      map[string]any `json:"output"`
	Error       string         `json:"error"`
}

// AutomationRun is one enrollment/run of an automation.
type AutomationRun struct {
	Object    string `json:"object"`
	Id        string `json:"id"`
	ContactId string `json:"contact_id"`
	// ContactEmail is the email of the contact the run is for; empty if that
	// contact was deleted.
	ContactEmail string `json:"contact_email"`
	// Status is "running" | "completed" | "failed" | "cancelled" | "skipped".
	Status      string `json:"status"`
	StartedAt   string `json:"started_at"`
	CompletedAt string `json:"completed_at"`
	CreatedAt   string `json:"created_at"`
	// AutomationId is present on GET /automations/:id/runs/:runId only.
	AutomationId string              `json:"automation_id,omitempty"`
	Steps        []AutomationRunStep `json:"steps,omitempty"`
	Error        string              `json:"error,omitempty"`
}

// AutomationsService handles the /automations endpoints.
type AutomationsService struct {
	client *Client
}

// Create creates an automation (Domain is required). POST /automations
func (s *AutomationsService) Create(params *CreateAutomationRequest) (*Automation, error) {
	return s.CreateWithContext(context.Background(), params)
}

// CreateWithContext creates an automation. POST /automations
func (s *AutomationsService) CreateWithContext(ctx context.Context, params *CreateAutomationRequest) (*Automation, error) {
	return request[Automation](ctx, s.client, http.MethodPost, "/automations", params, nil)
}

// Get retrieves an automation. GET /automations/:id
func (s *AutomationsService) Get(id string) (*Automation, error) {
	return s.GetWithContext(context.Background(), id)
}

// GetWithContext retrieves an automation. GET /automations/:id
func (s *AutomationsService) GetWithContext(ctx context.Context, id string) (*Automation, error) {
	return request[Automation](ctx, s.client, http.MethodGet, "/automations/"+esc(id), nil, nil)
}

// List lists automations. GET /automations
func (s *AutomationsService) List(params *ListParams) (*ListResponse[Automation], error) {
	return s.ListWithContext(context.Background(), params)
}

// ListWithContext lists automations. GET /automations
func (s *AutomationsService) ListWithContext(ctx context.Context, params *ListParams) (*ListResponse[Automation], error) {
	return request[ListResponse[Automation]](ctx, s.client, http.MethodGet, listPath("/automations", params), nil, nil)
}

// Update updates an automation. PATCH /automations/:id
func (s *AutomationsService) Update(id string, params *UpdateAutomationRequest) (*Automation, error) {
	return s.UpdateWithContext(context.Background(), id, params)
}

// UpdateWithContext updates an automation. PATCH /automations/:id
func (s *AutomationsService) UpdateWithContext(ctx context.Context, id string, params *UpdateAutomationRequest) (*Automation, error) {
	return request[Automation](ctx, s.client, http.MethodPatch, "/automations/"+esc(id), params, nil)
}

// AddStep appends a step to an automation; returns the created step.
// POST /automations/:id/steps
func (s *AutomationsService) AddStep(id string, params *AddAutomationStepRequest) (*AutomationStep, error) {
	return s.AddStepWithContext(context.Background(), id, params)
}

// AddStepWithContext appends a step to an automation.
func (s *AutomationsService) AddStepWithContext(ctx context.Context, id string, params *AddAutomationStepRequest) (*AutomationStep, error) {
	return request[AutomationStep](ctx, s.client, http.MethodPost, "/automations/"+esc(id)+"/steps", params, nil)
}

// DeleteStep deletes a step from an automation.
// DELETE /automations/:id/steps/:stepId
func (s *AutomationsService) DeleteStep(id, stepId string) (*DeleteAutomationStepResponse, error) {
	return s.DeleteStepWithContext(context.Background(), id, stepId)
}

// DeleteStepWithContext deletes a step from an automation.
func (s *AutomationsService) DeleteStepWithContext(ctx context.Context, id, stepId string) (*DeleteAutomationStepResponse, error) {
	return request[DeleteAutomationStepResponse](ctx, s.client, http.MethodDelete, "/automations/"+esc(id)+"/steps/"+esc(stepId), nil, nil)
}

// Runs lists an automation's runs. GET /automations/:id/runs
func (s *AutomationsService) Runs(id string, params *ListParams) (*ListResponse[AutomationRun], error) {
	return s.RunsWithContext(context.Background(), id, params)
}

// RunsWithContext lists an automation's runs. GET /automations/:id/runs
func (s *AutomationsService) RunsWithContext(ctx context.Context, id string, params *ListParams) (*ListResponse[AutomationRun], error) {
	return request[ListResponse[AutomationRun]](ctx, s.client, http.MethodGet, listPath("/automations/"+esc(id)+"/runs", params), nil, nil)
}

// GetRun retrieves a single automation run. GET /automations/:id/runs/:runId
func (s *AutomationsService) GetRun(id, runId string) (*AutomationRun, error) {
	return s.GetRunWithContext(context.Background(), id, runId)
}

// GetRunWithContext retrieves a single automation run.
func (s *AutomationsService) GetRunWithContext(ctx context.Context, id, runId string) (*AutomationRun, error) {
	return request[AutomationRun](ctx, s.client, http.MethodGet, "/automations/"+esc(id)+"/runs/"+esc(runId), nil, nil)
}

// Stop stops an automation — prevents new runs; in-progress runs finish.
// POST /automations/:id/stop
func (s *AutomationsService) Stop(id string) (*Automation, error) {
	return s.StopWithContext(context.Background(), id)
}

// StopWithContext stops an automation. POST /automations/:id/stop
func (s *AutomationsService) StopWithContext(ctx context.Context, id string) (*Automation, error) {
	return request[Automation](ctx, s.client, http.MethodPost, "/automations/"+esc(id)+"/stop", nil, nil)
}

// Remove deletes an automation. DELETE /automations/:id
func (s *AutomationsService) Remove(id string) (*RemovedResponse, error) {
	return s.RemoveWithContext(context.Background(), id)
}

// RemoveWithContext deletes an automation. DELETE /automations/:id
func (s *AutomationsService) RemoveWithContext(ctx context.Context, id string) (*RemovedResponse, error) {
	return request[RemovedResponse](ctx, s.client, http.MethodDelete, "/automations/"+esc(id), nil, nil)
}
