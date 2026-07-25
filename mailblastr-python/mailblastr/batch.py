"""Batch send resource — ``mailblastr.Batch.send([...])``."""

from . import http_client


class Batch:
    """``mailblastr.Batch`` — the /emails/batch endpoint."""

    @classmethod
    def send(cls, params, options=None):
        """Send up to 100 emails in one request. POST /emails/batch

        ``params`` is a list of :class:`mailblastr.Emails.SendParams` dicts.
        Batch items reject ``attachments`` and ``scheduled_at`` — send those
        individually via ``mailblastr.Emails.send``."""
        return http_client.request("POST", "/emails/batch", params, options)
