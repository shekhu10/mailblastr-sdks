import { test } from 'node:test';
import assert from 'node:assert/strict';
import { createHmac } from 'node:crypto';
import { MailBlastr, verifyWebhookSignature } from '../src/index';

interface Captured { url: string; method: string; headers: Record<string, string>; body: any }

// A mock fetch that records the request and returns a canned response.
function mockFetch(status: number, responseBody: unknown) {
  const calls: Captured[] = [];
  const fn = (async (url: string, init: any) => {
    calls.push({
      url,
      method: init.method,
      headers: init.headers,
      body: init.body ? JSON.parse(init.body) : undefined,
    });
    return {
      ok: status >= 200 && status < 300,
      status,
      text: async () => (responseBody === undefined ? '' : JSON.stringify(responseBody)),
      arrayBuffer: async () => {
        if (responseBody instanceof ArrayBuffer) return responseBody;
        const s = responseBody === undefined ? '' : JSON.stringify(responseBody);
        return new TextEncoder().encode(s).buffer;
      },
    } as unknown as Response;
  }) as unknown as typeof fetch;
  return { fn, calls };
}

test('constructor requires an API key', () => {
  // @ts-expect-error intentionally missing key
  assert.throws(() => new MailBlastr());
});

test('emails.send issues POST /emails with auth + body and returns { data }', async () => {
  const { fn, calls } = mockFetch(200, { id: 'e-1' });
  const mb = new MailBlastr('mb_test', { baseUrl: 'https://api.test', fetch: fn });
  const res = await mb.emails.send({ from: 'a@x.com', to: 'b@y.com', subject: 'Hi', html: '<p>x</p>' });
  assert.equal(calls[0].method, 'POST');
  assert.equal(calls[0].url, 'https://api.test/emails');
  assert.equal(calls[0].headers.Authorization, 'Bearer mb_test');
  assert.equal(calls[0].body.subject, 'Hi');
  assert.deepEqual(res, { data: { id: 'e-1' }, error: null });
});

test('emails.send forwards the Idempotency-Key header', async () => {
  const { fn, calls } = mockFetch(200, { id: 'e-2' });
  const mb = new MailBlastr('mb_test', { baseUrl: 'https://api.test', fetch: fn });
  await mb.emails.send({ from: 'a@x.com', to: 'b@y.com', subject: 'Hi', text: 't' }, { idempotencyKey: 'k1' });
  assert.equal(calls[0].headers['Idempotency-Key'], 'k1');
});

test('a non-2xx returns { error } with the MailBlastr error shape', async () => {
  const { fn } = mockFetch(422, { statusCode: 422, name: 'validation_error', message: 'bad' });
  const mb = new MailBlastr('mb_test', { baseUrl: 'https://api.test', fetch: fn });
  const res = await mb.emails.send({ from: 'a@x.com', to: 'b@y.com', subject: 'Hi', html: 'x' });
  assert.equal(res.data, null);
  assert.deepEqual(res.error, { statusCode: 422, name: 'validation_error', message: 'bad' });
});

test('emails: batch / get / update / cancel map to the right routes', async () => {
  const { fn, calls } = mockFetch(200, { ok: true });
  const mb = new MailBlastr('mb_test', { baseUrl: 'https://api.test', fetch: fn });
  await mb.emails.batch([{ from: 'a@x.com', to: 'b@y.com', subject: 's', text: 't' }]);
  await mb.emails.get('e1');
  await mb.emails.update('e1', { scheduled_at: '2030-01-01T00:00:00Z' });
  await mb.emails.cancel('e1');
  assert.deepEqual(calls.map((c) => `${c.method} ${c.url}`), [
    'POST https://api.test/emails/batch',
    'GET https://api.test/emails/e1',
    'PATCH https://api.test/emails/e1',
    'POST https://api.test/emails/e1/cancel',
  ]);
});

