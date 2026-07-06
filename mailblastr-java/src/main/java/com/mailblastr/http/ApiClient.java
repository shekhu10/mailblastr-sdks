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
 * {@code { statusCode, name, message }} error body.
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

    private MailblastrException errorFrom(HttpResult res) {
        int statusCode = res.statusCode();
        String name = "application_error";
        String message = "Request failed with status " + res.statusCode();
        try {
            Object parsed = Json.parse(res.bodyText());
            if (parsed instanceof Map) {
                Map<?, ?> m = (Map<?, ?>) parsed;
                Object sc = m.get("statusCode");
                if (sc instanceof Number) statusCode = ((Number) sc).intValue();
                Object n = m.get("name");
                if (n instanceof String) name = (String) n;
                Object msg = m.get("message");
                if (msg instanceof String) message = (String) msg;
            }
        } catch (RuntimeException ignored) {
            // Non-JSON error body — keep the fallbacks.
        }
        return new MailblastrException(statusCode, name, message);
    }
}
