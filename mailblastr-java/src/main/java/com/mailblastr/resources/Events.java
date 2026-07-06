package com.mailblastr.resources;

import com.mailblastr.ListParams;
import com.mailblastr.MailblastrResponse;
import com.mailblastr.http.ApiClient;
import com.mailblastr.requests.CreateEventRequest;
import com.mailblastr.requests.SendEventRequest;

/**
 * Custom automation events — {@code mailblastr.events()}. Sending an event
 * REQUIRES a {@code domain} (only that domain's automations trigger).
 */
public final class Events extends Resource {
    public Events(ApiClient api) { super(api); }

    /** Send a custom event that automations can trigger on. {@code POST /events/send} */
    public MailblastrResponse send(SendEventRequest request) {
        return api.request("POST", "/events/send", request);
    }

    public MailblastrResponse send(SendEventRequest request, String idempotencyKey) {
        return api.request("POST", "/events/send", request, idempotencyKey);
    }

    /** Create a custom-event definition (name + optional payload schema). {@code POST /events} */
    public MailblastrResponse create(CreateEventRequest request) {
        return api.request("POST", "/events", request);
    }

    public MailblastrResponse create(CreateEventRequest request, String idempotencyKey) {
        return api.request("POST", "/events", request, idempotencyKey);
    }

    /** List custom-event definitions. {@code GET /events} */
    public MailblastrResponse list() { return list(null); }

    public MailblastrResponse list(ListParams params) {
        return api.request("GET", "/events" + paginate(params));
    }

    /** Delete a custom-event definition. {@code DELETE /events/:id} */
    public MailblastrResponse remove(String id) {
        return api.request("DELETE", "/events/" + enc(id));
    }
}