test('domains resource maps every method', async () => {
  const { fn, calls } = mockFetch(200, {});
  const mb = new MailBlastr('mb_test', { baseUrl: 'https://api.test', fetch: fn });
  await mb.domains.create({ name: 'example.com' });
  await mb.domains.list();
  await mb.domains.get('d1');
  await mb.domains.update('d1', { open_tracking: false });
  await mb.domains.verify('d1');
  await mb.domains.remove('d1');
  assert.deepEqual(calls.map((c) => `${c.method} ${c.url}`), [
    'POST https://api.test/domains',
    'GET https://api.test/domains',
    'GET https://api.test/domains/d1',
    'PATCH https://api.test/domains/d1',
    'POST https://api.test/domains/d1/verify',
    'DELETE https://api.test/domains/d1',
  ]);
});

test('audiences + contacts (nested) map to the right routes', async () => {
  const { fn, calls } = mockFetch(200, {});
  const mb = new MailBlastr('mb_test', { baseUrl: 'https://api.test', fetch: fn });
  await mb.audiences.create({ name: 'Newsletter' });
  await mb.contacts.create({ audienceId: 'a1', email: 'x@y.com', first_name: 'X' });
  await mb.contacts.get({ audienceId: 'a1', id: 'x@y.com' });
  await mb.contacts.update({ audienceId: 'a1', id: 'c1', unsubscribed: true });
  await mb.contacts.remove({ audienceId: 'a1', id: 'c1' });
  await mb.contacts.list({ audienceId: 'a1' });
  assert.deepEqual(calls.map((c) => `${c.method} ${c.url}`), [
    'POST https://api.test/audiences',
    'POST https://api.test/audiences/a1/contacts',
    'GET https://api.test/audiences/a1/contacts/x%40y.com',
    'PATCH https://api.test/audiences/a1/contacts/c1',
    'DELETE https://api.test/audiences/a1/contacts/c1',
    'GET https://api.test/audiences/a1/contacts',
  ]);
  // contact create strips audienceId from the body
  assert.deepEqual(calls[1].body, { email: 'x@y.com', first_name: 'X' });
});

test('contacts (flat, no audienceId) hit the top-level /contacts routes', async () => {
  const { fn, calls } = mockFetch(200, {});
  const mb = new MailBlastr('mb_test', { baseUrl: 'https://api.test', fetch: fn });
  await mb.contacts.create({ email: 'x@y.com', first_name: 'X' });
  await mb.contacts.get({ id: 'x@y.com' });
  await mb.contacts.update({ id: 'c1', unsubscribed: true });
  await mb.contacts.remove({ id: 'c1' });
  await mb.contacts.list();
  await mb.contacts.list({ segment_id: 'seg1', limit: 10 });
  assert.deepEqual(calls.map((c) => `${c.method} ${c.url}`), [
    'POST https://api.test/contacts',
    'GET https://api.test/contacts/x%40y.com',
    'PATCH https://api.test/contacts/c1',
    'DELETE https://api.test/contacts/c1',
    'GET https://api.test/contacts',
    'GET https://api.test/contacts?limit=10&segment_id=seg1',
  ]);
  assert.deepEqual(calls[0].body, { email: 'x@y.com', first_name: 'X' });
});

test('contacts.batch / import / paginated list map to the right routes', async () => {
  const { fn, calls } = mockFetch(201, { object: 'list', imported: 2, updated: 0, skipped: 0, total: 2 });
  const mb = new MailBlastr('mb_test', { baseUrl: 'https://api.test', fetch: fn });
  await mb.contacts.batch({ audienceId: 'a1', contacts: [{ email: 'a@x.com' }, { email: 'b@x.com', first_name: 'B' }] });
  await mb.contacts.import({ audienceId: 'a1', csv: 'email\nc@x.com\n' });
  await mb.contacts.list({ audienceId: 'a1', limit: 50, after: 'cur1' });
  assert.deepEqual(calls.map((c) => `${c.method} ${c.url}`), [
    'POST https://api.test/audiences/a1/contacts/batch',
    'POST https://api.test/audiences/a1/contacts/import',
    'GET https://api.test/audiences/a1/contacts?limit=50&after=cur1',
  ]);
  assert.deepEqual(calls[0].body, { contacts: [{ email: 'a@x.com' }, { email: 'b@x.com', first_name: 'B' }] });
  assert.deepEqual(calls[1].body, { csv: 'email\nc@x.com\n' });
});

