"""Tests for the real HTTP layer (urllib monkeypatched at the urlopen level):
auth header, JSON encoding, error raising, and configuration."""

import io
import json
import unittest
import urllib.error
from unittest import mock

import mailblastr
from mailblastr import http_client
from mailblastr.exceptions import MailblastrError


class FakeResponse:
    def __init__(self, body: bytes):
        self._body = body

    def read(self):
        return self._body

    def __enter__(self):
        return self

    def __exit__(self, *args):
        return False


class TestHttpClient(unittest.TestCase):
    def setUp(self):
        mailblastr.api_key = "mb_test_key"
        mailblastr.base_url = http_client.DEFAULT_BASE_URL
        self.addCleanup(setattr, mailblastr, "api_key", None)
        self.addCleanup(setattr, mailblastr, "base_url", http_client.DEFAULT_BASE_URL)

    def test_request_sends_bearer_auth_json_and_user_agent(self):
        captured = {}

        def fake_urlopen(req, **kwargs):
            captured["req"] = req
            return FakeResponse(b'{"id": "em_1"}')

        with mock.patch("urllib.request.urlopen", side_effect=fake_urlopen):
            result = http_client.request(
                "POST",
                "/emails",
                {"from": "a@b.com", "to": "c@d.com", "subject": "s"},
                {"idempotency_key": "order-1"},
            )

        req = captured["req"]
        self.assertEqual(result, {"id": "em_1"})
        self.assertEqual(req.get_method(), "POST")
        self.assertEqual(req.full_url, "https://www.mailblastr.com/api/emails")
        self.assertEqual(req.get_header("Authorization"), "Bearer mb_test_key")
        self.assertEqual(req.get_header("Content-type"), "application/json")
        self.assertEqual(req.get_header("User-agent"), http_client.USER_AGENT)
        self.assertEqual(req.get_header("Idempotency-key"), "order-1")
        self.assertEqual(
            json.loads(req.data.decode("utf-8")),
            {"from": "a@b.com", "to": "c@d.com", "subject": "s"},
        )

    def test_get_request_has_no_body(self):
        captured = {}

        def fake_urlopen(req, **kwargs):
            captured["req"] = req
            return FakeResponse(b'{"object":"list","data":[]}')

        with mock.patch("urllib.request.urlopen", side_effect=fake_urlopen):
            http_client.request("GET", "/domains")
        self.assertIsNone(captured["req"].data)
        self.assertEqual(captured["req"].get_method(), "GET")

    def test_base_url_override(self):
        mailblastr.base_url = "http://localhost:3000/"  # trailing slash stripped
        captured = {}

        def fake_urlopen(req, **kwargs):
            captured["req"] = req
            return FakeResponse(b"{}")

        with mock.patch("urllib.request.urlopen", side_effect=fake_urlopen):
            http_client.request("GET", "/emails")
        self.assertEqual(captured["req"].full_url, "http://localhost:3000/emails")

    def test_non_2xx_raises_mailblastr_error_with_api_body(self):
        body = json.dumps(
            {"statusCode": 422, "name": "validation_error", "message": "domain is required"}
        ).encode("utf-8")
        response_body = io.BytesIO(body)
        err = urllib.error.HTTPError(
            "https://www.mailblastr.com/api/segments", 422, "Unprocessable", {}, response_body
        )
        with mock.patch("urllib.request.urlopen", side_effect=err):
            with self.assertRaises(MailblastrError) as ctx:
                http_client.request("POST", "/segments", {"name": "VIP"})
        e = ctx.exception
        self.assertEqual(e.status_code, 422)
        self.assertEqual(e.statusCode, 422)
        self.assertEqual(e.name, "validation_error")
        self.assertEqual(e.message, "domain is required")
        self.assertTrue(response_body.closed)

    def test_non_json_error_body_falls_back_to_status(self):
        err = urllib.error.HTTPError(
            "https://www.mailblastr.com/api/emails", 500, "Boom", {}, io.BytesIO(b"<html>oops</html>")
        )
        with mock.patch("urllib.request.urlopen", side_effect=err):
            with self.assertRaises(MailblastrError) as ctx:
                http_client.request("GET", "/emails")
        self.assertEqual(ctx.exception.status_code, 500)
        self.assertEqual(ctx.exception.name, "application_error")

    def test_network_error(self):
        err = urllib.error.URLError("connection refused")
        with mock.patch("urllib.request.urlopen", side_effect=err):
            with self.assertRaises(MailblastrError) as ctx:
                http_client.request("GET", "/emails")
        self.assertEqual(ctx.exception.status_code, 0)
        self.assertEqual(ctx.exception.name, "network_error")

    def test_missing_api_key_raises(self):
        mailblastr.api_key = None
        with self.assertRaises(MailblastrError) as ctx:
            http_client.request("GET", "/emails")
        self.assertEqual(ctx.exception.name, "missing_api_key")

    def test_request_raw_returns_bytes_and_has_no_content_type(self):
        captured = {}

        def fake_urlopen(req, **kwargs):
            captured["req"] = req
            return FakeResponse(b"\x89PNG...")

        with mock.patch("urllib.request.urlopen", side_effect=fake_urlopen):
            data = http_client.request_raw("GET", "/emails/receiving/rcv_1/raw")
        self.assertEqual(data, b"\x89PNG...")
        self.assertEqual(captured["req"].get_header("Authorization"), "Bearer mb_test_key")
        self.assertIsNone(captured["req"].get_header("Content-type"))

    def test_empty_response_returns_none(self):
        with mock.patch("urllib.request.urlopen", return_value=FakeResponse(b"")):
            self.assertIsNone(http_client.request("DELETE", "/webhooks/wh_1"))

    def test_passes_timeout_to_urlopen(self):
        captured = {}

        def fake_urlopen(req, **kwargs):
            captured.update(kwargs)
            return FakeResponse(b'{"ok": true}')

        mailblastr.timeout = 12.5
        self.addCleanup(setattr, mailblastr, "timeout", None)
        with mock.patch("urllib.request.urlopen", side_effect=fake_urlopen):
            http_client.request("GET", "/emails")
        self.assertEqual(captured.get("timeout"), 12.5)

    def test_retries_429_then_succeeds(self):
        mailblastr.max_retries = 2
        self.addCleanup(setattr, mailblastr, "max_retries", None)
        calls = {"n": 0}

        def fake_urlopen(req, **kwargs):
            calls["n"] += 1
            if calls["n"] == 1:
                raise urllib.error.HTTPError(
                    "https://www.mailblastr.com/api/emails", 429, "Too Many",
                    {"Retry-After": "0"}, io.BytesIO(b""),
                )
            return FakeResponse(b'{"id": "em_ok"}')

        with mock.patch("urllib.request.urlopen", side_effect=fake_urlopen):
            with mock.patch("time.sleep"):  # don't actually wait
                result = http_client.request("POST", "/emails", {"x": 1})
        self.assertEqual(calls["n"], 2)
        self.assertEqual(result, {"id": "em_ok"})

    def test_gives_up_after_max_retries_on_persistent_503(self):
        mailblastr.max_retries = 2
        self.addCleanup(setattr, mailblastr, "max_retries", None)
        calls = {"n": 0}
        error_bodies = []

        def fake_urlopen(req, **kwargs):
            calls["n"] += 1
            body = io.BytesIO(b"")
            error_bodies.append(body)
            raise urllib.error.HTTPError(
                "https://www.mailblastr.com/api/emails", 503, "Unavailable", {}, body,
            )

        with mock.patch("urllib.request.urlopen", side_effect=fake_urlopen):
            with mock.patch("time.sleep"):
                with self.assertRaises(MailblastrError) as ctx:
                    http_client.request("GET", "/emails")
        self.assertEqual(calls["n"], 3)  # initial + 2 retries
        self.assertEqual(ctx.exception.status_code, 503)
        self.assertTrue(all(body.closed for body in error_bodies))

    def test_partial_batch_send_is_never_retried(self):
        """A 429 from /emails/batch that already sent some emails must NOT be
        retried — retrying would send those recipients a second time."""
        mailblastr.max_retries = 2
        self.addCleanup(setattr, mailblastr, "max_retries", None)
        calls = {"n": 0}
        body = json.dumps(
            {
                "statusCode": 429,
                "name": "daily_quota_exceeded",
                "message": "Daily sending quota reached.",
                "limit": {"kind": "emails_daily", "used": 100, "limit": 100},
                "sent": [{"id": "em_1"}, {"id": "em_2"}],
                "sent_count": 2,
            }
        ).encode("utf-8")

        def fake_urlopen(req, **kwargs):
            calls["n"] += 1
            raise urllib.error.HTTPError(
                "https://www.mailblastr.com/api/emails/batch", 429, "Too Many",
                {"Retry-After": "0"}, io.BytesIO(body),
            )

        with mock.patch("urllib.request.urlopen", side_effect=fake_urlopen):
            with mock.patch("time.sleep"):
                with self.assertRaises(MailblastrError) as ctx:
                    http_client.request("POST", "/emails/batch", [{"x": 1}])
        self.assertEqual(calls["n"], 1)  # no retry
        e = ctx.exception
        self.assertEqual(e.name, "daily_quota_exceeded")
        self.assertEqual(e.sent_count, 2)
        self.assertEqual(e.sent, [{"id": "em_1"}, {"id": "em_2"}])

    def test_error_exposes_limit_body_and_retry_after(self):
        body = json.dumps(
            {
                "statusCode": 402,
                "name": "plan_limit_reached",
                "message": "Domain limit reached.",
                "limit": {"kind": "domains", "used": 1, "limit": 1},
            }
        ).encode("utf-8")
        err = urllib.error.HTTPError(
            "https://www.mailblastr.com/api/domains", 402, "Payment Required",
            {"Retry-After": "12"}, io.BytesIO(body),
        )
        with mock.patch("urllib.request.urlopen", side_effect=err):
            with self.assertRaises(MailblastrError) as ctx:
                http_client.request("POST", "/domains", {"name": "acme.com"})
        e = ctx.exception
        self.assertEqual(e.status_code, 402)
        self.assertEqual(e.limit["kind"], "domains")
        self.assertEqual(e.retry_after, 12.0)
        self.assertEqual(http_client._retry_after_seconds("3600"), 30.0)  # sleep is capped
        self.assertEqual(http_client._parse_retry_after("3600"), 3600.0)  # report is not
        self.assertEqual(e.body["message"], "Domain limit reached.")
        self.assertEqual(e.sent, [])
        self.assertEqual(e.sent_count, 0)

    def test_error_extra_in_an_unexpected_shape_keeps_the_envelope(self):
        """An additive field this version cannot use costs the caller that
        field only — never the envelope, and never the other extras.

        Ruby, PHP, Java, Go, Rust and .NET all read a malformed extra as
        absent; without the isinstance guard ``e.limit["kind"]`` would raise a
        TypeError from inside the caller's own error handler."""
        body = json.dumps(
            {
                "statusCode": 429,
                "name": "daily_quota_exceeded",
                "message": "over quota",
                "limit": "emails_daily",
                "reputation": "paused",
                "sent": [{"id": "em_1"}],
            }
        ).encode("utf-8")
        err = urllib.error.HTTPError(
            "https://www.mailblastr.com/api/emails", 429, "Too Many Requests",
            {}, io.BytesIO(body),
        )
        with mock.patch("urllib.request.urlopen", side_effect=err):
            with self.assertRaises(MailblastrError) as ctx:
                http_client.request("POST", "/emails", {"x": 1})
        e = ctx.exception
        self.assertEqual(e.status_code, 429)
        self.assertEqual(e.name, "daily_quota_exceeded")
        self.assertEqual(e.message, "over quota")
        self.assertIsNone(e.limit)
        self.assertIsNone(e.reputation)
        # A bad `limit` must not drop `sent`.
        self.assertEqual(e.sent_count, 1)
        # The raw values stay reachable for a caller that wants them.
        self.assertEqual(e.body["limit"], "emails_daily")

    def _capture_idempotency_header(self, key):
        """Issue a send with ``idempotency_key`` and return the header sent."""
        captured = {}

        def fake_urlopen(req, **kwargs):
            captured["req"] = req
            return FakeResponse(b'{"id": "em_1"}')

        with mock.patch("urllib.request.urlopen", side_effect=fake_urlopen):
            http_client.request("POST", "/emails", {"x": 1}, {"idempotency_key": key})
        return captured["req"].get_header("Idempotency-key")

    def test_idempotency_key_is_sent_verbatim(self):
        """The server trims and bounds the key; the SDK does not touch it."""
        self.assertEqual(self._capture_idempotency_header("  order-1  "), "  order-1  ")

    def test_over_long_idempotency_key_is_left_to_the_server(self):
        """A 256-character key is a 400 invalid_idempotency_key from the API,
        not a locally raised error — the server owns the bound."""
        too_long = "k" * (http_client.IDEMPOTENCY_KEY_MAX_LENGTH + 1)
        self.assertEqual(self._capture_idempotency_header(too_long), too_long)

    def test_max_length_idempotency_key_is_accepted(self):
        """255 is the documented maximum — the boundary must not be off by one."""
        self.assertEqual(http_client.IDEMPOTENCY_KEY_MAX_LENGTH, 255)
        key = "k" * http_client.IDEMPOTENCY_KEY_MAX_LENGTH
        self.assertEqual(self._capture_idempotency_header(key), key)

    def test_blank_idempotency_key_is_not_sent(self):
        """A falsy key means "no idempotency", not an error."""
        self.assertIsNone(self._capture_idempotency_header(""))

    def test_whitespace_only_idempotency_key_is_left_to_the_server(self):
        """Whitespace-only trims to length 0 server-side — a 400, not a local
        raise. It is still transmitted so the API reports it."""
        self.assertEqual(self._capture_idempotency_header("   "), "   ")

    def test_422_is_not_retried(self):
        mailblastr.max_retries = 2
        self.addCleanup(setattr, mailblastr, "max_retries", None)
        calls = {"n": 0}

        def fake_urlopen(req, **kwargs):
            calls["n"] += 1
            raise urllib.error.HTTPError(
                "https://www.mailblastr.com/api/emails", 422, "Bad", {}, io.BytesIO(b""),
            )

        with mock.patch("urllib.request.urlopen", side_effect=fake_urlopen):
            with self.assertRaises(MailblastrError):
                http_client.request("POST", "/emails", {"x": 1})
        self.assertEqual(calls["n"], 1)


if __name__ == "__main__":
    unittest.main()
