<?php

declare(strict_types=1);

namespace Mailblastr\Resources;

use Mailblastr\Client;

/**
 * API keys. The full secret token is returned only once, at creation.
 */
class ApiKeys extends Resource
{
    /**
     * Create an API key. POST /api-keys
     *
     * @param array $payload ['name' => …, 'permission' => 'full_access'|'sending_access',
     *                       'domain_id' => scope a sending_access key to one domain (legacy; prefer domain_ids),
     *                       'domain_ids' => list of domain ids to scope the key to (works with both
     *                       permissions); mutually exclusive with domain_id — providing both is a 422]
     */
    public function create(array $payload): array
    {
        return $this->client->request('POST', '/api-keys', $payload);
    }

    /** List API keys (non-secret display prefixes only). GET /api-keys */
    public function list(): array
    {
        return $this->client->request('GET', '/api-keys');
    }

    /** Revoke an API key. DELETE /api-keys/:id */
    public function remove(string $id): array
    {
        return $this->client->request('DELETE', '/api-keys/' . Client::e($id));
    }
}
