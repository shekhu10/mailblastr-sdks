<?php

declare(strict_types=1);

namespace Mailblastr;

use Mailblastr\Exceptions\MailblastrException;
use Mailblastr\Transport\CurlTransport;
use Mailblastr\Transport\TransportInterface;

/**
 * The MailBlastr API client. Construct via {@see Mailblastr::client()}.
 *
 * Exposes one property per resource:
 * `emails` (with nested `emails->attachments` and `emails->receiving`),
 * `batch`, `domains`, `audiences`, `contacts`, `contactProperties`,
 * `campaigns`, `segments`, `topics`, `templates`, `automations`, `webhooks`,
 * `logs`, `events`, `apiKeys` (list only — key lifecycle is dashboard-only),
 * `polls`.
 *
 * Successful calls return the decoded JSON response as an associative array.
 * Every non-2xx response throws {@see MailblastrException} carrying the API
 * error shape (statusCode / name / message).
 */
class Client
{
    public const DEFAULT_BASE_URL = 'https://www.mailblastr.com/api';
    public const USER_AGENT = 'mailblastr-php/' . Mailblastr::VERSION;

    /**
     * Longest `Idempotency-Key` the API accepts. The accepted range is
     * 1-255 characters measured after the server trims the value (the storage
     * column is VARCHAR(255)); anything outside it is a
     * 400 `invalid_idempotency_key`. The header is honoured by
     * `POST /emails` and `POST /emails/batch` ONLY — every other endpoint
     * ignores it, so a retry there creates a second resource.
     *
     * The SDK does not check the length itself: the server is the authority.
     */
    public const IDEMPOTENCY_KEY_MAX_LENGTH = 255;

    private string $apiKey;
    private string $baseUrl;
    private TransportInterface $transport;

    public readonly Resources\Emails $emails;
    public readonly Resources\Batch $batch;
    public readonly Resources\Domains $domains;
    public readonly Resources\Audiences $audiences;
    public readonly Resources\Contacts $contacts;
    public readonly Resources\ContactProperties $contactProperties;
    public readonly Resources\Campaigns $campaigns;
    public readonly Resources\Segments $segments;
    public readonly Resources\Topics $topics;
    public readonly Resources\Templates $templates;
    public readonly Resources\Automations $automations;
    public readonly Resources\Webhooks $webhooks;
    public readonly Resources\Logs $logs;
    public readonly Resources\Events $events;
    public readonly Resources\ApiKeys $apiKeys;
    public readonly Resources\Polls $polls;

    /**
     * @param string $apiKey  Your API key, e.g. 'mb_xxxxxxxxx'.
     * @param array  $options 'baseUrl' (string), 'transport' (TransportInterface),
     *                        'timeout' (int seconds per request, default 30; 0 = no timeout),
     *                        'maxRetries' (int automatic retries on 429/503, default 2; 0 disables).
     *                        'timeout'/'maxRetries' apply only to the default curl transport;
     *                        a supplied 'transport' is used as-is.
     */
    public function __construct(string $apiKey, array $options = [])
    {
        if ($apiKey === '') {
            throw new \InvalidArgumentException(
                'Mailblastr: an API key is required, e.g. Mailblastr::client("mb_...").'
            );
        }
        $this->apiKey = $apiKey;
        $this->baseUrl = rtrim((string) ($options['baseUrl'] ?? self::DEFAULT_BASE_URL), '/');
        $transport = $options['transport'] ?? new CurlTransport(
            (int) ($options['timeout'] ?? 30),
            (int) ($options['maxRetries'] ?? 2),
        );
        if (!$transport instanceof TransportInterface) {
            throw new \InvalidArgumentException('Mailblastr: "transport" must implement TransportInterface.');
        }
        $this->transport = $transport;

        $this->emails = new Resources\Emails($this);
        $this->batch = new Resources\Batch($this);
        $this->domains = new Resources\Domains($this);
        $this->audiences = new Resources\Audiences($this);
        $this->contacts = new Resources\Contacts($this);
        $this->contactProperties = new Resources\ContactProperties($this);
        $this->campaigns = new Resources\Campaigns($this);
        $this->segments = new Resources\Segments($this);
        $this->topics = new Resources\Topics($this);
        $this->templates = new Resources\Templates($this);
        $this->automations = new Resources\Automations($this);
        $this->webhooks = new Resources\Webhooks($this);
        $this->logs = new Resources\Logs($this);
        $this->events = new Resources\Events($this);
        $this->apiKeys = new Resources\ApiKeys($this);
        $this->polls = new Resources\Polls($this);
    }

