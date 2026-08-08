package com.mailblastr.http;

import com.mailblastr.MailblastrException;
import com.mailblastr.MailblastrResponse;
import com.mailblastr.json.Json;

import java.util.LinkedHashMap;
import java.util.Map;

/**
 * Internal request executor shared by all resources. Adds the standard
 * headers (Bearer auth, Content-Type, User-Agent, optional Idempotency-Key),
 * serializes bodies to JSON, and converts every non-2xx response into a
 * {@link MailblastrException} built from the API's
 * {@code { statusCode, name, message }} error envelope (keeping the full body
 * for the additive {@code limit} / {@code reputation} / {@code sent} fields).
 *
 * <p>The {@code User-Agent} header is mandatory on every MailBlastr endpoint —
 * a missing or blank one is rejected with HTTP 403 {@code validation_error}
 * before authentication even runs — so it is validated at construction time.
 */
public final class ApiClient {
    private final String apiKey;
    private final String baseUrl;
    private final String userAgent;
    private final HttpTransport transport;

    public ApiClient(String apiKey, String baseUrl, String userAgent, HttpTransport transport) {
        if (apiKey == null || apiKey.isEmpty()) {
            throw new IllegalArgumentException(
                    "Mailblastr: an API key is required, e.g. new Mailblastr(\"mb_...\").");
        }
        if (userAgent == null || userAgent.trim().isEmpty()) {
            throw new IllegalArgumentException(
                    "Mailblastr: a non-empty User-Agent is required — the API rejects "
                            + "requests without one with 403 validation_error.");
        }
        this.apiKey = apiKey;
        this.baseUrl = baseUrl.endsWith("/") ? baseUrl.substring(0, baseUrl.length() - 1) : baseUrl;
        this.userAgent = userAgent;
        this.transport = transport;
    }

    public MailblastrResponse request(String method, String path) {
        return request(method, path, null, null);
    }

    public MailblastrResponse request(String method, String path, Object body) {
        return request(method, path, body, null);
    }

    public MailblastrResponse request(String method, String path, Object body, String idempotencyKey) {
        HttpResult res = exec(method, path, body, idempotencyKey, true);
        return new MailblastrResponse(res.statusCode(), res.bodyText());
    }

    /** For endpoints that stream raw bytes (attachment downloads, raw MIME). */
    public byte[] requestRaw(String method, String path) {
        return exec(method, path, null, null, false).body();
    }

    private HttpResult exec(String method, String path, Object body, String idempotencyKey, boolean json) {
        Map<String, String> headers = new LinkedHashMap<>();
        headers.put("Authorization", "Bearer " + apiKey);
        if (json) headers.put("Content-Type", "application/json");
        headers.put("User-Agent", userAgent);
        if (idempotencyKey != null) headers.put("Idempotency-Key", idempotencyKey);

        String serialized = body == null ? null : Json.write(body);
        HttpResult res;
        try {
            res = transport.execute(method, baseUrl + path, headers, serialized);
        } catch (Exception e) {
            throw new MailblastrException(0, "network_error",
                    e.getMessage() == null ? e.getClass().getSimpleName() : e.getMessage());
        }
        if (res.statusCode() < 200 || res.statusCode() >= 300) {
            throw errorFrom(res);
        }
        return res;
    }

    @SuppressWarnings("unchecked")
    private MailblastrException errorFrom(HttpResult res) {
        int statusCode = res.statusCode();
        String name = "application_error";
        String message = "Request failed with status " + res.statusCode();
        Map<String, Object> body = null;
        try {
            Object parsed = Json.parse(res.bodyText());
            if (parsed instanceof Map) {
                Map<String, Object> m = (Map<String, Object>) parsed;
                body = m;
                // The real HTTP status wins: the envelope's `statusCode` always
                // mirrors it, and a handler may override the status a given
                // error name usually maps to.
                Object n = m.get("name");
                if (n instanceof String) name = (String) n;
                Object msg = m.get("message");
                if (msg instanceof String) {
                    message = (String) msg;
                } else {
                    // CSRF rejections use a non-envelope body: {"error":"csrf_failed"}.
                    Object err = m.get("error");
                    if (err instanceof String) {
                        name = (String) err;
                        message = "Request rejected: " + err;
                    }
                }
            }
        } catch (RuntimeException ignored) {
            // Non-JSON error body — keep the fallbacks.
        }
        return new MailblastrException(statusCode, name, message, body);
    }
}
