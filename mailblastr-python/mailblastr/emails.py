"""Emails resource — send, retrieve, attachments, and inbound (receiving)."""

from typing import Any, Dict, List, TypedDict, Union

from . import http_client
from .http_client import paginate, path_escape as _e


class Attachment(TypedDict, total=False):
    filename: str
    content: str  # base64-encoded file content (provide `content` OR `path`)
    path: str  # hosted URL to fetch the file from (provide `content` OR `path`)
    content_type: str
    content_id: str  # Content-ID for inline/related parts (cid: references)


class Tag(TypedDict):
    name: str
    value: str


# `from` is a Python keyword, so the params types use the functional syntax.
SendParams = TypedDict(
    "SendParams",
    {
        "from": str,  # required
        "to": Union[str, List[str]],  # required
        "subject": str,  # required
        "bcc": Union[str, List[str]],
        "cc": Union[str, List[str]],
        "reply_to": Union[str, List[str]],
        "html": str,
        "text": str,
        "preview_text": str,  # inbox preheader, max 150 chars
        "headers": Dict[str, str],
        "attachments": List[Attachment],
        "tags": List[Tag],
        "scheduled_at": str,  # ISO 8601 timestamp to schedule the send
        "topic_id": str,  # drop recipients unsubscribed from this topic
        "template_id": str,  # send using a saved template
        "template": Dict[str, Any],  # nested {id | alias, variables} reference
        "variables": Dict[str, Any],  # values for {{ placeholder }} variables
    },
    total=False,
)

UpdateParams = TypedDict("UpdateParams", {"scheduled_at": str})

ForwardParams = TypedDict(
    "ForwardParams",
    {
        "from": str,  # required: a verified sending address to forward from
        "to": Union[str, List[str]],  # required
        "subject": str,
    },
    total=False,
)

ReplyParams = TypedDict(
    "ReplyParams",
    {
        "from": str,  # required
        "html": str,
        "text": str,
        "subject": str,  # defaults to `Re: ...` (keeps the thread)
    },
    total=False,
)


class Emails:
    """``mailblastr.Emails`` — the /emails endpoints."""

    SendParams = SendParams
    UpdateParams = UpdateParams
    Attachment = Attachment
    Tag = Tag

    class Attachments:
        """Attachments of SENT emails — ``mailblastr.Emails.Attachments``."""

        @classmethod
        def list(cls, email_id):
            """List a sent email's attachments. GET /emails/:id/attachments"""
            return http_client.request("GET", f"/emails/{_e(email_id)}/attachments")

        @classmethod
        def get(cls, email_id, attachment_id):
            """Retrieve one attachment of a sent email.
            GET /emails/:id/attachments/:attachment_id"""
            return http_client.request(
                "GET", f"/emails/{_e(email_id)}/attachments/{_e(attachment_id)}"
            )

    class Receiving:
        """Inbound (received) email — ``mailblastr.Emails.Receiving``."""

        ForwardParams = ForwardParams
        ReplyParams = ReplyParams

        @classmethod
        def list(cls, params=None):
            """List received emails. GET /emails/receiving"""
            return http_client.request("GET", f"/emails/receiving{paginate(params)}")

        @classmethod
        def get(cls, email_id):
            """Retrieve a received email. GET /emails/receiving/:id"""
            return http_client.request("GET", f"/emails/receiving/{_e(email_id)}")

        @classmethod
        def attachments(cls, email_id):
            """List a received email's attachments. GET /emails/receiving/:id/attachments"""
            return http_client.request("GET", f"/emails/receiving/{_e(email_id)}/attachments")

        @classmethod
        def get_attachment(cls, email_id, attachment_id):
            """Download one attachment of a received email as raw ``bytes``.
            GET /emails/receiving/:id/attachments/:attachment_id (streams binary)."""
            return http_client.request_raw(
                "GET", f"/emails/receiving/{_e(email_id)}/attachments/{_e(attachment_id)}"
            )

        @classmethod
        def raw(cls, email_id):
            """Download the original RFC822/MIME message as raw ``bytes``.
            GET /emails/receiving/:id/raw (streams message/rfc822)."""
            return http_client.request_raw("GET", f"/emails/receiving/{_e(email_id)}/raw")

        @classmethod
        def forward(cls, email_id, params):
            """Forward a received email. POST /emails/receiving/:id/forward"""
            return http_client.request(
                "POST", f"/emails/receiving/{_e(email_id)}/forward", params
            )

        @classmethod
        def reply(cls, email_id, params):
            """Reply to a received email's sender, threaded into the same
            conversation. POST /emails/receiving/:id/reply"""
            return http_client.request("POST", f"/emails/receiving/{_e(email_id)}/reply", params)

        @classmethod
        def remove(cls, email_id):
            """Delete a received email. DELETE /emails/receiving/:id"""
            return http_client.request("DELETE", f"/emails/receiving/{_e(email_id)}")

    @classmethod
    def send(cls, params, options=None):
        """Send a single email. POST /emails

        Pass ``options={"idempotency_key": "..."}`` to safely retry."""
        return http_client.request("POST", "/emails", params, options)

    @classmethod
    def batch(cls, params, options=None):
        """Send up to 100 emails in one request. POST /emails/batch
        (alias of ``mailblastr.Batch.send``)."""
        return http_client.request("POST", "/emails/batch", params, options)

    @classmethod
    def list(cls, params=None):
        """List sent emails (trimmed list items — no status/html/text/events).
        GET /emails"""
        return http_client.request("GET", f"/emails{paginate(params)}")

    @classmethod
    def get(cls, email_id):
        """Retrieve a sent email and its events. GET /emails/:id"""
        return http_client.request("GET", f"/emails/{_e(email_id)}")

    @classmethod
    def update(cls, email_id, params):
        """Reschedule a scheduled email. PATCH /emails/:id"""
        return http_client.request("PATCH", f"/emails/{_e(email_id)}", params)

    @classmethod
    def cancel(cls, email_id):
        """Cancel a scheduled email. POST /emails/:id/cancel"""
        return http_client.request("POST", f"/emails/{_e(email_id)}/cancel")
