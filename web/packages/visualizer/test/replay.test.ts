/** Replay, layout, diff, highlighting and saved runs, without the wasm. */
import assert from 'node:assert/strict';
import { test } from 'node:test';
import {
  ArrayReplayer,
  CHECKPOINT,
  diffSnapshots,
  disksFor,
  fallbackSnippet,
  GridReplayer,
  hanoiAt,
  highlightGo,
  layoutSnapshot,
  makeSavedRun,
  parseMoves,
  readSavedRun,
  TableReplayer,
  TextReplayer,
  type Snapshot,
  type TraceEvent
} from '../dist/core.js';

const ev = (
  kind: number,
  indices: number[] = [],
  values: number[] = [],
  note?: string
): TraceEvent => ({
  kind,
  indices,
  values,
  note
});

test('array replay matches a direct application at every step, across checkpoints', () => {
  const initial = Array.from({ length: 40 }, (_, i) => (i * 37) % 40);
  const events: TraceEvent[] = [];
  const arr = initial.slice();
  for (let i = 0; i < CHECKPOINT * 3 + 5; i++) {
    const a = (i * 7) % 40;
    const b = (i * 13 + 1) % 40;
    events.push(ev(0, [a, b]));
    if (i % 3 === 0) {
      events.push(ev(1, [a, b]));
    } else if (i % 5 === 0) {
      events.push(ev(2, [a], [i]));
    }
  }
  const r = new ArrayReplayer(initial, events);
  const expected = initial.slice();
  for (let k = 0; k <= events.length; k++) {
    if (k % 97 === 0 || k === events.length) {
      assert.deepEqual(r.stateAt(k).values, expected, `step ${k}`);
    }
    const e = events[k];
    if (!e) break;
    if (e.kind === 1)
      [expected[e.indices[0]], expected[e.indices[1]]] = [
        expected[e.indices[1]],
        expected[e.indices[0]]
      ];
    if (e.kind === 2) expected[e.indices[0]] = e.values[0];
  }
  void arr;
  assert.deepEqual(r.final(), expected);
});

test('array marks and step highlights', () => {
  const r = new ArrayReplayer(
    [3, 1, 2],
    [
      ev(4, [0, 2], [], 'range'),
      ev(0, [1]),
      ev(4, [1], [], 'found'),
      ev(4, [2], [], 'sorted'),
      ev(4, [0, 1], [], 'flip')
    ]
  );
  const s = r.stateAt(2);
  assert.deepEqual(s.range, [0, 2]);
  assert.deepEqual(s.compared, [1]);
  const end = r.stateAt(5);
  assert.equal(end.found, 1);
  assert.equal(end.sorted[2], true);
  assert.deepEqual(end.flip, [0, 1]);
  assert.deepEqual(end.compared, []);
  // Out-of-range events are ignored, not thrown.
  assert.doesNotThrow(() => new ArrayReplayer([1], [ev(1, [0, 9]), ev(2, [5], [1])]).final());
});

test('grid replay: frontier, visited, current, path', () => {
  const r = new GridReplayer(9, [
    ev(5, [0], [0]),
    ev(3, [0], [0]),
    ev(5, [1], [1]),
    ev(5, [3], [1]),
    ev(3, [1], [1]),
    ev(4, [0, 1, 2], [], 'path')
  ]);
  const mid = r.stateAt(4);
  assert.deepEqual([...mid.frontier].sort(), [1, 3]);
  assert.deepEqual([...mid.visited].sort(), [0]);
  assert.equal(mid.current, 0);
  const s5 = r.stateAt(5);
  assert.equal(s5.current, 1);
  assert.equal(s5.cost.get(1), 1);
  assert.deepEqual(r.stateAt(6).path, [0, 1, 2]);
});

test('table replay: fills, reads, path', () => {
  const r = new TableReplayer(4, [
    ev(2, [0], [0]),
    ev(0, [0]),
    ev(2, [1], [1]),
    ev(4, [1, 0], [], 'path')
  ]);
  assert.deepEqual(r.stateAt(1).cells, [0, null, null, null]);
  assert.equal(r.stateAt(3).current, 1);
  assert.deepEqual(r.stateAt(2).reads, [0]);
  assert.deepEqual(r.stateAt(4).path, [1, 0]);
});

test('text replay: compare, window, found', () => {
  const r = new TextReplayer(6, [
    ev(3, [0, 1]),
    ev(0, [0], [0]),
    ev(4, [0, 1], [], 'found'),
    ev(4, [3, 4], [], 'found')
  ]);
  assert.deepEqual(r.stateAt(1).window, [0, 1]);
  assert.deepEqual(r.stateAt(2).compare, { text: 0, pattern: 0 });
  assert.deepEqual(r.stateAt(4).found, [
    [0, 1],
    [3, 4]
  ]);
});

test('hanoi moves replay', () => {
  const moves = parseMoves(['A -> C', 'A -> B', 'C -> B', 'bad line']);
  assert.equal(moves.length, 3);
  assert.equal(disksFor(3), 2);
  const s = hanoiAt(2, moves, 3);
  assert.deepEqual(s.pegs, [[], [2, 1], []]);
});

