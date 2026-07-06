package com.mailblastr.requests;

import com.mailblastr.json.JsonPayload;

import java.util.LinkedHashMap;
import java.util.Map;

/** A typed edge between two automation step keys. */
public final class AutomationConnection implements JsonPayload {
    private final String from;
    private final String to;
    private final String type;

    private AutomationConnection(String from, String to, String type) {
        this.from = from;
        this.to = to;
        this.type = type;
    }

    public static AutomationConnection of(String from, String to) {
        return new AutomationConnection(from, to, null);
    }

    /** {@code type} e.g. {@code "next"}, {@code "yes"}, {@code "no"}. */
    public static AutomationConnection of(String from, String to, String type) {
        return new AutomationConnection(from, to, type);
    }

    @Override
    public Map<String, Object> toMap() {
        Map<String, Object> m = new LinkedHashMap<>();
        m.put("from", from);
        m.put("to", to);
        if (type != null) m.put("type", type);
        return m;
    }
}
