package mailblastr

import (
	"net/http"
	"testing"
)

func TestContactsCreateFlatDomainFirst(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/contacts" {
			t.Errorf("%s %s, want POST /contacts", r.Method, r.URL.Path)
		}
		assertAuth(t, r)
		body := decodeBody(t, r)
		if body["domain"] != "yourdomain.com" {
			t.Errorf("domain = %v, want yourdomain.com", body["domain"])
		}
		if body["email"] != "user@example.com" {
			t.Errorf("email = %v", body["email"])
		}
		if body["first_name"] != "Ada" {
			t.Errorf("first_name = %v", body["first_name"])
		}
		w.Write([]byte(`{"object":"contact","id":"con_1"}`))
	})

	ref, err := client.Contacts.Create(&CreateContactRequest{
		Domain:    "yourdomain.com",
		Email:     "user@example.com",
		FirstName: "Ada",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if ref.Object != "contact" || ref.Id != "con_1" {
		t.Errorf("unexpected ref: %+v", ref)
	}
}

func TestContactsCreateNestedAudienceStripsDomain(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/audiences/aud_1/contacts" {
			t.Errorf("path = %s, want /audiences/aud_1/contacts", r.URL.Path)
		}
		body := decodeBody(t, r)
		if _, present := body["domain"]; present {
			t.Error("domain must not be sent on the nested audience route")
		}
		w.Write([]byte(`{"object":"contact","id":"con_2"}`))
	})

	_, err := client.Contacts.Create(&CreateContactRequest{
		AudienceId: "aud_1",
		Domain:     "yourdomain.com", // must be stripped
		Email:      "user@example.com",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
}

func TestContactsListFlatRequiresDomainQuery(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/contacts" {
			t.Errorf("path = %s", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("domain") != "yourdomain.com" {
			t.Errorf("domain query = %q", q.Get("domain"))
		}
		if q.Get("limit") != "50" || q.Get("segment_id") != "seg_1" {
			t.Errorf("query = %v", q)
		}
		w.Write([]byte(`{"object":"list","has_more":false,"data":[{"object":"contact","id":"con_1","email":"user@example.com","unsubscribed":false,"created_at":"2026-07-06T00:00:00Z"}]}`))
	})

	list, err := client.Contacts.List(&ListContactsRequest{
		Domain:    "yourdomain.com",
		Limit:     50,
		SegmentId: "seg_1",
	})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list.Data) != 1 || list.Data[0].Email != "user@example.com" {
		t.Errorf("unexpected list: %+v", list)
	}
}

func TestContactsGetByEmailWithDomain(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		// PathEscape leaves '@' intact (a valid path-segment character).
		if r.URL.Path != "/contacts/user@example.com" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if r.URL.Query().Get("domain") != "yourdomain.com" {
			t.Errorf("domain query = %q", r.URL.Query().Get("domain"))
		}
		w.Write([]byte(`{"object":"contact","id":"con_1","email":"user@example.com","unsubscribed":false,"created_at":"2026-07-06T00:00:00Z"}`))
	})

	got, err := client.Contacts.Get(&GetContactRequest{Id: "user@example.com", Domain: "yourdomain.com"})
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Id != "con_1" {
		t.Errorf("Id = %q", got.Id)
	}
}

func TestContactsUpdateTriState(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch || r.URL.Path != "/contacts/con_1" {
			t.Errorf("%s %s, want PATCH /contacts/con_1", r.Method, r.URL.Path)
		}
		body := decodeBody(t, r)
		if body["unsubscribed"] != true {
			t.Errorf("unsubscribed = %v, want true", body["unsubscribed"])
		}
		if _, present := body["first_name"]; present {
			t.Error("unset first_name must be omitted")
		}
		w.Write([]byte(`{"object":"contact","id":"con_1"}`))
	})

	_, err := client.Contacts.Update(&UpdateContactRequest{
		Id:           "con_1",
		Unsubscribed: Bool(true),
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
}

func TestContactsBatchOnConflict(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/audiences/aud_1/contacts/batch" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if r.URL.Query().Get("on_conflict") != "skip" {
			t.Errorf("on_conflict = %q", r.URL.Query().Get("on_conflict"))
		}
		body := decodeBody(t, r)
		contacts, ok := body["contacts"].([]any)
		if !ok || len(contacts) != 2 {
			t.Errorf("contacts = %v", body["contacts"])
		}
		w.Write([]byte(`{"object":"list","imported":1,"updated":0,"skipped":1,"total":2}`))
	})

	res, err := client.Contacts.Batch(&BatchContactsRequest{
		AudienceId: "aud_1",
		OnConflict: "skip",
		Contacts: []ContactInput{
			{Email: "a@example.com"},
			{Email: "b@example.com"},
		},
	})
	if err != nil {
		t.Fatalf("Batch: %v", err)
	}
	if res.Imported != 1 || res.Skipped != 1 || res.Total != 2 {
		t.Errorf("unexpected import result: %+v", res)
	}
}

func TestContactsSegmentMembership(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/contacts/con_1/segments/seg_1" {
			t.Errorf("%s %s, want POST /contacts/con_1/segments/seg_1", r.Method, r.URL.Path)
		}
		w.Write([]byte(`{"id":"seg_1"}`))
	})

	res, err := client.Contacts.AddToSegment("con_1", "seg_1")
	if err != nil {
		t.Fatalf("AddToSegment: %v", err)
	}
	if res.Id != "seg_1" {
		t.Errorf("Id = %q, want seg_1", res.Id)
	}
}
