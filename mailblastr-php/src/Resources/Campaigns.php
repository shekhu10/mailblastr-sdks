<?php

declare(strict_types=1);

namespace Mailblastr\Resources;

use Mailblastr\Client;

/**
 * Campaigns — bulk sends to a domain's contact pool (or a segment of it).
 * DOMAIN-FIRST: 'domain' is REQUIRED on create and picks the contact pool the
 * campaign targets (orthogonal to 'from', which may be a different verified
 * domain).
 */
class Campaigns extends Resource
{
    /**
     * Create a campaign. POST /campaigns
     *
     * @param array $payload domain (REQUIRED), from, subject, html/text,
     *                       reply_to, preview_text, name, segment_id, topic_id,
     *                       recurrence, recurrence_every, ab_test, followups,
     *                       list_to, unsubscribe_policy, send, scheduled_at.
     */
    public function create(array $payload): array
    {
        return $this->client->request('POST', '/campaigns', $payload);
    }

    /** Retrieve a campaign (includes statistics). GET /campaigns/:id */
    public function get(string $id): array
    {
        return $this->client->request('GET', '/campaigns/' . Client::e($id));
    }

    /**
     * List campaigns. GET /campaigns
     *
     * @param array $params Cursor pagination: limit, after, before.
     */
    public function list(array $params = []): array
    {
        return $this->client->request('GET', '/campaigns' . $this->paginationQuery($params));
    }

    /** Update a draft campaign. PATCH /campaigns/:id */
    public function update(string $id, array $payload): array
    {
        return $this->client->request('PATCH', '/campaigns/' . Client::e($id), $payload);
    }

    /**
     * Send now, or schedule with ['scheduled_at' => …]. POST /campaigns/:id/send
     */
    public function send(string $id, array $payload = []): array
    {
        return $this->client->request('POST', '/campaigns/' . Client::e($id) . '/send', $payload);
    }

    /** Cancel a scheduled campaign (returns it to draft). POST /campaigns/:id/cancel */
    public function cancel(string $id): array
    {
        return $this->client->request('POST', '/campaigns/' . Client::e($id) . '/cancel');
    }

    /** Per-campaign analytics (counts, engagement rates, top links). GET /campaigns/:id/stats */
    public function stats(string $id): array
    {
        return $this->client->request('GET', '/campaigns/' . Client::e($id) . '/stats');
    }

    /** A/B winner evaluation for an A/B campaign. GET /campaigns/:id/ab */
    public function ab(string $id): array
    {
        return $this->client->request('GET', '/campaigns/' . Client::e($id) . '/ab');
    }

    /** Delete a campaign. DELETE /campaigns/:id */
    public function remove(string $id): array
    {
        return $this->client->request('DELETE', '/campaigns/' . Client::e($id));
    }
}
