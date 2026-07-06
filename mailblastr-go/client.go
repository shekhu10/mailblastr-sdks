// Package mailblastr is the official Go SDK for the MailBlastr email API —
// send transactional and marketing email from your own verified domain.
//
// Usage:
//
//	client := mailblastr.NewClient("mb_xxxxxxxxx")
//	sent, err := client.Emails.Send(&mailblastr.SendEmailRequest{
//		From:    "Acme <hello@yourdomain.com>",
//		To:      []string{"user@example.com"},
//		Subject: "Hello from MailBlastr",
//		Html:    "<p>Your first email</p>",
//	})
//
// Every method has a context-aware variant (e.g. SendWithContext). Non-2xx
// API responses are returned as a *MailblastrError.
package mailblastr

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

const (
	// Version is the SDK version, sent in the User-Agent header.
	Version = "1.0.0"
	// DefaultBaseURL is the production MailBlastr API host.
	DefaultBaseURL = "https://api.mailblastr.com"

	defaultUserAgent = "mailblastr-go/" + Version
)

// MailblastrError is the error type returned for any non-2xx API response.
// It is parsed from the API's error shape: { statusCode, name, message }.
type MailblastrError struct {
	StatusCode int    `json:"statusCode"`
	Name       string `json:"name"`
	Message    string `json:"message"`
}

func (e *MailblastrError) Error() string {
	return fmt.Sprintf("mailblastr: [%d %s] %s", e.StatusCode, e.Name, e.Message)
}

// ObjectRef is the slim acknowledgement { object, id } returned by several
// create/update routes.
type ObjectRef struct {
	Object string `json:"object"`
	Id     string `json:"id"`
}

// IdResponse is the minimal { id } acknowledgement some routes return.
type IdResponse struct {
	Id string `json:"id"`
}

// RemovedResponse is returned by delete endpoints.
type RemovedResponse struct {
	Object  string `json:"object"`
	Id      string `json:"id"`
	Deleted bool   `json:"deleted"`
}

// ListResponse is the standard paginated list envelope.
type ListResponse[T any] struct {
	Object  string `json:"object"`
	HasMore bool   `json:"has_more"`
	Data    []T    `json:"data"`
}

// ListParams are the cursor-pagination params accepted by most List methods.
// Zero values are omitted from the query string.
type ListParams struct {
	Limit int
	// After is the id of the last item on the previous page.
	After string
	// Before is the id of the first item on the next page.
	Before string
}

func (p *ListParams) apply(q url.Values) {
	if p == nil {
		return
	}
	if p.Limit > 0 {
		q.Set("limit", strconv.Itoa(p.Limit))
	}
	if p.After != "" {
		q.Set("after", p.After)
	}
	if p.Before != "" {
		q.Set("before", p.Before)
	}
}

// listPath appends ListParams as a query string to base.
func listPath(base string, p *ListParams) string {
	q := url.Values{}
	p.apply(q)
	if enc := q.Encode(); enc != "" {
		return base + "?" + enc
	}
	return base
}

// RequestOptions carries per-request options for create calls.
type RequestOptions struct {
	// IdempotencyKey is sent as the Idempotency-Key header, letting you safely
	// retry a create (24h window).
	IdempotencyKey string
}

// Bool returns a pointer to b. Handy for optional tri-state fields.
func Bool(b bool) *bool { return &b }

// Int returns a pointer to i.
func Int(i int) *int { return &i }

// String returns a pointer to s.
func String(s string) *string { return &s }

// esc percent-encodes a path segment so ids like "a/../b" can't traverse the
// URL path.
func esc(s string) string { return url.PathEscape(s) }

