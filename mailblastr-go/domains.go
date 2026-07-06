package mailblastr

import (
	"context"
	"net/http"
)

// DomainRecord is one DNS record MailBlastr asks you to create.
type DomainRecord struct {
	// Record is the record's role: DKIM, SPF, DMARC, Tracking, or Receiving MX.
	Record string `json:"record"`
	Name   string `json:"name"`
	// Type is the DNS record type: CNAME, MX, TXT, or CAA.
	Type     string `json:"type"`
	Ttl      string `json:"ttl"`
	Status   string `json:"status"`
	Value    string `json:"value"`
	Priority int    `json:"priority,omitempty"`
}

// DomainCapabilities reports sending/receiving capability state
// ("enabled" | "disabled").
type DomainCapabilities struct {
	Sending   string `json:"sending"`
	Receiving string `json:"receiving"`
}

// Domain is a sending domain.
type Domain struct {
	Object string `json:"object"`
	Id     string `json:"id"`
	Name   string `json:"name"`
	// Status is the aggregated verification status: not_started, pending,
	// verified, partially_verified, partially_failed, failed,
	// temporary_failure, or revoked.
	Status string `json:"status"`
	Region string `json:"region"`
	// Zone is the DNS zone the registrar manages (e.g. "acme.com" when Name
	// is "replies.acme.com"). Record names should be entered relative to it.
	Zone      string         `json:"zone,omitempty"`
	CreatedAt string         `json:"created_at"`
	Records   []DomainRecord `json:"records"`
	// CustomReturnPath is the MAIL FROM subdomain (Return-Path); defaults to "send".
	CustomReturnPath string `json:"custom_return_path,omitempty"`
	OpenTracking     bool   `json:"open_tracking,omitempty"`
	ClickTracking    bool   `json:"click_tracking,omitempty"`
	// TrackingSubdomain is the open/click tracking subdomain label (empty =>
	// tracking on the shared host).
	TrackingSubdomain string `json:"tracking_subdomain,omitempty"`
	// Tls is the outbound TLS policy: "opportunistic" | "enforced".
	Tls              string              `json:"tls,omitempty"`
	Capabilities     *DomainCapabilities `json:"capabilities,omitempty"`
	TrackingDomain   string              `json:"tracking_domain,omitempty"`
	TrackingVerified bool                `json:"tracking_verified,omitempty"`
}

// CreateDomainRequest is the payload for POST /domains.
type CreateDomainRequest struct {
	Name   string `json:"name"`
	Region string `json:"region,omitempty"`
	// CustomReturnPath is the MAIL FROM subdomain (Return-Path); defaults to "send".
	CustomReturnPath string `json:"custom_return_path,omitempty"`
	OpenTracking     *bool  `json:"open_tracking,omitempty"`
	ClickTracking    *bool  `json:"click_tracking,omitempty"`
	// TrackingSubdomain is a custom tracking host label, e.g. "email" => email.<domain>.
	TrackingSubdomain string `json:"tracking_subdomain,omitempty"`
	// Tls is the outbound TLS policy: "opportunistic" | "enforced".
	Tls string `json:"tls,omitempty"`
	// Capabilities to enable, e.g. &DomainCapabilitiesInput{Receiving: "enabled"}.
	Capabilities *DomainCapabilitiesInput `json:"capabilities,omitempty"`
}

// DomainCapabilitiesInput selects capabilities on create/update.
type DomainCapabilitiesInput struct {
	Receiving string `json:"receiving,omitempty"`
}

// UpdateDomainRequest is the payload for PATCH /domains/:id.
type UpdateDomainRequest struct {
	OpenTracking      *bool  `json:"open_tracking,omitempty"`
	ClickTracking     *bool  `json:"click_tracking,omitempty"`
	TrackingSubdomain string `json:"tracking_subdomain,omitempty"`
	// CustomTracking enables/disables the custom open/click tracking host.
	CustomTracking *bool                    `json:"custom_tracking,omitempty"`
	Tls            string                   `json:"tls,omitempty"`
	Capabilities   *DomainCapabilitiesInput `json:"capabilities,omitempty"`
}

// DomainClaimRecord is the TXT record proving a claim.
type DomainClaimRecord struct {
	Type  string `json:"type"`
	Name  string `json:"name"`
	Value string `json:"value"`
	Ttl   string `json:"ttl"`
}

// DomainClaim is a claim on a domain already verified by another account.
type DomainClaim struct {
	Object        string            `json:"object"`
	Id            string            `json:"id"`
	Name          string            `json:"name"`
	DomainId      string            `json:"domain_id"`
	Region        string            `json:"region"`
	Status        string            `json:"status"`
	Record        DomainClaimRecord `json:"record"`
	BlockedReason string            `json:"blocked_reason"`
	FailureReason string            `json:"failure_reason"`
	CreatedAt     string            `json:"created_at"`
	ExpiresAt     string            `json:"expires_at"`
}

