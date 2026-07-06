package com.mailblastr.resources;

import com.mailblastr.ListParams;
import com.mailblastr.MailblastrResponse;
import com.mailblastr.http.ApiClient;
import com.mailblastr.requests.CreateCampaignRequest;
import com.mailblastr.requests.UpdateCampaignRequest;

import java.util.Collections;
import java.util.LinkedHashMap;
import java.util.Map;

/**
 * Campaigns (DOMAIN-FIRST) — bulk sends to a domain's contact pool or one of
 * its segments. {@code domain} is REQUIRED on create.
 */
public final class Campaigns extends Resource {
    public Campaigns(ApiClient api) { super(api); }

    /** {@code POST /campaigns} */
    public MailblastrResponse create(CreateCampaignRequest request) {
        return api.request("POST", "/campaigns", request);
    }

    /** {@code GET /campaigns/:id} */
    public MailblastrResponse get(String id) {
        return api.request("GET", "/campaigns/" + enc(id));
    }

    /** {@code GET /campaigns} */
    public MailblastrResponse list() { return list(null); }

    public MailblastrResponse list(ListParams params) {
        return api.request("GET", "/campaigns" + paginate(params));
    }

    /** {@code PATCH /campaigns/:id} */
    public MailblastrResponse update(String id, UpdateCampaignRequest request) {
        return api.request("PATCH", "/campaigns/" + enc(id), request);
    }

    /** Send now. {@code POST /campaigns/:id/send} */
    public MailblastrResponse send(String id) {
        return api.request("POST", "/campaigns/" + enc(id) + "/send", Collections.emptyMap());
    }

    /** Schedule the send. {@code POST /campaigns/:id/send} with {@code { scheduled_at }}. */
    public MailblastrResponse send(String id, String scheduledAt) {
        Map<String, Object> body = new LinkedHashMap<>();
        body.put("scheduled_at", scheduledAt);
        return api.request("POST", "/campaigns/" + enc(id) + "/send", body);
    }

    /** Cancel a scheduled campaign (returns it to draft). {@code POST /campaigns/:id/cancel} */
    public MailblastrResponse cancel(String id) {
        return api.request("POST", "/campaigns/" + enc(id) + "/cancel");
    }

    /** Per-campaign analytics (counts, engagement rates, top links). {@code GET /campaigns/:id/stats} */
    public MailblastrResponse stats(String id) {
        return api.request("GET", "/campaigns/" + enc(id) + "/stats");
    }

    /** A/B winner evaluation for an A/B campaign. {@code GET /campaigns/:id/ab} */
    public MailblastrResponse ab(String id) {
        return api.request("GET", "/campaigns/" + enc(id) + "/ab");
    }

    /** {@code DELETE /campaigns/:id} */
    public MailblastrResponse remove(String id) {
        return api.request("DELETE", "/campaigns/" + enc(id));
    }
}
