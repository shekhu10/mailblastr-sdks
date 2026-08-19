import base64
import hashlib
import hmac
import time

import mailblastr
from mailblastr.webhooks import _b64_lenient

from .helpers import RecordingTestCase


class PairIteratingHeaders:
    """Stand-in for Flask/werkzeug's ``request.headers``.

    Werkzeug's ``Headers.__iter__`` yields ``(name, value)`` TUPLES, not names —
    unlike a dict, Django's ``HttpHeaders`` or Starlette's ``Headers``. The
    package README passes ``request.headers`` straight into ``Webhooks.verify``,
    so this shape has to work.
    """

    def __init__(self, pairs):
        self._list = list(pairs)

    def __iter__(self):
        return iter(self._list)

    def items(self):
        return iter(self._list)

    def __getitem__(self, key):
        for name, value in self._list:
            if name.lower() == str(key).lower():
                return value
        raise KeyError(key)


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

    def test_failed_test_delivery_is_a_200_not_an_error(self):
        # A rejected test delivery still comes back HTTP 200, so this must not
        # raise — the outcome lives in "ok", and treating 200 as success would
        # silently report a broken endpoint as healthy.
        self.response = {
            "object": "webhook_test",
            "id": "wh_1",
            "ok": False,
            "error": "lookup_failed",
        }
        result = mailblastr.Webhooks.test("wh_1")
        self.assertCall("POST", "/webhooks/wh_1/test")
        self.assertFalse(result["ok"])
        self.assertEqual(result["error"], "lookup_failed")

    def test_successful_test_delivery_reports_the_endpoint_status(self):
        self.response = {
            "object": "webhook_test",
            "id": "wh_1",
            "ok": True,
            "status": 200,
        }
        result = mailblastr.Webhooks.test("wh_1")
        self.assertTrue(result["ok"])
        self.assertEqual(result["status"], 200)


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

    def test_flask_style_headers_that_iterate_pairs(self):
        """A container whose iteration yields (name, value) pairs — Flask's
        ``request.headers`` — must still be read. Iterating it as if each element
        were a NAME matched nothing, so every genuinely signed delivery came back
        `missing_headers` on the exact call the README documents."""
        secret = "topsecret"
        headers = PairIteratingHeaders(self.headers_for(secret.encode("utf-8")).items())
        result = mailblastr.Webhooks.verify(self.payload, headers, secret)
        self.assertEqual(result, {"valid": True})

    def test_unpadded_whsec_secret_verifies(self):
        """The signer derives its key with Node's lenient base64 decoder, which
        needs no `=` padding. Python's strict ``b64decode`` raised on the same
        suffix, the SDK fell back to the raw string, and a valid delivery was
        rejected as `no_match`."""
        key = b"abcd"
        unpadded = base64.b64encode(key).decode("ascii").rstrip("=")
        self.assertEqual(unpadded, "YWJjZA")  # length 6 — not a multiple of 4
        headers = self.headers_for(key)
        result = mailblastr.Webhooks.verify(self.payload, headers, "whsec_" + unpadded)
        self.assertEqual(result, {"valid": True})

    def test_url_safe_whsec_secret_verifies(self):
        """Node's base64 decoder also accepts the URL-safe `-`/`_` spellings;
        ``b64decode`` silently DISCARDS them, yielding a different key."""
        key = bytes(range(250, 256)) + b"\xfb\xff\xbe"
        urlsafe = base64.urlsafe_b64encode(key).decode("ascii").rstrip("=")
        self.assertTrue("-" in urlsafe or "_" in urlsafe)
        headers = self.headers_for(key)
        result = mailblastr.Webhooks.verify(self.payload, headers, "whsec_" + urlsafe)
        self.assertEqual(result, {"valid": True})

    def test_lenient_base64_matches_node_byte_for_byte(self):
        """Pinned against `node -e "Buffer.from(s, 'base64')"`, which is what the
        backend's secretToKey uses (lib/crypto.ts). A trailing single character
        carries no whole byte and is dropped; 2 or 3 decode as a partial group."""
        self.assertEqual(_b64_lenient("abcd"), b"\x69\xb7\x1d")
        self.assertEqual(_b64_lenient("hello"), b"\x85\xe9\x65")  # 5 chars -> drop 1
        self.assertEqual(_b64_lenient("abc"), b"\x69\xb7")
        self.assertEqual(_b64_lenient("ab"), b"\x69")
        self.assertEqual(_b64_lenient("YWJjZA"), b"abcd")
        self.assertEqual(_b64_lenient("ab-cd_ef"), b"\x69\xbf\x9c\x77\xf7\x9f")

    def test_unrecognized_secret_still_falls_back_to_raw_bytes(self):
        """A `whsec_` secret whose suffix decodes to nothing keeps the documented
        fallback: the WHOLE string as UTF-8, exactly like the server."""
        secret = "whsec_"
        headers = self.headers_for(secret.encode("utf-8"))
        self.assertEqual(mailblastr.Webhooks.verify(self.payload, headers, secret),
                         {"valid": True})
