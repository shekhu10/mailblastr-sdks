package com.mailblastr.requests;

import com.mailblastr.json.JsonPayload;

import java.util.Arrays;
import java.util.Collections;
import java.util.LinkedHashMap;
import java.util.Map;

/** Body of {@code PATCH /webhooks/:id}. */
public final class UpdateWebhookRequest implements JsonPayload {
    private final Map<String, Object> body;

    private UpdateWebhookRequest(Map<String, Object> body) { this.body = body; }

    public static Builder builder() { return new Builder(); }

    @Override
    public Map<String, Object> toMap() { return Collections.unmodifiableMap(body); }

    public static final class Builder {
        private final Map<String, Object> m = new LinkedHashMap<>();

        public Builder endpoint(String endpoint) { m.put("endpoint", endpoint); return this; }
        public Builder events(String... events) { m.put("events", Arrays.asList(events)); return this; }
        /** {@code "enabled"} or {@code "disabled"}. */
        public Builder status(String status) { m.put("status", status); return this; }

        public UpdateWebhookRequest build() { return new UpdateWebhookRequest(new LinkedHashMap<>(m)); }
    }
}
