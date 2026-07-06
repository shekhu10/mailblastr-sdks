package com.mailblastr.resources;

import com.mailblastr.ListParams;
import com.mailblastr.MailblastrResponse;
import com.mailblastr.http.ApiClient;
import com.mailblastr.requests.CreateTemplateRequest;
import com.mailblastr.requests.DuplicateTemplateRequest;
import com.mailblastr.requests.UpdateTemplateRequest;

import java.util.Collections;

/** Templates — {@code mailblastr.templates()}. */
public final class Templates extends Resource {
    public Templates(ApiClient api) { super(api); }

    /** Returns the slim ack {@code { object: 'template', id }}. {@code POST /templates} */
    public MailblastrResponse create(CreateTemplateRequest request) {
        return api.request("POST", "/templates", request);
    }

    /** {@code GET /templates/:id} */
    public MailblastrResponse get(String id) {
        return api.request("GET", "/templates/" + enc(id));
    }

    /** {@code GET /templates} */
    public MailblastrResponse list() { return list(null); }

    public MailblastrResponse list(ListParams params) {
        return api.request("GET", "/templates" + paginate(params));
    }

    /** Returns the slim ack. {@code PATCH /templates/:id} */
    public MailblastrResponse update(String id, UpdateTemplateRequest request) {
        return api.request("PATCH", "/templates/" + enc(id), request);
    }

    /** Duplicate a template. {@code POST /templates/:id/duplicate} */
    public MailblastrResponse duplicate(String id) {
        return api.request("POST", "/templates/" + enc(id) + "/duplicate", Collections.emptyMap());
    }

    public MailblastrResponse duplicate(String id, DuplicateTemplateRequest request) {
        return api.request("POST", "/templates/" + enc(id) + "/duplicate", request);
    }

    /** Publish a template (make its latest draft live). {@code POST /templates/:id/publish} */
    public MailblastrResponse publish(String id) {
        return api.request("POST", "/templates/" + enc(id) + "/publish");
    }

    /** {@code DELETE /templates/:id} */
    public MailblastrResponse remove(String id) {
        return api.request("DELETE", "/templates/" + enc(id));
    }
}
