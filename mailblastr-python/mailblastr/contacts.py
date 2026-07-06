"""Contacts resource — DOMAIN-FIRST flat /contacts API plus the nested
/audiences/:id/contacts variants, bulk imports, segment membership and topics.

Domain-first model: each sending domain has its own contact POOL — the same
address on two domains is two records with separate consent. On the flat
``/contacts`` API, ``domain`` is REQUIRED to create/list and disambiguates an
EMAIL id on get/update/remove. Pass ``audience_id`` instead to target a
specific audience via the nested API.
"""

from typing import Dict, List, TypedDict, Union

from . import http_client
from .http_client import build_query, path_escape as _e


class CreateParams(TypedDict, total=False):
    domain: str  # REQUIRED on the flat /contacts API (when audience_id is omitted)
    audience_id: str  # target a specific audience via the nested API instead
    email: str  # required
    first_name: str
    last_name: str
    unsubscribed: bool
    properties: Dict[str, Union[str, int, float]]


class UpdateParams(TypedDict, total=False):
    id: str  # required: contact id OR email
    domain: str  # disambiguates an EMAIL id across domains (flat API)
    audience_id: str  # omit for the flat /contacts/:id API
    first_name: str
    last_name: str
    unsubscribed: bool
    properties: Dict[str, Union[str, int, float]]


class ContactInput(TypedDict, total=False):
    email: str  # required
    first_name: str
    last_name: str
    unsubscribed: bool
    properties: Dict[str, Union[str, int, float]]


class UpdateTopicsParams(TypedDict):
    topics: List[Dict[str, str]]  # [{"id": ..., "subscription": "opt_in" | "opt_out"}]


class Contacts:
    """``mailblastr.Contacts`` — the /contacts (+ nested audience) endpoints."""

    CreateParams = CreateParams
    UpdateParams = UpdateParams
    ContactInput = ContactInput
    UpdateTopicsParams = UpdateTopicsParams

    @classmethod
    def create(cls, params):
        """Create a contact. Domain-first: pass ``domain`` to create in that
        domain's contact pool via the flat ``/contacts`` API (required there),
        or ``audience_id`` to target a specific audience via the nested API."""
        body = dict(params)
        audience_id = body.pop("audience_id", None)
        domain = body.pop("domain", None)
        if audience_id:
            # The nested audience route derives its pool from the path.
            return http_client.request("POST", f"/audiences/{_e(audience_id)}/contacts", body)
        if domain is not None:
            body["domain"] = domain
        return http_client.request("POST", "/contacts", body)

    @classmethod
    def get(cls, params):
        """Retrieve a contact by id or email — ``{"id", "domain"?, "audience_id"?}``.
        An id is exact; an EMAIL can exist in several domains' pools, so pass
        ``domain`` to pick the pool (omitted → the oldest match anywhere)."""
        cid = _e(params["id"])
        audience_id = params.get("audience_id")
        if audience_id:
            return http_client.request("GET", f"/audiences/{_e(audience_id)}/contacts/{cid}")
        qs = build_query({"domain": params.get("domain")})
        return http_client.request("GET", f"/contacts/{cid}{qs}")

    @classmethod
    def list(cls, params=None):
        """List contacts. Domain-first: ``domain`` is required on the flat
        ``/contacts`` list (names the pool); or pass ``audience_id`` for the
        nested list. ``segment_id`` filters either variant."""
        params = params or {}
        audience_id = params.get("audience_id")
        qs = build_query(
            {
                "domain": params.get("domain") if not audience_id else None,
                "limit": params.get("limit"),
                "after": params.get("after"),
                "before": params.get("before"),
                "segment_id": params.get("segment_id"),
            }
        )
        base = f"/audiences/{_e(audience_id)}/contacts" if audience_id else "/contacts"
        return http_client.request("GET", f"{base}{qs}")

    @classmethod
    def batch(cls, params):
        """Bulk-import contacts from a list — ``{"audience_id", "contacts",
        "on_conflict"?}``. Upserts by email; max 10,000 per call.
        ``on_conflict='skip'`` leaves existing contacts untouched.
        POST /audiences/:id/contacts/batch"""
        qs = build_query({"on_conflict": params.get("on_conflict")})
        return http_client.request(
            "POST",
            f"/audiences/{_e(params['audience_id'])}/contacts/batch{qs}",
            {"contacts": params["contacts"]},
        )

    @classmethod
    def import_csv(cls, params):
        """Bulk-import contacts from CSV text — ``{"audience_id", "csv",
        "on_conflict"?, "create_properties"?}``. Upserts by email. By default
        every non-builtin CSV column is auto-registered as a custom property;
        pass ``create_properties=False`` for strict mode.
        POST /audiences/:id/contacts/import"""
        qs = build_query(
            {
                "on_conflict": params.get("on_conflict"),
                "create_properties": (
                    "false" if params.get("create_properties") is False else None
                ),
            }
        )
        return http_client.request(
            "POST",
            f"/audiences/{_e(params['audience_id'])}/contacts/import{qs}",
            {"csv": params["csv"]},
        )

    @classmethod
    def update(cls, params):
        """Update a contact. On the flat API, pass ``domain`` when ``id`` is
        an EMAIL (disambiguates across pools). PATCH /contacts/:id"""
        body = dict(params)
        cid = _e(body.pop("id"))
        audience_id = body.pop("audience_id", None)
        domain = body.pop("domain", None)
        if audience_id:
            return http_client.request(
                "PATCH", f"/audiences/{_e(audience_id)}/contacts/{cid}", body
            )
        if domain is not None:
            body["domain"] = domain
        return http_client.request("PATCH", f"/contacts/{cid}", body)

    @classmethod
    def remove(cls, params):
        """Delete a contact — ``{"id", "domain"?, "audience_id"?}``. On the
        flat API, pass ``domain`` when ``id`` is an EMAIL."""
        cid = _e(params["id"])
        audience_id = params.get("audience_id")
        if audience_id:
            return http_client.request("DELETE", f"/audiences/{_e(audience_id)}/contacts/{cid}")
        qs = build_query({"domain": params.get("domain")})
        return http_client.request("DELETE", f"/contacts/{cid}{qs}")

    @classmethod
    def add_to_segment(cls, contact_id, segment_id):
        """Add a contact to a segment. POST /contacts/:id/segments/:segment_id"""
        return http_client.request(
            "POST", f"/contacts/{_e(contact_id)}/segments/{_e(segment_id)}"
        )

    @classmethod
    def remove_from_segment(cls, contact_id, segment_id):
        """Remove a contact from a segment. DELETE /contacts/:id/segments/:segment_id"""
        return http_client.request(
            "DELETE", f"/contacts/{_e(contact_id)}/segments/{_e(segment_id)}"
        )

    @classmethod
    def list_segments(cls, contact_id):
        """List the segments a contact belongs to. GET /contacts/:id/segments"""
        return http_client.request("GET", f"/contacts/{_e(contact_id)}/segments")

    @classmethod
    def get_topics(cls, contact_id):
        """Get a contact's topic subscriptions. GET /contacts/:id/topics"""
        return http_client.request("GET", f"/contacts/{_e(contact_id)}/topics")

    @classmethod
    def update_topics(cls, contact_id, params):
        """Update a contact's topic subscriptions. PATCH /contacts/:id/topics"""
        return http_client.request("PATCH", f"/contacts/{_e(contact_id)}/topics", params)
