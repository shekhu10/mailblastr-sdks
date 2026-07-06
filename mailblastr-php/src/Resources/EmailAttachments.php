<?php

declare(strict_types=1);

namespace Mailblastr\Resources;

use Mailblastr\Client;

/**
 * Sent-email attachments — `$mailblastr->emails->attachments`.
 */
class EmailAttachments extends Resource
{
    /** List a sent email's attachments. GET /emails/:id/attachments */
    public function list(string $emailId): array
    {
        return $this->client->request('GET', '/emails/' . Client::e($emailId) . '/attachments');
    }

    /** Retrieve one attachment of a sent email (metadata + download_url). GET /emails/:id/attachments/:attachmentId */
    public function get(string $emailId, string $attachmentId): array
    {
        return $this->client->request(
            'GET',
            '/emails/' . Client::e($emailId) . '/attachments/' . Client::e($attachmentId)
        );
    }
}
