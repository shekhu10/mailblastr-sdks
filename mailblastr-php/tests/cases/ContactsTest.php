<?php

declare(strict_types=1);

// ---- create: flat, domain-first ----
[$mb, $t] = make_client();
$mb->contacts->create([
    'domain' => 'example.com',
    'email' => 'user@example.com',
    'first_name' => 'Ada',
]);
check_same('contacts.create flat: method', 'POST', $t->last()['method']);
check_same('contacts.create flat: path', '/contacts', $t->lastPath());
check_same('contacts.create flat: domain kept in body', [
    'domain' => 'example.com',
    'email' => 'user@example.com',
    'first_name' => 'Ada',
], $t->lastJson());

// ---- create: nested audience route strips audienceId + domain from body ----
$mb->contacts->create([
    'audienceId' => 'aud_1',
    'domain' => 'example.com',
    'email' => 'user@example.com',
]);
check_same('contacts.create nested: path', '/audiences/aud_1/contacts', $t->lastPath());
check_same('contacts.create nested: body has neither audienceId nor domain', ['email' => 'user@example.com'], $t->lastJson());

// ---- get: by email + domain ----
$mb->contacts->get(['id' => 'user@example.com', 'domain' => 'example.com']);
check_same('contacts.get: email id encoded + domain query', '/contacts/user%40example.com?domain=example.com', $t->lastPath());

$mb->contacts->get(['id' => 'cont_1', 'audienceId' => 'aud_1']);
check_same('contacts.get nested: path', '/audiences/aud_1/contacts/cont_1', $t->lastPath());

// ---- list: flat requires domain; nested ignores it ----
$mb->contacts->list(['domain' => 'example.com', 'limit' => 10, 'segment_id' => 'seg_1']);
check_same('contacts.list flat: path', '/contacts?domain=example.com&limit=10&segment_id=seg_1', $t->lastPath());

$mb->contacts->list(['audienceId' => 'aud_1', 'domain' => 'example.com', 'limit' => 10]);
check_same('contacts.list nested: domain omitted from query', '/audiences/aud_1/contacts?limit=10', $t->lastPath());

// ---- batch import ----
$mb->contacts->batch([
    'audienceId' => 'aud_1',
    'on_conflict' => 'skip',
    'contacts' => [['email' => 'a@b.com'], ['email' => 'c@d.com']],
]);
check_same('contacts.batch: path', '/audiences/aud_1/contacts/batch?on_conflict=skip', $t->lastPath());
check_same('contacts.batch: body wraps contacts', ['contacts' => [['email' => 'a@b.com'], ['email' => 'c@d.com']]], $t->lastJson());

// ---- CSV import ----
$mb->contacts->import([
    'audienceId' => 'aud_1',
    'csv' => "email,company\na@b.com,Acme",
    'create_properties' => false,
]);
check_same('contacts.import: strict-mode query', '/audiences/aud_1/contacts/import?create_properties=false', $t->lastPath());
check_same('contacts.import: csv body', ['csv' => "email,company\na@b.com,Acme"], $t->lastJson());

// ---- update: flat by email keeps domain in body ----
$mb->contacts->update([
    'id' => 'user@example.com',
    'domain' => 'example.com',
    'unsubscribed' => true,
]);
check_same('contacts.update flat: method', 'PATCH', $t->last()['method']);
check_same('contacts.update flat: path', '/contacts/user%40example.com', $t->lastPath());
check_same('contacts.update flat: body', ['domain' => 'example.com', 'unsubscribed' => true], $t->lastJson());

// ---- remove: nested + flat with domain ----
$mb->contacts->remove(['id' => 'cont_1', 'audienceId' => 'aud_1']);
check_same('contacts.remove nested: method', 'DELETE', $t->last()['method']);
check_same('contacts.remove nested: path', '/audiences/aud_1/contacts/cont_1', $t->lastPath());

$mb->contacts->remove(['id' => 'user@example.com', 'domain' => 'example.com']);
check_same('contacts.remove flat: path', '/contacts/user%40example.com?domain=example.com', $t->lastPath());

// ---- segments + topics helpers ----
$mb->contacts->addToSegment('cont_1', 'seg_1');
check_same('contacts.addToSegment: method', 'POST', $t->last()['method']);
check_same('contacts.addToSegment: path', '/contacts/cont_1/segments/seg_1', $t->lastPath());

$mb->contacts->removeFromSegment('cont_1', 'seg_1');
check_same('contacts.removeFromSegment: method', 'DELETE', $t->last()['method']);

$mb->contacts->listSegments('cont_1');
check_same('contacts.listSegments: path', '/contacts/cont_1/segments', $t->lastPath());

$mb->contacts->getTopics('cont_1');
check_same('contacts.getTopics: path', '/contacts/cont_1/topics', $t->lastPath());

$mb->contacts->updateTopics('cont_1', ['topics' => [['id' => 'top_1', 'subscription' => 'opt_in']]]);
check_same('contacts.updateTopics: method', 'PATCH', $t->last()['method']);
check_same('contacts.updateTopics: path', '/contacts/cont_1/topics', $t->lastPath());
check_same('contacts.updateTopics: body', ['topics' => [['id' => 'top_1', 'subscription' => 'opt_in']]], $t->lastJson());
