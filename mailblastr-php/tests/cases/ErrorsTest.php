<?php

declare(strict_types=1);

use Mailblastr\Exceptions\MailblastrException;

// ---- API error shape is surfaced on the exception ----
[$mb, $t] = make_client();
$t->queue(422, ['statusCode' => 422, 'name' => 'validation_error', 'message' => 'domain is required']);
$caught = null;
try {
    $mb->segments->create(['name' => 'VIP']);
} catch (MailblastrException $e) {
    $caught = $e;
}
check('errors: MailblastrException thrown on 422', $caught instanceof MailblastrException);
check_same('errors: statusCode', 422, $caught?->getStatusCode());
check_same('errors: name', 'validation_error', $caught?->getName());
check_same('errors: message', 'domain is required', $caught?->getMessage());

// ---- non-JSON error body falls back to HTTP status defaults ----
[$mb, $t] = make_client();
$t->queue(500, 'Internal Server Error');
$caught = null;
try {
    $mb->emails->get('em_1');
} catch (MailblastrException $e) {
    $caught = $e;
}
check('errors: exception on non-JSON 500', $caught instanceof MailblastrException);
check_same('errors: fallback statusCode', 500, $caught?->getStatusCode());
check_same('errors: fallback name', 'application_error', $caught?->getName());
check_same('errors: fallback message', 'Request failed with status 500', $caught?->getMessage());

// ---- raw (binary) routes also raise parsed API errors ----
[$mb, $t] = make_client();
$t->queue(404, ['statusCode' => 404, 'name' => 'not_found', 'message' => 'Attachment not found']);
$caught = null;
try {
    $mb->emails->receiving->getAttachment('rcv_1', 'att_missing');
} catch (MailblastrException $e) {
    $caught = $e;
}
check_same('errors: raw route name', 'not_found', $caught?->getName());
check_same('errors: raw route statusCode', 404, $caught?->getStatusCode());

// ---- empty API key rejected client-side ----
$caught = null;
try {
    \Mailblastr\Mailblastr::client('');
} catch (\InvalidArgumentException $e) {
    $caught = $e;
}
check('errors: empty API key throws InvalidArgumentException', $caught instanceof \InvalidArgumentException);

// ---- custom baseUrl (trailing slash trimmed) ----
$transport = new \Mailblastr\Tests\FakeTransport();
$mb = \Mailblastr\Mailblastr::client('mb_test_key', [
    'transport' => $transport,
    'baseUrl' => 'https://api.staging.mailblastr.com/',
]);
$mb->emails->get('em_1');
check_same('client: custom baseUrl', 'https://api.staging.mailblastr.com/emails/em_1', $transport->last()['url']);

// ---- default base URL ----
[$mb, $t] = make_client();
$mb->emails->get('em_1');
check('client: default base URL', str_starts_with($t->last()['url'], 'https://api.mailblastr.com/'));
check('client: user-agent header present', in_array('User-Agent: mailblastr-php/' . \Mailblastr\Mailblastr::VERSION, $t->last()['headers'], true));
