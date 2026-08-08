package com.mailblastr.resources;

import com.mailblastr.ListParams;
import com.mailblastr.MailblastrResponse;
import com.mailblastr.http.ApiClient;
import com.mailblastr.requests.CreateEventRequest;
import com.mailblastr.requests.SendEventRequest;
import com.mailblastr.requests.UpdateEventRequest;

/**
 * Custom automation events — {@code mailblastr.events()}. Sending an event
 * REQUIRES a {@code domain} (only that domain's automations trigger).
 *
 * <p>None of these routes honour {@code Idempotency-Key} — only
 * {@code POST /emails} and {@code POST /emails/batch} do. A retried event send
 * ingests a second event and can enroll the contact twice, so dedupe on your
 * side before calling.
 */
public final class Events extends Resource {
    public Events(ApiClient api) { super(api); }

    /** Send a custom event that automations can trigger on. {@code POST /events/send} */
    public MailblastrResponse send(SendEventRequest request) {
        return api.request("POST", "/events/send", request);
    }

    /**
     * @deprecated {@code POST /events/send} does not read {@code Idempotency-Key}
     *     — the header is accepted but ignored, so a retry ingests a second
     *     event instead of replaying the first response. Use
     *     {@link #send(SendEventRequest)} and dedupe client-side.
     */
    @Deprecated
    public MailblastrResponse send(SendEventRequest request, String idempotencyKey) {
        return api.request("POST", "/events/send", request, idempotencyKey);
    }

    /** Create a custom-event definition (name + optional payload schema). {@code POST /events} */
    public MailblastrResponse create(CreateEventRequest request) {
        return api.request("POST", "/events", request);
    }

    /**
     * @deprecated {@code POST /events} does not read {@code Idempotency-Key} —
     *     see {@link #send(SendEventRequest, String)}. Use
     *     {@link #create(CreateEventRequest)}; a duplicate event name is
     *     already rejected with a {@code 422 validation_error}.
     */
    @Deprecated
    public MailblastrResponse create(CreateEventRequest request, String idempotencyKey) {
        return api.request("POST", "/events", request, idempotencyKey);
    }

    /** List custom-event definitions. {@code GET /events} */
    public MailblastrResponse list() { return list(null); }

    public MailblastrResponse list(ListParams params) {
        return api.request("GET", "/events" + paginate(params));
    }

    /**
     * Update a custom-event definition's payload schema. The event NAME is
     * immutable (automations reference it) — create a new event to rename.
     * {@code PATCH /events/:id}
     */
    public MailblastrResponse update(String id, UpdateEventRequest request) {
        return api.request("PATCH", "/events/" + enc(id), request);
    }

    /** Delete a custom-event definition. {@code DELETE /events/:id} */
    public MailblastrResponse remove(String id) {
        return api.request("DELETE", "/events/" + enc(id));
    }
}
