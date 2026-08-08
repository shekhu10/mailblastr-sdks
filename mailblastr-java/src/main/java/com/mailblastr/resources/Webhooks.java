package com.mailblastr.resources;

import com.mailblastr.ListParams;
import com.mailblastr.MailblastrResponse;
import com.mailblastr.http.ApiClient;
import com.mailblastr.requests.CreateWebhookRequest;
import com.mailblastr.requests.UpdateWebhookRequest;

import javax.crypto.Mac;
import javax.crypto.spec.SecretKeySpec;
import java.nio.charset.StandardCharsets;
import java.security.GeneralSecurityException;
import java.security.MessageDigest;
import java.util.Base64;
import java.util.Map;

/** Webhooks — {@code mailblastr.webhooks()} — including local Svix-style signature verification. */
public final class Webhooks extends Resource {
    public Webhooks(ApiClient api) { super(api); }

    /** Create a webhook. The signing secret is shown ONCE, only here. {@code POST /webhooks} */
    public MailblastrResponse create(CreateWebhookRequest request) {
        return api.request("POST", "/webhooks", request);
    }

    /** {@code GET /webhooks/:id} */
    public MailblastrResponse get(String id) {
        return api.request("GET", "/webhooks/" + enc(id));
    }

    /**
     * List webhooks. Unlike domains/api-keys/topics this route always applies a
     * limit — with no pagination params you get the first 20.
     * {@code GET /webhooks}
     */
    public MailblastrResponse list() { return list(null); }

    public MailblastrResponse list(ListParams params) {
        return api.request("GET", "/webhooks" + paginate(params));
    }

    /** Returns the slim ack {@code { object: 'webhook', id }}. {@code PATCH /webhooks/:id} */
    public MailblastrResponse update(String id, UpdateWebhookRequest request) {
        return api.request("PATCH", "/webhooks/" + enc(id), request);
    }

    /**
     * Rotate the signing secret — the new plaintext {@code signing_secret} is
     * returned ONCE; the old secret stops verifying immediately.
     * {@code POST /webhooks/:id/rotate}
     */
    public MailblastrResponse rotate(String id) {
        return api.request("POST", "/webhooks/" + enc(id) + "/rotate");
    }

    /**
     * Send a synchronous, single-attempt test delivery and return the
     * endpoint's live result.
     *
     * <p><strong>A failed delivery is still HTTP 200</strong> — this method
     * does not throw for one. Inspect {@code ok} on the response
     * ({@code res.getBoolean("ok")}) and read {@code error} when it is false.
     * {@code POST /webhooks/:id/test}
     */
    public MailblastrResponse test(String id) {
        return api.request("POST", "/webhooks/" + enc(id) + "/test");
    }

    /** {@code DELETE /webhooks/:id} */
    public MailblastrResponse remove(String id) {
        return api.request("DELETE", "/webhooks/" + enc(id));
    }

    /** Instance alias of {@link #verifyWebhookSignature(String, Map, String)}. */
    public VerifyWebhookResult verify(String payload, Map<String, String> headers, String secret) {
        return verifyWebhookSignature(payload, headers, secret);
    }

    /** Instance alias of {@link #verifyWebhookSignature(String, Map, String, long)}. */
    public VerifyWebhookResult verify(String payload, Map<String, String> headers, String secret, long toleranceSec) {
        return verifyWebhookSignature(payload, headers, secret, toleranceSec);
    }

    /** Verify with the default 5-minute timestamp tolerance. */
    public static VerifyWebhookResult verifyWebhookSignature(String payload, Map<String, String> headers, String secret) {
        return verifyWebhookSignature(payload, headers, secret, 300);
    }

