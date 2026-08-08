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

    /**
     * List campaigns. With no pagination params the route returns ALL of them
     * and {@code has_more:false}. {@code GET /campaigns}
     */
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

    /**
     * Send (or schedule) with an IANA {@code schedule_timezone}, persisted
     * onto the campaign so daily batching evaluates batch-days in that zone.
     * Either argument may be {@code null} to omit it.
     * {@code POST /campaigns/:id/send}
     */
    public MailblastrResponse send(String id, String scheduledAt, String scheduleTimezone) {
        Map<String, Object> body = new LinkedHashMap<>();
        if (scheduledAt != null) body.put("scheduled_at", scheduledAt);
        if (scheduleTimezone != null) body.put("schedule_timezone", scheduleTimezone);
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

    /**
     * Who opened / clicked / replied, per contact. Each of the three lists is
     * hard-capped at 500 rows and this route is not paginated.
     * {@code GET /campaigns/:id/engagement}
     */
    public MailblastrResponse engagement(String id) {
        return api.request("GET", "/campaigns/" + enc(id) + "/engagement");
    }

    /**
     * A/B winner evaluation for an A/B campaign. A campaign that is not an A/B
     * test answers {@code 422 validation_error}. Note the deliberately
     * camelCase {@code zScore} / {@code pValue} fields in the result.
     * {@code GET /campaigns/:id/ab}
     */
    public MailblastrResponse ab(String id) {
        return api.request("GET", "/campaigns/" + enc(id) + "/ab");
    }

    /** {@code DELETE /campaigns/:id} */
    public MailblastrResponse remove(String id) {
        return api.request("DELETE", "/campaigns/" + enc(id));
    }
}
