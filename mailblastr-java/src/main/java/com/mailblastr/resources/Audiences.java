package com.mailblastr.resources;

import com.mailblastr.ListParams;
import com.mailblastr.MailblastrResponse;
import com.mailblastr.http.ApiClient;

import java.util.LinkedHashMap;
import java.util.Map;

/** Audiences — {@code mailblastr.audiences()}. */
public final class Audiences extends Resource {
    public Audiences(ApiClient api) { super(api); }

    /** {@code POST /audiences} */
    public MailblastrResponse create(String name) {
        Map<String, Object> body = new LinkedHashMap<>();
        body.put("name", name);
        return api.request("POST", "/audiences", body);
    }

    /** {@code GET /audiences/:id} */
    public MailblastrResponse get(String id) {
        return api.request("GET", "/audiences/" + enc(id));
    }

    /** {@code GET /audiences} */
    public MailblastrResponse list() { return list(null); }

    public MailblastrResponse list(ListParams params) {
        return api.request("GET", "/audiences" + paginate(params));
    }

    /**
     * Import contacts from a link-shared Google Sheet — header columns become
     * contact properties; rows land in a fresh segment.
     * {@code POST /audiences/:id/contacts/import-sheet}
     */
    public MailblastrResponse importSheet(String audienceId, String url) {
        return importSheet(audienceId, url, null);
    }

    public MailblastrResponse importSheet(String audienceId, String url, String segmentName) {
        Map<String, Object> body = new LinkedHashMap<>();
        body.put("url", url);
        if (segmentName != null) body.put("segment_name", segmentName);
        return api.request("POST", "/audiences/" + enc(audienceId) + "/contacts/import-sheet", body);
    }

    /** Rename an audience. {@code PATCH /audiences/:id} */
    public MailblastrResponse update(String id, String name) {
        Map<String, Object> body = new LinkedHashMap<>();
        body.put("name", name);
        return api.request("PATCH", "/audiences/" + enc(id), body);
    }

    /** {@code DELETE /audiences/:id} */
    public MailblastrResponse remove(String id) {
        return api.request("DELETE", "/audiences/" + enc(id));
    }
}
