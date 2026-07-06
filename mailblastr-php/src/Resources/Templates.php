<?php

declare(strict_types=1);

namespace Mailblastr\Resources;

use Mailblastr\Client;

/**
 * Email templates (draft/publish lifecycle, {{ variables }}, aliases).
 */
class Templates extends Resource
{
    /**
     * Create a template. POST /templates
     * Returns the slim ack ['object' => 'template', 'id' => …].
     *
     * @param array $payload name (required), alias, subject, from, reply_to,
     *                       html, text, variables ([['key' => …, 'type' => …,
     *                       'fallback_value' => …], …]).
     */
    public function create(array $payload): array
    {
        return $this->client->request('POST', '/templates', $payload);
    }

    /** Retrieve a template. GET /templates/:id */
    public function get(string $id): array
    {
        return $this->client->request('GET', '/templates/' . Client::e($id));
    }

    /**
     * List templates. GET /templates
     *
     * @param array $params Cursor pagination: limit, after, before.
     */
    public function list(array $params = []): array
    {
        return $this->client->request('GET', '/templates' . $this->paginationQuery($params));
    }

    /**
     * Update a template (creates a new draft version). PATCH /templates/:id
     * Returns the slim ack ['object' => 'template', 'id' => …].
     */
    public function update(string $id, array $payload): array
    {
        return $this->client->request('PATCH', '/templates/' . Client::e($id), $payload);
    }

    /**
     * Duplicate a template. POST /templates/:id/duplicate
     *
     * @param array $payload Optional ['name' => …, 'alias' => …].
     */
    public function duplicate(string $id, array $payload = []): array
    {
        return $this->client->request('POST', '/templates/' . Client::e($id) . '/duplicate', $payload);
    }

    /** Publish a template (make its latest draft live). POST /templates/:id/publish */
    public function publish(string $id): array
    {
        return $this->client->request('POST', '/templates/' . Client::e($id) . '/publish');
    }

    /** Delete a template. DELETE /templates/:id */
    public function remove(string $id): array
    {
        return $this->client->request('DELETE', '/templates/' . Client::e($id));
    }
}
