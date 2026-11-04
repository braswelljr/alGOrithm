/**
 * measure() through the real algorithm.wasm, and the growth helpers the Big O
 * page draws with. Run `make web/wasm` at the repository root first.
 */
import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import { before, describe, test } from 'node:test';
import { fileURLToPath } from 'node:url';
import {
  growth,
  growthClasses,
  predict,
  startAlgorithm,
  type AlgorithmApi
} from '../../src/visualizer/core/index.ts';

const asset = (name: string) => fileURLToPath(new URL(`../../public/${name}`, import.meta.url));
let api: AlgorithmApi;

before(async () => {
  await import(asset('wasm_exec.js'));
  api = await startAlgorithm(await readFile(asset('algorithm.wasm')));
});

describe('measure', () => {
  test('merge sort fits n log n and agrees with its stated cost', () => {
    const r = api.measure('merge-sort', [64, 128, 256, 512]);
    assert.ok(r.ok, r.ok ? '' : r.failure.error);
    if (!r.ok) return;
    const m = r.value;
    assert.equal(m.best, 'O(n log n)');
    assert.equal(m.agrees, true);
    assert.equal(m.samples.length, 4);
    assert.ok(Math.abs(predict(m.fits[0], 512) - m.samples[3].ops) / m.samples[3].ops < 0.2);
  });

  test('bubble sort on sorted values is one pass', () => {
    const r = api.measure('bubble-sort', [], 'sorted');
    assert.ok(r.ok && r.value.best === 'O(n)');
  });

  test('a structure cannot be measured, and says why', () => {
    const r = api.measure('stack');
    assert.ok(!r.ok && /cannot be measured/.test(r.failure.error));
  });

  test('every entry the catalog lists carries a why field, many of them filled', () => {
    const filled = api.catalog().entries.filter((e) => e.why !== '');
    assert.ok(filled.length >= 120, `${filled.length} entries explain their cost`);
  });
});

test('growth classes are in order of growth at large n', () => {
  const at = growthClasses.map((c) => growth(c.class, 64));
  for (let i = 1; i < at.length; i++) assert.ok(at[i] > at[i - 1], growthClasses[i].class);
  assert.ok(Number.isNaN(growth('O(n!)', 5)));
});
