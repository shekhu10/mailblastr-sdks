package com.mailblastr.resources;

import com.mailblastr.ListParams;
import com.mailblastr.MailblastrResponse;
import com.mailblastr.http.ApiClient;
import com.mailblastr.http.Query;
import com.mailblastr.requests.CreateTopicRequest;
import com.mailblastr.requests.UpdateTopicRequest;

/** Topics (DOMAIN-FIRST) — {@code domain} is REQUIRED on create and list. */
public final class Topics extends Resource {
    public Topics(ApiClient api) { super(api); }

    /** {@code POST /topics} — {@code domain} is required on the request. */
    public MailblastrResponse create(CreateTopicRequest request) {
        return api.request("POST", "/topics", request);
    }

    /** {@code GET /topics/:id} */
    public MailblastrResponse get(String id) {
        return api.request("GET", "/topics/" + enc(id));
    }

    /**
     * List a domain's topics. With no pagination params the route skips paging
     * and answers with the whole pool — capped at <strong>1,000</strong> rows,
     * with {@code has_more} true if that ceiling bites.
     * {@code GET /topics?domain=}
     */
    public MailblastrResponse list(String domain) { return list(domain, null); }

    public MailblastrResponse list(String domain, ListParams params) {
        Query q = new Query().add("domain", domain);
        if (params != null) params.applyTo(q);
        return api.request("GET", "/topics" + q);
    }

    /** {@code PATCH /topics/:id} */
    public MailblastrResponse update(String id, UpdateTopicRequest request) {
        return api.request("PATCH", "/topics/" + enc(id), request);
    }

    /** {@code DELETE /topics/:id} */
    public MailblastrResponse remove(String id) {
        return api.request("DELETE", "/topics/" + enc(id));
    }
}