test('segments resource maps every method (incl. contacts preview)', async () => {
  const { fn, calls } = mockFetch(200, {});
  const mb = new MailBlastr('mb_test', { baseUrl: 'https://api.test', fetch: fn });
  await mb.segments.create({ audience_id: 'a1', name: 'Gmail users', filter: { status: 'subscribed', email_contains: '@gmail.com' } });
  await mb.segments.list();
  await mb.segments.get('s1');
  await mb.segments.contacts('s1');
  await mb.segments.update('s1', { name: 'Renamed' });
  await mb.segments.remove('s1');
  assert.deepEqual(calls.map((c) => `${c.method} ${c.url}`), [
    'POST https://api.test/segments',
    'GET https://api.test/segments',
    'GET https://api.test/segments/s1',
    'GET https://api.test/segments/s1/contacts',
    'PATCH https://api.test/segments/s1',
    'DELETE https://api.test/segments/s1',
  ]);
});

test('broadcasts.create forwards segment_id', async () => {
  const { fn, calls } = mockFetch(200, { id: 'b-1' });
  const mb = new MailBlastr('mb_test', { baseUrl: 'https://api.test', fetch: fn });
  await mb.broadcasts.create({ audience_id: 'a1', from: 'f@x.com', subject: 's', html: 'x', segment_id: 's1' });
  assert.equal(calls[0].body.segment_id, 's1');
});

test('templates resource maps every method', async () => {
  const { fn, calls } = mockFetch(200, {});
  const mb = new MailBlastr('mb_test', { baseUrl: 'https://api.test', fetch: fn });
  await mb.templates.create({ name: 'Welcome', subject: 'Hi {{name}}', html: '<p>Hello {{name}}</p>' });
  await mb.templates.list();
  await mb.templates.get('t1');
  await mb.templates.update('t1', { subject: 'Yo {{name}}' });
  await mb.templates.remove('t1');
  assert.deepEqual(calls.map((c) => `${c.method} ${c.url}`), [
    'POST https://api.test/templates',
    'GET https://api.test/templates',
    'GET https://api.test/templates/t1',
    'PATCH https://api.test/templates/t1',
    'DELETE https://api.test/templates/t1',
  ]);
});

test('emails.send forwards template_id + variables in the body', async () => {
  const { fn, calls } = mockFetch(200, { id: 'e-9' });
  const mb = new MailBlastr('mb_test', { baseUrl: 'https://api.test', fetch: fn });
  await mb.emails.send({ from: 'a@x.com', to: 'b@y.com', subject: 'ignored', template_id: 't1', variables: { name: 'Ada' } });
  assert.equal(calls[0].body.template_id, 't1');
  assert.deepEqual(calls[0].body.variables, { name: 'Ada' });
});

test('broadcasts + apiKeys map to the right routes', async () => {
  const { fn, calls } = mockFetch(200, {});
  const mb = new MailBlastr('mb_test', { baseUrl: 'https://api.test', fetch: fn });
  await mb.broadcasts.create({ audience_id: 'a1', from: 'f@x.com', subject: 's', html: 'x' });
  await mb.broadcasts.send('b1', { scheduled_at: '2030-01-01T00:00:00Z' });
  await mb.broadcasts.cancel('b1');
  await mb.apiKeys.create({ name: 'CI', permission: 'sending_access', domain_id: 'd1' });
  await mb.apiKeys.remove('k1');
  assert.deepEqual(calls.map((c) => `${c.method} ${c.url}`), [
    'POST https://api.test/broadcasts',
    'POST https://api.test/broadcasts/b1/send',
    'POST https://api.test/broadcasts/b1/cancel',
    'POST https://api.test/api-keys',
    'DELETE https://api.test/api-keys/k1',
  ]);
  assert.equal(calls[3].body.domain_id, 'd1');
});

test('emails.list + sent-attachments map to the right routes', async () => {
  const { fn, calls } = mockFetch(200, {});
  const mb = new MailBlastr('mb_test', { baseUrl: 'https://api.test', fetch: fn });
  await mb.emails.list({ limit: 10, after: 'cur1' });
  await mb.emails.listAttachments('e1');
  await mb.emails.getAttachment('e1', 'att1');
  assert.deepEqual(calls.map((c) => `${c.method} ${c.url}`), [
    'GET https://api.test/emails?limit=10&after=cur1',
    'GET https://api.test/emails/e1/attachments',
    'GET https://api.test/emails/e1/attachments/att1',
  ]);
});

