<?php

declare(strict_types=1);

namespace Mailblastr\Resources;

/**
 * API keys — listing only, by design.
 *
 * Keys are created, re-scoped and revoked in the MailBlastr dashboard by a
 * signed-in user: POST /api-keys, PATCH /api-keys/:id and DELETE /api-keys/:id
 * answer 403 `dashboard_only` to every API-key caller, whatever its
 * permission. Exposing only list() means a leaked key cannot mint itself a
 * replacement or widen its own access.
 */
class ApiKeys extends Resource
{
    /**
     * List API keys (non-secret display prefixes only; revoked keys are
     * excluded). GET /api-keys
     *
     * @param array $params Optional cursor pagination (limit, after, before).
     *                      With no params every key is returned.
     */
    public function list(array $params = []): array
    {
        return $this->client->request('GET', '/api-keys' . $this->paginationQuery($params));
    }
}
