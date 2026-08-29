import mailblastr

from .helpers import RecordingTestCase


class TestEmails(RecordingTestCase):
    def test_send(self):
        params = {
            "from": "Acme <hello@acme.com>",
            "to": ["user@example.com"],
            "subject": "Hi",
            "html": "<p>Hi</p>",
        }
        result = mailblastr.Emails.send(params)
        self.assertCall("POST", "/emails", params)
        self.assertEqual(result, self.response)

    def test_send_with_idempotency_key(self):
        params = {"from": "a@b.com", "to": "c@d.com", "subject": "s", "text": "t"}
        mailblastr.Emails.send(params, options={"idempotency_key": "order-123"})
        self.assertCall("POST", "/emails", params, {"idempotency_key": "order-123"})

    def test_batch_alias(self):
        payloads = [{"from": "a@b.com", "to": "x@y.com", "subject": "s", "text": "t"}]
        mailblastr.Emails.batch(payloads)
        self.assertCall("POST", "/emails/batch", payloads)

    def test_list_with_pagination(self):
        mailblastr.Emails.list({"limit": 20, "after": "em_1"})
        self.assertCall("GET", "/emails?limit=20&after=em_1")

    def test_list_no_params(self):
        mailblastr.Emails.list()
        self.assertCall("GET", "/emails")

    def test_list_with_status_and_search_filters(self):
        mailblastr.Emails.list({"status": "bounced", "search": "acme.com"})
        self.assertCall("GET", "/emails?status=bounced&search=acme.com")

    def test_list_with_mailbox_folder(self):
        mailblastr.Emails.list({"folder": "outbox"})
        self.assertCall("GET", "/emails?folder=outbox")

    def test_list_with_source_filters(self):
        mailblastr.Emails.list({"campaign_id": "cmp_1", "domain_id": "dom_1"})
        self.assertCall("GET", "/emails?campaign_id=cmp_1&domain_id=dom_1")

    def test_sources(self):
        mailblastr.Emails.sources()
        self.assertCall("GET", "/emails/sources")

    def test_get(self):
        mailblastr.Emails.get("em_123")
        self.assertCall("GET", "/emails/em_123")

    def test_get_path_escapes_id(self):
        mailblastr.Emails.get("../api-keys")
        self.assertCall("GET", "/emails/..%2Fapi-keys")

    def test_update_reschedule(self):
        mailblastr.Emails.update("em_123", {"scheduled_at": "2026-08-01T09:00:00Z"})
        self.assertCall("PATCH", "/emails/em_123", {"scheduled_at": "2026-08-01T09:00:00Z"})

    def test_cancel(self):
        mailblastr.Emails.cancel("em_123")
        self.assertCall("POST", "/emails/em_123/cancel")


class TestBatch(RecordingTestCase):
    def test_send(self):
        payloads = [
            {"from": "a@b.com", "to": "x@y.com", "subject": "1", "text": "t"},
            {"from": "a@b.com", "to": "z@y.com", "subject": "2", "text": "t"},
        ]
        mailblastr.Batch.send(payloads)
        self.assertCall("POST", "/emails/batch", payloads)


class TestEmailAttachments(RecordingTestCase):
    def test_list(self):
        mailblastr.Emails.list_attachments("em_1")
        self.assertCall("GET", "/emails/em_1/attachments")

    def test_get(self):
        mailblastr.Emails.get_attachment("em_1", "att_1")
        self.assertCall("GET", "/emails/em_1/attachments/att_1")


class TestReceiving(RecordingTestCase):
    def test_list(self):
        mailblastr.Emails.Receiving.list({"limit": 5})
        self.assertCall("GET", "/emails/receiving?limit=5")

    def test_get(self):
        mailblastr.Emails.Receiving.get("rcv_1")
        self.assertCall("GET", "/emails/receiving/rcv_1")

    def test_addresses(self):
        mailblastr.Emails.Receiving.list_addresses()
        self.assertCall("GET", "/emails/receiving/addresses")

    def test_list_filtered_by_received_for(self):
        mailblastr.Emails.Receiving.list({"received_for": "hi@acme.com"})
        self.assertCall("GET", "/emails/receiving?received_for=hi%40acme.com")

    def test_attachments(self):
        mailblastr.Emails.Receiving.list_attachments("rcv_1")
        self.assertCall("GET", "/emails/receiving/rcv_1/attachments")

    def test_attachments_paginated(self):
        mailblastr.Emails.Receiving.list_attachments("rcv_1", {"limit": 10, "after": "2"})
        self.assertCall("GET", "/emails/receiving/rcv_1/attachments?limit=10&after=2")

    def test_get_attachment_is_binary(self):
        data = mailblastr.Emails.Receiving.get_attachment("rcv_1", "att_9")
        self.assertEqual(self.last["path"], "/emails/receiving/rcv_1/attachments/att_9")
        self.assertEqual(self.last["method"], "GET")
        self.assertTrue(self.last.get("raw"))
        self.assertEqual(data, self.raw_response)

    def test_raw_is_binary(self):
        data = mailblastr.Emails.Receiving.get_raw("rcv_1")
        self.assertEqual(self.last["path"], "/emails/receiving/rcv_1/raw")
        self.assertTrue(self.last.get("raw"))
        self.assertEqual(data, self.raw_response)

    def test_forward(self):
        payload = {"from": "me@acme.com", "to": "team@acme.com"}
        mailblastr.Emails.Receiving.forward("rcv_1", payload)
        self.assertCall("POST", "/emails/receiving/rcv_1/forward", payload)

    def test_reply(self):
        payload = {"from": "me@acme.com", "html": "<p>Thanks</p>"}
        mailblastr.Emails.Receiving.reply("rcv_1", payload)
        self.assertCall("POST", "/emails/receiving/rcv_1/reply", payload)

    def test_remove(self):
        mailblastr.Emails.Receiving.remove("rcv_1")
        self.assertCall("DELETE", "/emails/receiving/rcv_1")
