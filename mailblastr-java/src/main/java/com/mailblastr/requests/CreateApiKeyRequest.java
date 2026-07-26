package com.mailblastr.requests;

import com.mailblastr.json.JsonPayload;

import java.util.ArrayList;
import java.util.Collections;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

/** Body of {@code POST /api-keys}. The full token is returned only once, at creation. */
public final class CreateApiKeyRequest implements JsonPayload {
    private final Map<String, Object> body;

    private CreateApiKeyRequest(Map<String, Object> body) { this.body = body; }

    public static Builder builder() { return new Builder(); }

    @Override
    public Map<String, Object> toMap() { return Collections.unmodifiableMap(body); }

    public static final class Builder {
        private final Map<String, Object> m = new LinkedHashMap<>();

        public Builder name(String name) { m.put("name", name); return this; }
        /** {@code "full_access"} or {@code "sending_access"}. */
        public Builder permission(String permission) { m.put("permission", permission); return this; }
        /** Scope a {@code sending_access} key to one domain (legacy; prefer {@link #domainIds}). */
        public Builder domainId(String domainId) { m.put("domain_id", domainId); return this; }
        /**
         * Scope the key to one or more domains. Only valid with {@code sending_access} —
     * full-access keys always work across all your domains (the API rejects the
     * combination with a validation_error).
         * Mutually exclusive with {@link #domainId} — providing both is a 422.
         */
        public Builder domainIds(List<String> domainIds) { m.put("domain_ids", new ArrayList<>(domainIds)); return this; }

        public CreateApiKeyRequest build() { return new CreateApiKeyRequest(new LinkedHashMap<>(m)); }
    }
}
