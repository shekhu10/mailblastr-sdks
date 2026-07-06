"""Exceptions raised by the MailBlastr SDK."""


class MailblastrError(Exception):
    """An API (or transport) error.

    Carries the ``{statusCode, name, message}`` body returned by the
    MailBlastr API on non-2xx responses. ``status_code`` is 0 for
    transport-level failures (``name`` = ``network_error``) and for
    client-side configuration errors (``name`` = ``missing_api_key``).
    """

    def __init__(self, status_code=0, name="application_error", message=""):
        super().__init__(message or name)
        self.status_code = status_code
        self.name = name
        self.message = message

    @property
    def statusCode(self):  # noqa: N802 — mirrors the API error body key
        """Alias matching the API error body key ``statusCode``."""
        return self.status_code

    def __str__(self):
        return f"[{self.name}] {self.message} (status {self.status_code})"
