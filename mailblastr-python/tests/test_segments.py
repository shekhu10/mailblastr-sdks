import mailblastr

from .helpers import RecordingTestCase


class TestSegments(RecordingTestCase):
    def test_create_domain_required_in_body(self):
        params = {"domain": "acme.com", "name": "VIP", "filter": {"status": "subscribed"}}
        mailblastr.Segments.create(params)
        self.assertCall("POST", "/segments", params)

    def test_get(self):
        mailblastr.Segments.get("seg_1")
        self.assertCall("GET", "/segments/seg_1")

    def test_list_domain_scoped(self):
        mailblastr.Segments.list({"domain": "acme.com", "limit": 10, "after": "seg_9"})
        self.assertCall("GET", "/segments?domain=acme.com&limit=10&after=seg_9")

    def test_list_requires_domain(self):
        with self.assertRaises(KeyError):
            mailblastr.Segments.list({})

    def test_contacts_preview(self):
        mailblastr.Segments.contacts("seg_1")
        self.assertCall("GET", "/segments/seg_1/contacts")

    def test_contacts_preview_paginated(self):
        mailblastr.Segments.contacts("seg_1", {"limit": 100, "after": "con_9"})
        self.assertCall("GET", "/segments/seg_1/contacts?limit=100&after=con_9")

    def test_update(self):
        mailblastr.Segments.update("seg_1", {"name": "VIP+"})
        self.assertCall("PATCH", "/segments/seg_1", {"name": "VIP+"})

    def test_remove(self):
        mailblastr.Segments.remove("seg_1")
        self.assertCall("DELETE", "/segments/seg_1")


class TestTopics(RecordingTestCase):
    def test_create_domain_required_in_body(self):
        params = {"domain": "acme.com", "name": "Product updates", "default_subscription": "opt_in"}
        mailblastr.Topics.create(params)
        self.assertCall("POST", "/topics", params)

    def test_list_domain_scoped(self):
        mailblastr.Topics.list({"domain": "acme.com", "limit": 50})
        self.assertCall("GET", "/topics?domain=acme.com&limit=50")

    def test_get_update_remove(self):
        mailblastr.Topics.get("top_1")
        self.assertCall("GET", "/topics/top_1")
        mailblastr.Topics.update("top_1", {"visibility": "private"})
        self.assertCall("PATCH", "/topics/top_1", {"visibility": "private"})
        mailblastr.Topics.remove("top_1")
        self.assertCall("DELETE", "/topics/top_1")