test('emails.receiving maps every method', async () => {
  const { fn, calls } = mockFetch(200, {});
  const mb = new MailBlastr('mb_test', { baseUrl: 'https://api.test', fetch: fn });
  await mb.emails.receiving.list({ limit: 5 });
  await mb.emails.receiving.get('r1');
  await mb.emails.receiving.listAttachments('r1');
  await mb.emails.receiving.getAttachment('r1', 'att2');
  await mb.emails.receiving.forward('r1', { from: 'me@x.com', to: 'fwd@x.com' });
  await mb.emails.receiving.remove('r1');
  assert.deepEqual(calls.map((c) => `${c.method} ${c.url}`), [
    'GET https://api.test/emails/receiving?limit=5',
    'GET https://api.test/emails/receiving/r1',
    'GET https://api.test/emails/receiving/r1/attachments',
    'GET https://api.test/emails/receiving/r1/attachments/att2',
    'POST https://api.test/emails/receiving/r1/forward',
    'DELETE https://api.test/emails/receiving/r1',
  ]);
  assert.deepEqual(calls[4].body, { from: 'me@x.com', to: 'fwd@x.com' });
});

test('emails.receiving.getAttachment returns raw bytes (ArrayBuffer)', async () => {
  const { fn } = mockFetch(200, 'BINARY');
  const mb = new MailBlastr('mb_test', { baseUrl: 'https://api.test', fetch: fn });
  const res = await mb.emails.receiving.getAttachment('r1', 'att2');
  assert.equal(res.error, null);
  assert.ok(res.data instanceof ArrayBuffer);
});

test('batch.send posts the payload array to /emails/batch', async () => {
  const { fn, calls } = mockFetch(200, { data: [] });
  const mb = new MailBlastr('mb_test', { baseUrl: 'https://api.test', fetch: fn });
  await mb.batch.send([{ from: 'a@x.com', to: 'b@y.com', subject: 's', text: 't' }]);
  assert.equal(calls[0].method, 'POST');
  assert.equal(calls[0].url, 'https://api.test/emails/batch');
  assert.ok(Array.isArray(calls[0].body));
});

test('domains claim methods map to the right routes', async () => {
  const { fn, calls } = mockFetch(200, {});
  const mb = new MailBlastr('mb_test', { baseUrl: 'https://api.test', fetch: fn });
  await mb.domains.claim({ name: 'example.com' });
  await mb.domains.getClaim('d1');
  await mb.domains.verifyClaim('d1');
  assert.deepEqual(calls.map((c) => `${c.method} ${c.url}`), [
    'POST https://api.test/domains/claim',
    'GET https://api.test/domains/d1/claim',
    'POST https://api.test/domains/d1/claim/verify',
  ]);
});

test('contacts segment + topic methods map to the right routes', async () => {
  const { fn, calls } = mockFetch(200, {});
  const mb = new MailBlastr('mb_test', { baseUrl: 'https://api.test', fetch: fn });
  await mb.contacts.addToSegment('c1', 's1');
  await mb.contacts.removeFromSegment('c1', 's1');
  await mb.contacts.listSegments('c1');
  await mb.contacts.getTopics('c1');
  await mb.contacts.updateTopics('c1', { topics: [{ id: 't1', subscription: 'opt_in' }] });
  assert.deepEqual(calls.map((c) => `${c.method} ${c.url}`), [
    'POST https://api.test/contacts/c1/segments/s1',
    'DELETE https://api.test/contacts/c1/segments/s1',
    'GET https://api.test/contacts/c1/segments',
    'GET https://api.test/contacts/c1/topics',
    'PATCH https://api.test/contacts/c1/topics',
  ]);
  assert.deepEqual(calls[4].body, { topics: [{ id: 't1', subscription: 'opt_in' }] });
});