// Client is the MailBlastr API client. Construct it with NewClient; the
// exported BaseURL, HTTPClient, and UserAgent fields may be overridden before
// first use.
type Client struct {
	apiKey string

	// BaseURL is the API host (default DefaultBaseURL). No trailing slash.
	BaseURL string
	// HTTPClient is the underlying *http.Client (default http.DefaultClient).
	HTTPClient *http.Client
	// UserAgent is sent with every request.
	UserAgent string

	Emails            *EmailsService
	Batch             *BatchService
	Domains           *DomainsService
	Audiences         *AudiencesService
	Contacts          *ContactsService
	ContactProperties *ContactPropertiesService
	Campaigns         *CampaignsService
	Segments          *SegmentsService
	Topics            *TopicsService
	Templates         *TemplatesService
	Automations       *AutomationsService
	Webhooks          *WebhooksService
	Events            *EventsService
	ApiKeys           *ApiKeysService
	Logs              *LogsService
	Polls             *PollsService
}

// NewClient creates a MailBlastr API client authenticated with the given API
// key (e.g. "mb_xxxxxxxxx").
func NewClient(apiKey string) *Client {
	c := &Client{
		apiKey:     apiKey,
		BaseURL:    DefaultBaseURL,
		HTTPClient: http.DefaultClient,
		UserAgent:  defaultUserAgent,
	}
	c.Emails = &EmailsService{client: c, Receiving: &ReceivingService{client: c}}
	c.Batch = &BatchService{client: c}
	c.Domains = &DomainsService{client: c}
	c.Audiences = &AudiencesService{client: c}
	c.Contacts = &ContactsService{client: c}
	c.ContactProperties = &ContactPropertiesService{client: c}
	c.Campaigns = &CampaignsService{client: c}
	c.Segments = &SegmentsService{client: c}
	c.Topics = &TopicsService{client: c}
	c.Templates = &TemplatesService{client: c}
	c.Automations = &AutomationsService{client: c}
	c.Webhooks = &WebhooksService{client: c}
	c.Events = &EventsService{client: c}
	c.ApiKeys = &ApiKeysService{client: c}
	c.Logs = &LogsService{client: c}
	c.Polls = &PollsService{client: c}
	return c
}

func (c *Client) newRequest(ctx context.Context, method, path string, body any, opts *RequestOptions) (*http.Request, error) {
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("User-Agent", c.UserAgent)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if opts != nil && opts.IdempotencyKey != "" {
		req.Header.Set("Idempotency-Key", opts.IdempotencyKey)
	}
	return req, nil
}

func (c *Client) httpClient() *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}
	return http.DefaultClient
}

func parseAPIError(status int, body []byte) *MailblastrError {
	apiErr := &MailblastrError{
		StatusCode: status,
		Name:       "application_error",
		Message:    fmt.Sprintf("request failed with status %d", status),
	}
	var parsed MailblastrError
	if json.Unmarshal(body, &parsed) == nil {
		if parsed.StatusCode != 0 {
			apiErr.StatusCode = parsed.StatusCode
		}
		if parsed.Name != "" {
			apiErr.Name = parsed.Name
		}
		if parsed.Message != "" {
			apiErr.Message = parsed.Message
		}
	}
	return apiErr
}

// do executes the request; on 2xx it unmarshals the JSON body into v (when v
// is non-nil), on non-2xx it returns a *MailblastrError.
func (c *Client) do(req *http.Request, v any) error {
	resp, err := c.httpClient().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return parseAPIError(resp.StatusCode, body)
	}
	if v == nil || len(body) == 0 {
		return nil
	}
	return json.Unmarshal(body, v)
}

// request is the shared typed JSON call helper used by every service.
func request[T any](ctx context.Context, c *Client, method, path string, body any, opts *RequestOptions) (*T, error) {
	req, err := c.newRequest(ctx, method, path, body, opts)
	if err != nil {
		return nil, err
	}
	out := new(T)
	if err := c.do(req, out); err != nil {
		return nil, err
	}
	return out, nil
}

// requestRaw is like request but for endpoints that stream raw binary bytes
// (e.g. a received-email attachment download). On error the JSON error body
// is parsed like request.
func requestRaw(ctx context.Context, c *Client, method, path string) ([]byte, error) {
	req, err := c.newRequest(ctx, method, path, nil, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, parseAPIError(resp.StatusCode, body)
	}
	return body, nil
}
