package com.mailblastr.requests;

import com.mailblastr.json.JsonPayload;

import java.util.LinkedHashMap;
import java.util.Map;

/** A name/value tag attached to a sent email. */
public final class Tag implements JsonPayload {
    private final String name;
    private final String value;

    private Tag(String name, String value) {
        this.name = name;
        this.value = value;
    }

    public static Tag of(String name, String value) { return new Tag(name, value); }

    public String getTagName() { return name; }
    public String getValue() { return value; }

    @Override
    public Map<String, Object> toMap() {
        Map<String, Object> m = new LinkedHashMap<>();
        m.put("name", name);
        m.put("value", value);
        return m;
    }
}