test('contactProperties resource maps every method', async () => {
  const { fn, calls } = mockFetch(200, {});
  const mb = new MailBlastr('mb_test', { baseUrl: 'https://api.test', fetch: fn });
  await mb.contactProperties.create({ name: 'Plan', type: 'string' });
  await mb.contactProperties.list();
  await mb.contactProperties.get('p1');
  await mb.contactProperties.update('p1', { fallback_value: 'free' });
  await mb.contactProperties.remove('p1');
  assert.deepEqual(calls.map((c) => `${c.method} ${c.url}`), [
    'POST https://api.test/contact-properties',
    'GET https://api.test/contact-properties',
    'GET https://api.test/contact-properties/p1',
    'PATCH https://api.test/contact-properties/p1',
    'DELETE https://api.test/contact-properties/p1',
  ]);
});

test('topics resource maps every method', async () => {
  const { fn, calls } = mockFetch(200, {});
  const mb = new MailBlastr('mb_test', { baseUrl: 'https://api.test', fetch: fn });
  await mb.topics.create({ name: 'Product updates', default_subscription: 'opt_in' });
  await mb.topics.list();
  await mb.topics.get('t1');
  await mb.topics.update('t1', { description: 'News' });
  await mb.topics.remove('t1');
  assert.deepEqual(calls.map((c) => `${c.method} ${c.url}`), [
    'POST https://api.test/topics',
    'GET https://api.test/topics',
    'GET https://api.test/topics/t1',
    'PATCH https://api.test/topics/t1',
    'DELETE https://api.test/topics/t1',
  ]);
});

test('templates duplicate + publish map to the right routes', async () => {
  const { fn, calls } = mockFetch(200, {});
  const mb = new MailBlastr('mb_test', { baseUrl: 'https://api.test', fetch: fn });
  await mb.templates.duplicate('t1');
  await mb.templates.publish('t1');
  assert.deepEqual(calls.map((c) => `${c.method} ${c.url}`), [
    'POST https://api.test/templates/t1/duplicate',
    'POST https://api.test/templates/t1/publish',
  ]);
});

test('automations resource maps every method', async () => {
  const { fn, calls } = mockFetch(200, {});
  const mb = new MailBlastr('mb_test', { baseUrl: 'https://api.test', fetch: fn });
  await mb.automations.create({ name: 'Onboarding', audience_id: 'a1', trigger: 'contact.created' });
  await mb.automations.list();
  await mb.automations.get('au1');
  await mb.automations.update('au1', { status: 'active' });
  await mb.automations.addStep('au1', { type: 'send_email', config: { template_id: 't1' } });
  await mb.automations.runs('au1');
  await mb.automations.getRun('au1', 'run1');
  await mb.automations.remove('au1');
  assert.deepEqual(calls.map((c) => `${c.method} ${c.url}`), [
    'POST https://api.test/automations',
    'GET https://api.test/automations',
    'GET https://api.test/automations/au1',
    'PATCH https://api.test/automations/au1',
    'POST https://api.test/automations/au1/steps',
    'GET https://api.test/automations/au1/runs',
    'GET https://api.test/automations/au1/runs/run1',
    'DELETE https://api.test/automations/au1',
  ]);
});

test('webhooks resource maps every method', async () => {
  const { fn, calls } = mockFetch(200, {});
  const mb = new MailBlastr('mb_test', { baseUrl: 'https://api.test', fetch: fn });
  await mb.webhooks.create({ endpoint: 'https://hooks.x.com/mb', events: ['email.delivered'] });
  await mb.webhooks.list();
  await mb.webhooks.get('wh1');
  await mb.webhooks.update('wh1', { status: 'disabled' });
  await mb.webhooks.remove('wh1');
  assert.deepEqual(calls.map((c) => `${c.method} ${c.url}`), [
    'POST https://api.test/webhooks',
    'GET https://api.test/webhooks',
    'GET https://api.test/webhooks/wh1',
    'PATCH https://api.test/webhooks/wh1',
    'DELETE https://api.test/webhooks/wh1',
  ]);
});

test('broadcasts.stats + ab map to the right routes', async () => {
  const { fn, calls } = mockFetch(200, {});
  const mb = new MailBlastr('mb_test', { baseUrl: 'https://api.test', fetch: fn });
  await mb.broadcasts.stats('b1');
  await mb.broadcasts.ab('b1');
  assert.deepEqual(calls.map((c) => `${c.method} ${c.url}`), [
    'GET https://api.test/broadcasts/b1/stats',
    'GET https://api.test/broadcasts/b1/ab',
  ]);
});

