/** System designs: the schema, the per-step state, the layout and the words. */
import assert from 'node:assert/strict';
import { test } from 'node:test';
import {
  architectureAt,
  architectureSchema,
  describeArchitecture,
  layoutArchitecture
} from '../../src/visualizer/core/index.ts';

const design = architectureSchema.parse({
  components: [
    { id: 'client', kind: 'client', label: 'Browser' },
    { id: 'api', kind: 'service', label: 'API' },
    { id: 'cache', kind: 'cache', label: 'Cache' },
    { id: 'db', kind: 'db', label: 'Database' }
  ],
  links: [
    { from: 'client', to: 'api' },
    { from: 'api', to: 'cache', label: 'lookup' },
    { from: 'api', to: 'db' }
  ],
  frames: [
    {
      note: 'a miss',
      requests: [
        { link: 0, id: 1, status: 'ok' },
        { link: 1, id: 1, status: 'miss' },
        { link: 2, id: 1, status: 'ok' }
      ],
      loads: [{ component: 'cache', load: 1, queue: 0, hitRate: 0 }]
    },
    {
      note: 'a hit',
      requests: [
        { link: 0, id: 2, status: 'ok' },
        { link: 1, id: 2, status: 'hit' }
      ],
      loads: [{ component: 'cache', load: 1, queue: 0, hitRate: 50 }]
    }
  ]
});

test('state at each step: the frame, the latest loads, the running totals', () => {
  const before = architectureAt(design, 0);
  assert.equal(before.frame, undefined);
  assert.equal(before.loads.size, 0);
  assert.equal(describeArchitecture(design, before), 'The design before any request.');

  const first = architectureAt(design, 1);
  assert.equal(first.frame?.note, 'a miss');
  assert.equal(first.totals.miss, 1);
  assert.equal(first.loads.get('cache')?.hitRate, 0);

  const last = architectureAt(design, 99);
  assert.equal(last.step, 2);
  assert.equal(last.totals.ok, 3);
  assert.equal(last.totals.hit, 1);
  assert.equal(last.loads.get('cache')?.hitRate, 50);
  assert.match(
    describeArchitecture(design, last),
    /^a hit\. Browser to API \(ok\); API to Cache \(hit\)\.$/
  );
});

test('layout puts components in the order requests flow', () => {
  const pos = layoutArchitecture(design);
  assert.ok(pos.get('client')!.x < pos.get('api')!.x);
  assert.ok(pos.get('api')!.x < pos.get('cache')!.x);
  assert.equal(pos.get('cache')!.x, pos.get('db')!.x);
  assert.notEqual(pos.get('cache')!.y, pos.get('db')!.y);
});

test('layout survives a cycle', () => {
  const loop = architectureSchema.parse({
    components: [
      { id: 'a', kind: 'queue', label: 'A' },
      { id: 'b', kind: 'worker', label: 'B' }
    ],
    links: [
      { from: 'a', to: 'b' },
      { from: 'b', to: 'a' }
    ],
    frames: []
  });
  assert.equal(layoutArchitecture(loop).size, 2);
});

test('the schema rejects a request on a missing link or a load on a missing component', () => {
  const bad = (patch: object) =>
    architectureSchema.safeParse({
      components: [{ id: 'a', kind: 'service', label: 'A' }],
      links: [],
      frames: [{ note: 'x', requests: [], loads: [], ...patch }]
    }).success;
  assert.equal(bad({ requests: [{ link: 0, id: 1, status: 'ok' }] }), false);
  assert.equal(bad({ loads: [{ component: 'nope', load: 1, queue: 0, hitRate: -1 }] }), false);
  assert.equal(bad({}), true);
});
