package mailblastr

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Webhook is a delivery endpoint subscribed to events.
type Webhook struct {
	Object   string   `json:"object"`
	Id       string   `json:"id"`
	Endpoint string   `json:"endpoint"`
	Events   []string `json:"events"`
	Status   string   `json:"status"`
	// HasSecret reports whether a signing secret is set. (The secret itself is
	// returned ONLY on create + rotate, never on get/list.)
	HasSecret bool `json:"has_secret,omitempty"`
	// LastDeliveryAt is the timestamp of the last delivery attempt.
	LastDeliveryAt string `json:"last_delivery_at,omitempty"`
	// LastDeliveryStatus is the HTTP status of the last delivery attempt (0
	// until first delivery).
	LastDeliveryStatus int `json:"last_delivery_status,omitempty"`
	// FailureCount is the consecutive delivery failure count.
	FailureCount int    `json:"failure_count,omitempty"`
	CreatedAt    string `json:"created_at"`
}

// CreateWebhookRequest is the payload for POST /webhooks.
type CreateWebhookRequest struct {
	Endpoint string   `json:"endpoint"`
	Events   []string `json:"events"`
	// Secret is an optional caller-supplied signing secret. When omitted,
	// MailBlastr generates one (returned once).
	Secret string `json:"secret,omitempty"`
}

// CreateWebhookResponse carries the signing secret, shown ONCE (create and
// rotate only).
type CreateWebhookResponse struct {
	Object        string `json:"object"`
	Id            string `json:"id"`
	SigningSecret string `json:"signing_secret"`
}

// UpdateWebhookRequest is the payload for PATCH /webhooks/:id.
type UpdateWebhookRequest struct {
	Endpoint string   `json:"endpoint,omitempty"`
	Events   []string `json:"events,omitempty"`
	// Status is "enabled" | "disabled".
	Status string `json:"status,omitempty"`
}

// WebhookTestResult is the endpoint's live result from a synchronous test
// delivery.
type WebhookTestResult struct {
	Object string `json:"object"`
	Id     string `json:"id"`
	// Delivered reports whether the endpoint accepted the test delivery.
	Delivered bool `json:"delivered,omitempty"`
	// Status is the HTTP status the endpoint returned.
	Status int    `json:"status,omitempty"`
	Error  string `json:"error,omitempty"`
}

// VerifyWebhookResult is the outcome of verifying a webhook delivery signature.
type VerifyWebhookResult struct {
	// Valid is true when the signature matches and (when checked) the
	// timestamp is fresh.
	Valid bool
	// Reason is a machine reason when Valid is false (e.g. "missing_headers",
	// "no_match").
	Reason string
}

// VerifyWebhookOptions configures VerifyWebhookSignature.
type VerifyWebhookOptions struct {
	// ToleranceSec is the max allowed clock skew in seconds. 0 means the
	// default of 300; a negative value skips the timestamp check entirely.
	ToleranceSec int
}

// WebhooksService handles the /webhooks endpoints.
type WebhooksService struct {
	client *Client
}

// Create creates a webhook. The signing secret is shown ONCE, only here.
// POST /webhooks
func (s *WebhooksService) Create(params *CreateWebhookRequest) (*CreateWebhookResponse, error) {
	return s.CreateWithContext(context.Background(), params)
}

// CreateWithContext creates a webhook. POST /webhooks
func (s *WebhooksService) CreateWithContext(ctx context.Context, params *CreateWebhookRequest) (*CreateWebhookResponse, error) {
	return request[CreateWebhookResponse](ctx, s.client, http.MethodPost, "/webhooks", params, nil)
}

// Get retrieves a webhook. GET /webhooks/:id
func (s *WebhooksService) Get(id string) (*Webhook, error) {
	return s.GetWithContext(context.Background(), id)
}

// GetWithContext retrieves a webhook. GET /webhooks/:id
func (s *WebhooksService) GetWithContext(ctx context.Context, id string) (*Webhook, error) {
	return request[Webhook](ctx, s.client, http.MethodGet, "/webhooks/"+esc(id), nil, nil)
}

// List lists webhooks. GET /webhooks
func (s *WebhooksService) List(params *ListParams) (*ListResponse[Webhook], error) {
	return s.ListWithContext(context.Background(), params)
}

// ListWithContext lists webhooks. GET /webhooks
func (s *WebhooksService) ListWithContext(ctx context.Context, params *ListParams) (*ListResponse[Webhook], error) {
	return request[ListResponse[Webhook]](ctx, s.client, http.MethodGet, listPath("/webhooks", params), nil, nil)
}

// Update updates a webhook; returns the slim ack { object, id }.
// PATCH /webhooks/:id
func (s *WebhooksService) Update(id string, params *UpdateWebhookRequest) (*ObjectRef, error) {
	return s.UpdateWithContext(context.Background(), id, params)
}

// UpdateWithContext updates a webhook. PATCH /webhooks/:id
func (s *WebhooksService) UpdateWithContext(ctx context.Context, id string, params *UpdateWebhookRequest) (*ObjectRef, error) {
	return request[ObjectRef](ctx, s.client, http.MethodPatch, "/webhooks/"+esc(id), params, nil)
}

