package mailblastr

import (
	"net/http"
	"testing"
)

func TestCampaignsCreateDomainRequired(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/campaigns" {
			t.Errorf("%s %s, want POST /campaigns", r.Method, r.URL.Path)
		}
		assertAuth(t, r)
		body := decodeBody(t, r)
		if body["domain"] != "yourdomain.com" {
			t.Errorf("domain = %v, want yourdomain.com", body["domain"])
		}
		if body["from"] != "Acme <hello@yourdomain.com>" {
			t.Errorf("from = %v", body["from"])
		}
		if body["subject"] != "Launch" {
			t.Errorf("subject = %v", body["subject"])
		}
		if body["segment_id"] != "seg_1" {
			t.Errorf("segment_id = %v", body["segment_id"])
		}
		ab, ok := body["ab_test"].(map[string]any)
		if !ok || ab["enabled"] != true || ab["subject_b"] != "Launch v2" {
			t.Errorf("ab_test = %v", body["ab_test"])
		}
		w.Write([]byte(`{"id":"cmp_1"}`))
	})

	res, err := client.Campaigns.Create(&CreateCampaignRequest{
		Domain:    "yourdomain.com",
		From:      "Acme <hello@yourdomain.com>",
		Subject:   "Launch",
		Html:      "<p>We launched!</p>",
		SegmentId: "seg_1",
		AbTest:    &CampaignAbTest{Enabled: true, SubjectB: "Launch v2", Metric: "open"},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if res.Id != "cmp_1" {
		t.Errorf("Id = %q, want cmp_1", res.Id)
	}
}

func TestCampaignsSendScheduled(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/campaigns/cmp_1/send" {
			t.Errorf("%s %s, want POST /campaigns/cmp_1/send", r.Method, r.URL.Path)
		}
		body := decodeBody(t, r)
		if body["scheduled_at"] != "2026-08-01T09:00:00Z" {
			t.Errorf("scheduled_at = %v", body["scheduled_at"])
		}
		w.Write([]byte(`{"id":"cmp_1"}`))
	})

	res, err := client.Campaigns.Send("cmp_1", &SendCampaignRequest{ScheduledAt: "2026-08-01T09:00:00Z"})
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	if res.Id != "cmp_1" {
		t.Errorf("Id = %q", res.Id)
	}
}

func TestCampaignsSendNowNilParamsSendsEmptyObject(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		body := decodeBody(t, r)
		if len(body) != 0 {
			t.Errorf("body = %v, want empty object", body)
		}
		w.Write([]byte(`{"id":"cmp_1"}`))
	})

	if _, err := client.Campaigns.Send("cmp_1", nil); err != nil {
		t.Fatalf("Send: %v", err)
	}
}

func TestCampaignsStats(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/campaigns/cmp_1/stats" {
			t.Errorf("%s %s, want GET /campaigns/cmp_1/stats", r.Method, r.URL.Path)
		}
		w.Write([]byte(`{"object":"campaign_stats","campaign_id":"cmp_1","links":[{"url":"https://example.com","clicks":3}]}`))
	})

	stats, err := client.Campaigns.Stats("cmp_1")
	if err != nil {
		t.Fatalf("Stats: %v", err)
	}
	if stats.CampaignId != "cmp_1" || len(stats.Links) != 1 {
		t.Errorf("unexpected stats: %+v", stats)
	}
}
