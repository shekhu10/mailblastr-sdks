package com.mailblastr.resources;

import com.mailblastr.ListParams;
import com.mailblastr.MailblastrResponse;
import com.mailblastr.http.ApiClient;
import com.mailblastr.requests.AutomationStep;
import com.mailblastr.requests.CreateAutomationRequest;
import com.mailblastr.requests.UpdateAutomationRequest;

/**
 * Automations (DOMAIN-FIRST) — {@code domain} is REQUIRED on create; only
 * {@code events().send(...)} calls carrying the same domain trigger them.
 */
public final class Automations extends Resource {
    public Automations(ApiClient api) { super(api); }

    /** {@code POST /automations} — {@code domain} is required on the request. */
    public MailblastrResponse create(CreateAutomationRequest request) {
        return api.request("POST", "/automations", request);
    }

    /** {@code GET /automations/:id} */
    public MailblastrResponse get(String id) {
        return api.request("GET", "/automations/" + enc(id));
    }

    /** {@code GET /automations} */
    public MailblastrResponse list() { return list(null); }

    public MailblastrResponse list(ListParams params) {
        return api.request("GET", "/automations" + paginate(params));
    }

    /** {@code PATCH /automations/:id} */
    public MailblastrResponse update(String id, UpdateAutomationRequest request) {
        return api.request("PATCH", "/automations/" + enc(id), request);
    }

    /** Append a step. {@code POST /automations/:id/steps} */
    public MailblastrResponse addStep(String id, AutomationStep step) {
        return api.request("POST", "/automations/" + enc(id) + "/steps", step);
    }

    /** Delete a step. {@code DELETE /automations/:id/steps/:stepId} */
    public MailblastrResponse deleteStep(String id, String stepId) {
        return api.request("DELETE", "/automations/" + enc(id) + "/steps/" + enc(stepId));
    }

    /** List an automation's runs. {@code GET /automations/:id/runs} */
    public MailblastrResponse runs(String id) { return runs(id, null); }

    public MailblastrResponse runs(String id, ListParams params) {
        return api.request("GET", "/automations/" + enc(id) + "/runs" + paginate(params));
    }

    /** Retrieve a single run with its step trace. {@code GET /automations/:id/runs/:runId} */
    public MailblastrResponse getRun(String id, String runId) {
        return api.request("GET", "/automations/" + enc(id) + "/runs/" + enc(runId));
    }

    /** Stop an automation — prevents new runs; in-progress runs finish. {@code POST /automations/:id/stop} */
    public MailblastrResponse stop(String id) {
        return api.request("POST", "/automations/" + enc(id) + "/stop");
    }

    /** {@code DELETE /automations/:id} */
    public MailblastrResponse remove(String id) {
        return api.request("DELETE", "/automations/" + enc(id));
    }
}
