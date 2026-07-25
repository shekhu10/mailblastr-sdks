package com.mailblastr.resources;

import com.mailblastr.MailblastrResponse;
import com.mailblastr.http.ApiClient;
import com.mailblastr.requests.BatchEmailRequest;
import com.mailblastr.requests.SendEmailRequest;

import java.util.List;

/** Batch send — {@code mailblastr.batch().send(...)}. */
public final class Batch extends Resource {
    public Batch(ApiClient api) { super(api); }

    /**
     * Send up to 100 emails in one request. {@code POST /emails/batch}
     *
     * @deprecated use {@link #sendEmails(List)} — batch items reject
     *     {@code attachments} and {@code scheduled_at} (send those individually
     *     via {@code emails().send}), which {@link BatchEmailRequest} enforces
     *     at compile time.
     */
    @Deprecated
    public MailblastrResponse send(List<SendEmailRequest> requests) {
        return api.request("POST", "/emails/batch", requests);
    }

    /** @deprecated use {@link #sendEmails(List, String)} — see {@link #send(List)}. */
    @Deprecated
    public MailblastrResponse send(List<SendEmailRequest> requests, String idempotencyKey) {
        return api.request("POST", "/emails/batch", requests, idempotencyKey);
    }

    /**
     * Send up to 100 emails in one request. Batch items reject
     * {@code attachments} and {@code scheduled_at} — send those individually
     * via {@code emails().send}. {@code POST /emails/batch}
     */
    public MailblastrResponse sendEmails(List<BatchEmailRequest> requests) {
        return api.request("POST", "/emails/batch", requests);
    }

    /** Like {@link #sendEmails(List)}, with an Idempotency-Key (24h retry window). */
    public MailblastrResponse sendEmails(List<BatchEmailRequest> requests, String idempotencyKey) {
        return api.request("POST", "/emails/batch", requests, idempotencyKey);
    }
}
