package com.mailblastr.resources;

import com.mailblastr.MailblastrResponse;
import com.mailblastr.http.ApiClient;
import com.mailblastr.http.Query;
import com.mailblastr.requests.ListLogsParams;

/** API request logs — {@code mailblastr.logs()} (read-only). */
public final class Logs extends Resource {
    public Logs(ApiClient api) { super(api); }

    /** {@code GET /logs} */
    public MailblastrResponse list() { return list(null); }

    /**
     * List API request logs. Cursor-paginated with optional server-side
     * {@code method} / {@code status} filters.
     */
    public MailblastrResponse list(ListLogsParams params) {
        Query q = new Query();
        if (params != null) {
            q.add("limit", params.getLimit())
             .add("after", params.getAfter())
             .add("before", params.getBefore())
             .add("method", params.getMethod())
             .add("status", params.getStatus());
        }
        return api.request("GET", "/logs" + q);
    }

    /** Retrieve one log entry (includes request/response bodies). {@code GET /logs/:id} */
    public MailblastrResponse get(String id) {
        return api.request("GET", "/logs/" + enc(id));
    }
}
