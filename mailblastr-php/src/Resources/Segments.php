<?php

declare(strict_types=1);

namespace Mailblastr\Resources;

use Mailblastr\Client;

/**
 * Segments. DOMAIN-FIRST: 'domain' is REQUIRED on create and list — segment
 * names are unique within a domain (reusable across domains), and every domain
 * carries an auto-created "General" (all contacts) segment.
 */
class Segments extends Resource
{
    /**
     * Create a segment on a sending domain. POST /segments
     *
     * @param array $payload ['domain' => REQUIRED, 'name' => …,
     *                       'filter' => ['status' => 'all'|'subscribed'|'unsubscribed',
     *                                    'email_contains' => …,
     *                                    'property_filters' => [['key' => …, 'operator' => 'eq'|'contains'|'exists', 'value' => …], …]]]
     */
    public function create(array $payload): array
    {
        return $this->client->request('POST', '/segments', $payload);
    }

    /** Retrieve a segment. GET /segments/:id */
    public function get(string $id): array
    {
        return $this->client->request('GET', '/segments/' . Client::e($id));
    }

    /**
     * List a domain's segments ('domain' is REQUIRED; includes its
     * auto-created "General" segment). GET /segments?domain=…
     *
     * @param array $params ['domain' => REQUIRED, 'limit' => …, 'after' => …, 'before' => …]
     */
    public function list(array $params): array
    {
        return $this->client->request('GET', '/segments' . $this->paginationQuery($params, ['domain']));
    }

    /** Preview the contacts a segment currently resolves to. GET /segments/:id/contacts */
    public function contacts(string $id): array
    {
        return $this->client->request('GET', '/segments/' . Client::e($id) . '/contacts');
    }

    /** Update a segment. PATCH /segments/:id */
    public function update(string $id, array $payload): array
    {
        return $this->client->request('PATCH', '/segments/' . Client::e($id), $payload);
    }

    /** Delete a segment. DELETE /segments/:id */
    public function remove(string $id): array
    {
        return $this->client->request('DELETE', '/segments/' . Client::e($id));
    }
}
