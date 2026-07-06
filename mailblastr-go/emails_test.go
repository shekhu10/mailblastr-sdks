package mailblastr

import (
	"context"
	"net/http"
	"testing"
)

func TestEmailsSend(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/emails" {
			t.Errorf("path = %s, want /emails", r.URL.Path)
		}
		assertAuth(t, r)
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", ct)
		}
		body := decodeBody(t, r)
		if body["from"] != "Acme <hello@yourdomain.com>" {
			t.Errorf("from = %v", body["from"])
		}
		to, ok := body["to"].([]any)
		if !ok || len(to) != 1 || to[0] != "user@example.com" {
			t.Errorf("to = %v", body["to"])
		}
		if body["subject"] != "Hello" {
			t.Errorf("subject = %v", body["subject"])
		}
		if body["html"] != "<p>Hi</p>" {
			t.Errorf("html = %v", body["html"])
		}
		if _, present := body["text"]; present {
			t.Error("empty text should be omitted from the body")
		}
		w.Write([]byte(`{"id":"em_123"}`))
	})

	sent, err := client.Emails.Send(&SendEmailRequest{
		From:    "Acme <hello@yourdomain.com>",
		To:      []string{"user@example.com"},
		Subject: "Hello",
		Html:    "<p>Hi</p>",
	})
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	if sent.Id != "em_123" {
		t.Errorf("Id = %q, want em_123", sent.Id)
	}
}

func TestEmailsSendWithOptionsIdempotencyKey(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Idempotency-Key"); got != "order-123" {
			t.Errorf("Idempotency-Key = %q, want order-123", got)
		}
		w.Write([]byte(`{"id":"em_1"}`))
	})

	_, err := client.Emails.SendWithOptions(context.Background(),
		&SendEmailRequest{From: "a@b.com", To: []string{"c@d.com"}, Subject: "s", Text: "t"},
		&RequestOptions{IdempotencyKey: "order-123"})
	if err != nil {
		t.Fatalf("SendWithOptions: %v", err)
	}
}

func TestEmailsListPagination(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/emails" {
			t.Errorf("path = %s, want /emails", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("limit") != "20" || q.Get("after") != "em_9" {
			t.Errorf("query = %v", q)
		}
		w.Write([]byte(`{"object":"list","has_more":true,"data":[{"object":"email","id":"em_10","from":"a@b.com","to":["c@d.com"],"subject":"s","last_event":"delivered","created_at":"2026-07-06T00:00:00Z"}]}`))
	})

	list, err := client.Emails.List(&ListParams{Limit: 20, After: "em_9"})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if !list.HasMore || len(list.Data) != 1 || list.Data[0].Id != "em_10" {
		t.Errorf("unexpected list: %+v", list)
	}
}

func TestBatchSend(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/emails/batch" {
			t.Errorf("%s %s, want POST /emails/batch", r.Method, r.URL.Path)
		}
		w.Write([]byte(`{"data":[{"id":"em_1"},{"id":"em_2"}]}`))
	})

	res, err := client.Batch.Send([]*SendEmailRequest{
		{From: "a@b.com", To: []string{"x@y.com"}, Subject: "1", Text: "one"},
		{From: "a@b.com", To: []string{"z@y.com"}, Subject: "2", Text: "two"},
	})
	if err != nil {
		t.Fatalf("Batch.Send: %v", err)
	}
	if len(res.Data) != 2 || res.Data[1].Id != "em_2" {
		t.Errorf("unexpected batch response: %+v", res)
	}
}

func TestReceivingGetAttachmentRawBytes(t *testing.T) {
	raw := []byte{0x25, 0x50, 0x44, 0x46} // %PDF
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/emails/receiving/rcv_1/attachments/att_1" {
			t.Errorf("path = %s", r.URL.Path)
		}
		assertAuth(t, r)
		w.Header().Set("Content-Type", "application/pdf")
		w.Write(raw)
	})

	got, err := client.Emails.Receiving.GetAttachment("rcv_1", "att_1")
	if err != nil {
		t.Fatalf("GetAttachment: %v", err)
	}
	if string(got) != string(raw) {
		t.Errorf("bytes = %v, want %v", got, raw)
	}
}