test('broadcasts.create forwards recurrence + ab_test + new fields', async () => {
  const { fn, calls } = mockFetch(200, { id: 'b-1' });
  const mb = new MailBlastr('mb_test', { baseUrl: 'https://api.test', fetch: fn });
  await mb.broadcasts.create({
    audience_id: 'a1', from: 'f@x.com', subject: 's', html: 'x',
    reply_to: 'r@x.com', preview_text: 'pv',
    recurrence: 'weekly', recurrence_every: 2,
    ab_test: { enabled: true, subject_b: 'B', metric: 'click', test_pct: 30 },
  });
  assert.equal(calls[0].body.recurrence, 'weekly');
  assert.equal(calls[0].body.recurrence_every, 2);
  assert.equal(calls[0].body.preview_text, 'pv');
  assert.equal(calls[0].body.reply_to, 'r@x.com');
  assert.deepEqual(calls[0].body.ab_test, { enabled: true, subject_b: 'B', metric: 'click', test_pct: 30 });
});

test('audiences.update maps to PATCH /audiences/:id', async () => {
  const { fn, calls } = mockFetch(200, {});
  const mb = new MailBlastr('mb_test', { baseUrl: 'https://api.test', fetch: fn });
  await mb.audiences.update('a1', { name: 'Renamed' });
  assert.equal(calls[0].method, 'PATCH');
  assert.equal(calls[0].url, 'https://api.test/audiences/a1');
  assert.deepEqual(calls[0].body, { name: 'Renamed' });
});

test('domains dns helpers map to the right routes', async () => {
  const { fn, calls } = mockFetch(200, {});
  const mb = new MailBlastr('mb_test', { baseUrl: 'https://api.test', fetch: fn });
  await mb.domains.detectDns('d1');
  await mb.domains.applyCloudflareDns('d1', { token: 'cf' });
  await mb.domains.applyGoDaddyDns('d1', { key: 'k', secret: 's' });
  assert.deepEqual(calls.map((c) => `${c.method} ${c.url}`), [
    'GET https://api.test/domains/d1/dns/detect',
    'POST https://api.test/domains/d1/dns/cloudflare',
    'POST https://api.test/domains/d1/dns/godaddy',
  ]);
  assert.deepEqual(calls[1].body, { token: 'cf' });
  assert.deepEqual(calls[2].body, { key: 'k', secret: 's' });
});

test('segments.create forwards filter.property_filters', async () => {
  const { fn, calls } = mockFetch(200, {});
  const mb = new MailBlastr('mb_test', { baseUrl: 'https://api.test', fetch: fn });
  await mb.segments.create({
    audience_id: 'a1', name: 'Pro users',
    filter: { status: 'subscribed', property_filters: [{ key: 'plan', operator: 'eq', value: 'pro' }] },
  });
  assert.deepEqual(calls[0].body.filter.property_filters, [{ key: 'plan', operator: 'eq', value: 'pro' }]);
});

test('emails.receiving.getRaw returns raw bytes (ArrayBuffer)', async () => {
  const { fn, calls } = mockFetch(200, 'RAWMIME');
  const mb = new MailBlastr('mb_test', { baseUrl: 'https://api.test', fetch: fn });
  const res = await mb.emails.receiving.getRaw('r1');
  assert.equal(calls[0].url, 'https://api.test/emails/receiving/r1/raw');
  assert.equal(res.error, null);
  assert.ok(res.data instanceof ArrayBuffer);
});

test('automations.deleteStep maps to DELETE /automations/:id/steps/:stepId', async () => {
  const { fn, calls } = mockFetch(200, { id: 'st1', deleted: true });
  const mb = new MailBlastr('mb_test', { baseUrl: 'https://api.test', fetch: fn });
  await mb.automations.deleteStep('au1', 'st1');
  assert.equal(calls[0].method, 'DELETE');
  assert.equal(calls[0].url, 'https://api.test/automations/au1/steps/st1');
});

