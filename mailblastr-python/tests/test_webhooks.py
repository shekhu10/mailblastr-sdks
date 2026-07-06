import base64
import hashlib
import hmac
import time

import mailblastr

from .helpers import RecordingTestCase


def _sign(secret_key: bytes, msg_id: str, timestamp: str, payload: str) -> str:
    signed = f"{msg_id}.{timestamp}.{payload}"
    return base64.b64encode(
        hmac.new(secret_key, signed.encode("utf-8"), hashlib.sha256).digest()
    ).decode("ascii")


class TestWebhooksApi(RecordingTestCase):
    def test_create(self):
        params = {
            "endpoint": "https://yourapp.com/hooks",
            "events": ["email.delivered", "email.bounced"],
        }
        mailblastr.Webhooks.create(params)
        self.assertCall("POST", "/webhooks", params)

    def test_get_list_update_remove(self):
        mailblastr.Webhooks.get("wh_1")
        self.assertCall("GET", "/webhooks/wh_1")
        mailblastr.Webhooks.list({"limit": 10})
        self.assertCall("GET", "/webhooks?limit=10")
        mailblastr.Webhooks.update("wh_1", {"status": "disabled"})
        self.assertCall("PATCH", "/webhooks/wh_1", {"status": "disabled"})
        mailblastr.Webhooks.remove("wh_1")
        self.assertCall("DELETE", "/webhooks/wh_1")

    def test_rotate_and_test(self):
        mailblastr.Webhooks.rotate("wh_1")
        self.assertCall("POST", "/webhooks/wh_1/rotate")
        mailblastr.Webhooks.test("wh_1")
        self.assertCall("POST", "/webhooks/wh_1/test")


class TestVerifySignature(RecordingTestCase):
    payload = '{"type":"email.delivered","data":{"id":"em_1"}}'

    def headers_for(self, secret_key: bytes, ts=None, msg_id="msg_1"):
        ts = str(int(time.time())) if ts is None else ts
        sig = _sign(secret_key, msg_id, ts, self.payload)
        return {
            "svix-id": msg_id,
            "svix-timestamp": ts,
            "svix-signature": f"v1,{sig}",
        }

    def test_valid_raw_secret(self):
        secret = "topsecret"
        headers = self.headers_for(secret.encode("utf-8"))
        result = mailblastr.Webhooks.verify(self.payload, headers, secret)
        self.assertEqual(result, {"valid": True})

    def test_valid_whsec_prefixed_secret(self):
        key = b"0123456789abcdef0123456789abcdef"
        secret = "whsec_" + base64.b64encode(key).decode("ascii")
        headers = self.headers_for(key)
        result = mailblastr.verify_webhook_signature(self.payload, headers, secret)
        self.assertEqual(result, {"valid": True})

    def test_case_insensitive_headers_and_multi_sig(self):
        secret = "topsecret"
        ts = str(int(time.time()))
        good = _sign(secret.encode(), "msg_1", ts, self.payload)
        headers = {
            "Svix-Id": "msg_1",
            "SVIX-TIMESTAMP": ts,
            "Svix-Signature": f"v1,badbadbad v1,{good}",
        }
        self.assertTrue(mailblastr.Webhooks.verify(self.payload, headers, secret)["valid"])

    def test_tampered_payload_rejected(self):
        secret = "topsecret"
        headers = self.headers_for(secret.encode("utf-8"))
        result = mailblastr.Webhooks.verify(self.payload + " ", headers, secret)
        self.assertEqual(result, {"valid": False, "reason": "no_match"})

    def test_missing_headers(self):
        result = mailblastr.Webhooks.verify(self.payload, {"svix-id": "msg_1"}, "s")
        self.assertEqual(result, {"valid": False, "reason": "missing_headers"})

    def test_missing_secret(self):
        headers = self.headers_for(b"x")
        result = mailblastr.Webhooks.verify(self.payload, headers, "")
        self.assertEqual(result, {"valid": False, "reason": "missing_secret"})

    def test_stale_timestamp_rejected(self):
        secret = "topsecret"
        stale = str(int(time.time()) - 3600)
        headers = self.headers_for(secret.encode("utf-8"), ts=stale)
        result = mailblastr.Webhooks.verify(self.payload, headers, secret)
        self.assertEqual(result, {"valid": False, "reason": "timestamp_out_of_tolerance"})

    def test_stale_timestamp_accepted_when_tolerance_disabled(self):
        secret = "topsecret"
        stale = str(int(time.time()) - 3600)
        headers = self.headers_for(secret.encode("utf-8"), ts=stale)
        result = mailblastr.Webhooks.verify(self.payload, headers, secret, tolerance_sec=0)
        self.assertEqual(result, {"valid": True})

    def test_invalid_timestamp(self):
        secret = "topsecret"
        headers = self.headers_for(secret.encode("utf-8"), ts="not-a-number")
        result = mailblastr.Webhooks.verify(self.payload, headers, secret)
        self.assertEqual(result, {"valid": False, "reason": "invalid_timestamp"})

    def test_bytes_payload_accepted(self):
        secret = "topsecret"
        headers = self.headers_for(secret.encode("utf-8"))
        result = mailblastr.Webhooks.verify(self.payload.encode("utf-8"), headers, secret)
        self.assertEqual(result, {"valid": True})
