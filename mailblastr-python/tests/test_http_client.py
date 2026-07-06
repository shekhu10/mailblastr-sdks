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

        def fake_urlopen(req):
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
        self.assertEqual(req.full_url, "https://api.mailblastr.com/emails")
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

        def fake_urlopen(req):
            captured["req"] = req
            return FakeResponse(b'{"object":"list","data":[]}')

        with mock.patch("urllib.request.urlopen", side_effect=fake_urlopen):
            http_client.request("GET", "/domains")
        self.assertIsNone(captured["req"].data)
        self.assertEqual(captured["req"].get_method(), "GET")

    def test_base_url_override(self):
        mailblastr.base_url = "http://localhost:3000/"  # trailing slash stripped
        captured = {}

        def fake_urlopen(req):
            captured["req"] = req
            return FakeResponse(b"{}")

        with mock.patch("urllib.request.urlopen", side_effect=fake_urlopen):
            http_client.request("GET", "/emails")
        self.assertEqual(captured["req"].full_url, "http://localhost:3000/emails")

    def test_non_2xx_raises_mailblastr_error_with_api_body(self):
        body = json.dumps(
            {"statusCode": 422, "name": "validation_error", "message": "domain is required"}
        ).encode("utf-8")
        err = urllib.error.HTTPError(
            "https://api.mailblastr.com/segments", 422, "Unprocessable", {}, io.BytesIO(body)
        )
        with mock.patch("urllib.request.urlopen", side_effect=err):
            with self.assertRaises(MailblastrError) as ctx:
                http_client.request("POST", "/segments", {"name": "VIP"})
        e = ctx.exception
        self.assertEqual(e.status_code, 422)
        self.assertEqual(e.statusCode, 422)
        self.assertEqual(e.name, "validation_error")
        self.assertEqual(e.message, "domain is required")

    def test_non_json_error_body_falls_back_to_status(self):
        err = urllib.error.HTTPError(
            "https://api.mailblastr.com/emails", 500, "Boom", {}, io.BytesIO(b"<html>oops</html>")
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

        def fake_urlopen(req):
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


if __name__ == "__main__":
    unittest.main()
