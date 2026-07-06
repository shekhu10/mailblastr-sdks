package com.mailblastr.requests;

import com.mailblastr.json.JsonPayload;

import java.util.ArrayList;
import java.util.Collections;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

/**
 * A segment's filter: subscription {@code status} ({@code all} |
 * {@code subscribed} | {@code unsubscribed}), an {@code email_contains}
 * substring, and custom-property predicates ({@code eq} | {@code contains} |
 * {@code exists}).
 */
public final class SegmentFilter implements JsonPayload {
    private final Map<String, Object> body;

    private SegmentFilter(Map<String, Object> body) { this.body = body; }

    public static Builder builder() { return new Builder(); }

    @Override
    public Map<String, Object> toMap() { return Collections.unmodifiableMap(body); }

    public static final class Builder {
        private final Map<String, Object> m = new LinkedHashMap<>();

        /** {@code "all"}, {@code "subscribed"}, or {@code "unsubscribed"}. */
        public Builder status(String status) { m.put("status", status); return this; }
        public Builder emailContains(String emailContains) { m.put("email_contains", emailContains); return this; }

        /** Add a custom-property predicate. {@code value} is required for eq/contains. */
        @SuppressWarnings("unchecked")
        public Builder propertyFilter(String key, String operator, Object value) {
            Map<String, Object> f = new LinkedHashMap<>();
            f.put("key", key);
            f.put("operator", operator);
            if (value != null) f.put("value", value);
            ((List<Object>) m.computeIfAbsent("property_filters", k -> new ArrayList<>())).add(f);
            return this;
        }

        /** Add an {@code exists} predicate (no value). */
        public Builder propertyExists(String key) { return propertyFilter(key, "exists", null); }

        public SegmentFilter build() { return new SegmentFilter(new LinkedHashMap<>(m)); }
    }
}
