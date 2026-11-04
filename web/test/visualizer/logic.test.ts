import assert from 'node:assert/strict';
import { test } from 'node:test';
import {
  circuitAt,
  describeCircuit,
  describeLogic,
  groupsShown,
  isWaveform,
  logicSteps,
  toggledInput,
  wavesOf
} from '../../src/visualizer/core/replay/circuit.ts';
import { outputSchema } from '../../src/visualizer/core/schema.ts';

// A half adder as the Go registry sends it: inputs A, B; gates x (xor), a (and); outputs S, C.
const halfAdder = outputSchema.parse({
  circuit: {
    name: 'half adder',
    gates: [
      { name: 'A', kind: 'input', in: null, layer: 0, row: 0 },
      { name: 'B', kind: 'input', in: null, layer: 0, row: 1 },
      { name: 'x', kind: 'xor', in: [0, 1], layer: 1, row: 0 },
      { name: 'a', kind: 'and', in: [0, 1], layer: 1, row: 1 },
      { name: 'S', kind: 'output', in: [2], layer: 2, row: 0 },
      { name: 'C', kind: 'output', in: [3], layer: 2, row: 1 }
    ],
    frames: [
      [1, 1, -1, -1, -1, -1],
      [1, 1, 0, 1, -1, -1],
      [1, 1, 0, 1, 0, 1]
    ],
    phases: null
  }
}).circuit!;

test('circuit replay: levels, changes and words per gate delay', () => {
  const s0 = circuitAt(halfAdder, 0);
  assert.equal(s0.delay, 0);
  assert.match(describeCircuit(halfAdder, s0), /inputs are applied \(A=1 B=1\)/);
  const s1 = circuitAt(halfAdder, 1);
  assert.deepEqual([...s1.changed].sort(), [2, 3]);
  assert.match(describeCircuit(halfAdder, s1), /x \(xor\) became 0; a \(and\) became 1/);
  const s9 = circuitAt(halfAdder, 9); // clamped to the last frame
  assert.equal(s9.frame, 2);
  assert.match(describeCircuit(halfAdder, s9), /Outputs: S=0 C=1/);
});

test('flipping an input edits the run input the way the entry reads it', () => {
  const levels = circuitAt(halfAdder, 2).levels;
  // One bit per input.
  assert.deepEqual(toggledInput(halfAdder, { values: [1, 1] }, 0, levels)?.values, [0, 1]);
  // An expression: the key.
  assert.equal(toggledInput(halfAdder, { text: 'A ^ B' }, 1, levels)?.key, 'A=1 B=0');
  // Not an input gate: nothing.
  assert.equal(toggledInput(halfAdder, { values: [1, 1] }, 2, levels), undefined);
  // A bus: A0, A1, B0, B1 as two numbers.
  const adder = {
    ...halfAdder,
    gates: halfAdder.gates.map((g, i) => ({ ...g, name: ['A0', 'A1', 'x', 'a', 'S', 'C'][i] }))
  };
  assert.deepEqual(toggledInput(adder, { values: [2, 3] }, 1, levels)?.values, [0, 3]);
  // A sequence of inputs cannot be flipped from one frame.
  const seq = {
    ...halfAdder,
    phases: [
      { frame: 0, label: 'S=1' },
      { frame: 2, label: 'S=0' }
    ]
  };
  assert.equal(toggledInput(seq, { text: '10 00' }, 0, levels), undefined);
});

test('truth tables step row by row, then Karnaugh groups', () => {
  const t = outputSchema.parse({
    logic: {
      vars: ['A', 'B'],
      out: [0, 1, 1, 0],
      kmap: {
        rows: 2,
        cols: 2,
        rowVars: 'A',
        colVars: 'B',
        rowLabels: ['0', '1'],
        colLabels: ['0', '1'],
        cells: [0, 1, 1, 0],
        minterm: [0, 1, 2, 3],
        groups: [[1], [2]],
        terms: ['!A & B', 'A & !B'],
        sop: '!A & B | A & !B'
      }
    }
  }).logic!;
  assert.equal(logicSteps(t), 6);
  assert.match(describeLogic(t, 2), /Row 1: A=0 B=1 gives 1/);
  assert.equal(groupsShown(t.kmap, 4, 4), 0);
  assert.equal(groupsShown(t.kmap, 4, 6), 2);
  assert.match(describeLogic(t, 6), /Together: !A & B \| A & !B/);
});

test('timing diagrams are rebuilt per tick from high and low spans', () => {
  const spans = outputSchema.parse({
    timeline: [
      { lane: 'CLK', label: '0', start: 0, end: 1, state: 'low' },
      { lane: 'CLK', label: '1', start: 1, end: 2, state: 'high' },
      { lane: 'Q', label: '0', start: 0, end: 1, state: 'low' },
      { lane: 'Q', label: '1', start: 1, end: 2, state: 'high' }
    ]
  }).timeline;
  assert.ok(isWaveform(spans));
  assert.ok(!isWaveform([{ lane: 'P1', label: 'P1', start: 0, end: 3, state: 'running' }]));
  const { waves, ticks } = wavesOf(spans);
  assert.equal(ticks, 2);
  assert.deepEqual(
    waves.map((w) => w.name),
    ['CLK', 'Q']
  );
  assert.deepEqual(waves[1].levels, [false, true]);
});

test('bit rows parse with roles, nil fields as empty', () => {
  const o = outputSchema.parse({
    bits: [
      { label: 'single', fields: [{ name: 'sign', bits: '0', role: 'sign' }] },
      { label: 'does not fit', fields: null, note: 'too big' }
    ]
  });
  assert.equal(o.bits[0].fields[0].role, 'sign');
  assert.deepEqual(o.bits[1].fields, []);
});
