<?php

declare(strict_types=1);

namespace Mailblastr\Resources;

use Mailblastr\Client;

/**
 * Custom events — the triggers for automations.
 */
class Events extends Resource
{
    /**
     * Send a custom event that automations can trigger on. POST /events/send
     *
     * DOMAIN-FIRST: 'domain' is REQUIRED — only automations belonging to that
     * domain are triggered, and contacts auto-created by the event land in the
     * domain's own contact pool.
     *
     * @param array $payload ['event' => name ('name' accepted as an alias),
     *                       'domain' => REQUIRED,
     *                       'contact_id' => … OR 'email' => …,
     *                       'payload' => [ … ] ('data' accepted as an alias)]
     * @param array $options 'idempotencyKey' to safely retry.
     */
    public function send(array $payload, array $options = []): array
    {
        return $this->client->request('POST', '/events/send', $payload, $options);
    }

    /**
     * Create a custom-event definition (name + optional payload schema).
     * POST /events
     *
     * @param array $payload ['name' => …, 'schema' => ['key' => 'string'|'number'|'boolean'|'date', …]]
     */
    public function create(array $payload, array $options = []): array
    {
        return $this->client->request('POST', '/events', $payload, $options);
    }

    /**
     * List custom-event definitions. GET /events
     *
     * @param array $params Cursor pagination: limit, after, before.
     */
    public function list(array $params = []): array
    {
        return $this->client->request('GET', '/events' . $this->paginationQuery($params));
    }

    /** Delete a custom-event definition. DELETE /events/:id */
    public function remove(string $id): array
    {
        return $this->client->request('DELETE', '/events/' . Client::e($id));
    }
}
