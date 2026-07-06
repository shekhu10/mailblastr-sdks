package com.mailblastr.requests;

import com.mailblastr.json.JsonPayload;

import java.util.Collections;
import java.util.LinkedHashMap;
import java.util.Map;

/** Body of {@code PATCH /topics/:id}. */
public final class UpdateTopicRequest implements JsonPayload {
    private final Map<String, Object> body;

    private UpdateTopicRequest(Map<String, Object> body) { this.body = body; }

    public static Builder builder() { return new Builder(); }

    @Override
    public Map<String, Object> toMap() { return Collections.unmodifiableMap(body); }

    public static final class Builder {
        private final Map<String, Object> m = new LinkedHashMap<>();

        public Builder name(String name) { m.put("name", name); return this; }
        public Builder description(String description) { m.put("description", description); return this; }
        /** {@code "public"} or {@code "private"}. */
        public Builder visibility(String visibility) { m.put("visibility", visibility); return this; }

        public UpdateTopicRequest build() { return new UpdateTopicRequest(new LinkedHashMap<>(m)); }
    }
}
