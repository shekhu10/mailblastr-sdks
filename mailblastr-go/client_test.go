package mailblastr

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

const testAPIKey = "mb_test_key"

// newTestClient spins up an httptest server and points a client at it.
func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c := NewClient(testAPIKey)
	c.BaseURL = srv.URL
	return c
}

// decodeBody unmarshals the request body into a map for assertions.
func decodeBody(t *testing.T, r *http.Request) map[string]any {
	t.Helper()
	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		t.Fatalf("decode request body: %v", err)
	}
	return body
}

func assertAuth(t *testing.T, r *http.Request) {
	t.Helper()
	if got := r.Header.Get("Authorization"); got != "Bearer "+testAPIKey {
		t.Errorf("Authorization = %q, want %q", got, "Bearer "+testAPIKey)
	}
}

func TestErrorParsing(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(`{"statusCode":422,"name":"validation_error","message":"domain is required"}`))
	})

	_, err := client.Segments.Create(&CreateSegmentRequest{Name: "VIP"})
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	var apiErr *MailblastrError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *MailblastrError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != 422 {
		t.Errorf("StatusCode = %d, want 422", apiErr.StatusCode)
	}
	if apiErr.Name != "validation_error" {
		t.Errorf("Name = %q, want validation_error", apiErr.Name)
	}
	if apiErr.Message != "domain is required" {
		t.Errorf("Message = %q, want %q", apiErr.Message, "domain is required")
	}
}

func TestErrorParsingNonJSONBody(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("Bad Gateway"))
	})

	_, err := client.Emails.Get("em_1")
	var apiErr *MailblastrError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *MailblastrError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != 502 {
		t.Errorf("StatusCode = %d, want 502", apiErr.StatusCode)
	}
	if apiErr.Name != "application_error" {
		t.Errorf("Name = %q, want application_error", apiErr.Name)
	}
}

func TestRetryOn429(t *testing.T) {
	var calls atomic.Int32
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			// Retry-After: 0 makes the retry immediate (keeps the test fast)
			// and exercises the Retry-After delta-seconds path.
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"statusCode":429,"name":"rate_limited","message":"slow down"}`))
			return
		}
		w.Write([]byte(`{"object":"email","id":"em_1"}`))
	})

	email, err := client.Emails.Get("em_1")
	if err != nil {
		t.Fatalf("Get after retry: %v", err)
	}
	if got := calls.Load(); got != 2 {
		t.Errorf("server calls = %d, want 2 (one 429 then a successful retry)", got)
	}
	if email == nil || email.Id != "em_1" {
		t.Errorf("email = %+v, want id em_1 from the retried response", email)
	}
}

func TestNoRetryOn500(t *testing.T) {
	var calls atomic.Int32
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"statusCode":500,"name":"server_error","message":"boom"}`))
	})

	if _, err := client.Emails.Get("em_1"); err == nil {
		t.Fatal("expected an error on 500")
	}
	if got := calls.Load(); got != 1 {
		t.Errorf("server calls = %d, want 1 (500 is not retryable)", got)
	}
}

func TestPathEscaping(t *testing.T) {
	var gotPath string
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath()
		w.Write([]byte(`{"object":"email","id":"x"}`))
	})

	if _, err := client.Emails.Get("../api-keys"); err != nil {
		t.Fatalf("Get: %v", err)
	}
	if gotPath != "/emails/..%2Fapi-keys" {
		t.Errorf("path = %q, want escaped id (no traversal)", gotPath)
	}
}
