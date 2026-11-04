/**
 * Real-world uses in the catalog and example() through the real
 * algorithm.wasm. Run `make web/wasm` at the repository root first.
 */
import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import { before, describe, test } from 'node:test';
import { fileURLToPath } from 'node:url';
import { startAlgorithm, type AlgorithmApi } from '../dist/core.js';

const asset = (name: string) => fileURLToPath(new URL(`../assets/${name}`, import.meta.url));
let api: AlgorithmApi;

before(async () => {
  await import(asset('wasm_exec.js'));
  api = await startAlgorithm(await readFile(asset('algorithm.wasm')));
});

describe('real-world code', () => {
  test('entries carry their uses, and every linked example program is embedded', () => {
    const entries = api.catalog().entries;
    const withUses = entries.filter((e) => e.realWorld.length);
    assert.ok(withUses.length >= 150, `${withUses.length} entries with uses`);
    for (const e of withUses)
      for (const u of e.realWorld) {
        assert.ok(u.title && u.where, `${e.id}: a use needs a title and a sentence`);
        if (!u.example) continue;
        const p = api.example(u.example);
        assert.ok(p.ok, `${e.id}: ${u.example} ${p.ok ? '' : p.failure.error}`);
        if (p.ok) {
          assert.match(p.value.code, /^\/\/ Command /m);
          assert.equal(p.value.command, `go run ./${u.example}`);
        }
      }
  });

  test('an unknown or escaping path is refused', () => {
    for (const path of ['examples/nope', '../pkg', ''])
      assert.equal(api.example(path).ok, false, path);
  });
});
