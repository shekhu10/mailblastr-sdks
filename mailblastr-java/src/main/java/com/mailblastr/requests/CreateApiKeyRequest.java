package com.mailblastr.requests;

import com.mailblastr.json.JsonPayload;

import java.util.Collections;
import java.util.LinkedHashMap;
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
        /** Scope a {@code sending_access} key to one domain (ignored otherwise). */
        public Builder domainId(String domainId) { m.put("domain_id", domainId); return this; }

        public CreateApiKeyRequest build() { return new CreateApiKeyRequest(new LinkedHashMap<>(m)); }
    }
}