test('webhooks.rotate + test map to the right routes', async () => {
  const { fn, calls } = mockFetch(200, {});
  const mb = new MailBlastr('mb_test', { baseUrl: 'https://api.test', fetch: fn });
  await mb.webhooks.rotate('wh1');
  await mb.webhooks.test('wh1');
  assert.deepEqual(calls.map((c) => `${c.method} ${c.url}`), [
    'POST https://api.test/webhooks/wh1/rotate',
    'POST https://api.test/webhooks/wh1/test',
  ]);
});

test('webhooks.verify validates a correct Svix-style signature', () => {
  const mb = new MailBlastr('mb_test', { baseUrl: 'https://api.test', fetch: mockFetch(200, {}).fn });
  const secret = 'whsec_' + Buffer.from('supersecretkeybytes!!').toString('base64');
  const id = 'msg_123';
  const timestamp = String(Math.floor(Date.now() / 1000));
  const payload = JSON.stringify({ type: 'email.delivered', data: { id: 'e1' } });
  // Mirror the backend: HMAC over `${id}.${timestamp}.${body}` with the decoded key.
  const key = Buffer.from(secret.slice('whsec_'.length), 'base64');
  const sig = 'v1,' + createHmac('sha256', key).update(`${id}.${timestamp}.${payload}`).digest('base64');

  const ok = mb.webhooks.verify(payload, { 'svix-id': id, 'svix-timestamp': timestamp, 'svix-signature': sig }, secret);
  assert.deepEqual(ok, { valid: true });

  // Wrong payload → no match.
  const bad = mb.webhooks.verify('{}', { 'svix-id': id, 'svix-timestamp': timestamp, 'svix-signature': sig }, secret);
  assert.equal(bad.valid, false);
  assert.equal(bad.reason, 'no_match');

  // Missing headers.
  const missing = mb.webhooks.verify(payload, {}, secret);
  assert.deepEqual(missing, { valid: false, reason: 'missing_headers' });

  // Reads headers case-insensitively + accepts a raw secret (no whsec_ prefix).
  const rawSecret = 'plainsecret';
  const sig2 = 'v1,' + createHmac('sha256', Buffer.from(rawSecret, 'utf8')).update(`${id}.${timestamp}.${payload}`).digest('base64');
  const ci = mb.webhooks.verify(payload, { 'Svix-Id': id, 'Svix-Timestamp': timestamp, 'Svix-Signature': sig2 }, rawSecret);
  assert.deepEqual(ci, { valid: true });
});

test('verifyWebhookSignature enforces timestamp tolerance (and 0 disables it)', () => {
  const secret = 'whsec_' + Buffer.from('key').toString('base64');
  const id = 'msg_1';
  const stale = String(Math.floor(Date.now() / 1000) - 10_000);
  const payload = '{}';
  const key = Buffer.from(secret.slice('whsec_'.length), 'base64');
  const sig = 'v1,' + createHmac('sha256', key).update(`${id}.${stale}.${payload}`).digest('base64');
  const headers = { 'svix-id': id, 'svix-timestamp': stale, 'svix-signature': sig };

  const rejected = verifyWebhookSignature(payload, headers, secret);
  assert.equal(rejected.valid, false);
  assert.equal(rejected.reason, 'timestamp_out_of_tolerance');

  const accepted = verifyWebhookSignature(payload, headers, secret, { toleranceSec: 0 });
  assert.deepEqual(accepted, { valid: true });
});

test('logs + events map to the right routes', async () => {
  const { fn, calls } = mockFetch(200, {});
  const mb = new MailBlastr('mb_test', { baseUrl: 'https://api.test', fetch: fn });
  await mb.logs.list({ limit: 20, before: 'cur2' });
  await mb.logs.get('l1');
  await mb.events.send({ name: 'signup.completed', email: 'c@x.com', data: { plan: 'pro' } });
  assert.deepEqual(calls.map((c) => `${c.method} ${c.url}`), [
    'GET https://api.test/logs?limit=20&before=cur2',
    'GET https://api.test/logs/l1',
    'POST https://api.test/events',
  ]);
  assert.deepEqual(calls[2].body, { name: 'signup.completed', email: 'c@x.com', data: { plan: 'pro' } });
});