// ClaimDomainRequest is the payload for POST /domains/claim.
type ClaimDomainRequest struct {
	Name   string `json:"name"`
	Region string `json:"region,omitempty"`
}

// DetectDnsResponse describes a domain's DNS provider and the one-click apply
// methods available.
type DetectDnsResponse struct {
	Provider    any              `json:"provider"`
	Nameservers []string         `json:"nameservers"`
	Methods     []map[string]any `json:"methods"`
}

// CloudflareDnsRequest applies DNS records via the Cloudflare API.
type CloudflareDnsRequest struct {
	Token string `json:"token"`
}

// GoDaddyDnsRequest applies DNS records via the GoDaddy API.
type GoDaddyDnsRequest struct {
	Key    string `json:"key"`
	Secret string `json:"secret"`
}

// NamecheapDnsRequest applies DNS records via the Namecheap API. Namecheap
// must have the calling server's IP whitelisted in the account's API settings.
type NamecheapDnsRequest struct {
	ApiUser  string `json:"apiUser"`
	ApiKey   string `json:"apiKey"`
	UserName string `json:"userName,omitempty"`
}

// DomainsService handles the /domains endpoints.
type DomainsService struct {
	client *Client
}

// Create creates a domain. POST /domains
func (s *DomainsService) Create(params *CreateDomainRequest) (*Domain, error) {
	return s.CreateWithContext(context.Background(), params)
}

// CreateWithContext creates a domain. POST /domains
func (s *DomainsService) CreateWithContext(ctx context.Context, params *CreateDomainRequest) (*Domain, error) {
	return request[Domain](ctx, s.client, http.MethodPost, "/domains", params, nil)
}

// Get retrieves a domain. GET /domains/:id
func (s *DomainsService) Get(id string) (*Domain, error) {
	return s.GetWithContext(context.Background(), id)
}

// GetWithContext retrieves a domain. GET /domains/:id
func (s *DomainsService) GetWithContext(ctx context.Context, id string) (*Domain, error) {
	return request[Domain](ctx, s.client, http.MethodGet, "/domains/"+esc(id), nil, nil)
}

// List lists domains. GET /domains
func (s *DomainsService) List(params *ListParams) (*ListResponse[Domain], error) {
	return s.ListWithContext(context.Background(), params)
}

// ListWithContext lists domains. GET /domains
func (s *DomainsService) ListWithContext(ctx context.Context, params *ListParams) (*ListResponse[Domain], error) {
	return request[ListResponse[Domain]](ctx, s.client, http.MethodGet, listPath("/domains", params), nil, nil)
}

// Update updates a domain; returns the slim ack { object, id }. PATCH /domains/:id
func (s *DomainsService) Update(id string, params *UpdateDomainRequest) (*ObjectRef, error) {
	return s.UpdateWithContext(context.Background(), id, params)
}

// UpdateWithContext updates a domain. PATCH /domains/:id
func (s *DomainsService) UpdateWithContext(ctx context.Context, id string, params *UpdateDomainRequest) (*ObjectRef, error) {
	return request[ObjectRef](ctx, s.client, http.MethodPatch, "/domains/"+esc(id), params, nil)
}

// Verify kicks off DNS verification; returns the slim ack { object, id }.
// POST /domains/:id/verify
func (s *DomainsService) Verify(id string) (*ObjectRef, error) {
	return s.VerifyWithContext(context.Background(), id)
}

// VerifyWithContext kicks off DNS verification. POST /domains/:id/verify
func (s *DomainsService) VerifyWithContext(ctx context.Context, id string) (*ObjectRef, error) {
	return request[ObjectRef](ctx, s.client, http.MethodPost, "/domains/"+esc(id)+"/verify", nil, nil)
}

// Claim claims a domain already verified elsewhere. POST /domains/claim
func (s *DomainsService) Claim(params *ClaimDomainRequest) (*DomainClaim, error) {
	return s.ClaimWithContext(context.Background(), params)
}

// ClaimWithContext claims a domain already verified elsewhere. POST /domains/claim
func (s *DomainsService) ClaimWithContext(ctx context.Context, params *ClaimDomainRequest) (*DomainClaim, error) {
	return request[DomainClaim](ctx, s.client, http.MethodPost, "/domains/claim", params, nil)
}

// GetClaim retrieves a domain's claim record. GET /domains/:id/claim
func (s *DomainsService) GetClaim(id string) (*DomainClaim, error) {
	return s.GetClaimWithContext(context.Background(), id)
}

