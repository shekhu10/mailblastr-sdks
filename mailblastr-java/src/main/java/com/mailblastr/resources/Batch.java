package com.mailblastr.resources;

import com.mailblastr.MailblastrResponse;
import com.mailblastr.http.ApiClient;
import com.mailblastr.requests.SendEmailRequest;

import java.util.List;

/** Batch send — {@code mailblastr.batch().send(...)}. */
public final class Batch extends Resource {
    public Batch(ApiClient api) { super(api); }

    /** Send up to 100 emails in one request. {@code POST /emails/batch} */
    public MailblastrResponse send(List<SendEmailRequest> requests) {
        return api.request("POST", "/emails/batch", requests);
    }

    public MailblastrResponse send(List<SendEmailRequest> requests, String idempotencyKey) {
        return api.request("POST", "/emails/batch", requests, idempotencyKey);
    }
}
