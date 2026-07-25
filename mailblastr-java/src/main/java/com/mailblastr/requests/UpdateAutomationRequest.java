package com.mailblastr.requests;

import com.mailblastr.json.JsonPayload;

import java.util.ArrayList;
import java.util.Collections;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

/** Body of {@code PATCH /automations/:id}. */
public final class UpdateAutomationRequest implements JsonPayload {
    private final Map<String, Object> body;

    private UpdateAutomationRequest(Map<String, Object> body) { this.body = body; }

    public static Builder builder() { return new Builder(); }

    @Override
    public Map<String, Object> toMap() { return Collections.unmodifiableMap(body); }

    public static final class Builder {
        private final Map<String, Object> m = new LinkedHashMap<>();

        public Builder name(String name) { m.put("name", name); return this; }
        /** {@code "enabled"} or {@code "disabled"}. */
        public Builder status(String status) { m.put("status", status); return this; }
        /** Re-point at another of your domains (disabled automations only). */
        public Builder domain(String domain) { m.put("domain", domain); return this; }
        /**
         * Update the {@code mailblastr:schedule} trigger's schedule. Only
         * valid on automations with that trigger.
         *
         * @param at ISO 8601 instant the automation fires (future, at most 366 days ahead)
         * @param timezone IANA timezone the schedule was picked in, e.g. {@code "America/New_York"}
         */
        public Builder triggerConfig(String at, String timezone) {
            Map<String, Object> config = new LinkedHashMap<>();
            config.put("at", at);
            config.put("timezone", timezone);
            m.put("trigger_config", config);
            return this;
        }

        @SuppressWarnings("unchecked")
        public Builder connection(AutomationConnection connection) {
            ((List<Object>) m.computeIfAbsent("connections", k -> new ArrayList<>())).add(connection);
            return this;
        }

        public Builder connections(List<AutomationConnection> connections) {
            m.put("connections", new ArrayList<>(connections));
            return this;
        }

        public UpdateAutomationRequest build() { return new UpdateAutomationRequest(new LinkedHashMap<>(m)); }
    }
}
