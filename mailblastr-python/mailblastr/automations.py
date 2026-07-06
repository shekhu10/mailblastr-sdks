"""Automations resource — DOMAIN-FIRST: every automation belongs to one of
your sending domains (``domain`` is REQUIRED on create). Only ``Events.send``
calls with the same ``domain`` trigger it."""

from typing import Any, Dict, List, TypedDict

from . import http_client
from .http_client import paginate, path_escape as _e


class CreateParams(TypedDict, total=False):
    name: str  # required
    domain: str  # REQUIRED: the sending domain this automation belongs to
    trigger: str  # e.g. 'contact.created', 'email.opened', or a custom event name
    status: str  # 'enabled' | 'disabled' (default 'disabled')
    steps: List[Dict[str, Any]]  # optional inline step graph
    connections: List[Dict[str, str]]  # optional typed edges between step keys


class UpdateParams(TypedDict, total=False):
    name: str
    status: str  # 'enabled' | 'disabled'
    domain: str  # re-point at another domain (disabled automations only)
    connections: List[Dict[str, str]]


class AddStepParams(TypedDict, total=False):
    type: str  # required, e.g. 'send_email' | 'wait' | 'condition'
    config: Dict[str, Any]
    key: str


class Automations:
    """``mailblastr.Automations`` — the /automations endpoints."""

    CreateParams = CreateParams
    UpdateParams = UpdateParams
    AddStepParams = AddStepParams

    @classmethod
    def create(cls, params):
        """Create an automation (``domain`` is required). POST /automations"""
        return http_client.request("POST", "/automations", params)

    @classmethod
    def get(cls, automation_id):
        """Retrieve an automation (with steps/connections/enrollments).
        GET /automations/:id"""
        return http_client.request("GET", f"/automations/{_e(automation_id)}")

    @classmethod
    def list(cls, params=None):
        """List automations. GET /automations"""
        return http_client.request("GET", f"/automations{paginate(params)}")

    @classmethod
    def update(cls, automation_id, params):
        """Update an automation. PATCH /automations/:id"""
        return http_client.request("PATCH", f"/automations/{_e(automation_id)}", params)

    @classmethod
    def add_step(cls, automation_id, params):
        """Append a step to an automation. POST /automations/:id/steps"""
        return http_client.request("POST", f"/automations/{_e(automation_id)}/steps", params)

    @classmethod
    def delete_step(cls, automation_id, step_id):
        """Delete a step from an automation.
        DELETE /automations/:id/steps/:step_id"""
        return http_client.request(
            "DELETE", f"/automations/{_e(automation_id)}/steps/{_e(step_id)}"
        )

    @classmethod
    def runs(cls, automation_id, params=None):
        """List an automation's runs. GET /automations/:id/runs"""
        return http_client.request(
            "GET", f"/automations/{_e(automation_id)}/runs{paginate(params)}"
        )

    @classmethod
    def get_run(cls, automation_id, run_id):
        """Retrieve a single automation run (with its step trace).
        GET /automations/:id/runs/:run_id"""
        return http_client.request(
            "GET", f"/automations/{_e(automation_id)}/runs/{_e(run_id)}"
        )

    @classmethod
    def stop(cls, automation_id):
        """Stop an automation — prevents new runs; in-progress runs finish.
        POST /automations/:id/stop"""
        return http_client.request("POST", f"/automations/{_e(automation_id)}/stop")

    @classmethod
    def remove(cls, automation_id):
        """Delete an automation. DELETE /automations/:id"""
        return http_client.request("DELETE", f"/automations/{_e(automation_id)}")
