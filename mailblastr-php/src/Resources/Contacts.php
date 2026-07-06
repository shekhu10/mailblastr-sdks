<?php

declare(strict_types=1);

namespace Mailblastr\Resources;

use Mailblastr\Client;

/**
 * Contacts are DOMAIN-FIRST: each sending domain has its own contact pool (the
 * same address on two domains is two records with separate consent).
 *
 * - Pass 'domain' to use the flat `/contacts` API (REQUIRED there on
 *   create/list; on get/update/remove it disambiguates an EMAIL id across pools).
 * - Pass 'audienceId' instead to target a specific audience via the nested
 *   `/audiences/:id/contacts` API.
 */
class Contacts extends Resource
{
    /**
     * Create a contact. POST /contacts (flat, 'domain' required) or
     * POST /audiences/:id/contacts (when 'audienceId' is given).
     * Returns the slim ack ['object' => 'contact', 'id' => …].
     *
     * @param array $params domain|audienceId + email (required), first_name,
     *                      last_name, unsubscribed, properties.
     */
    public function create(array $params): array
    {
        $audienceId = $params['audienceId'] ?? null;
        unset($params['audienceId']);
        if ($audienceId !== null) {
            // The nested audience route derives its pool from the path; only
            // the flat route takes `domain` in the body.
            unset($params['domain']);
            return $this->client->request('POST', '/audiences/' . Client::e((string) $audienceId) . '/contacts', $params);
        }
        return $this->client->request('POST', '/contacts', $params);
    }

    /**
     * Retrieve a contact by id or email. An id is exact; an EMAIL can exist in
     * several domains' pools, so pass 'domain' to pick the pool (omitted → the
     * oldest match anywhere).
     *
     * @param array $params ['id' => contact id or email, 'domain' => …, 'audienceId' => …]
     */
    public function get(array $params): array
    {
        $id = Client::e((string) ($params['id'] ?? ''));
        if (!empty($params['audienceId'])) {
            return $this->client->request(
                'GET',
                '/audiences/' . Client::e((string) $params['audienceId']) . '/contacts/' . $id
            );
        }
        $qs = isset($params['domain']) ? '?domain=' . rawurlencode((string) $params['domain']) : '';
        return $this->client->request('GET', '/contacts/' . $id . $qs);
    }

    /**
     * List contacts. Domain-first: 'domain' is required on the flat `/contacts`
     * list (names the pool); pass 'audienceId' for the nested list instead.
     * 'segment_id' restricts to members of that segment (both APIs).
     *
     * @param array $params domain|audienceId, limit, after, before, segment_id.
     */
    public function list(array $params = []): array
    {
        $audienceId = $params['audienceId'] ?? null;
        $query = [];
        if ($audienceId === null && isset($params['domain'])) {
            $query['domain'] = $params['domain'];
        }
        foreach (['limit', 'after', 'before', 'segment_id'] as $key) {
            if (isset($params[$key])) {
                $query[$key] = $params[$key];
            }
        }
        $base = $audienceId !== null
            ? '/audiences/' . Client::e((string) $audienceId) . '/contacts'
            : '/contacts';
        return $this->client->request('GET', $base . $this->client->query($query));
    }

    /**
     * Bulk-import contacts from an array. Upserts by email; max 10,000 per
     * call. POST /audiences/:id/contacts/batch
     *
     * @param array $params ['audienceId' => …, 'contacts' => [ … ],
     *                      'on_conflict' => 'upsert'|'skip' ('skip' leaves
     *                      existing contacts untouched; default 'upsert')]
     */
    public function batch(array $params): array
    {
        $qs = isset($params['on_conflict'])
            ? '?on_conflict=' . rawurlencode((string) $params['on_conflict'])
            : '';
        return $this->client->request(
            'POST',
            '/audiences/' . Client::e((string) $params['audienceId']) . '/contacts/batch' . $qs,
            ['contacts' => $params['contacts'] ?? []]
        );
    }

