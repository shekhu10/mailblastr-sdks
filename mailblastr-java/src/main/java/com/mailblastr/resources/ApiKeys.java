package com.mailblastr.resources;

import com.mailblastr.ListParams;
import com.mailblastr.MailblastrResponse;
import com.mailblastr.http.ApiClient;

/**
 * API keys — {@code mailblastr.apiKeys()}. Read-only by design.
 *
 * <p>Keys are minted, re-scoped and revoked in the MailBlastr dashboard, and
 * only there: those routes accept a signed-in dashboard session, never an API
 * key. This SDK therefore deliberately exposes no method for them — a leaked
 * key cannot mint itself a replacement, widen its own permission or revoke the
 * keys around it. All it can do is read the inventory below.
 *
 * <p>{@code token} on a listed key is only the 8-character display prefix; the
 * full secret exists solely in the dashboard, at the moment the key is created.
 */
public final class ApiKeys extends Resource {
    public ApiKeys(ApiClient api) { super(api); }

    /**
     * List non-revoked keys. With no pagination params the route returns ALL
     * of them and {@code has_more:false}. {@code GET /api-keys}
     */
    public MailblastrResponse list() { return list(null); }

    public MailblastrResponse list(ListParams params) {
        return api.request("GET", "/api-keys" + paginate(params));
    }
}
