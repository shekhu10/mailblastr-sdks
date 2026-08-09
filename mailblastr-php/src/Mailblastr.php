<?php

declare(strict_types=1);

namespace Mailblastr;

/**
 * Entry point for the MailBlastr PHP SDK.
 *
 * ```php
 * $mailblastr = Mailblastr\Mailblastr::client('mb_xxxxxxxxx');
 *
 * $mailblastr->emails->send([
 *     'from' => 'Acme <hello@yourdomain.com>',
 *     'to' => ['user@example.com'],
 *     'subject' => 'Hello',
 *     'html' => '<p>Hi 👋</p>',
 * ]);
 * ```
 */
final class Mailblastr
{
    public const VERSION = '3.0.0';

    /**
     * Create a MailBlastr API client.
     *
     * @param string $apiKey  Your API key, e.g. 'mb_xxxxxxxxx'.
     * @param array  $options Optional config:
     *                        - 'baseUrl'   (string) override the API host
     *                        - 'transport' (Transport\TransportInterface) swap the HTTP transport (e.g. for tests)
     */
    public static function client(string $apiKey, array $options = []): Client
    {
        return new Client($apiKey, $options);
    }
}
