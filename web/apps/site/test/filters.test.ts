/** The catalog filters: kind and level choose one of, tags must all match. */
import assert from 'node:assert/strict';
import { test } from 'node:test';
import {
  facets,
  filterCatalog,
  isFiltered,
  matches,
  noFilters,
  words
} from '../src/lib/filters.ts';

const entry = (id: string, kind: string, level: string, tags: string[]) =>
  ({ id, kind, level, tags }) as never;
const catalog = {
  entries: [
    entry('quick', 'algorithm', 'intermediate', ['comparison', 'divide-and-conquer']),
    entry('bubble', 'algorithm', 'beginner', ['comparison', 'stable', 'in-place']),
    entry('stack', 'data-structure', 'beginner', ['lifo']),
    entry('race', 'simulation', 'advanced', ['concurrency', 'os'])
  ],
  tags: [{ tag: 'lifo', meaning: 'last in, first out', count: 1 }]
} as never;

test('kind and level match any chosen value; every chosen tag must be present', () => {
  const f = { kind: ['algorithm'], level: [], tag: ['comparison', 'stable'] };
  assert.deepEqual(
    filterCatalog(catalog, f).entries.map((e: { id: string }) => e.id),
    ['bubble']
  );
  assert.equal(matches(entry('x', 'algorithm', 'beginner', ['comparison']), f), false);
  assert.equal(filterCatalog(catalog, noFilters), catalog, 'no filters is the catalog itself');
  assert.equal(isFiltered(noFilters), false);
});

test('each chip counts what it would leave, levels in teaching order, tags with meanings', () => {
  const levels = facets(catalog, noFilters, 'level');
  assert.deepEqual(
    levels.map((l) => l.value),
    ['beginner', 'intermediate', 'advanced']
  );
  assert.equal(levels[0].count, 2);
  const withComparison = { ...noFilters, tag: ['comparison'] };
  const tags = facets(catalog, withComparison, 'tag');
  assert.equal(tags.find((t) => t.value === 'stable')?.count, 1, 'adding stable narrows to bubble');
  assert.equal(
    tags.find((t) => t.value === 'lifo')?.count,
    0,
    'lifo and comparison share no entry'
  );
  assert.equal(tags.find((t) => t.value === 'comparison')?.active, true);
  assert.equal(
    facets(catalog, noFilters, 'tag').find((t) => t.value === 'lifo')?.meaning,
    'last in, first out'
  );
  assert.equal(words('data-structure'), 'data structure');
});
