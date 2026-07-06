"""Domains resource — create, verify, claim, and one-click DNS providers."""

from typing import Dict, TypedDict

from . import http_client
from .http_client import paginate, path_escape as _e


class CreateParams(TypedDict, total=False):
    name: str  # required
    region: str
    custom_return_path: str  # MAIL FROM subdomain (Return-Path); defaults to 'send'
    open_tracking: bool
    click_tracking: bool
    tracking_subdomain: str  # custom tracking host label, e.g. 'email'
    tls: str  # 'opportunistic' | 'enforced'
    capabilities: Dict[str, str]  # e.g. {"receiving": "enabled"}


class UpdateParams(TypedDict, total=False):
    open_tracking: bool
    click_tracking: bool
    tracking_subdomain: str
    custom_tracking: bool
    tls: str  # 'opportunistic' | 'enforced'
    capabilities: Dict[str, str]


class ClaimParams(TypedDict, total=False):
    name: str  # required
    region: str


class Domains:
    """``mailblastr.Domains`` — the /domains endpoints."""

    CreateParams = CreateParams
    UpdateParams = UpdateParams
    ClaimParams = ClaimParams

    @classmethod
    def create(cls, params):
        """Add a sending domain. POST /domains"""
        return http_client.request("POST", "/domains", params)

    @classmethod
    def get(cls, domain_id):
        """Retrieve a domain (with its DNS records). GET /domains/:id"""
        return http_client.request("GET", f"/domains/{_e(domain_id)}")

    @classmethod
    def list(cls, params=None):
        """List domains. GET /domains"""
        return http_client.request("GET", f"/domains{paginate(params)}")

    @classmethod
    def update(cls, domain_id, params):
        """Update domain settings. PATCH /domains/:id"""
        return http_client.request("PATCH", f"/domains/{_e(domain_id)}", params)

    @classmethod
    def verify(cls, domain_id):
        """Trigger DNS verification. POST /domains/:id/verify"""
        return http_client.request("POST", f"/domains/{_e(domain_id)}/verify")

    @classmethod
    def claim(cls, params):
        """Claim a domain already verified elsewhere. POST /domains/claim"""
        return http_client.request("POST", "/domains/claim", params)

    @classmethod
    def get_claim(cls, domain_id):
        """Retrieve a domain's claim record. GET /domains/:id/claim"""
        return http_client.request("GET", f"/domains/{_e(domain_id)}/claim")

    @classmethod
    def verify_claim(cls, domain_id):
        """Verify a domain claim. POST /domains/:id/claim/verify"""
        return http_client.request("POST", f"/domains/{_e(domain_id)}/claim/verify")

    @classmethod
    def detect_dns(cls, domain_id):
        """Detect the domain's DNS provider and available one-click apply
        methods. GET /domains/:id/dns/detect"""
        return http_client.request("GET", f"/domains/{_e(domain_id)}/dns/detect")

    @classmethod
    def apply_cloudflare_dns(cls, domain_id, params):
        """Apply this domain's DNS records via the Cloudflare API, then
        auto-verify. POST /domains/:id/dns/cloudflare — ``{"token": ...}``"""
        return http_client.request("POST", f"/domains/{_e(domain_id)}/dns/cloudflare", params)

    @classmethod
    def apply_godaddy_dns(cls, domain_id, params):
        """Apply this domain's DNS records via the GoDaddy API, then
        auto-verify. POST /domains/:id/dns/godaddy — ``{"key", "secret"}``"""
        return http_client.request("POST", f"/domains/{_e(domain_id)}/dns/godaddy", params)

    @classmethod
    def apply_namecheap_dns(cls, domain_id, params):
        """Apply this domain's DNS records via the Namecheap API (existing
        records preserved), then auto-verify.
        POST /domains/:id/dns/namecheap — ``{"apiUser", "apiKey", "userName"?}``"""
        return http_client.request("POST", f"/domains/{_e(domain_id)}/dns/namecheap", params)

    @classmethod
    def remove(cls, domain_id):
        """Delete a domain. DELETE /domains/:id"""
        return http_client.request("DELETE", f"/domains/{_e(domain_id)}")
