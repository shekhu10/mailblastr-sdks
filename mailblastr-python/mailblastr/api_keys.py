"""API keys resource."""

from typing import TypedDict

from . import http_client
from .http_client import path_escape as _e


class CreateParams(TypedDict, total=False):
    name: str  # required
    permission: str  # 'full_access' | 'sending_access'
    domain_id: str  # scope a sending_access key to one domain


class ApiKeys:
    """``mailblastr.ApiKeys`` — the /api-keys endpoints."""

    CreateParams = CreateParams

    @classmethod
    def create(cls, params):
        """Create an API key. The full token is returned ONLY here.
        POST /api-keys"""
        return http_client.request("POST", "/api-keys", params)

    @classmethod
    def list(cls):
        """List API keys (non-secret prefixes only). GET /api-keys"""
        return http_client.request("GET", "/api-keys")

    @classmethod
    def remove(cls, api_key_id):
        """Delete an API key. DELETE /api-keys/:id"""
        return http_client.request("DELETE", f"/api-keys/{_e(api_key_id)}")
