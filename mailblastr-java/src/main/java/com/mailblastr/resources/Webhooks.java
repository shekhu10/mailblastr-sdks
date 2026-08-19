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
     * A {@code whsec_}-prefixed secret is base64-decoded exactly as leniently
     * as the signer decodes it — unpadded and URL-safe ({@code -}/{@code _})
     * spellings all derive the same key; other secrets are used as raw UTF-8
     * key bytes.
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

    /** Base64 alphabet, plus the URL-safe spellings of the last two characters. */
    private static final String B64_ALPHABET =
            "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/-_";

    /**
     * Decode base64 the way the signer does, not the way Java does.
     *
     * <p>The backend derives its key with Node's {@code Buffer.from(suffix,
     * 'base64')} (lib/crypto.ts {@code secretToKey}), which is LENIENT: it
     * ignores characters outside the alphabet, accepts the URL-safe {@code -}
     * and {@code _} spellings of {@code +} and {@code /}, needs no {@code =}
     * padding, and drops a lone 1-character remainder (which carries no whole
     * byte). {@code Base64.getDecoder()} tolerates only the missing padding: it
     * throws on {@code -}/{@code _}, on any other out-of-alphabet character,
     * and on a trailing 1-character group — and {@link #secretToKey(String)}
     * then quietly fell back to the raw UTF-8 bytes of the whole
     * {@code whsec_…} string. So a caller-supplied secret (the {@code secret}
     * field on {@code POST /webhooks}, which is stored verbatim) written in the
     * URL-safe alphabet produced a DIFFERENT key here than at the signer, and
     * every authentic delivery came back {@code no_match} — silently dropping
     * all real webhook traffic. This mirrors the signer instead, as the Node,
     * Python and PHP SDKs already do.
     */
    private static byte[] b64Lenient(String text) {
        StringBuilder cleaned = new StringBuilder(text.length());
        for (int i = 0; i < text.length(); i++) {
            char c = text.charAt(i);
            if (B64_ALPHABET.indexOf(c) < 0) continue;
            cleaned.append(c == '-' ? '+' : c == '_' ? '/' : c);
        }
        // A single leftover character encodes no whole byte; Node discards it
        // rather than failing the decode. Java's MIME decoder would throw
        // ("Last unit does not have enough valid bits") and lose the key.
        if (cleaned.length() % 4 == 1) cleaned.setLength(cleaned.length() - 1);
        try {
            // The MIME decoder, unlike getDecoder(), tolerates absent padding.
            return Base64.getMimeDecoder().decode(cleaned.toString());
        } catch (IllegalArgumentException ignored) { // defensive — input is pre-cleaned
            return new byte[0];
        }
    }

    /**
     * Derive the HMAC key from a {@code whsec_}-prefixed secret (leniently
     * base64-decode the suffix, see {@link #b64Lenient(String)}); a secret
     * without the prefix — or one whose suffix decodes to no bytes at all — is
     * used as raw UTF-8 bytes.
     */
    private static byte[] secretToKey(String secret) {
        if (secret.startsWith("whsec_")) {
            byte[] decoded = b64Lenient(secret.substring("whsec_".length()));
            if (decoded.length > 0) return decoded;
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
