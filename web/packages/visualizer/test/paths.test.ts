/**
 * The learning paths from the real algorithm.wasm: every step names an entry in
 * the catalog and runs on its suggested input. Run `make web/wasm` first.
 */
import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import { before, test } from 'node:test';
import { fileURLToPath } from 'node:url';
import { startAlgorithm, type AlgorithmApi } from '../dist/core.js';

const asset = (name: string) => fileURLToPath(new URL(`../assets/${name}`, import.meta.url));
let api: AlgorithmApi;

before(async () => {
  await import(asset('wasm_exec.js'));
  api = await startAlgorithm(await readFile(asset('algorithm.wasm')));
});

test('every learning path step names a catalog entry and runs', () => {
  const paths = api.paths();
  assert.ok(paths.ok, paths.ok ? '' : paths.failure.error);
  if (!paths.ok) return;
  assert.ok(paths.value.length > 5, `${paths.value.length} paths`);
  const entries = new Map(api.catalog().entries.map((e) => [e.id, e]));
  for (const p of paths.value) {
    for (const s of p.steps) {
      const e = entries.get(s.entry);
      assert.ok(e, `${p.id}: no entry ${s.entry}`);
      if (!e) continue;
      assert.ok(s.command.startsWith('algorithm run '), `${p.id}/${s.entry}: ${s.command}`);
      const r = api.run(e.id, s.input ?? e.example, -1);
      assert.ok(r.ok, `${p.id}/${s.entry}: ${r.ok ? '' : r.failure.error}`);
    }
  }
});
