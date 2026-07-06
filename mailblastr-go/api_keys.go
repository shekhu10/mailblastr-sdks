package mailblastr

import (
	"context"
	"net/http"
)

// ApiKey is an API key as returned by ApiKeys.List (GET /api-keys).
type ApiKey struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	// Token is the non-secret display prefix of the key (e.g.
	// "mb_live_abcd…"); empty for legacy keys with no stored prefix. The full
	// secret is returned only once, at creation.
	Token string `json:"token"`
	// Permission is derived from the key's scopes: "full_access" |
	// "sending_access".
	Permission string `json:"permission"`
	// DomainId is set when the key is scoped to a single sending domain.
	DomainId  string `json:"domain_id"`
	CreatedAt string `json:"created_at"`
	// LastUsedAt is the last time the key authenticated a request; empty if
	// never used.
	LastUsedAt string `json:"last_used_at"`
}

// CreateApiKeyRequest is the payload for POST /api-keys.
type CreateApiKeyRequest struct {
	Name string `json:"name"`
	// Permission is "full_access" | "sending_access".
	Permission string `json:"permission,omitempty"`
	// DomainId scopes a sending_access key to one domain (ignored otherwise).
	DomainId string `json:"domain_id,omitempty"`
}

// CreateApiKeyResponse carries the full secret Token, returned only once.
type CreateApiKeyResponse struct {
	Object   string `json:"object"`
	Id       string `json:"id"`
	Token    string `json:"token"`
	DomainId string `json:"domain_id"`
}

// ApiKeysService handles the /api-keys endpoints.
type ApiKeysService struct {
	client *Client
}

// Create creates an API key. The full secret Token is returned only here.
// POST /api-keys
func (s *ApiKeysService) Create(params *CreateApiKeyRequest) (*CreateApiKeyResponse, error) {
	return s.CreateWithContext(context.Background(), params)
}

// CreateWithContext creates an API key. POST /api-keys
func (s *ApiKeysService) CreateWithContext(ctx context.Context, params *CreateApiKeyRequest) (*CreateApiKeyResponse, error) {
	return request[CreateApiKeyResponse](ctx, s.client, http.MethodPost, "/api-keys", params, nil)
}

// List lists API keys. GET /api-keys
func (s *ApiKeysService) List() (*ListResponse[ApiKey], error) {
	return s.ListWithContext(context.Background())
}

// ListWithContext lists API keys. GET /api-keys
func (s *ApiKeysService) ListWithContext(ctx context.Context) (*ListResponse[ApiKey], error) {
	return request[ListResponse[ApiKey]](ctx, s.client, http.MethodGet, "/api-keys", nil, nil)
}

// Remove deletes an API key. DELETE /api-keys/:id
func (s *ApiKeysService) Remove(id string) (*RemovedResponse, error) {
	return s.RemoveWithContext(context.Background(), id)
}

// RemoveWithContext deletes an API key. DELETE /api-keys/:id
func (s *ApiKeysService) RemoveWithContext(ctx context.Context, id string) (*RemovedResponse, error) {
	return request[RemovedResponse](ctx, s.client, http.MethodDelete, "/api-keys/"+esc(id), nil, nil)
}
