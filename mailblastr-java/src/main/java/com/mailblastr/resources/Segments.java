package com.mailblastr.resources;

import com.mailblastr.ListParams;
import com.mailblastr.MailblastrResponse;
import com.mailblastr.http.ApiClient;
import com.mailblastr.http.Query;
import com.mailblastr.requests.CreateSegmentRequest;
import com.mailblastr.requests.UpdateSegmentRequest;

/**
 * Segments (DOMAIN-FIRST) — {@code domain} is REQUIRED on create and list.
 * Segment names are unique within a domain; every domain carries an
 * auto-created "General" segment.
 */
public final class Segments extends Resource {
    public Segments(ApiClient api) { super(api); }

    /** {@code POST /segments} — {@code domain} is required on the request. */
    public MailblastrResponse create(CreateSegmentRequest request) {
        return api.request("POST", "/segments", request);
    }

    /** {@code GET /segments/:id} */
    public MailblastrResponse get(String id) {
        return api.request("GET", "/segments/" + enc(id));
    }

    /** List a domain's segments. {@code GET /segments?domain=} */
    public MailblastrResponse list(String domain) { return list(domain, null); }

    public MailblastrResponse list(String domain, ListParams params) {
        Query q = new Query().add("domain", domain);
        if (params != null) params.applyTo(q);
        return api.request("GET", "/segments" + q);
    }

    /** Preview the contacts a segment currently resolves to. {@code GET /segments/:id/contacts} */
    public MailblastrResponse contacts(String id) {
        return api.request("GET", "/segments/" + enc(id) + "/contacts");
    }

    /** {@code PATCH /segments/:id} */
    public MailblastrResponse update(String id, UpdateSegmentRequest request) {
        return api.request("PATCH", "/segments/" + enc(id), request);
    }

    /** {@code DELETE /segments/:id} */
    public MailblastrResponse remove(String id) {
        return api.request("DELETE", "/segments/" + enc(id));
    }
}
