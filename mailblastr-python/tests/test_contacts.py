import mailblastr

from .helpers import RecordingTestCase


class TestContacts(RecordingTestCase):
    def test_create_flat_requires_domain_in_body(self):
        mailblastr.Contacts.create(
            {"domain": "acme.com", "email": "user@example.com", "first_name": "Ada"}
        )
        self.assertCall(
            "POST",
            "/contacts",
            {"email": "user@example.com", "first_name": "Ada", "domain": "acme.com"},
        )

    def test_create_nested_audience_strips_routing_keys(self):
        mailblastr.Contacts.create({"audience_id": "aud_1", "email": "user@example.com"})
        self.assertCall("POST", "/audiences/aud_1/contacts", {"email": "user@example.com"})

    def test_get_by_id(self):
        mailblastr.Contacts.get({"id": "con_1"})
        self.assertCall("GET", "/contacts/con_1")

    def test_get_by_email_and_domain(self):
        mailblastr.Contacts.get({"id": "user@example.com", "domain": "acme.com"})
        self.assertCall("GET", "/contacts/user%40example.com?domain=acme.com")

    def test_get_nested_audience(self):
        mailblastr.Contacts.get({"id": "con_1", "audience_id": "aud_1"})
        self.assertCall("GET", "/audiences/aud_1/contacts/con_1")

    def test_list_flat_domain(self):
        mailblastr.Contacts.list({"domain": "acme.com", "limit": 50})
        self.assertCall("GET", "/contacts?domain=acme.com&limit=50")

    def test_list_audience_with_segment_filter(self):
        mailblastr.Contacts.list({"audience_id": "aud_1", "segment_id": "seg_1"})
        self.assertCall("GET", "/audiences/aud_1/contacts?segment_id=seg_1")

    def test_batch_import(self):
        contacts = [{"email": "a@b.com"}, {"email": "c@d.com", "first_name": "C"}]
        mailblastr.Contacts.batch(
            {"audience_id": "aud_1", "contacts": contacts, "on_conflict": "skip"}
        )
        self.assertCall(
            "POST", "/audiences/aud_1/contacts/batch?on_conflict=skip", {"contacts": contacts}
        )

    def test_import_csv(self):
        mailblastr.Contacts.import_csv(
            {
                "audience_id": "aud_1",
                "csv": "email,company\na@b.com,Acme",
                "on_conflict": "skip",
                "create_properties": False,
            }
        )
        self.assertCall(
            "POST",
            "/audiences/aud_1/contacts/import?on_conflict=skip&create_properties=false",
            {"csv": "email,company\na@b.com,Acme"},
        )

    def test_import_csv_defaults(self):
        mailblastr.Contacts.import_csv({"audience_id": "aud_1", "csv": "email\na@b.com"})
        self.assertCall("POST", "/audiences/aud_1/contacts/import", {"csv": "email\na@b.com"})

    def test_update_flat_email_with_domain(self):
        mailblastr.Contacts.update(
            {"id": "user@example.com", "domain": "acme.com", "unsubscribed": True}
        )
        self.assertCall(
            "PATCH",
            "/contacts/user%40example.com",
            {"unsubscribed": True, "domain": "acme.com"},
        )

    def test_update_flat_by_id_no_domain(self):
        mailblastr.Contacts.update({"id": "con_1", "first_name": "Ada"})
        self.assertCall("PATCH", "/contacts/con_1", {"first_name": "Ada"})

    def test_update_nested_audience(self):
        mailblastr.Contacts.update({"id": "con_1", "audience_id": "aud_1", "last_name": "L"})
        self.assertCall("PATCH", "/audiences/aud_1/contacts/con_1", {"last_name": "L"})

    def test_remove_flat_with_domain(self):
        mailblastr.Contacts.remove({"id": "user@example.com", "domain": "acme.com"})
        self.assertCall("DELETE", "/contacts/user%40example.com?domain=acme.com")

    def test_remove_nested(self):
        mailblastr.Contacts.remove({"id": "con_1", "audience_id": "aud_1"})
        self.assertCall("DELETE", "/audiences/aud_1/contacts/con_1")

    def test_segment_membership(self):
        mailblastr.Contacts.add_to_segment("con_1", "seg_1")
        self.assertCall("POST", "/contacts/con_1/segments/seg_1")
        mailblastr.Contacts.remove_from_segment("con_1", "seg_1")
        self.assertCall("DELETE", "/contacts/con_1/segments/seg_1")
        mailblastr.Contacts.list_segments("con_1")
        self.assertCall("GET", "/contacts/con_1/segments")

    def test_topics(self):
        mailblastr.Contacts.get_topics("con_1")
        self.assertCall("GET", "/contacts/con_1/topics")
        payload = {"topics": [{"id": "top_1", "subscription": "opt_in"}]}
        mailblastr.Contacts.update_topics("con_1", payload)
        self.assertCall("PATCH", "/contacts/con_1/topics", payload)
