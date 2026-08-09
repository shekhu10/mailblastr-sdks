<?php

declare(strict_types=1);

namespace Mailblastr\Resources;

use Mailblastr\Client;

/**
 * Emails — send, list, retrieve, reschedule, cancel. Sub-resource:
 * `$mailblastr->emails->receiving` (inbound email).
 */
class Emails extends Resource
{
    /** Inbound (received) email sub-resource. */
    public readonly ReceivingEmails $receiving;

    public function __construct(Client $client)
    {
        parent::__construct($client);
        $this->receiving = new ReceivingEmails($client);
    }

    /**
     * Send a single email. POST /emails
     *
     * @param array $payload from (<=320 chars), to (1-50 recipients), subject,
     *                       html/text, cc (<=50), bcc (<=50), reply_to,
     *                       preview_text (<=150 chars), headers, attachments
     *                       (each with 'filename' + 'content' base64 OR 'path'
     *                       URL; <=25MB decoded each, <=40MB decoded total),
     *                       scheduled_at (ISO 8601 or a relative phrase like
     *                       'in 1 min'; at most 30 days ahead), topic_id,
     *                       template_id, template, variables.
     *                       'tags' is NOT a supported field — sending it is a 422.
     * @param array $options 'idempotencyKey' (1-255 characters) to safely retry.
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
     * @param array $options  'idempotencyKey' (1-255 characters) to safely retry.
     */
    public function batch(array $payloads, array $options = []): array
    {
        return $this->client->request('POST', '/emails/batch', $payloads, $options);
    }

    /**
     * List sent emails (trimmed list items — no status/html/text/events, and
     * unset cc/bcc/reply_to come back as null rather than []). GET /emails
     *
     * @param array $params Cursor pagination (limit, after, before) plus
     *                      optional server-side filters: campaign_id,
     *                      automation_id, source ('individual' restricts to
     *                      one-off API sends), domain_id, status (matched
     *                      case-insensitively against the item's last_event),
     *                      search (matches recipients, subject or sender;
     *                      'q' is accepted as an alias).
     */
    public function list(array $params = []): array
    {
        return $this->client->request('GET', '/emails' . $this->paginationQuery(
            $params,
            ['campaign_id', 'automation_id', 'source', 'domain_id', 'status', 'search', 'q']
        ));
    }

    /**
     * Per-source send metrics — one row per campaign, automation, and one for
     * individual (one-off) sends. Not paginated. GET /emails/sources
     */
    public function sources(): array
    {
        return $this->client->request('GET', '/emails/sources');
    }

    /** Retrieve a sent email and its events. GET /emails/:id */
    public function get(string $id): array
    {
        return $this->client->request('GET', '/emails/' . Client::e($id));
    }

    /** List a sent email's attachments. GET /emails/:id/attachments */
    public function listAttachments(string $id): array
    {
        return $this->client->request('GET', '/emails/' . Client::e($id) . '/attachments');
    }

    /** Retrieve one attachment of a sent email (metadata + download_url). GET /emails/:id/attachments/:attachmentId */
    public function getAttachment(string $id, string $attachmentId): array
    {
        return $this->client->request(
            'GET',
            '/emails/' . Client::e($id) . '/attachments/' . Client::e($attachmentId)
        );
    }

    /**
     * Reschedule a scheduled email. PATCH /emails/:id
     *
     * @param array $payload ['scheduled_at' => ISO 8601 timestamp or a relative
     *                       phrase like 'in 1 min'; must be in the future and
     *                       at most 30 days ahead]
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