    /**
     * Perform a JSON API request.
     *
     * @param string            $method  HTTP method.
     * @param string            $path    API path beginning with '/', already URL-encoded.
     * @param array|object|null $body    JSON body (null sends none; [] sends {}).
     * @param array             $options 'idempotencyKey' (string) — sent verbatim as
     *                                   Idempotency-Key. Honoured by POST /emails and
     *                                   POST /emails/batch ONLY; 1-255 characters
     *                                   ({@see self::IDEMPOTENCY_KEY_MAX_LENGTH}),
     *                                   validated server-side.
     *
     * @return array The decoded JSON response.
     *
     * @throws MailblastrException on any non-2xx response or network failure.
     */
    public function request(string $method, string $path, array|object|null $body = null, array $options = []): array
    {
        $json = null;
        if ($body !== null) {
            // An empty PHP array must serialize as a JSON object ({}), not [] —
            // every empty-body endpoint here (campaigns->send, templates->duplicate, …)
            // expects an object. List bodies (e.g. /emails/batch) are never empty.
            if (is_array($body) && $body === []) {
                $body = new \stdClass();
            }
            $json = json_encode($body, JSON_UNESCAPED_SLASHES | JSON_UNESCAPED_UNICODE);
            if ($json === false) {
                throw new \InvalidArgumentException('Mailblastr: request body could not be JSON-encoded.');
            }
        }

        $headers = [
            'Authorization: Bearer ' . $this->apiKey,
            'Content-Type: application/json',
            'Accept: application/json',
            'User-Agent: ' . self::USER_AGENT,
        ];
        if (!empty($options['idempotencyKey'])) {
            $headers[] = 'Idempotency-Key: ' . $options['idempotencyKey'];
        }

        $res = $this->transport->request($method, $this->baseUrl . $path, $headers, $json);

        $decoded = $res['body'] !== '' ? json_decode($res['body'], true) : null;

        if ($res['status'] < 200 || $res['status'] >= 300) {
            throw MailblastrException::fromResponse($res['status'], is_array($decoded) ? $decoded : []);
        }

        return is_array($decoded) ? $decoded : [];
    }

    /**
     * Like {@see request()}, but for endpoints that stream raw bytes (e.g. a
     * received-email attachment download). Returns the raw body string; error
     * responses are parsed as JSON and thrown like `request`.
     *
     * @throws MailblastrException on any non-2xx response or network failure.
     */
    public function requestRaw(string $method, string $path, array $options = []): string
    {
        $headers = [
            'Authorization: Bearer ' . $this->apiKey,
            'User-Agent: ' . self::USER_AGENT,
        ];
        if (!empty($options['idempotencyKey'])) {
            $headers[] = 'Idempotency-Key: ' . $options['idempotencyKey'];
        }

        $res = $this->transport->request($method, $this->baseUrl . $path, $headers, null);

        if ($res['status'] < 200 || $res['status'] >= 300) {
            $decoded = $res['body'] !== '' ? json_decode($res['body'], true) : null;
            throw MailblastrException::fromResponse($res['status'], is_array($decoded) ? $decoded : []);
        }

        return $res['body'];
    }

    /**
     * Build a query string ('' or '?a=b&c=d') from params. Nulls are dropped;
     * booleans become 'true'/'false'.
     */
    public function query(array $params): string
    {
        $clean = [];
        foreach ($params as $key => $value) {
            if ($value === null) {
                continue;
            }
            if (is_bool($value)) {
                $value = $value ? 'true' : 'false';
            }
            $clean[$key] = $value;
        }
        if ($clean === []) {
            return '';
        }
        return '?' . http_build_query($clean, '', '&', PHP_QUERY_RFC3986);
    }

    /**
     * Percent-encode a single path segment, so an id like '../api-keys' can't
     * traverse the URL path and forge a request to a different endpoint.
     */
    public static function e(string $segment): string
    {
        return rawurlencode($segment);
    }
}
