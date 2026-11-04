import assert from 'node:assert/strict';
import { test } from 'node:test';
import { pathsChart } from '../src/lib/learn.ts';

test('the learning-path map has a node per path and an arrow per prerequisite', () => {
  const chart = pathsChart([
    { id: 'sorting-from-zero', title: 'Sorting from zero', prerequisites: [] },
    {
      id: 'graphs',
      title: 'Graphs "step" by step',
      prerequisites: ['sorting-from-zero', 'missing']
    }
  ]);
  assert.match(chart, /^flowchart LR/);
  assert.match(chart, /p_sorting_from_zero\["Sorting from zero"\]/);
  assert.match(chart, /p_graphs\["Graphs #quot;step#quot; by step"\]/);
  assert.match(chart, /p_sorting_from_zero --> p_graphs/);
  assert.doesNotMatch(chart, /missing/, 'an unknown prerequisite draws no arrow');
  assert.match(chart, /click p_graphs "\/learn\/graphs\/"/);
});
