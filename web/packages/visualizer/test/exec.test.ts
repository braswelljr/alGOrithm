/**
 * Machine runs (Output.exec): the schema's checks, and every compiler and
 * architecture entry through the real wasm, frame by frame. Run
 * `make web/wasm` at the repository root first.
 */
import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import { before, test } from 'node:test';
import { fileURLToPath } from 'node:url';
import { execSchema, startAlgorithm, type AlgorithmApi } from '../dist/core.js';

const asset = (name: string) => fileURLToPath(new URL(`../assets/${name}`, import.meta.url));
let api: AlgorithmApi;

before(async () => {
  await import(asset('wasm_exec.js'));
  api = await startAlgorithm(await readFile(asset('algorithm.wasm')));
});

const frame = (over: Record<string, unknown> = {}) => ({ note: 'n', pc: 0, focus: -1, ...over });

test('the exec schema fills Go nulls and rejects what points nowhere', () => {
  const ok = execSchema.parse({
    view: 'vm',
    code: ['PUSH 1'],
    frames: [frame({ stack: null, vars: null })]
  });
  assert.deepEqual(ok.frames[0].stack, []);
  assert.deepEqual(ok.frames[0].vars, {});
  assert.equal(ok.frames[0].cycle, 0);
  assert.equal(
    execSchema.safeParse({ view: 'vm', code: ['A'], frames: [frame({ pc: 3 })] }).success,
    false
  );
  assert.equal(
    execSchema.safeParse({ view: 'heap', objects: ['A'], frames: [frame({ refs: [[2]] })] })
      .success,
    false,
    'a reference to object 2 of 1'
  );
  assert.equal(
    execSchema.safeParse({ view: 'pipeline', rows: ['a'], cells: [], frames: [] }).success,
    false,
    'a row without cells'
  );
  assert.equal(execSchema.safeParse({ view: 'radio', frames: [] }).success, false);
});

test('every compiler and architecture entry runs as a machine the views can draw', () => {
  const entries = api
    .catalog()
    .entries.filter((e) => e.category === 'compilers' || e.category === 'architecture');
  assert.ok(entries.length >= 30, `${entries.length} entries`);
  const views = new Set<string>();
  for (const e of entries) {
    const r = api.run(e.id, e.example, 0);
    assert.ok(r.ok, `${e.id}: ${r.ok ? '' : r.failure.error}`);
    const x = r.value.output.exec;
    if (!x) continue;
    views.add(x.view);
    assert.ok(x.frames.length > 0, `${e.id} has no frames`);
    for (const f of x.frames) assert.ok(f.note.length > 0, `${e.id}: a frame without words`);
    if (x.view === 'cpu') {
      const phases = x.frames.slice(0, 4).map((f) => f.phase);
      assert.deepEqual(phases, ['fetch', 'decode', 'execute', 'writeback'], e.id);
      assert.equal(x.frames[0].registers.length, 8, e.id);
    }
    if (x.view === 'pipeline') {
      assert.equal(x.frames.length, x.cells[0].length, `${e.id}: one frame per cycle`);
    }
    if (x.view === 'tokens') {
      assert.equal(x.tokens.at(-1)?.kind, 'eof', e.id);
      for (const t of x.tokens) assert.equal((x.source ?? '').slice(t.start, t.end), t.text, e.id);
    }
    if (x.view === 'heap') {
      for (const f of x.frames) assert.equal(f.colors.length, x.objects.length, e.id);
    }
  }
  assert.deepEqual([...views].sort(), ['cpu', 'heap', 'pipeline', 'tokens', 'vm']);
});

test('the stack VM and the interpreter print the same', () => {
  const vm = api.run(
    'stack-vm',
    api.catalog().entries.find((e) => e.id === 'stack-vm')!.example,
    0
  );
  const tw = api.run(
    'tree-walking-interpreter',
    api.catalog().entries.find((e) => e.id === 'tree-walking-interpreter')!.example,
    0
  );
  assert.ok(vm.ok && tw.ok);
  assert.deepEqual(vm.value.output.lines, tw.value.output.lines);
  assert.deepEqual(vm.value.output.lines, ['1', '2', '6', '24']);
  const last = vm.value.output.exec!.frames.at(-1)!;
  assert.deepEqual(last.output, ['1', '2', '6', '24']);
});