test('layout: binary tree in order, forest side by side, rows and sets', () => {
  const tree: Snapshot = {
    layout: 'tree',
    nodes: [
      { id: 0, label: '8', parent: -1 },
      { id: 1, label: '3', parent: 0 },
      { id: 2, label: '10', parent: 0 }
    ],
    edges: [
      { from: 0, to: 1, label: 'left' },
      { from: 0, to: 2, label: 'right' }
    ]
  };
  const p = layoutSnapshot(tree);
  assert.ok(p.get(1)!.x < p.get(0)!.x && p.get(0)!.x < p.get(2)!.x);
  assert.ok(p.get(1)!.y > p.get(0)!.y);
  const lone: Snapshot = {
    layout: 'tree',
    nodes: [
      { id: 0, label: '5', parent: -1 },
      { id: 1, label: '2', parent: 0 }
    ],
    edges: [{ from: 0, to: 1, label: 'left' }]
  };
  const q = layoutSnapshot(lone);
  assert.ok(q.get(1)!.x < q.get(0)!.x, 'a lone left child sits to the left');
  const forest: Snapshot = {
    layout: 'tree',
    nodes: [
      { id: 0, label: 'a', parent: -1, role: 'root' },
      { id: 1, label: 'b', parent: 0 },
      { id: 2, label: 'c', parent: -1, role: 'root' }
    ],
    edges: [{ from: 0, to: 1 }]
  };
  const f = layoutSnapshot(forest);
  assert.equal(f.size, 3);
  assert.ok(f.get(2)!.x > f.get(0)!.x);
  const row = layoutSnapshot({
    layout: 'row',
    nodes: [
      { id: 0, label: 'x', parent: -1 },
      { id: 1, label: 'y', parent: -1 }
    ],
    edges: []
  });
  assert.equal(row.get(0)!.y, row.get(1)!.y);
  const set = layoutSnapshot({
    layout: 'set',
    nodes: Array.from({ length: 10 }, (_, i) => ({ id: i, label: String(i), parent: -1 })),
    edges: []
  });
  assert.ok(set.get(9)!.y > set.get(0)!.y);
});

test('snapshot diff finds the added node and link', () => {
  const a: Snapshot = { layout: 'row', nodes: [{ id: 0, label: '1', parent: -1 }], edges: [] };
  const b: Snapshot = {
    layout: 'row',
    nodes: [
      { id: 0, label: '1', parent: -1 },
      { id: 1, label: '2', parent: -1 }
    ],
    edges: [{ from: 0, to: 1, label: 'next' }]
  };
  const d = diffSnapshots(a, b);
  assert.deepEqual([...d.added], [1]);
  assert.deepEqual([...d.newEdges], ['0-1']);
  assert.deepEqual(diffSnapshots(b, a).removed, ['2']);
  assert.equal(diffSnapshots(undefined, b).added.size, 0);
});

test('Go highlighting keeps every character and finds the kinds', () => {
  const code =
    'package sorting\n\n// Quick sorts.\nfunc Quick(s []int) []int {\n\tx := "a//b" + `raw\nline` // c\n\treturn s[0x1F:]\n}\n';
  const lines = highlightGo(code);
  assert.equal(lines.map((l) => l.map((t) => t.text).join('')).join('\n'), code);
  const kinds = new Set(lines.flat().map((t) => t.kind));
  for (const k of ['keyword', 'comment', 'string', 'type', 'func', 'number'])
    assert.ok(kinds.has(k as never), k);
});

test('saved runs round-trip and reject junk', () => {
  const run = makeSavedRun('quick-sort', { values: [3, 1] });
  const back = readSavedRun(JSON.stringify(run));
  assert.ok(back.ok);
  assert.equal(back.ok && back.run.entry, 'quick-sort');
  assert.equal(readSavedRun('nope').ok, false);
  assert.equal(readSavedRun('{"format":"other"}').ok, false);
  const snip = fallbackSnippet(
    { id: 'quick-sort', import: 'x/sorting', package: 'sorting', func: 'Quick' },
    { values: [3, 1], target: 2 }
  );
  assert.equal(snip.cli, 'algorithm run quick-sort 3 1 --target 2');
});

test('timelines play as time passes', async () => {
  const { lanesOf, spansUntil, timelineEnd, describeTime } = await import('../dist/core.js');
  const spans = [
    { lane: 'P1', label: 'P1', start: 0, end: 3, state: 'running' },
    { lane: 'idle', label: '', start: 3, end: 4, state: 'idle' },
    { lane: 'P2', label: 'P2', start: 4, end: 6 }
  ];
  assert.deepEqual(lanesOf(spans), ['P1', 'P2', 'idle']);
  assert.equal(timelineEnd(spans), 6);
  assert.deepEqual(
    spansUntil(spans, 2).map((s) => [s.lane, s.end]),
    [['P1', 2]]
  );
  assert.match(describeTime(spans, 4), /P2 running/);
});

test('table marks: faults and hits land on their cells', () => {
  const r = new TableReplayer(4, [ev(2, [0], [7]), ev(4, [0], [], 'fault'), ev(4, [1], [], 'hit')]);
  const s = r.stateAt(3);
  assert.equal(s.marks.get(0), 'fault');
  assert.equal(s.marks.get(1), 'hit');
});