// Rotate rotates the signing secret. The new plaintext SigningSecret is
// returned ONCE (reveal-once); the old secret stops verifying immediately.
// POST /webhooks/:id/rotate
func (s *WebhooksService) Rotate(id string) (*CreateWebhookResponse, error) {
	return s.RotateWithContext(context.Background(), id)
}

// RotateWithContext rotates the signing secret. POST /webhooks/:id/rotate
func (s *WebhooksService) RotateWithContext(ctx context.Context, id string) (*CreateWebhookResponse, error) {
	return request[CreateWebhookResponse](ctx, s.client, http.MethodPost, "/webhooks/"+esc(id)+"/rotate", nil, nil)
}

// Test sends a synchronous test delivery and returns the endpoint's live
// result. POST /webhooks/:id/test
func (s *WebhooksService) Test(id string) (*WebhookTestResult, error) {
	return s.TestWithContext(context.Background(), id)
}

// TestWithContext sends a synchronous test delivery. POST /webhooks/:id/test
func (s *WebhooksService) TestWithContext(ctx context.Context, id string) (*WebhookTestResult, error) {
	return request[WebhookTestResult](ctx, s.client, http.MethodPost, "/webhooks/"+esc(id)+"/test", nil, nil)
}

// Remove deletes a webhook. DELETE /webhooks/:id
func (s *WebhooksService) Remove(id string) (*RemovedResponse, error) {
	return s.RemoveWithContext(context.Background(), id)
}

// RemoveWithContext deletes a webhook. DELETE /webhooks/:id
func (s *WebhooksService) RemoveWithContext(ctx context.Context, id string) (*RemovedResponse, error) {
	return request[RemovedResponse](ctx, s.client, http.MethodDelete, "/webhooks/"+esc(id), nil, nil)
}

// Verify verifies a webhook delivery's Svix-style signature against your
// endpoint's signing secret. See VerifyWebhookSignature — this is a pure
// local computation, no HTTP request is made.
func (s *WebhooksService) Verify(payload []byte, headers http.Header, secret string, opts *VerifyWebhookOptions) VerifyWebhookResult {
	return VerifyWebhookSignature(payload, headers, secret, opts)
}

// secretToKey derives the HMAC key from a "whsec_"-prefixed secret
// (base64-decodes the suffix); a secret without the prefix is used as raw
// UTF-8 bytes.
func secretToKey(secret string) []byte {
	if strings.HasPrefix(secret, "whsec_") {
		if raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(secret, "whsec_")); err == nil && len(raw) > 0 {
			return raw
		}
	}
	return []byte(secret)
}

// VerifyWebhookSignature verifies a MailBlastr webhook delivery's Svix-style
// signature ("<id>.<timestamp>.<body>" -> base64 HMAC-SHA256, tagged "v1,").
//
// payload MUST be the exact raw request body the server sent (do not
// re-serialize the parsed JSON — whitespace differences break the signature).
// headers is typically r.Header from your handler; the svix-id,
// svix-timestamp, and svix-signature headers are read case-insensitively. A
// header may carry multiple space-separated signatures; any one matching
// makes the delivery valid.
//
// The timestamp freshness check defaults to a 5-minute tolerance; pass
// &VerifyWebhookOptions{ToleranceSec: -1} to skip it. This is a pure local
// computation — it makes no HTTP request.
func VerifyWebhookSignature(payload []byte, headers http.Header, secret string, opts *VerifyWebhookOptions) VerifyWebhookResult {
	id := headers.Get("svix-id")
	timestamp := headers.Get("svix-timestamp")
	sigHeader := headers.Get("svix-signature")
	if id == "" || timestamp == "" || sigHeader == "" {
		return VerifyWebhookResult{Valid: false, Reason: "missing_headers"}
	}
	if secret == "" {
		return VerifyWebhookResult{Valid: false, Reason: "missing_secret"}
	}

	// Optional timestamp freshness check (default 5-minute tolerance;
	// negative disables).
	toleranceSec := 300
	if opts != nil && opts.ToleranceSec != 0 {
		toleranceSec = opts.ToleranceSec
	}
	if toleranceSec > 0 {
		ts, err := strconv.ParseInt(timestamp, 10, 64)
		if err != nil {
			return VerifyWebhookResult{Valid: false, Reason: "invalid_timestamp"}
		}
		skew := time.Now().Unix() - ts
		if skew < 0 {
			skew = -skew
		}
		if skew > int64(toleranceSec) {
			return VerifyWebhookResult{Valid: false, Reason: "timestamp_out_of_tolerance"}
		}
	}

	mac := hmac.New(sha256.New, secretToKey(secret))
	mac.Write([]byte(id + "." + timestamp + "."))
	mac.Write(payload)
	expected := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	// The header may contain multiple space-separated "v1,<sig>" entries; any
	// match wins.
	for _, part := range strings.Fields(sigHeader) {
		sig := strings.TrimPrefix(part, "v1,")
		if hmac.Equal([]byte(sig), []byte(expected)) {
			return VerifyWebhookResult{Valid: true}
		}
	}
	return VerifyWebhookResult{Valid: false, Reason: "no_match"}
}
