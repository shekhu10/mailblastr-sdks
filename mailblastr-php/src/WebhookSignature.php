<?php

declare(strict_types=1);

namespace Mailblastr;

/**
 * Pure Svix-style webhook signature verification — a local computation, no
 * HTTP request. Also reachable as `$mailblastr->webhooks->verify(...)`.
 *
 * Signing scheme: `{id}.{timestamp}.{body}` → base64 HMAC-SHA256, tagged `v1,`.
 */
final class WebhookSignature
{
    /**
     * Verify a webhook delivery's signature against your endpoint's signing secret.
     *
     * `$payload` MUST be the exact raw request body string the server sent (do
     * not re-serialize parsed JSON — whitespace differences break the signature).
     * `$headers` accepts the svix-id / svix-timestamp / svix-signature headers,
     * read case-insensitively (values may be strings or arrays of strings). A
     * signature header may carry multiple space-separated `v1,<sig>` entries;
     * any one matching makes the delivery valid.
     *
     * @param string $payload The exact raw request body.
     * @param array  $headers The delivery headers (name => value|value[]).
     * @param string $secret  The endpoint's signing secret ('whsec_…' or raw).
     * @param array  $options 'toleranceSec' (int) — max allowed clock skew in
     *                        seconds (default 300; pass 0 to skip the check).
     *
     * @return array{valid: bool, reason?: string}
     */
    public static function verify(string $payload, array $headers, string $secret, array $options = []): array
    {
        $id = self::readHeader($headers, 'svix-id');
        $timestamp = self::readHeader($headers, 'svix-timestamp');
        $sigHeader = self::readHeader($headers, 'svix-signature');
        if ($id === null || $id === '' || $timestamp === null || $timestamp === '' || $sigHeader === null || $sigHeader === '') {
            return ['valid' => false, 'reason' => 'missing_headers'];
        }
        if ($secret === '') {
            return ['valid' => false, 'reason' => 'missing_secret'];
        }

        // Optional timestamp freshness check (default 5-minute tolerance; 0 disables).
        $toleranceSec = array_key_exists('toleranceSec', $options) ? (int) $options['toleranceSec'] : 300;
        if ($toleranceSec > 0) {
            if (!is_numeric($timestamp)) {
                return ['valid' => false, 'reason' => 'invalid_timestamp'];
            }
            $skew = abs(time() - (int) floor((float) $timestamp));
            if ($skew > $toleranceSec) {
                return ['valid' => false, 'reason' => 'timestamp_out_of_tolerance'];
            }
        }

        $signed = $id . '.' . $timestamp . '.' . $payload;
        $expected = base64_encode(hash_hmac('sha256', $signed, self::secretToKey($secret), true));

        // The header may contain multiple space-separated `v1,<sig>` entries; any match wins.
        foreach (explode(' ', $sigHeader) as $part) {
            $trimmed = trim($part);
            if ($trimmed === '') {
                continue;
            }
            $sig = str_starts_with($trimmed, 'v1,') ? substr($trimmed, 3) : $trimmed;
            if (hash_equals($expected, $sig)) {
                return ['valid' => true];
            }
        }
        return ['valid' => false, 'reason' => 'no_match'];
    }

    /** Case-insensitively read a single header value (first if it's an array). */
    private static function readHeader(array $headers, string $name): ?string
    {
        foreach ($headers as $key => $value) {
            if (strtolower((string) $key) === $name) {
                if (is_array($value)) {
                    return isset($value[0]) ? (string) $value[0] : null;
                }
                return (string) $value;
            }
        }
        return null;
    }

    /**
     * Derive the HMAC key from a `whsec_`-prefixed secret (base64-decode the
     * suffix); a secret without the prefix is used as raw UTF-8 bytes.
     *
     * The decode is deliberately LENIENT, mirroring the key the server actually
     * SIGNS with: its own secretToKey uses Node's `Buffer.from(suffix,
     * 'base64')`, which ignores characters outside the alphabet and accepts the
     * URL-safe one. Decoding strictly here returned false for a caller-supplied
     * secret (POST /webhooks accepts one) whose suffix was not valid standard
     * base64, fell back to the raw string as the key, and then rejected every
     * genuine delivery as `no_match`.
     */
    private static function secretToKey(string $secret): string
    {
        if (str_starts_with($secret, 'whsec_')) {
            $suffix = strtr(substr($secret, strlen('whsec_')), '-_', '+/');
            $decoded = base64_decode($suffix, false);
            if ($decoded !== false && $decoded !== '') {
                return $decoded;
            }
        }
        return $secret;
    }
}
