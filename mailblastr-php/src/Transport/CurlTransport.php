<?php

declare(strict_types=1);

namespace Mailblastr\Transport;

use Mailblastr\Exceptions\MailblastrException;

/**
 * Default HTTP transport backed by the curl extension (no composer runtime deps).
 */
class CurlTransport implements TransportInterface
{
    public function __construct(private readonly int $timeoutSeconds = 30)
    {
    }

    public function request(string $method, string $url, array $headers, ?string $body): array
    {
        $ch = curl_init($url);
        if ($ch === false) {
            throw new MailblastrException('Mailblastr: failed to initialize curl.', 0, 'network_error');
        }

        curl_setopt_array($ch, [
            CURLOPT_CUSTOMREQUEST => $method,
            CURLOPT_RETURNTRANSFER => true,
            CURLOPT_HTTPHEADER => $headers,
            CURLOPT_TIMEOUT => $this->timeoutSeconds,
            CURLOPT_FOLLOWLOCATION => false,
        ]);
        if ($body !== null) {
            curl_setopt($ch, CURLOPT_POSTFIELDS, $body);
        }

        $responseBody = curl_exec($ch);
        if ($responseBody === false) {
            $error = curl_error($ch);
            curl_close($ch);
            throw new MailblastrException('Network error: ' . $error, 0, 'network_error');
        }

        $status = (int) curl_getinfo($ch, CURLINFO_RESPONSE_CODE);
        curl_close($ch);

        return ['status' => $status, 'body' => (string) $responseBody];
    }
}
