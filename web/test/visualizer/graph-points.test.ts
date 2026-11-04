import assert from 'node:assert/strict';
import { test } from 'node:test';
import {
  describeGraphEvent,
  describePointsEvent,
  GraphReplayer,
  layers,
  layoutGraph,
  outputSchema,
  PointsReplayer
} from '../../src/visualizer/core/index.ts';

const ev = (kind: number, indices: number[], values: number[] = [], note?: string) => ({
  kind,
  indices,
  values,
  note
});
const [COMPARE, , SET, VISIT, MARK, PUSH] = [0, 1, 2, 3, 4, 5];

test('graph replay: frontier, visits, distances, path, order, groups, tree, flow, cycle', () => {
  const events = [
    ev(PUSH, [0]),
    ev(VISIT, [0]),
    ev(SET, [1], [4]),
    ev(PUSH, [1]),
    ev(VISIT, [1]),
    ev(MARK, [0, 1], [], 'path'),
    ev(MARK, [2], [], 'order'),
    ev(MARK, [0], [], 'order'),
    ev(MARK, [0, 2], [1], 'group'),
    ev(MARK, [0, 9], [], 'tree'),
    ev(MARK, [1], [3], 'flow'),
    ev(MARK, [0, 1, 2], [], 'cycle'),
    ev(VISIT, [99])
  ];
  const r = new GraphReplayer(3, 2, events);
  const s = r.stateAt(r.length);
  assert.deepEqual([...s.visited], [0, 1]);
  assert.equal(s.frontier.size, 0);
  assert.equal(s.distance.get(1), 4);
  assert.deepEqual(s.path, [0, 1]);
  assert.deepEqual(
    [...s.order.entries()],
    [
      [2, 1],
      [0, 2]
    ]
  );
  assert.equal(s.group.get(2), 1);
  assert.deepEqual([...s.tree], [0], 'an edge index out of range is ignored');
  assert.equal(s.flow.get(1), 3);
  assert.equal(s.cycle.size, 3);
  assert.equal(r.stateAt(1).frontier.has(0), true, 'state at a step replays only that far');
  assert.match(
    describeGraphEvent(events[2], ['A', 'B', 'C'], [{ from: 0, to: 1 }]),
    /distance to B is now 4/
  );
});

test('graph layout: a DAG in layers, anything else on a circle', () => {
  const dag = {
    directed: true,
    nodes: ['a', 'b', 'c'],
    edges: [
      { from: 0, to: 1, weight: 1 },
      { from: 1, to: 2, weight: 1 }
    ]
  };
  assert.deepEqual(layers(dag), [0, 1, 2]);
  const ys = layoutGraph(dag).map((p) => p.y);
  assert.ok(ys[0] < ys[1] && ys[1] < ys[2]);
  const cyc = { ...dag, edges: [...dag.edges, { from: 2, to: 0, weight: 1 }] };
  assert.equal(layers(cyc), null);
  assert.equal(layoutGraph({ ...dag, directed: false }).length, 3);
});

test('points replay: compared, visited, hull, pair, inside', () => {
  const events = [
    ev(COMPARE, [0, 1]),
    ev(VISIT, [2]),
    ev(MARK, [0, 2], [], 'hull'),
    ev(MARK, [0, 2, 1], [], 'hull'),
    ev(MARK, [1, 2], [], 'pair'),
    ev(MARK, [3], [], 'inside')
  ];
  const r = new PointsReplayer(4, events);
  const s = r.stateAt(r.length);
  assert.deepEqual(s.compared, [0, 1]);
  assert.deepEqual(s.hull, [0, 2, 1], 'the hull grows as it is marked');
  assert.deepEqual(s.pair, [1, 2]);
  assert.ok(s.inside.has(3));
  assert.deepEqual(r.stateAt(3).hull, [0, 2]);
  assert.match(
    describePointsEvent(events[4], [
      { x: 0, y: 0 },
      { x: 1, y: 1 },
      { x: 2, y: 2 },
      { x: 3, y: 3 }
    ]),
    /closest pair/
  );
});

test('output schema: graph and points parse; a bad edge is refused', () => {
  const ok = outputSchema.parse({
    graph: { directed: true, nodes: ['A', 'B'], edges: [{ from: 0, to: 1, weight: 5 }] },
    points: [{ x: 1, y: 2 }]
  });
  assert.equal(ok.graph?.edges[0].weight, 5);
  assert.equal(ok.points.length, 1);
  assert.deepEqual(outputSchema.parse({ points: null }).points, []);
  assert.equal(
    outputSchema.safeParse({ graph: { nodes: ['A'], edges: [{ from: 0, to: 3, weight: 1 }] } })
      .success,
    false
  );
});

test('a nodes-by-nodes table is a matrix: the table animates, not the graph', async () => {
  const { runKind } = await import('../../src/visualizer/index.ts');
  const base = { id: 'x', events: [{ kind: 2, indices: [3], values: [1] }], nanos: 0 };
  const graph = { directed: false, nodes: ['A', 'B'], edges: [{ from: 0, to: 1, weight: 1 }] };
  const table = { rows: 2, cols: 2, cells: [0, 1, 1, 0], rowLabels: [], colLabels: [] };
  const out = (o: object) => ({ ...base, output: outputSchema.parse(o) });
  assert.equal(runKind(out({ graph, table })), 'table');
  assert.equal(runKind(out({ graph })), 'graph');
  assert.equal(runKind(out({ table: { ...table, rows: 1, cols: 4 } }), 'backtracking'), 'table');
  assert.equal(runKind(out({ table }), 'backtracking'), 'board');
});

test('an event-driven timeline reveals chosen items and outlines the compared one', async () => {
  const { chosenAt } = await import('../../src/visualizer/core/index.ts');
  const events = [
    ev(COMPARE, [0]),
    ev(MARK, [0], [], 'chosen'),
    ev(COMPARE, [1]),
    ev(COMPARE, [2]),
    ev(MARK, [2], [], 'chosen')
  ];
  assert.deepEqual([...chosenAt(events, 2).chosen], [0]);
  const s = chosenAt(events, 4);
  assert.equal(s.compared, 2);
  assert.deepEqual([...chosenAt(events, 5).chosen], [0, 2]);
  assert.equal(chosenAt(events, 0).compared, null);
});

test('grid labels parse and must match the cells', () => {
  const g = { rows: 1, cols: 2, cost: [1, 1], start: 0, end: 1 };
  assert.deepEqual(outputSchema.parse({ grid: { ...g, labels: ['C', 'A'] } }).grid?.labels, [
    'C',
    'A'
  ]);
  assert.deepEqual(outputSchema.parse({ grid: g }).grid?.labels, []);
  assert.equal(outputSchema.safeParse({ grid: { ...g, labels: ['C'] } }).success, false);
});

test("Kruskal's considered edge lasts one step; segments and choices parse", () => {
  const r = new GraphReplayer(3, 2, [
    ev(MARK, [1], [], 'edge'),
    ev(MARK, [1], [], 'tree'),
    ev(VISIT, [0])
  ]);
  assert.equal(r.stateAt(1).considering, 1);
  assert.equal(r.stateAt(2).considering, null, 'the next event clears it');
  assert.ok(r.stateAt(2).tree.has(1));
  assert.deepEqual(
    outputSchema.parse({
      segments: [
        [0, 1],
        [2, 3]
      ]
    }).segments,
    [
      [0, 1],
      [2, 3]
    ]
  );
  assert.deepEqual(outputSchema.parse({}).segments, []);
});
