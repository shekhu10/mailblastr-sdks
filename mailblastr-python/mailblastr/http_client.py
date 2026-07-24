"""Internal HTTP layer for the MailBlastr SDK. Stdlib only (urllib + json)."""

import email.utils
import json
import time
import urllib.error
import urllib.request
from urllib.parse import quote, urlencode

from .exceptions import MailblastrError

DEFAULT_BASE_URL = "https://www.mailblastr.com/api"

# Keep in sync with pyproject.toml "version".
VERSION = "1.0.0"
USER_AGENT = f"mailblastr-python/{VERSION}"

# Defaults for the timeout + retry policy. Override per client via the
# module-level `mailblastr.timeout` (seconds) and `mailblastr.max_retries`.
DEFAULT_TIMEOUT = 30.0  # seconds
DEFAULT_MAX_RETRIES = 2
# Only 429 (rate limited) and 503 (unavailable) are retried: the server
# guarantees neither was applied, so a retry can't duplicate a side-effect
# (e.g. a double send). Honors Retry-After.
_RETRYABLE_STATUS = (429, 503)


def path_escape(value):
    """Percent-encode a path segment so an id like ``../api-keys`` can't
    traverse the URL path and forge a request to a different endpoint."""
    return quote(str(value), safe="")


def build_query(params):
    """Build ``''`` or ``'?a=b&c=d'`` from a dict, dropping ``None`` values."""
    if not params:
        return ""
    filtered = {k: str(v) for k, v in params.items() if v is not None}
    return f"?{urlencode(filtered)}" if filtered else ""


def paginate(params):
    """Build the ``?limit=&after=&before=`` query string from pagination params."""
    if not params:
        return ""
    return build_query(
        {
            "limit": params.get("limit"),
            "after": params.get("after"),
            "before": params.get("before"),
        }
    )


def _config():
    # Late import: reads the module-level `mailblastr.api_key` / `base_url`
    # set by the caller (resend-python style configuration).
    import mailblastr

    api_key = getattr(mailblastr, "api_key", None)
    if not api_key or not isinstance(api_key, str):
        raise MailblastrError(
            0,
            "missing_api_key",
            'mailblastr.api_key must be set, e.g. mailblastr.api_key = "mb_xxxxxxxxx"',
        )
    base_url = (getattr(mailblastr, "base_url", None) or DEFAULT_BASE_URL).rstrip("/")
    timeout = getattr(mailblastr, "timeout", None)
    timeout = DEFAULT_TIMEOUT if timeout is None else float(timeout)
    max_retries = getattr(mailblastr, "max_retries", None)
    max_retries = DEFAULT_MAX_RETRIES if max_retries is None else max(0, int(max_retries))
    return api_key, base_url, timeout, max_retries


def _retry_after_seconds(header):
    """Parse a Retry-After header (delta-seconds or HTTP-date) → seconds, capped
    at 30s. Returns None when absent/unparseable."""
    if not header:
        return None
    try:
        return min(max(0.0, float(header)), 30.0)
    except (TypeError, ValueError):
        pass
    parsed = email.utils.parsedate_to_datetime(header)
    if parsed is not None:
        delta = parsed.timestamp() - time.time()
        return min(max(0.0, delta), 30.0)
    return None


def _headers(api_key, json_body, options):
    headers = {
        "Authorization": f"Bearer {api_key}",
        "User-Agent": USER_AGENT,
    }
    if json_body:
        headers["Content-Type"] = "application/json"
    if options and options.get("idempotency_key"):
        headers["Idempotency-Key"] = options["idempotency_key"]
    return headers


def _error_from(status, raw):
    parsed = None
    if raw:
        try:
            parsed = json.loads(raw)
        except (ValueError, UnicodeDecodeError):
            parsed = None
    body = parsed if isinstance(parsed, dict) else {}
    return MailblastrError(
        body.get("statusCode", status),
        body.get("name", "application_error"),
        body.get("message", f"Request failed with status {status}"),
    )


def request(method, path, body=None, options=None):
    """Make a JSON API request. Returns the parsed JSON body (``None`` when the
    response is empty). Raises :class:`MailblastrError` on any non-2xx status."""
    api_key, base_url, timeout, max_retries = _config()
    data = None if body is None else json.dumps(body).encode("utf-8")
    req = urllib.request.Request(
        base_url + path, data=data, headers=_headers(api_key, True, options), method=method
    )
    raw = _send(req, timeout, max_retries)
    if not raw:
        return None
    try:
        return json.loads(raw)
    except ValueError:
        return raw.decode("utf-8", "replace")


def request_raw(method, path, options=None):
    """Make a request to an endpoint that streams raw binary bytes (attachment /
    raw MIME downloads). Returns ``bytes``. Raises :class:`MailblastrError` on
    any non-2xx status (the error body is JSON even on binary routes)."""
    api_key, base_url, timeout, max_retries = _config()
    req = urllib.request.Request(
        base_url + path, data=None, headers=_headers(api_key, False, options), method=method
    )
    return _send(req, timeout, max_retries)


def _send(req, timeout, max_retries):
    """Send with a socket timeout, retrying only 429/503 (Retry-After aware)."""
    attempt = 0
    while True:
        try:
            with urllib.request.urlopen(req, timeout=timeout) as res:
                return res.read()
        except urllib.error.HTTPError as err:
            should_retry = err.code in _RETRYABLE_STATUS and attempt < max_retries
            retry_after = None
            raw = b""
            try:
                if should_retry:
                    retry_after = _retry_after_seconds(err.headers.get("Retry-After") if err.headers else None)
                else:
                    raw = err.read()
            finally:
                # HTTPError is also a file-like response. Leaving it open leaks
                # its socket/file descriptor on every 4xx/5xx and each retry.
                err.close()
            if should_retry:
                time.sleep(retry_after if retry_after is not None else min(30.0, 0.5 * (2 ** attempt)))
                attempt += 1
                continue
            raise _error_from(err.code, raw) from None
        except urllib.error.URLError as err:
            # Includes socket.timeout (raised as URLError.reason on timeout).
            raise MailblastrError(0, "network_error", str(getattr(err, "reason", err))) from None
