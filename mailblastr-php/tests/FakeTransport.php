<?php

declare(strict_types=1);

namespace Mailblastr\Tests;

use Mailblastr\Transport\TransportInterface;

/**
 * In-memory transport for tests: records every request (method/url/headers/body)
 * and replays queued responses. Defaults to `200 {}` when the queue is empty.
 */
class FakeTransport implements TransportInterface
{
    /** @var array<int, array{method: string, url: string, headers: array, body: ?string}> */
    public array $requests = [];

    /** @var array<int, array{status: int, body: string}> */
    private array $queue = [];

    /** Queue the next response (arrays are JSON-encoded; strings pass through raw). */
    public function queue(int $status, array|string $body): void
    {
        $this->queue[] = [
            'status' => $status,
            'body' => is_string($body) ? $body : (string) json_encode($body),
        ];
    }

    public function request(string $method, string $url, array $headers, ?string $body): array
    {
        $this->requests[] = [
            'method' => $method,
            'url' => $url,
            'headers' => $headers,
            'body' => $body,
        ];
        $next = array_shift($this->queue);
        return $next ?? ['status' => 200, 'body' => '{}'];
    }

    /** The most recent recorded request. */
    public function last(): array
    {
        if ($this->requests === []) {
            throw new \RuntimeException('FakeTransport: no requests recorded.');
        }
        return $this->requests[count($this->requests) - 1];
    }

    /** The path portion (with query) of the most recent request URL. */
    public function lastPath(): string
    {
        $url = $this->last()['url'];
        $pos = strpos($url, '://');
        if ($pos === false) {
            return $url;
        }
        $slash = strpos($url, '/', $pos + 3);
        return $slash === false ? '' : substr($url, $slash);
    }

    /** The decoded JSON body of the most recent request (null when none). */
    public function lastJson(): mixed
    {
        $body = $this->last()['body'];
        return $body === null ? null : json_decode($body, true);
    }
}
