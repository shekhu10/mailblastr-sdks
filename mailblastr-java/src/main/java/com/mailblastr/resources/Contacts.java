package com.mailblastr.resources;

import com.mailblastr.MailblastrResponse;
import com.mailblastr.http.ApiClient;
import com.mailblastr.http.Query;
import com.mailblastr.requests.BatchContactsRequest;
import com.mailblastr.requests.CreateContactRequest;
import com.mailblastr.requests.ImportContactsRequest;
import com.mailblastr.requests.ListContactsParams;
import com.mailblastr.requests.UpdateContactRequest;
import com.mailblastr.requests.UpdateContactTopicsRequest;

import java.util.LinkedHashMap;
import java.util.Map;

/**
 * Contacts (DOMAIN-FIRST) — {@code mailblastr.contacts()}. Each sending
 * domain has its own contact pool: the flat {@code /contacts} API requires a
 * {@code domain}, while the nested {@code /audiences/:id/contacts} API derives
 * the pool from the audience in the path. The same email address on two
 * domains is two records with separate consent.
 */
public final class Contacts extends Resource {
    public Contacts(ApiClient api) { super(api); }

    /**
     * Create a contact. Returns the slim ack {@code { object: 'contact', id }}.
     * Pass {@code domain} on the builder for the flat API (required there) or
     * {@code audienceId} for the nested API.
     */
    public MailblastrResponse create(CreateContactRequest request) {
        if (request.getAudienceId() != null) {
            return api.request("POST", "/audiences/" + enc(request.getAudienceId()) + "/contacts", request.toMap());
        }
        Map<String, Object> body = new LinkedHashMap<>(request.toMap());
        body.put("domain", request.getDomain());
        return api.request("POST", "/contacts", body);
    }

    /** Retrieve a contact by id (exact — no domain needed). {@code GET /contacts/:id} */
    public MailblastrResponse get(String id) {
        return api.request("GET", "/contacts/" + enc(id));
    }

    /**
     * Retrieve a contact by id or EMAIL; {@code domain} picks the pool when an
     * email exists in several domains' pools. {@code GET /contacts/:id?domain=}
     */
    public MailblastrResponse get(String id, String domain) {
        Query q = new Query().add("domain", domain);
        return api.request("GET", "/contacts/" + enc(id) + q);
    }

    /** Retrieve a contact via the nested API. {@code GET /audiences/:audienceId/contacts/:id} */
    public MailblastrResponse getByAudience(String audienceId, String id) {
        return api.request("GET", "/audiences/" + enc(audienceId) + "/contacts/" + enc(id));
    }

    /** List a domain pool's contacts (shortcut for the flat API). */
    public MailblastrResponse list(String domain) {
        return list(ListContactsParams.builder().domain(domain).build());
    }

    /**
     * List contacts. DOMAIN-FIRST: {@code domain} is required on the flat
     * {@code GET /contacts} list (names the pool); pass {@code audienceId} to
     * use the nested list instead. {@code segmentId} works on both.
     */
    public MailblastrResponse list(ListContactsParams params) {
        Query q = new Query();
        if (params.getDomain() != null && params.getAudienceId() == null) q.add("domain", params.getDomain());
        q.add("limit", params.getLimit());
        q.add("after", params.getAfter());
        q.add("before", params.getBefore());
        q.add("segment_id", params.getSegmentId());
        String base = params.getAudienceId() != null
                ? "/audiences/" + enc(params.getAudienceId()) + "/contacts"
                : "/contacts";
        return api.request("GET", base + q);
    }

    /**
     * Bulk-import contacts from a list (upsert by email; max 10,000 per call).
     * {@code POST /audiences/:id/contacts/batch}
     */
    public MailblastrResponse batch(BatchContactsRequest request) {
        Query q = new Query().add("on_conflict", request.getOnConflict());
        Map<String, Object> body = new LinkedHashMap<>();
        body.put("contacts", request.getContacts());
        return api.request("POST", "/audiences/" + enc(request.getAudienceId()) + "/contacts/batch" + q, body);
    }

    /**
     * Bulk-import contacts from CSV text (upsert by email; non-builtin columns
     * auto-registered as custom properties unless {@code createProperties(false)}).
     * {@code POST /audiences/:id/contacts/import}
     */
    public MailblastrResponse importCsv(ImportContactsRequest request) {
        Query q = new Query().add("on_conflict", request.getOnConflict());
        if (Boolean.FALSE.equals(request.getCreateProperties())) q.add("create_properties", "false");
        Map<String, Object> body = new LinkedHashMap<>();
        body.put("csv", request.getCsv());
        return api.request("POST", "/audiences/" + enc(request.getAudienceId()) + "/contacts/import" + q, body);
    }

    /**
     * Update a contact. Returns the slim ack. On the flat API, set
     * {@code domain} on the builder when the id is an EMAIL.
     */
    public MailblastrResponse update(UpdateContactRequest request) {
        if (request.getAudienceId() != null) {
            return api.request("PATCH",
                    "/audiences/" + enc(request.getAudienceId()) + "/contacts/" + enc(request.getId()),
                    request.toMap());
        }
        Map<String, Object> body = new LinkedHashMap<>(request.toMap());
        if (request.getDomain() != null) body.put("domain", request.getDomain());
        return api.request("PATCH", "/contacts/" + enc(request.getId()), body);
    }

    /** Delete a contact by id. {@code DELETE /contacts/:id} */
    public MailblastrResponse remove(String id) {
        return api.request("DELETE", "/contacts/" + enc(id));
    }

    /** Delete a contact by id or EMAIL + domain. {@code DELETE /contacts/:id?domain=} */
    public MailblastrResponse remove(String id, String domain) {
        Query q = new Query().add("domain", domain);
        return api.request("DELETE", "/contacts/" + enc(id) + q);
    }

    /** Delete a contact via the nested API. {@code DELETE /audiences/:audienceId/contacts/:id} */
    public MailblastrResponse removeByAudience(String audienceId, String id) {
        return api.request("DELETE", "/audiences/" + enc(audienceId) + "/contacts/" + enc(id));
    }

    /** Add a contact to a segment. {@code POST /contacts/:id/segments/:segmentId} */
    public MailblastrResponse addToSegment(String id, String segmentId) {
        return api.request("POST", "/contacts/" + enc(id) + "/segments/" + enc(segmentId));
    }

    /** Remove a contact from a segment. {@code DELETE /contacts/:id/segments/:segmentId} */
    public MailblastrResponse removeFromSegment(String id, String segmentId) {
        return api.request("DELETE", "/contacts/" + enc(id) + "/segments/" + enc(segmentId));
    }

    /** List the segments a contact belongs to. {@code GET /contacts/:id/segments} */
    public MailblastrResponse listSegments(String id) {
        return api.request("GET", "/contacts/" + enc(id) + "/segments");
    }

    /** Get a contact's topic subscriptions. {@code GET /contacts/:id/topics} */
    public MailblastrResponse getTopics(String id) {
        return api.request("GET", "/contacts/" + enc(id) + "/topics");
    }

    /** Update a contact's topic subscriptions. {@code PATCH /contacts/:id/topics} */
    public MailblastrResponse updateTopics(String id, UpdateContactTopicsRequest request) {
        return api.request("PATCH", "/contacts/" + enc(id) + "/topics", request);
    }
}
