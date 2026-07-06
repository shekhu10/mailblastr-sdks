"""Segments resource — DOMAIN-FIRST: segments belong to a sending domain.
``domain`` is REQUIRED on create and list; names are unique within a domain
(reusable across domains) and every domain carries an auto-created "General"
segment."""

from typing import Any, Dict, TypedDict

from . import http_client
from .http_client import build_query, path_escape as _e


class CreateParams(TypedDict, total=False):
    domain: str  # REQUIRED: the sending domain this segment belongs to
    name: str  # required
    filter: Dict[str, Any]  # {status?, email_contains?, property_filters?}


class UpdateParams(TypedDict, total=False):
    name: str
    filter: Dict[str, Any]


class ListParams(TypedDict, total=False):
    domain: str  # REQUIRED: the sending domain whose segments to list
    limit: int
    after: str
    before: str


class Segments:
    """``mailblastr.Segments`` — the /segments endpoints."""

    CreateParams = CreateParams
    UpdateParams = UpdateParams
    ListParams = ListParams

    @classmethod
    def create(cls, params):
        """Create a segment on a sending domain (``domain`` is required).
        POST /segments"""
        return http_client.request("POST", "/segments", params)

    @classmethod
    def get(cls, segment_id):
        """Retrieve a segment. GET /segments/:id"""
        return http_client.request("GET", f"/segments/{_e(segment_id)}")

    @classmethod
    def list(cls, params):
        """List a domain's segments (``domain`` is required; includes the
        auto-created "General" segment). GET /segments?domain=..."""
        qs = build_query(
            {
                "domain": params["domain"],
                "limit": params.get("limit"),
                "after": params.get("after"),
                "before": params.get("before"),
            }
        )
        return http_client.request("GET", f"/segments{qs}")

    @classmethod
    def contacts(cls, segment_id):
        """Preview the contacts a segment currently resolves to.
        GET /segments/:id/contacts"""
        return http_client.request("GET", f"/segments/{_e(segment_id)}/contacts")

    @classmethod
    def update(cls, segment_id, params):
        """Update a segment. PATCH /segments/:id"""
        return http_client.request("PATCH", f"/segments/{_e(segment_id)}", params)

    @classmethod
    def remove(cls, segment_id):
        """Delete a segment. DELETE /segments/:id"""
        return http_client.request("DELETE", f"/segments/{_e(segment_id)}")
