<?php

declare(strict_types=1);

namespace Mailblastr\Resources;

use Mailblastr\Client;

/**
 * Emails — send, list, retrieve, reschedule, cancel. Sub-resources:
 * `$mailblastr->emails->attachments` (sent-email attachments) and
 * `$mailblastr->emails->receiving` (inbound email).
 */
class Emails extends Resource
{
    /** Sent-email attachments sub-resource. */
    public readonly EmailAttachments $attachments;
    /** Inbound (received) email sub-resource. */
    public readonly ReceivingEmails $receiving;

    public function __construct(Client $client)
    {
        parent::__construct($client);
        $this->attachments = new EmailAttachments($client);
        $this->receiving = new ReceivingEmails($client);
    }

    /**
     * Send a single email. POST /emails
     *
     * @param array $payload from, to, subject, html/text, cc, bcc, reply_to,
     *                       preview_text, headers, attachments (each with
     *                       'filename' + 'content' base64 OR 'path' URL),
     *                       tags, scheduled_at, topic_id, template_id,
     *                       template, variables.
     * @param array $options 'idempotencyKey' to safely retry.
     */
    public function send(array $payload, array $options = []): array
    {
        return $this->client->request('POST', '/emails', $payload, $options);
    }

    /**
     * Send up to 100 emails in one request. POST /emails/batch
     * (alias of `$mailblastr->batch->send()`).
     *
     * Batch items reject 'attachments' and 'scheduled_at' — send those
     * individually via `send()`.
     *
     * @param array $payloads A list of send-email payloads.
     */
    public function batch(array $payloads, array $options = []): array
    {
        return $this->client->request('POST', '/emails/batch', $payloads, $options);
    }

    /**
     * List sent emails (trimmed list items — no status/html/text/events).
     * GET /emails
     *
     * @param array $params Cursor pagination (limit, after, before) plus
     *                      optional server-side filters: campaign_id,
     *                      automation_id, source ('individual' restricts to
     *                      one-off API sends), domain_id.
     */
    public function list(array $params = []): array
    {
        return $this->client->request('GET', '/emails' . $this->paginationQuery($params, ['campaign_id', 'automation_id', 'source', 'domain_id']));
    }

    /** Retrieve a sent email and its events. GET /emails/:id */
    public function get(string $id): array
    {
        return $this->client->request('GET', '/emails/' . Client::e($id));
    }

    /**
     * Reschedule a scheduled email. PATCH /emails/:id
     *
     * @param array $payload ['scheduled_at' => ISO 8601 timestamp]
     */
    public function update(string $id, array $payload): array
    {
        return $this->client->request('PATCH', '/emails/' . Client::e($id), $payload);
    }

    /** Cancel a scheduled email. POST /emails/:id/cancel */
    public function cancel(string $id): array
    {
        return $this->client->request('POST', '/emails/' . Client::e($id) . '/cancel');
    }
}
