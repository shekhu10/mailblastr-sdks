package mailblastr

import (
	"net/http"
	"testing"
)

func TestSegmentsCreateDomainRequired(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/segments" {
			t.Errorf("%s %s, want POST /segments", r.Method, r.URL.Path)
		}
		assertAuth(t, r)
		body := decodeBody(t, r)
		if body["domain"] != "yourdomain.com" {
			t.Errorf("domain = %v, want yourdomain.com", body["domain"])
		}
		if body["name"] != "VIP" {
			t.Errorf("name = %v, want VIP", body["name"])
		}
		filter, ok := body["filter"].(map[string]any)
		if !ok || filter["status"] != "subscribed" {
			t.Errorf("filter = %v", body["filter"])
		}
		w.Write([]byte(`{"object":"segment","id":"seg_1","audience_id":"aud_1","name":"VIP","filter":{"status":"subscribed","email_contains":null,"property_filters":[]},"created_at":"2026-07-06T00:00:00Z","updated_at":"2026-07-06T00:00:00Z"}`))
	})

	seg, err := client.Segments.Create(&CreateSegmentRequest{
		Domain: "yourdomain.com",
		Name:   "VIP",
		Filter: &SegmentFilter{Status: "subscribed"},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if seg.Id != "seg_1" || seg.Name != "VIP" || seg.Filter.Status != "subscribed" {
		t.Errorf("unexpected segment: %+v", seg)
	}
}

func TestSegmentsListDomainQuery(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/segments" {
			t.Errorf("path = %s", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("domain") != "yourdomain.com" {
			t.Errorf("domain = %q, want yourdomain.com", q.Get("domain"))
		}
		if q.Get("limit") != "10" {
			t.Errorf("limit = %q, want 10", q.Get("limit"))
		}
		w.Write([]byte(`{"object":"list","has_more":false,"data":[{"object":"segment","id":"seg_general","audience_id":"aud_1","name":"General","filter":{"status":"all","email_contains":null,"property_filters":[]},"created_at":"2026-07-06T00:00:00Z","updated_at":"2026-07-06T00:00:00Z"}]}`))
	})

	list, err := client.Segments.List(&ListSegmentsRequest{Domain: "yourdomain.com", Limit: 10})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list.Data) != 1 || list.Data[0].Name != "General" {
		t.Errorf("unexpected list: %+v", list)
	}
}

func TestSegmentsContactsPreview(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/segments/seg_1/contacts" {
			t.Errorf("%s %s, want GET /segments/seg_1/contacts", r.Method, r.URL.Path)
		}
		w.Write([]byte(`{"object":"list","has_more":false,"data":[]}`))
	})

	list, err := client.Segments.Contacts("seg_1")
	if err != nil {
		t.Fatalf("Contacts: %v", err)
	}
	if list.Object != "list" || len(list.Data) != 0 {
		t.Errorf("unexpected list: %+v", list)
	}
}
