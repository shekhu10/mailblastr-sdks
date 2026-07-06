"""Events resource — send custom events that trigger automations, plus
custom-event definitions. DOMAIN-FIRST: ``domain`` is REQUIRED on
``Events.send`` (only automations belonging to that domain are triggered)."""

from typing import Any, Dict, TypedDict

from . import http_client
from .http_client import paginate, path_escape as _e


class SendParams(TypedDict, total=False):
    event: str  # the custom event name automations trigger on (`name` = alias)
    name: str  # alias for `event`
    domain: str  # REQUIRED: the sending domain this event belongs to
    contact_id: str  # identify the contact by id (contact_id OR email)
    email: str  # identify the contact by email (contact_id OR email)
    payload: Dict[str, Any]  # arbitrary event payload (`data` = alias)
    data: Dict[str, Any]  # alias for `payload`


class CreateParams(TypedDict, total=False):
    name: str  # required (cannot start with the reserved `mailblastr:` prefix)
    schema: Dict[str, str]  # flat key -> 'string' | 'number' | 'boolean' | 'date'


class Events:
    """``mailblastr.Events`` — the /events endpoints."""

    SendParams = SendParams
    CreateParams = CreateParams

    @classmethod
    def send(cls, params, options=None):
        """Send a custom event that automations can trigger on (``domain`` is
        required). POST /events/send"""
        return http_client.request("POST", "/events/send", params, options)

    @classmethod
    def create(cls, params, options=None):
        """Create a custom-event definition (name + optional payload schema).
        POST /events"""
        return http_client.request("POST", "/events", params, options)

    @classmethod
    def list(cls, params=None):
        """List custom-event definitions. GET /events"""
        return http_client.request("GET", f"/events{paginate(params)}")

    @classmethod
    def remove(cls, event_id):
        """Delete a custom-event definition. DELETE /events/:id"""
        return http_client.request("DELETE", f"/events/{_e(event_id)}")
