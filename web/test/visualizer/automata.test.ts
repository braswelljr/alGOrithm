/**
 * Automata and cellular automata: the layout and step helpers on their own,
 * then every automata entry through the real algorithm.wasm (run
 * `make web/wasm` at the repository root first).
 */
import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import { before, describe, test } from 'node:test';
import { fileURLToPath } from 'node:url';
import {
  automatonStep,
  describeAutomaton,
  generationEnds,
  layoutAutomaton,
  machineSchema,
  startAlgorithm,
  TableReplayer,
  type AlgorithmApi,
  type Automaton,
  type TraceEvent
} from '../../src/visualizer/core/index.ts';

const tiny: Automaton = machineSchema.parse({
  states: [
    { name: 'q0', start: true },
    { name: 'q1' },
    { name: 'q2', accept: true },
    { name: 'lost' }
  ],
  edges: [
    { from: 0, to: 1, label: 'a' },
    { from: 1, to: 2, label: 'b' },
    { from: 2, to: 2, label: 'b' }
  ],
  input: 'abb',
  steps: [
    { active: [0], head: 0, note: 'start in q0' },
    { active: [1], edges: [0], head: 1, note: 'read a: q0 to q1' },
    { active: [2], edges: [1], head: 2, note: 'read b: q1 to q2' },
    { active: [2], edges: [2], head: 3, note: 'read b: q2 to q2; accept' }
  ],
  accepted: true
});

test('states are laid out by distance from the start; unreachable ones last', () => {
  const pos = layoutAutomaton(tiny);
  assert.ok(pos[0]!.x < pos[1]!.x && pos[1]!.x < pos[2]!.x, 'left to right along the path');
  assert.ok(pos[3]!.x > pos[2]!.x, 'the unreachable state goes after the reachable ones');
  assert.deepEqual(layoutAutomaton(tiny), pos, 'the layout is deterministic');
});

test('a step index clamps to the run, and each step is described', () => {
  assert.equal(automatonStep(tiny, -3)?.note, 'start in q0');
  assert.equal(automatonStep(tiny, 99)?.head, 3);
  assert.match(describeAutomaton(tiny, 1), /Read a: q0 to q1\. In q1\./);
});

test('a machine naming a state that does not exist is rejected', () => {
  const bad = machineSchema.safeParse({ ...tiny, steps: [{ active: [9] }] });
  assert.equal(bad.success, false);
});

test('generation marks split a cellular run into generations', () => {
  const set = (i: number, v: number): TraceEvent => ({
    kind: 2,
    indices: [i],
    values: [v],
    note: ''
  });
  const gen: TraceEvent = { kind: 4, indices: [], values: [], note: 'generation' };
  const events = [set(0, 1), set(1, 0), gen, set(1, 1), gen];
  const ends = generationEnds(events);
  assert.deepEqual(ends, [3, 5]);
  const rep = new TableReplayer(2, events);
  assert.deepEqual(rep.stateAt(ends[0]!).cells, [1, 0]);
  assert.deepEqual(rep.stateAt(ends[1]!).cells, [1, 1]);
});

describe('automata through algorithm.wasm', () => {
  let api: AlgorithmApi;
  const asset = (name: string) => fileURLToPath(new URL(`../../public/${name}`, import.meta.url));
  before(async () => {
    await import(asset('wasm_exec.js'));
    api = await startAlgorithm(await readFile(asset('algorithm.wasm')));
  });

  test('every automata entry runs its example, and a machine run is well formed', () => {
    const entries = api.catalog().entries.filter((e) => e.category === 'automata');
    assert.ok(entries.length >= 25, `${entries.length} automata entries`);
    for (const e of entries) {
      const r = api.run(e.id, e.example, 0);
      assert.ok(r.ok, `${e.id}: ${r.ok ? '' : r.failure.error}`);
      if (!r.ok) continue;
      const m = r.value.output.machine;
      if (m) {
        assert.ok(m.steps.length > 0, `${e.id} has steps`);
        assert.ok(m.steps[0]!.head === 0, `${e.id} starts before the input`);
      }
    }
  });

  test('the busy beaver ends with six ones on its tape after 14 moves', () => {
    const r = api.run('busy-beaver', {}, 0);
    assert.ok(r.ok);
    if (!r.ok) return;
    const m = r.value.output.machine!;
    assert.equal(m.steps.length - 1, 14);
    assert.equal([...m.steps.at(-1)!.tape].filter((c) => c === '1').length, 6);
  });

  test('the Game of Life plays one generation a step', () => {
    const r = api.run(
      'game-of-life',
      { text: '.....\n..#..\n..#..\n..#..\n.....', options: { gens: 2 } },
      0
    );
    assert.ok(r.ok);
    if (!r.ok) return;
    const ends = generationEnds(r.value.events);
    assert.equal(ends.length, 3, 'the first board and two generations');
    const t = r.value.output.table!;
    const rep = new TableReplayer(t.cells.length, r.value.events);
    const sideways = rep.stateAt(ends[1]!).cells;
    assert.deepEqual(sideways.slice(10, 15), [0, 1, 1, 1, 0], 'the blinker turns sideways');
    assert.deepEqual(rep.stateAt(ends[2]!).cells, rep.stateAt(ends[0]!).cells, 'and back');
  });
});
