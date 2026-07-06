package mailblastr

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"testing"
	"time"
)

// sign computes a Svix-style "v1,<base64 hmac>" signature the way the
// MailBlastr backend does, for building test deliveries.
func sign(t *testing.T, secret, id, timestamp string, payload []byte) string {
	t.Helper()
	mac := hmac.New(sha256.New, secretToKey(secret))
	fmt.Fprintf(mac, "%s.%s.", id, timestamp)
	mac.Write(payload)
	return "v1," + base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func deliveryHeaders(id, timestamp, signature string) http.Header {
	h := http.Header{}
	h.Set("Svix-Id", id)
	h.Set("Svix-Timestamp", timestamp)
	h.Set("Svix-Signature", signature)
	return h
}

func TestVerifyWebhookSignatureValid(t *testing.T) {
	secret := "whsec_" + base64.StdEncoding.EncodeToString([]byte("super-secret-key"))
	payload := []byte(`{"type":"email.delivered","data":{"id":"em_1"}}`)
	id := "msg_1"
	ts := strconv.FormatInt(time.Now().Unix(), 10)

	headers := deliveryHeaders(id, ts, sign(t, secret, id, ts, payload))
	res := VerifyWebhookSignature(payload, headers, secret, nil)
	if !res.Valid {
		t.Fatalf("expected valid signature, got reason %q", res.Reason)
	}
}

func TestVerifyWebhookSignatureRawSecret(t *testing.T) {
	// A secret without the whsec_ prefix is used as raw UTF-8 bytes.
	secret := "plain-secret"
	payload := []byte(`{"ok":true}`)
	id := "msg_2"
	ts := strconv.FormatInt(time.Now().Unix(), 10)

	headers := deliveryHeaders(id, ts, sign(t, secret, id, ts, payload))
	if res := VerifyWebhookSignature(payload, headers, secret, nil); !res.Valid {
		t.Fatalf("expected valid signature, got reason %q", res.Reason)
	}
}

func TestVerifyWebhookSignatureMultipleSignatures(t *testing.T) {
	secret := "plain-secret"
	payload := []byte(`{"ok":true}`)
	id := "msg_3"
	ts := strconv.FormatInt(time.Now().Unix(), 10)

	good := sign(t, secret, id, ts, payload)
	headers := deliveryHeaders(id, ts, "v1,bogus= "+good)
	if res := VerifyWebhookSignature(payload, headers, secret, nil); !res.Valid {
		t.Fatalf("any matching signature should win, got reason %q", res.Reason)
	}
}

func TestVerifyWebhookSignatureTampered(t *testing.T) {
	secret := "plain-secret"
	payload := []byte(`{"amount":100}`)
	id := "msg_4"
	ts := strconv.FormatInt(time.Now().Unix(), 10)

	headers := deliveryHeaders(id, ts, sign(t, secret, id, ts, payload))
	res := VerifyWebhookSignature([]byte(`{"amount":999}`), headers, secret, nil)
	if res.Valid {
		t.Fatal("tampered payload must not verify")
	}
	if res.Reason != "no_match" {
		t.Errorf("reason = %q, want no_match", res.Reason)
	}
}

func TestVerifyWebhookSignatureStaleTimestamp(t *testing.T) {
	secret := "plain-secret"
	payload := []byte(`{"ok":true}`)
	id := "msg_5"
	ts := strconv.FormatInt(time.Now().Add(-time.Hour).Unix(), 10)

	headers := deliveryHeaders(id, ts, sign(t, secret, id, ts, payload))
	res := VerifyWebhookSignature(payload, headers, secret, nil)
	if res.Valid || res.Reason != "timestamp_out_of_tolerance" {
		t.Fatalf("expected timestamp_out_of_tolerance, got valid=%v reason=%q", res.Valid, res.Reason)
	}

	// Disabling the check (negative tolerance) makes the same delivery valid.
	res = VerifyWebhookSignature(payload, headers, secret, &VerifyWebhookOptions{ToleranceSec: -1})
	if !res.Valid {
		t.Fatalf("expected valid with timestamp check disabled, got reason %q", res.Reason)
	}
}

func TestVerifyWebhookSignatureMissingHeaders(t *testing.T) {
	res := VerifyWebhookSignature([]byte(`{}`), http.Header{}, "secret", nil)
	if res.Valid || res.Reason != "missing_headers" {
		t.Fatalf("expected missing_headers, got valid=%v reason=%q", res.Valid, res.Reason)
	}
}

func TestVerifyWebhookSignatureMissingSecret(t *testing.T) {
	headers := deliveryHeaders("msg_6", strconv.FormatInt(time.Now().Unix(), 10), "v1,abc")
	res := VerifyWebhookSignature([]byte(`{}`), headers, "", nil)
	if res.Valid || res.Reason != "missing_secret" {
		t.Fatalf("expected missing_secret, got valid=%v reason=%q", res.Valid, res.Reason)
	}
}

func TestWebhooksCreateReturnsSecretOnce(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/webhooks" {
			t.Errorf("%s %s, want POST /webhooks", r.Method, r.URL.Path)
		}
		body := decodeBody(t, r)
		if body["endpoint"] != "https://yourapp.com/hooks" {
			t.Errorf("endpoint = %v", body["endpoint"])
		}
		events, ok := body["events"].([]any)
		if !ok || len(events) != 2 {
			t.Errorf("events = %v", body["events"])
		}
		w.Write([]byte(`{"object":"webhook","id":"wh_1","signing_secret":"whsec_abc123"}`))
	})

	res, err := client.Webhooks.Create(&CreateWebhookRequest{
		Endpoint: "https://yourapp.com/hooks",
		Events:   []string{"email.delivered", "email.bounced"},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if res.SigningSecret != "whsec_abc123" {
		t.Errorf("SigningSecret = %q", res.SigningSecret)
	}
}
