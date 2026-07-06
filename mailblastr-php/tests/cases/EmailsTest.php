<?php

declare(strict_types=1);

// ---- emails->send ----
[$mb, $t] = make_client();
$t->queue(200, ['id' => 'em_1']);
$res = $mb->emails->send([
    'from' => 'Acme <hello@yourdomain.com>',
    'to' => ['user@example.com'],
    'subject' => 'Hello',
    'html' => '<p>Hi</p>',
], ['idempotencyKey' => 'order-123']);
check_same('emails.send: method', 'POST', $t->last()['method']);
check_same('emails.send: path', '/emails', $t->lastPath());
check_same('emails.send: body', [
    'from' => 'Acme <hello@yourdomain.com>',
    'to' => ['user@example.com'],
    'subject' => 'Hello',
    'html' => '<p>Hi</p>',
], $t->lastJson());
check_auth('emails.send', $t->last());
check('emails.send: idempotency header', in_array('Idempotency-Key: order-123', $t->last()['headers'], true));
check_same('emails.send: response', ['id' => 'em_1'], $res);

// ---- batch->send (JSON list body) ----
[$mb, $t] = make_client();
$t->queue(200, ['data' => [['id' => 'em_1'], ['id' => 'em_2']]]);
$mb->batch->send([
    ['from' => 'a@x.com', 'to' => 'u1@e.com', 'subject' => 's1', 'text' => 'b1'],
    ['from' => 'a@x.com', 'to' => 'u2@e.com', 'subject' => 's2', 'text' => 'b2'],
]);
check_same('batch.send: method', 'POST', $t->last()['method']);
check_same('batch.send: path', '/emails/batch', $t->lastPath());
check('batch.send: body is a JSON array', str_starts_with((string) $t->last()['body'], '['));

// emails->batch alias hits the same endpoint
$mb->emails->batch([['from' => 'a@x.com', 'to' => 'u@e.com', 'subject' => 's', 'text' => 'b']]);
check_same('emails.batch alias: path', '/emails/batch', $t->lastPath());

// ---- list with pagination ----
[$mb, $t] = make_client();
$mb->emails->list(['limit' => 2, 'after' => 'em_1']);
check_same('emails.list: method', 'GET', $t->last()['method']);
check_same('emails.list: path', '/emails?limit=2&after=em_1', $t->lastPath());
check_same('emails.list: no body on GET', null, $t->last()['body']);

// ---- get with a path-traversal id is escaped ----
$mb->emails->get('../api-keys');
check_same('emails.get: traversal id is percent-encoded', '/emails/..%2Fapi-keys', $t->lastPath());

// ---- update (reschedule) + cancel ----
$mb->emails->update('em_1', ['scheduled_at' => '2026-08-01T00:00:00Z']);
check_same('emails.update: method', 'PATCH', $t->last()['method']);
check_same('emails.update: path', '/emails/em_1', $t->lastPath());
check_same('emails.update: body', ['scheduled_at' => '2026-08-01T00:00:00Z'], $t->lastJson());

$mb->emails->cancel('em_1');
check_same('emails.cancel: method', 'POST', $t->last()['method']);
check_same('emails.cancel: path', '/emails/em_1/cancel', $t->lastPath());
check_same('emails.cancel: no body', null, $t->last()['body']);

// ---- attachments sub-resource ----
$mb->emails->attachments->list(emailId: 'em_1');
check_same('emails.attachments.list: path', '/emails/em_1/attachments', $t->lastPath());
$mb->emails->attachments->get(emailId: 'em_1', attachmentId: 'att_9');
check_same('emails.attachments.get: path', '/emails/em_1/attachments/att_9', $t->lastPath());

// ---- receiving sub-resource ----
[$mb, $t] = make_client();
$mb->emails->receiving->list(['limit' => 5]);
check_same('receiving.list: path', '/emails/receiving?limit=5', $t->lastPath());

$mb->emails->receiving->get('rcv_1');
check_same('receiving.get: path', '/emails/receiving/rcv_1', $t->lastPath());

$mb->emails->receiving->listAttachments('rcv_1');
check_same('receiving.listAttachments: path', '/emails/receiving/rcv_1/attachments', $t->lastPath());

$t->queue(200, 'RAWBYTES');
$bytes = $mb->emails->receiving->getAttachment('rcv_1', 'att_1');
check_same('receiving.getAttachment: raw bytes returned', 'RAWBYTES', $bytes);
check_same('receiving.getAttachment: path', '/emails/receiving/rcv_1/attachments/att_1', $t->lastPath());
check_auth('receiving.getAttachment', $t->last());

$t->queue(200, "From: a@b.c\r\n\r\nbody");
$raw = $mb->emails->receiving->getRaw('rcv_1');
check_same('receiving.getRaw: raw MIME returned', "From: a@b.c\r\n\r\nbody", $raw);
check_same('receiving.getRaw: path', '/emails/receiving/rcv_1/raw', $t->lastPath());

$mb->emails->receiving->forward('rcv_1', ['from' => 'me@yourdomain.com', 'to' => 'team@you.com']);
check_same('receiving.forward: method', 'POST', $t->last()['method']);
check_same('receiving.forward: path', '/emails/receiving/rcv_1/forward', $t->lastPath());
check_same('receiving.forward: body', ['from' => 'me@yourdomain.com', 'to' => 'team@you.com'], $t->lastJson());

$mb->emails->receiving->reply('rcv_1', ['from' => 'me@yourdomain.com', 'text' => 'thanks!']);
check_same('receiving.reply: path', '/emails/receiving/rcv_1/reply', $t->lastPath());

$mb->emails->receiving->remove('rcv_1');
check_same('receiving.remove: method', 'DELETE', $t->last()['method']);
check_same('receiving.remove: path', '/emails/receiving/rcv_1', $t->lastPath());
