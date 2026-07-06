package com.mailblastr.resources;

import com.mailblastr.MailblastrResponse;
import com.mailblastr.http.ApiClient;
import com.mailblastr.requests.CreateApiKeyRequest;

/** API keys — {@code mailblastr.apiKeys()}. The full token is returned only on create. */
public final class ApiKeys extends Resource {
    public ApiKeys(ApiClient api) { super(api); }

    /** {@code POST /api-keys} */
    public MailblastrResponse create(CreateApiKeyRequest request) {
        return api.request("POST", "/api-keys", request);
    }

    /** {@code GET /api-keys} */
    public MailblastrResponse list() {
        return api.request("GET", "/api-keys");
    }

    /** {@code DELETE /api-keys/:id} */
    public MailblastrResponse remove(String id) {
        return api.request("DELETE", "/api-keys/" + enc(id));
    }
}