    /**
     * Verify a webhook delivery's Svix-style signature against your endpoint's
     * signing secret. Mirrors the backend signing scheme:
     * {@code "<svix-id>.<svix-timestamp>.<body>"} → base64 HMAC-SHA256, tagged
     * {@code v1,}.
     *
     * <p>{@code payload} MUST be the exact raw request body string the server
     * sent — do not re-serialize parsed JSON (whitespace differences break the
     * signature). {@code headers} must contain the {@code svix-id},
     * {@code svix-timestamp} and {@code svix-signature} headers (read
     * case-insensitively). A signature header may carry multiple
     * space-separated signatures; any one matching makes the delivery valid.
     * A {@code whsec_}-prefixed secret is base64-decoded; other secrets are
     * used as raw UTF-8 key bytes.
     *
     * <p>This is a pure local computation — it makes no HTTP request.
     *
     * @param toleranceSec max allowed clock skew in seconds (0 disables the check)
     */
    public static VerifyWebhookResult verifyWebhookSignature(
            String payload, Map<String, String> headers, String secret, long toleranceSec) {
        String id = readHeader(headers, "svix-id");
        String timestamp = readHeader(headers, "svix-timestamp");
        String sigHeader = readHeader(headers, "svix-signature");
        if (isBlank(id) || isBlank(timestamp) || isBlank(sigHeader)) {
            return new VerifyWebhookResult(false, "missing_headers");
        }
        if (isBlank(secret)) {
            return new VerifyWebhookResult(false, "missing_secret");
        }

        // Optional timestamp freshness check (default 5-minute tolerance; 0 disables).
        if (toleranceSec > 0) {
            long ts;
            try {
                ts = Long.parseLong(timestamp.trim());
            } catch (NumberFormatException e) {
                return new VerifyWebhookResult(false, "invalid_timestamp");
            }
            long skew = Math.abs(System.currentTimeMillis() / 1000L - ts);
            if (skew > toleranceSec) {
                return new VerifyWebhookResult(false, "timestamp_out_of_tolerance");
            }
        }

        String signed = id + "." + timestamp + "." + (payload == null ? "" : payload);
        String expected;
        try {
            Mac mac = Mac.getInstance("HmacSHA256");
            mac.init(new SecretKeySpec(secretToKey(secret), "HmacSHA256"));
            expected = Base64.getEncoder().encodeToString(mac.doFinal(signed.getBytes(StandardCharsets.UTF_8)));
        } catch (GeneralSecurityException e) {
            throw new IllegalStateException("HmacSHA256 unavailable", e);
        }

        // The header may contain multiple space-separated `v1,<sig>` entries; any match wins.
        for (String part : sigHeader.split(" ")) {
            String trimmed = part.trim();
            if (trimmed.isEmpty()) continue;
            String sig = trimmed.startsWith("v1,") ? trimmed.substring(3) : trimmed;
            if (MessageDigest.isEqual(
                    sig.getBytes(StandardCharsets.UTF_8),
                    expected.getBytes(StandardCharsets.UTF_8))) {
                return new VerifyWebhookResult(true, null);
            }
        }
        return new VerifyWebhookResult(false, "no_match");
    }

    /**
     * Derive the HMAC key from a {@code whsec_}-prefixed secret (base64-decode
     * the suffix); a secret without the prefix is used as raw UTF-8 bytes.
     */
    private static byte[] secretToKey(String secret) {
        if (secret.startsWith("whsec_")) {
            try {
                byte[] decoded = Base64.getDecoder().decode(secret.substring("whsec_".length()));
                if (decoded.length > 0) return decoded;
            } catch (IllegalArgumentException ignored) {
                // fall through to raw bytes
            }
        }
        return secret.getBytes(StandardCharsets.UTF_8);
    }

    /** Case-insensitively read a single header value. */
    private static String readHeader(Map<String, String> headers, String name) {
        if (headers == null) return null;
        for (Map.Entry<String, String> e : headers.entrySet()) {
            if (e.getKey() != null && e.getKey().equalsIgnoreCase(name)) return e.getValue();
        }
        return null;
    }

    private static boolean isBlank(String s) { return s == null || s.isEmpty(); }
}
