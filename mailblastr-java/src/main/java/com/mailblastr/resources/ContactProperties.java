package com.mailblastr.resources;

import com.mailblastr.ListParams;
import com.mailblastr.MailblastrResponse;
import com.mailblastr.http.ApiClient;
import com.mailblastr.requests.CreateContactPropertyRequest;

import java.util.Collections;

/** Custom contact properties (merge tags) — {@code mailblastr.contactProperties()}. */
public final class ContactProperties extends Resource {
    public ContactProperties(ApiClient api) { super(api); }

    /** Returns the slim ack {@code { object: 'contact_property', id }}. {@code POST /contact-properties} */
    public MailblastrResponse create(CreateContactPropertyRequest request) {
        return api.request("POST", "/contact-properties", request);
    }

    /** {@code GET /contact-properties/:id} */
    public MailblastrResponse get(String id) {
        return api.request("GET", "/contact-properties/" + enc(id));
    }

    /**
     * List the registry. With no pagination params ALL properties are returned
     * and {@code has_more:false}. {@code GET /contact-properties}
     */
    public MailblastrResponse list() { return list(null); }

    public MailblastrResponse list(ListParams params) {
        return api.request("GET", "/contact-properties" + paginate(params));
    }

    /**
     * Update the fallback value (the only mutable field — key/type are
     * immutable; pass {@code null} to clear). {@code PATCH /contact-properties/:id}
     */
    public MailblastrResponse update(String id, Object fallbackValue) {
        return api.request("PATCH", "/contact-properties/" + enc(id),
                Collections.singletonMap("fallback_value", fallbackValue));
    }

    /** {@code DELETE /contact-properties/:id} */
    public MailblastrResponse remove(String id) {
        return api.request("DELETE", "/contact-properties/" + enc(id));
    }
}