// GetClaimWithContext retrieves a domain's claim record. GET /domains/:id/claim
func (s *DomainsService) GetClaimWithContext(ctx context.Context, id string) (*DomainClaim, error) {
	return request[DomainClaim](ctx, s.client, http.MethodGet, "/domains/"+esc(id)+"/claim", nil, nil)
}

// VerifyClaim verifies a domain claim. POST /domains/:id/claim/verify
func (s *DomainsService) VerifyClaim(id string) (*DomainClaim, error) {
	return s.VerifyClaimWithContext(context.Background(), id)
}

// VerifyClaimWithContext verifies a domain claim. POST /domains/:id/claim/verify
func (s *DomainsService) VerifyClaimWithContext(ctx context.Context, id string) (*DomainClaim, error) {
	return request[DomainClaim](ctx, s.client, http.MethodPost, "/domains/"+esc(id)+"/claim/verify", nil, nil)
}

// DetectDns detects a domain's DNS provider and the one-click apply methods
// available. GET /domains/:id/dns/detect
func (s *DomainsService) DetectDns(id string) (*DetectDnsResponse, error) {
	return s.DetectDnsWithContext(context.Background(), id)
}

// DetectDnsWithContext detects a domain's DNS provider. GET /domains/:id/dns/detect
func (s *DomainsService) DetectDnsWithContext(ctx context.Context, id string) (*DetectDnsResponse, error) {
	return request[DetectDnsResponse](ctx, s.client, http.MethodGet, "/domains/"+esc(id)+"/dns/detect", nil, nil)
}

// ApplyCloudflareDns applies this domain's DNS records via the Cloudflare
// API, then auto-verifies. POST /domains/:id/dns/cloudflare
func (s *DomainsService) ApplyCloudflareDns(id string, params *CloudflareDnsRequest) (map[string]any, error) {
	return s.ApplyCloudflareDnsWithContext(context.Background(), id, params)
}

// ApplyCloudflareDnsWithContext applies DNS records via the Cloudflare API.
func (s *DomainsService) ApplyCloudflareDnsWithContext(ctx context.Context, id string, params *CloudflareDnsRequest) (map[string]any, error) {
	out, err := request[map[string]any](ctx, s.client, http.MethodPost, "/domains/"+esc(id)+"/dns/cloudflare", params, nil)
	if err != nil {
		return nil, err
	}
	return *out, nil
}

// ApplyGoDaddyDns applies this domain's DNS records via the GoDaddy API, then
// auto-verifies. POST /domains/:id/dns/godaddy
func (s *DomainsService) ApplyGoDaddyDns(id string, params *GoDaddyDnsRequest) (map[string]any, error) {
	return s.ApplyGoDaddyDnsWithContext(context.Background(), id, params)
}

// ApplyGoDaddyDnsWithContext applies DNS records via the GoDaddy API.
func (s *DomainsService) ApplyGoDaddyDnsWithContext(ctx context.Context, id string, params *GoDaddyDnsRequest) (map[string]any, error) {
	out, err := request[map[string]any](ctx, s.client, http.MethodPost, "/domains/"+esc(id)+"/dns/godaddy", params, nil)
	if err != nil {
		return nil, err
	}
	return *out, nil
}

// ApplyNamecheapDns applies this domain's DNS records via the Namecheap API
// (existing records are preserved), then auto-verifies.
// POST /domains/:id/dns/namecheap
func (s *DomainsService) ApplyNamecheapDns(id string, params *NamecheapDnsRequest) (map[string]any, error) {
	return s.ApplyNamecheapDnsWithContext(context.Background(), id, params)
}

// ApplyNamecheapDnsWithContext applies DNS records via the Namecheap API.
func (s *DomainsService) ApplyNamecheapDnsWithContext(ctx context.Context, id string, params *NamecheapDnsRequest) (map[string]any, error) {
	out, err := request[map[string]any](ctx, s.client, http.MethodPost, "/domains/"+esc(id)+"/dns/namecheap", params, nil)
	if err != nil {
		return nil, err
	}
	return *out, nil
}

// Remove deletes a domain. DELETE /domains/:id
func (s *DomainsService) Remove(id string) (*RemovedResponse, error) {
	return s.RemoveWithContext(context.Background(), id)
}

// RemoveWithContext deletes a domain. DELETE /domains/:id
func (s *DomainsService) RemoveWithContext(ctx context.Context, id string) (*RemovedResponse, error) {
	return request[RemovedResponse](ctx, s.client, http.MethodDelete, "/domains/"+esc(id), nil, nil)
}
