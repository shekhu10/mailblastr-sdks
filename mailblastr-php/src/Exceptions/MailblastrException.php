<?php

declare(strict_types=1);

namespace Mailblastr\Exceptions;

/**
 * Thrown for every non-2xx API response (and transport-level failures).
 *
 * Carries the API error shape: `statusCode`, `name`, `message`.
 */
class MailblastrException extends \Exception
{
    public function __construct(
        string $message,
        private readonly int $statusCode = 0,
        private readonly string $errorName = 'application_error',
    ) {
        parent::__construct($message);
    }

    /**
     * Build an exception from a decoded API error body
     * (`{ statusCode, name, message }`), falling back to the HTTP status.
     */
    public static function fromResponse(int $httpStatus, array $body): self
    {
        $message = isset($body['message']) && is_string($body['message'])
            ? $body['message']
            : "Request failed with status {$httpStatus}";
        $statusCode = isset($body['statusCode']) && is_int($body['statusCode'])
            ? $body['statusCode']
            : $httpStatus;
        $name = isset($body['name']) && is_string($body['name'])
            ? $body['name']
            : 'application_error';

        return new self($message, $statusCode, $name);
    }

    /** The HTTP status code (0 for network-level errors). */
    public function getStatusCode(): int
    {
        return $this->statusCode;
    }

    /** The machine-readable API error name, e.g. 'validation_error'. */
    public function getName(): string
    {
        return $this->errorName;
    }
}
