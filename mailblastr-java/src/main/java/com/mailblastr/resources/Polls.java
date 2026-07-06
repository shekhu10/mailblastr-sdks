package com.mailblastr.resources;

import com.mailblastr.ListParams;
import com.mailblastr.MailblastrResponse;
import com.mailblastr.http.ApiClient;

/** Read-only results of the in-email poll widget — {@code mailblastr.polls()}. */
public final class Polls extends Resource {
    public Polls(ApiClient api) { super(api); }

    /** One summary row per email that has poll responses. {@code GET /polls} */
    public MailblastrResponse list() { return list(null); }

    public MailblastrResponse list(ListParams params) {
        return api.request("GET", "/polls" + paginate(params));
    }

    /** The aggregated answer breakdown for one email. {@code GET /polls/:emailId} */
    public MailblastrResponse get(String emailId) {
        return api.request("GET", "/polls/" + enc(emailId));
    }
}
