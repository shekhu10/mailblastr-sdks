"""Emails resource — send, retrieve, attachments, and inbound (receiving)."""

from typing import Any, Dict, List, TypedDict, Union

from . import http_client
from .http_client import build_query, paginate, path_escape as _e


class Attachment(TypedDict, total=False):
    filename: str
    content: str  # base64-encoded file content (provide `content` OR `path`)
    path: str  # hosted URL to fetch the file from (provide `content` OR `path`)
    content_type: str
    content_id: str  # Content-ID for inline/related parts (cid: references)


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
        "scheduled_at": str,  # ISO 8601 timestamp to schedule the send
        "topic_id": str,  # drop recipients unsubscribed from this topic
        "template_id": str,  # send using a saved template
        "template": Dict[str, Any],  # nested {id | alias, variables} reference
        "variables": Dict[str, Any],  # values for {{ placeholder }} variables
    },
    total=False,
)

UpdateParams = TypedDict("UpdateParams", {"scheduled_at": str})


class ListParams(TypedDict, total=False):
    limit: int
    after: str
    before: str
    campaign_id: str  # only emails sent by this campaign (incl. its follow-ups)
    automation_id: str  # only emails sent by this automation (ignored with campaign_id)
    source: str  # 'individual' restricts to one-off API sends
    domain_id: str  # only emails sent from this sending domain (domain id)
    status: str  # match the value reads expose as `last_event`, e.g. 'delivered'
    search: str  # substring match over recipients, subject and sender


class ReceivingListParams(TypedDict, total=False):
    limit: int
    after: str
    before: str
    received_for: str  # only messages received for this address


class AttachmentListParams(TypedDict, total=False):
    limit: int
    after: str
    before: str

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
    ListParams = ListParams
    Attachment = Attachment

    class Attachments:
        """Attachments of SENT emails — ``mailblastr.Emails.Attachments``."""

        @classmethod
        def list(cls, email_id):
            """List a sent email's attachments (not paginated — every
            attachment is returned). GET /emails/:id/attachments"""
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
        ListParams = ReceivingListParams
        AttachmentListParams = AttachmentListParams

        @classmethod
        def list(cls, params=None):
            """List received emails, optionally filtered by ``received_for``
            (only messages received for that address). GET /emails/receiving

            Called with no ``limit`` and no cursor this returns up to 1000
            messages in one response; pass ``limit`` for normal 1-100 paging."""
            params = params or {}
            qs = build_query(
                {
                    "limit": params.get("limit"),
                    "after": params.get("after"),
                    "before": params.get("before"),
                    "received_for": params.get("received_for"),
                }
            )
            return http_client.request("GET", f"/emails/receiving{qs}")

        @classmethod
        def addresses(cls):
            """Per-address inbound stats (total, replies, interested and the
            last received time for each receiving address). Not paginated.
            GET /emails/receiving/addresses"""
            return http_client.request("GET", "/emails/receiving/addresses")

        @classmethod
        def get(cls, email_id):
            """Retrieve a received email. GET /emails/receiving/:id"""
            return http_client.request("GET", f"/emails/receiving/{_e(email_id)}")

        @classmethod
        def attachments(cls, email_id, params=None):
            """List a received email's attachments.
            GET /emails/receiving/:id/attachments

            With no ``limit`` and no ``after`` cursor every attachment is
            returned; supplying either applies normal paging."""
            return http_client.request(
                "GET", f"/emails/receiving/{_e(email_id)}/attachments{paginate(params)}"
            )

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

        Pass ``options={"idempotency_key": "..."}`` to safely retry: replaying
        the key returns the original response instead of sending twice. The key
        must be 1-255 characters after the server trims it
        (:data:`~mailblastr.IDEMPOTENCY_KEY_MAX_LENGTH`); the server, not this
        SDK, rejects anything else with 400 ``invalid_idempotency_key``.
        ``tags`` is not a supported field — sending it is a 422."""
        return http_client.request("POST", "/emails", params, options)

    @classmethod
    def batch(cls, params, options=None):
        """Send up to 100 emails in one request. POST /emails/batch
        (alias of ``mailblastr.Batch.send``).

        Batch items reject ``attachments`` and ``scheduled_at`` — send those
        individually via :meth:`send`. ``options={"idempotency_key": "..."}``
        is honoured here too (same 1-255 rule as :meth:`send`); with a key set, a failure part
        way through reports the emails that already went out on the raised
        error's ``sent`` / ``sent_count``."""
        return http_client.request("POST", "/emails/batch", params, options)

    @classmethod
    def list(cls, params=None):
        """List sent emails (trimmed list items — no status/html/text/events,
        and unset cc/bcc/reply_to come back as ``None`` rather than ``[]``).
        Optional filters: ``campaign_id``, ``automation_id``,
        ``source='individual'``, ``domain_id``, ``status`` (matched against the
        value reads expose as ``last_event``) and ``search`` (recipients,
        subject, sender). GET /emails"""
        params = params or {}
        qs = build_query(
            {
                "limit": params.get("limit"),
                "after": params.get("after"),
                "before": params.get("before"),
                "campaign_id": params.get("campaign_id"),
                "automation_id": params.get("automation_id"),
                "source": params.get("source"),
                "domain_id": params.get("domain_id"),
                "status": params.get("status"),
                "search": params.get("search"),
            }
        )
        return http_client.request("GET", f"/emails{qs}")

    @classmethod
    def sources(cls):
        """Per-source send metrics — one row per campaign, automation and
        individual sending, with delivered/opened/clicked counts. Not
        paginated. GET /emails/sources"""
        return http_client.request("GET", "/emails/sources")

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
