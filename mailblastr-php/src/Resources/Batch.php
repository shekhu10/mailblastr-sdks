<?php

declare(strict_types=1);

namespace Mailblastr\Resources;

/**
 * Batch send — `$mailblastr->batch->send([...])`.
 */
class Batch extends Resource
{
    /**
     * Send up to 100 emails in one request. POST /emails/batch
     *
     * Batch items reject 'attachments' and 'scheduled_at' — send those
     * individually via `$mailblastr->emails->send()`.
     *
     * @param array $payloads A list of send-email payloads.
     * @param array $options  'idempotencyKey' to safely retry.
     */
    public function send(array $payloads, array $options = []): array
    {
        return $this->client->request('POST', '/emails/batch', $payloads, $options);
    }
}
