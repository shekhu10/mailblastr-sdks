package com.mailblastr.resources;

import com.mailblastr.ListParams;
import com.mailblastr.MailblastrResponse;
import com.mailblastr.http.ApiClient;
import com.mailblastr.requests.ForwardEmailRequest;
import com.mailblastr.requests.ReplyEmailRequest;
import com.mailblastr.requests.SendEmailRequest;

import java.util.Collections;
import java.util.List;
import java.util.Map;

/** Sent email — {@code mailblastr.emails()} — plus the inbound sub-resource {@code emails().receiving()}. */
public final class Emails extends Resource {
    private final Receiving receiving;

    public Emails(ApiClient api) {
        super(api);
        this.receiving = new Receiving(api);
    }

    /** Inbound (received) email sub-resource. */
    public Receiving receiving() { return receiving; }

    /** Send a single email. {@code POST /emails} */
    public MailblastrResponse send(SendEmailRequest request) {
        return api.request("POST", "/emails", request);
    }

    /** Send a single email with an Idempotency-Key (24h retry window). */
    public MailblastrResponse send(SendEmailRequest request, String idempotencyKey) {
        return api.request("POST", "/emails", request, idempotencyKey);
    }

    /** Send up to 100 emails in one request. {@code POST /emails/batch} (alias of {@code batch().send}). */
    public MailblastrResponse batch(List<SendEmailRequest> requests) {
        return api.request("POST", "/emails/batch", requests);
    }

    public MailblastrResponse batch(List<SendEmailRequest> requests, String idempotencyKey) {
        return api.request("POST", "/emails/batch", requests, idempotencyKey);
    }

    /** List sent emails (trimmed items — no status/html/text/events). {@code GET /emails} */
    public MailblastrResponse list() { return list(null); }

    public MailblastrResponse list(ListParams params) {
        return api.request("GET", "/emails" + paginate(params));
    }

    /** Retrieve a sent email and its events. {@code GET /emails/:id} */
    public MailblastrResponse get(String id) {
        return api.request("GET", "/emails/" + enc(id));
    }

    /** List a sent email's attachments. {@code GET /emails/:id/attachments} */
    public MailblastrResponse listAttachments(String id) {
        return api.request("GET", "/emails/" + enc(id) + "/attachments");
    }

    /** Retrieve one attachment's metadata. {@code GET /emails/:id/attachments/:attachmentId} */
    public MailblastrResponse getAttachment(String id, String attachmentId) {
        return api.request("GET", "/emails/" + enc(id) + "/attachments/" + enc(attachmentId));
    }

    /** Reschedule a scheduled email. {@code PATCH /emails/:id} */
    public MailblastrResponse update(String id, String scheduledAt) {
        return api.request("PATCH", "/emails/" + enc(id),
                Collections.singletonMap("scheduled_at", scheduledAt));
    }

    /** Cancel a scheduled email. {@code POST /emails/:id/cancel} */
    public MailblastrResponse cancel(String id) {
        return api.request("POST", "/emails/" + enc(id) + "/cancel");
    }

    /** Inbound (received) email — {@code mailblastr.emails().receiving()}. */
    public static final class Receiving extends Resource {
        Receiving(ApiClient api) { super(api); }

        /** List received emails. {@code GET /emails/receiving} */
        public MailblastrResponse list() { return list(null); }

        public MailblastrResponse list(ListParams params) {
            return api.request("GET", "/emails/receiving" + paginate(params));
        }

        /** Retrieve a received email. {@code GET /emails/receiving/:id} */
        public MailblastrResponse get(String id) {
            return api.request("GET", "/emails/receiving/" + enc(id));
        }

        /** List a received email's attachments. {@code GET /emails/receiving/:id/attachments} */
        public MailblastrResponse listAttachments(String id) {
            return api.request("GET", "/emails/receiving/" + enc(id) + "/attachments");
        }

        /**
         * Download one attachment as raw bytes (the route streams the binary
         * file, not JSON). {@code GET /emails/receiving/:id/attachments/:attachmentId}
         */
        public byte[] getAttachment(String id, String attachmentId) {
            return api.requestRaw("GET", "/emails/receiving/" + enc(id) + "/attachments/" + enc(attachmentId));
        }

        /** Download the original RFC822/MIME message as raw bytes. {@code GET /emails/receiving/:id/raw} */
        public byte[] getRaw(String id) {
            return api.requestRaw("GET", "/emails/receiving/" + enc(id) + "/raw");
        }

        /** Forward a received email. {@code POST /emails/receiving/:id/forward} */
        public MailblastrResponse forward(String id, ForwardEmailRequest request) {
            return api.request("POST", "/emails/receiving/" + enc(id) + "/forward", request);
        }

        /** Reply to a received email's sender, threaded. {@code POST /emails/receiving/:id/reply} */
        public MailblastrResponse reply(String id, ReplyEmailRequest request) {
            return api.request("POST", "/emails/receiving/" + enc(id) + "/reply", request);
        }

        /** Delete a received email. {@code DELETE /emails/receiving/:id} */
        public MailblastrResponse remove(String id) {
            return api.request("DELETE", "/emails/receiving/" + enc(id));
        }
    }
}
