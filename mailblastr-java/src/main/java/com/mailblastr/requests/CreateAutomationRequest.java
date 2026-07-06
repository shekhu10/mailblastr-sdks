package com.mailblastr.requests;

import com.mailblastr.json.JsonPayload;

import java.util.ArrayList;
import java.util.Collections;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

/**
 * Body of {@code POST /automations} (DOMAIN-FIRST):
 * {@link Builder#domain(String)} is REQUIRED — the sending domain this
 * automation belongs to. Only {@code events().send(...)} calls carrying the
 * same {@code domain} trigger it.
 */
public final class CreateAutomationRequest implements JsonPayload {
    private final Map<String, Object> body;

    private CreateAutomationRequest(Map<String, Object> body) { this.body = body; }

    public static Builder builder() { return new Builder(); }

    @Override
    public Map<String, Object> toMap() { return Collections.unmodifiableMap(body); }

    public static final class Builder {
        private final Map<String, Object> m = new LinkedHashMap<>();

        public Builder name(String name) { m.put("name", name); return this; }
        /** REQUIRED. The sending domain this automation belongs to. */
        public Builder domain(String domain) { m.put("domain", domain); return this; }
        /**
         * The event that starts a run: {@code contact.created}, an engagement
         * event ({@code email.opened}, {@code email.clicked}, {@code email.replied},
         * {@code email.bounced}, {@code email.delivered}), or any custom event
         * name sent via {@code events().send(...)}. Usually supplied as a
         * steps[0] trigger step instead.
         */
        public Builder trigger(String trigger) { m.put("trigger", trigger); return this; }
        /** Initial status: {@code "enabled"} or {@code "disabled"} (default disabled). */
        public Builder status(String status) { m.put("status", status); return this; }

        /** Append an inline step; give steps a {@code key} to reference them from connections. */
        @SuppressWarnings("unchecked")
        public Builder step(AutomationStep step) {
            ((List<Object>) m.computeIfAbsent("steps", k -> new ArrayList<>())).add(step);
            return this;
        }

        public Builder steps(List<AutomationStep> steps) {
            m.put("steps", new ArrayList<>(steps));
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

        public CreateAutomationRequest build() { return new CreateAutomationRequest(new LinkedHashMap<>(m)); }
    }
}
