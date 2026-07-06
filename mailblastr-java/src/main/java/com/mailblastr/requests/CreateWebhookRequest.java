package com.mailblastr.requests;

import com.mailblastr.json.JsonPayload;

import java.util.Arrays;
import java.util.Collections;
import java.util.LinkedHashMap;
import java.util.Map;

/**
 * Body of {@code POST /webhooks}. The plaintext {@code signing_secret} is
 * returned ONCE, only on create (and on rotate).
 */
public final class CreateWebhookRequest implements JsonPayload {
    private final Map<String, Object> body;

    private CreateWebhookRequest(Map<String, Object> body) { this.body = body; }

    public static Builder builder() { return new Builder(); }

    @Override
    public Map<String, Object> toMap() { return Collections.unmodifiableMap(body); }

    public static final class Builder {
        private final Map<String, Object> m = new LinkedHashMap<>();

        public Builder endpoint(String endpoint) { m.put("endpoint", endpoint); return this; }
        /** Event names to deliver, e.g. {@code email.delivered}, {@code email.bounced}. */
        public Builder events(String... events) { m.put("events", Arrays.asList(events)); return this; }
        /** Optional caller-supplied signing secret. When omitted, MailBlastr generates one. */
        public Builder secret(String secret) { m.put("secret", secret); return this; }

        public CreateWebhookRequest build() { return new CreateWebhookRequest(new LinkedHashMap<>(m)); }
    }
}
