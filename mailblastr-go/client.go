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
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	// Version is the SDK version, sent in the User-Agent header.
	Version = "1.1.0"
	// DefaultBaseURL is the production MailBlastr API host.
	DefaultBaseURL = "https://www.mailblastr.com/api"

	defaultUserAgent = "mailblastr-go/" + Version

	// DefaultTimeout is the per-request timeout applied to every HTTP call.
	DefaultTimeout = 30 * time.Second
	// DefaultMaxRetries is how many times a retryable (429/503) response is
	// retried, i.e. up to DefaultMaxRetries+1 total attempts.
	DefaultMaxRetries = 2

	// maxBackoff caps how long the client will wait between retries.
	maxBackoff = 30 * time.Second
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
	// HTTPClient is the underlying *http.Client. NewClient gives the client its
	// own instance (not http.DefaultClient) so tuning it never affects other
	// callers.
	HTTPClient *http.Client
	// UserAgent is sent with every request.
	UserAgent string

	// Timeout bounds every HTTP request (JSON and raw/binary paths alike),
	// applied per attempt via a request context. Default DefaultTimeout (30s).
	// A value of 0 disables the timeout.
	Timeout time.Duration
	// MaxRetries is how many times a 429 or 503 response is retried before the
	// error is returned. Default DefaultMaxRetries (2); 0 disables retries.
	// Only 429/503 are retried — never other 5xx, network errors, or timeouts,
	// so a non-idempotent write is never silently duplicated.
	MaxRetries int

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
		HTTPClient: &http.Client{Timeout: DefaultTimeout},
		UserAgent:  defaultUserAgent,
		Timeout:    DefaultTimeout,
		MaxRetries: DefaultMaxRetries,
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

// newRequest builds the request and returns the marshaled body bytes alongside
// it so the do-request chokepoint can re-create the body reader on each retry
// (a request body is consumed once per send).
func (c *Client) newRequest(ctx context.Context, method, path string, body any, opts *RequestOptions) (*http.Request, []byte, error) {
	var bodyBytes []byte
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, nil, err
		}
		bodyBytes = b
		reader = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, reader)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("User-Agent", c.UserAgent)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if opts != nil && opts.IdempotencyKey != "" {
		req.Header.Set("Idempotency-Key", opts.IdempotencyKey)
	}
	return req, bodyBytes, nil
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

// maxRetries returns the configured retry budget, clamping negatives to 0.
func (c *Client) maxRetries() int {
	if c.MaxRetries < 0 {
		return 0
	}
	return c.MaxRetries
}

// isRetryable reports whether a status code may be safely retried. Only 429 and
// 503 qualify: the server guarantees the request was not applied, so retrying
// cannot duplicate a non-idempotent side effect (e.g. sending an email twice).
func isRetryable(status int) bool {
	return status == http.StatusTooManyRequests || status == http.StatusServiceUnavailable
}

// capDelay clamps a wait duration to [0, maxBackoff].
func capDelay(d time.Duration) time.Duration {
	if d < 0 {
		return 0
	}
	if d > maxBackoff {
		return maxBackoff
	}
	return d
}

// backoffDelay computes how long to wait before the next retry. It honors the
// Retry-After header — delta-seconds (integer or decimal) or an HTTP-date —
// capping at maxBackoff and treating negatives as 0. When Retry-After is absent
// or unparseable it falls back to exponential backoff min(30s, 0.5s*2^attempt),
// where attempt is 0 for the first retry.
func backoffDelay(retryAfter string, attempt int) time.Duration {
	if h := strings.TrimSpace(retryAfter); h != "" {
		if secs, err := strconv.ParseFloat(h, 64); err == nil {
			return capDelay(time.Duration(secs * float64(time.Second)))
		}
		if t, err := http.ParseTime(h); err == nil {
			return capDelay(time.Until(t))
		}
	}
	return capDelay(time.Duration(float64(500*time.Millisecond) * math.Pow(2, float64(attempt))))
}

// doRequest is the single HTTP chokepoint shared by the JSON (do) and raw
// (requestRaw) paths, so both get the per-request timeout and the 429/503
// retry loop. reqBody is the buffered request body (nil for bodyless calls);
// it is re-wrapped in a fresh reader before each attempt because a body is
// consumed once per send.
func (c *Client) doRequest(req *http.Request, reqBody []byte) (int, []byte, error) {
	maxRetries := c.maxRetries()
	for attempt := 0; ; attempt++ {
		if reqBody != nil {
			req.Body = io.NopCloser(bytes.NewReader(reqBody))
			req.ContentLength = int64(len(reqBody))
		}

		status, header, body, err := c.attempt(req)
		if err != nil {
			// Network errors and timeouts are not retried.
			return 0, nil, err
		}
		if !isRetryable(status) || attempt >= maxRetries {
			return status, body, nil
		}
		time.Sleep(backoffDelay(header.Get("Retry-After"), attempt))
	}
}

// attempt performs a single HTTP round-trip with the per-request timeout applied
// as a context deadline, fully reading (and closing) the response body before it
// returns so the deadline covers the body read too.
func (c *Client) attempt(req *http.Request) (int, http.Header, []byte, error) {
	ctx := req.Context()
	if c.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.Timeout)
		defer cancel()
	}
	resp, err := c.httpClient().Do(req.WithContext(ctx))
	if err != nil {
		return 0, nil, nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, nil, err
	}
	return resp.StatusCode, resp.Header, body, nil
}

// do executes the request through the chokepoint; on 2xx it unmarshals the JSON
// body into v (when v is non-nil), on non-2xx it returns a *MailblastrError.
func (c *Client) do(req *http.Request, reqBody []byte, v any) error {
	status, body, err := c.doRequest(req, reqBody)
	if err != nil {
		return err
	}
	if status < 200 || status >= 300 {
		return parseAPIError(status, body)
	}
	if v == nil || len(body) == 0 {
		return nil
	}
	return json.Unmarshal(body, v)
}

// request is the shared typed JSON call helper used by every service.
func request[T any](ctx context.Context, c *Client, method, path string, body any, opts *RequestOptions) (*T, error) {
	req, reqBody, err := c.newRequest(ctx, method, path, body, opts)
	if err != nil {
		return nil, err
	}
	out := new(T)
	if err := c.do(req, reqBody, out); err != nil {
		return nil, err
	}
	return out, nil
}

// requestRaw is like request but for endpoints that stream raw binary bytes
// (e.g. a received-email attachment download). On error the JSON error body
// is parsed like request.
func requestRaw(ctx context.Context, c *Client, method, path string) ([]byte, error) {
	req, _, err := c.newRequest(ctx, method, path, nil, nil)
	if err != nil {
		return nil, err
	}
	status, body, err := c.doRequest(req, nil)
	if err != nil {
		return nil, err
	}
	if status < 200 || status >= 300 {
		return nil, parseAPIError(status, body)
	}
	return body, nil
}