    /**
     * Bulk-import contacts from CSV text (header row optional). Upserts by
     * email. By default every non-builtin CSV column is auto-registered as a
     * custom property (so `company`, `plan`, … survive and become {{merge}}
     * tags). Pass 'create_properties' => false for strict mode.
     * POST /audiences/:id/contacts/import
     *
     * @param array $params ['audienceId' => …, 'csv' => …,
     *                      'on_conflict' => 'upsert'|'skip',
     *                      'create_properties' => bool]
     */
    public function import(array $params): array
    {
        $query = [];
        if (isset($params['on_conflict'])) {
            $query['on_conflict'] = $params['on_conflict'];
        }
        if (($params['create_properties'] ?? null) === false) {
            $query['create_properties'] = 'false';
        }
        return $this->client->request(
            'POST',
            '/audiences/' . Client::e((string) $params['audienceId']) . '/contacts/import' . $this->client->query($query),
            ['csv' => $params['csv'] ?? '']
        );
    }

    /**
     * Update a contact. Returns the slim ack ['object' => 'contact', 'id' => …].
     * On the flat API, pass 'domain' when 'id' is an EMAIL (disambiguates
     * across pools).
     *
     * @param array $params ['id' => contact id or email, 'audienceId' => …,
     *                      'domain' => …] + first_name, last_name,
     *                      unsubscribed, properties.
     */
    public function update(array $params): array
    {
        $id = (string) ($params['id'] ?? '');
        $audienceId = $params['audienceId'] ?? null;
        unset($params['id'], $params['audienceId']);
        if ($audienceId !== null) {
            // The nested audience route derives its pool from the path; the
            // flat route takes an optional `domain` in the body.
            unset($params['domain']);
            return $this->client->request(
                'PATCH',
                '/audiences/' . Client::e((string) $audienceId) . '/contacts/' . Client::e($id),
                $params
            );
        }
        return $this->client->request('PATCH', '/contacts/' . Client::e($id), $params);
    }

    /**
     * Delete a contact. On the flat API, pass 'domain' when 'id' is an EMAIL
     * (disambiguates across pools).
     *
     * @param array $params ['id' => contact id or email, 'domain' => …, 'audienceId' => …]
     */
    public function remove(array $params): array
    {
        $id = Client::e((string) ($params['id'] ?? ''));
        if (!empty($params['audienceId'])) {
            return $this->client->request(
                'DELETE',
                '/audiences/' . Client::e((string) $params['audienceId']) . '/contacts/' . $id
            );
        }
        $qs = isset($params['domain']) ? '?domain=' . rawurlencode((string) $params['domain']) : '';
        return $this->client->request('DELETE', '/contacts/' . $id . $qs);
    }

    /** Add a contact to a segment. POST /contacts/:id/segments/:segmentId */
    public function addToSegment(string $id, string $segmentId): array
    {
        return $this->client->request(
            'POST',
            '/contacts/' . Client::e($id) . '/segments/' . Client::e($segmentId)
        );
    }

    /** Remove a contact from a segment. DELETE /contacts/:id/segments/:segmentId */
    public function removeFromSegment(string $id, string $segmentId): array
    {
        return $this->client->request(
            'DELETE',
            '/contacts/' . Client::e($id) . '/segments/' . Client::e($segmentId)
        );
    }

    /** List the segments a contact belongs to. GET /contacts/:id/segments */
    public function listSegments(string $id): array
    {
        return $this->client->request('GET', '/contacts/' . Client::e($id) . '/segments');
    }

    /** Get a contact's topic subscriptions. GET /contacts/:id/topics */
    public function getTopics(string $id): array
    {
        return $this->client->request('GET', '/contacts/' . Client::e($id) . '/topics');
    }

    /**
     * Update a contact's topic subscriptions. PATCH /contacts/:id/topics
     *
     * @param array $payload ['topics' => [['id' => …, 'subscription' => 'opt_in'|'opt_out'], …]]
     */
    public function updateTopics(string $id, array $payload): array
    {
        return $this->client->request('PATCH', '/contacts/' . Client::e($id) . '/topics', $payload);
    }
}
