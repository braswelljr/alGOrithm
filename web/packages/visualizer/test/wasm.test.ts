/**
 * The real algorithm.wasm, the one the package ships, driven through the
 * package's own API in Node. Run `make web/wasm` at the repository root first.
 */
import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import { before, describe, test } from 'node:test';
import { fileURLToPath } from 'node:url';
import {
  ArrayReplayer,
  catalogSchema,
  startAlgorithm,
  type AlgorithmApi,
  type Entry,
  type Input
} from '../dist/core.js';

const asset = (name: string) => fileURLToPath(new URL(`../assets/${name}`, import.meta.url));
let api: AlgorithmApi;

before(async () => {
  await import(asset('wasm_exec.js'));
  api = await startAlgorithm(await readFile(asset('algorithm.wasm')));
});

/** The entry's own example, or (from an older build without one) an input its kind accepts. */
function sample(e: Entry): Input {
  if (Object.keys(e.example).length) return e.example;
  switch (e.input) {
    case 'ints':
      return { values: [5, 3, 9, 1, 3] };
    case 'sorted-ints':
    case 'any-ints':
      return { values: [5, 3, 9, 1, 3], target: 9 };
    case 'text':
      return { text: 'Hello World', key: 'lemon', shift: 3 };
    case 'two-texts':
      return { text: 'kitten', other: 'sitting' };
    case 'number':
      return { values: [5] };
    case 'two-numbers':
      return { values: [2, 3] };
    case 'ops':
      return { values: [4, 2], target: 2, ops: [{ name: 'pop' }, { name: 'push', value: '7' }] };
    case 'grid':
      return { text: 'S..#\n.#..\n...E' };
    case 'pattern':
      return { text: 'abracadabra', other: 'abra' };
    case 'amount':
      return { values: [1, 2, 5], target: 11 };
    default:
      return {};
  }
}

describe('algorithm.wasm', () => {
  test('publishes a catalog the schema accepts', () => {
    const c = api.catalog();
    assert.ok(catalogSchema.safeParse(c).success);
    assert.ok(c.entries.length > 20, `${c.entries.length} entries`);
    assert.equal(new Set(c.entries.map((e) => e.id)).size, c.entries.length, 'unique ids');
  });

  test('every entry runs and its result parses', () => {
    for (const e of api.catalog().entries) {
      const r = api.run(e.id, sample(e), 0);
      assert.ok(r.ok || r.failure.error, `${e.id}: ${r.ok ? '' : r.failure.error}`);
      if (e.traced && r.ok) assert.ok(r.value.events.length > 0, `${e.id} reported no events`);
    }
  });

  test('replaying a sort from its initial array gives its result', () => {
    const sorts = api
      .catalog()
      .entries.filter((e) => e.traced && e.input === 'ints' && /^sort/.test(e.category));
    assert.ok(sorts.length > 5, 'the sorts are in the catalog');
    for (const e of sorts) {
      // Bogo sort gives up after `tries` shuffles; give it room to finish 8 values.
      const options: Record<string, number> = {};
      for (const o of e.options ?? []) if (o.name === 'tries' && o.max) options[o.name] = o.max;
      const r = api.run(e.id, { values: [9, -2, 7, 7, 0, 41, -8, 3], options }, 0);
      assert.ok(r.ok, `${e.id}: ${r.ok ? '' : r.failure.error}`);
      // A truncated trace stops before the end, so it cannot replay to the result.
      if (!r.ok || r.value.truncated || !r.value.output.values.length || r.value.output.table)
        continue;
      const replayed = new ArrayReplayer(r.value.output.initial, r.value.events).final();
      assert.deepEqual(replayed, r.value.output.values, e.id);
    }
  });

  test('grids, tables and matches come back in shape', () => {
    const c = api.catalog();
    for (const e of c.entries) {
      const r = api.run(e.id, sample(e), 0);
      if (!r.ok) continue;
      const o = r.value.output;
      if (e.input === 'grid') assert.ok(o.grid && o.path.length > 0, `${e.id} found no path`);
      if (e.input === 'pattern' && !Object.keys(e.example).length)
        assert.deepEqual(o.matches, [0, 7], e.id);
      if (o.timeline.length)
        assert.ok(
          o.timeline.every((s) => s.end >= s.start),
          `${e.id} has a backward span`
        );
      if (o.track)
        assert.ok(
          o.track.path.every((t) => t < o.track!.size),
          `${e.id} leaves the disk`
        );
    }
  });

  test('bad input fails with a message, not a throw', () => {
    const r = api.run('nope', {}, 0);
    assert.equal(r.ok, false);
    const p = api.parse('csv', 'a,b\n1,x\n', { column: '2' });
    assert.equal(p.ok, false);
    assert.ok(!p.ok && p.failure.problems.length > 0);
  });

  test('parse, preview, generate and compare', () => {
    const p = api.parse('text', '3, 1, 2');
    assert.deepEqual(p.ok && p.value, [3, 1, 2]);
    assert.ok(api.preview('a,b\n1,2\n', true).ok);
    const g = api.generate('random', 10, 1);
    assert.ok(g.ok && g.value.length === 10);
    const sorts = api
      .catalog()
      .entries.filter((e) => e.input === 'ints' && e.traced)
      .slice(0, 3)
      .map((e) => e.id);
    const c = api.compare(sorts, { values: [3, 2, 1] });
    assert.ok(c.ok && c.value.length === sorts.length);
  });

  test('source, snippet and generateGrid (new in the per-category build)', (t) => {
    const first = api.catalog().entries[0];
    const s = api.source(first.id);
    if (!s.ok && /has no source/.test(s.failure.error)) {
      t.skip('this wasm build predates source()');
      return;
    }
    assert.ok(s.ok && s.value.code.includes('package '), 'source has Go code');
    const sn = api.snippet(first.id, sample(first));
    assert.ok(sn.ok && sn.value.cli.includes(first.id));
    const g = api.generateGrid('maze', 11, 21, 1);
    assert.ok(g.ok && g.value.includes('S') && g.value.includes('E'));
  });

  test('quiz: the same seed gives the same questions, each with its answer among the choices', () => {
    const entries = api
      .catalog()
      .entries.filter((e) => e.traced)
      .slice(0, 25);
    let asked = 0;
    for (const e of entries) {
      const a = api.quiz(e.id, e.example, 4, 7);
      const b = api.quiz(e.id, e.example, 4, 7);
      assert.ok(a.ok && b.ok, e.id);
      if (!a.ok || !b.ok) continue;
      assert.deepEqual(a.value, b.value, `${e.id}: same seed, same questions`);
      for (const q of a.value) {
        assert.ok(
          q.answer < q.choices.length && new Set(q.choices).size === q.choices.length,
          e.id
        );
      }
      asked += a.value.length;
    }
    assert.ok(asked > 20, `${asked} questions`);
    const bad = api.quiz('no-such-entry', {}, 3, 1);
    assert.equal(bad.ok, false);
  });
});
