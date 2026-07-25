<?php

declare(strict_types=1);

namespace Mailblastr\Resources;

use Mailblastr\Client;

/**
 * Automations — multi-step workflows triggered by events.
 * DOMAIN-FIRST: 'domain' is REQUIRED on create; only `events->send()` calls
 * with the same 'domain' trigger the automation, so the same event name across
 * several products can never trigger the wrong one.
 */
class Automations extends Resource
{
    /**
     * Create an automation. POST /automations
     *
     * @param array $payload ['name' => …, 'domain' => REQUIRED,
     *                       'trigger' => 'contact.created' | 'mailblastr:schedule' |
     *                                    'email.opened' | 'email.clicked' |
     *                                    'email.replied' | 'email.bounced' |
     *                                    'email.delivered' | any custom event name,
     *                       'trigger_config' => ['at' => ISO 8601 instant,
     *                                            'timezone' => IANA name]
     *                                           (required with the
     *                                           'mailblastr:schedule' trigger;
     *                                           not accepted on any other),
     *                       'status' => 'enabled'|'disabled' (default 'disabled'),
     *                       'steps' => [['key' => …, 'type' => …, 'config' => …], …],
     *                       'connections' => [['from' => …, 'to' => …, 'type' => …], …]]
     */
    public function create(array $payload): array
    {
        return $this->client->request('POST', '/automations', $payload);
    }

    /** Retrieve an automation (steps, connections, enrollments). GET /automations/:id */
    public function get(string $id): array
    {
        return $this->client->request('GET', '/automations/' . Client::e($id));
    }

    /**
     * List automations. GET /automations
     *
     * @param array $params Cursor pagination: limit, after, before.
     */
    public function list(array $params = []): array
    {
        return $this->client->request('GET', '/automations' . $this->paginationQuery($params));
    }

    /**
     * Update an automation (name/status/domain/trigger_config/connections).
     * 'trigger_config' (['at' => …, 'timezone' => …]) updates the
     * 'mailblastr:schedule' trigger's schedule (only valid on automations with
     * that trigger). PATCH /automations/:id
     */
    public function update(string $id, array $payload): array
    {
        return $this->client->request('PATCH', '/automations/' . Client::e($id), $payload);
    }

    /**
     * Append a step to an automation. POST /automations/:id/steps
     *
     * @param array $payload ['type' => …, 'config' => …, 'key' => …]
     */
    public function addStep(string $id, array $payload): array
    {
        return $this->client->request('POST', '/automations/' . Client::e($id) . '/steps', $payload);
    }

    /** Delete a step from an automation. DELETE /automations/:id/steps/:stepId */
    public function deleteStep(string $id, string $stepId): array
    {
        return $this->client->request(
            'DELETE',
            '/automations/' . Client::e($id) . '/steps/' . Client::e($stepId)
        );
    }

    /**
     * List an automation's runs. GET /automations/:id/runs
     *
     * @param array $params Cursor pagination: limit, after, before.
     */
    public function runs(string $id, array $params = []): array
    {
        return $this->client->request(
            'GET',
            '/automations/' . Client::e($id) . '/runs' . $this->paginationQuery($params)
        );
    }

    /** Retrieve a single automation run (with its step trace). GET /automations/:id/runs/:runId */
    public function getRun(string $id, string $runId): array
    {
        return $this->client->request(
            'GET',
            '/automations/' . Client::e($id) . '/runs/' . Client::e($runId)
        );
    }

    /** Stop an automation — prevents new runs; in-progress runs finish. POST /automations/:id/stop */
    public function stop(string $id): array
    {
        return $this->client->request('POST', '/automations/' . Client::e($id) . '/stop');
    }

    /** Delete an automation. DELETE /automations/:id */
    public function remove(string $id): array
    {
        return $this->client->request('DELETE', '/automations/' . Client::e($id));
    }
}
