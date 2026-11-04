/**
 * The command palette's search (src/hooks/use-command-search.ts): a name that
 * starts with the words beats one that contains them, each kind keeps its best
 * few with the full count, and a scope searches only its kinds.
 */
import assert from 'node:assert/strict';
import { test } from 'node:test';
import { searchIndex } from '../src/hooks/use-command-search.ts';
import { openScope, SCOPES } from '../src/lib/command-palette/scope.ts';
import type { SearchIndex } from '../src/lib/search-index.ts';

const entry = (id: string, name: string, tags: string[] = []) => ({
  id,
  name,
  category: 'sorting',
  categoryName: 'Sorting',
  topic: 'Algorithms',
  summary: `${name} summary`,
  level: 'beginner',
  kind: 'algorithm',
  time: 'O(n log n)',
  space: 'O(1)',
  input: 'ints',
  tags
});

const index: SearchIndex = {
  entries: [
    entry('bubble-sort', 'Bubble Sort'),
    entry('quick-sort', 'Quick Sort', ['divide-and-conquer']),
    entry('quickselect', 'Quickselect'),
    ...Array.from({ length: 8 }, (_, i) => entry(`sort-${i}`, `Sort ${i}`))
  ],
  paths: [
    {
      id: 'sorting-from-zero',
      title: 'Sorting from zero',
      summary: 'Sorts',
      level: 'beginner',
      steps: 6
    }
  ],
  categories: [
    { id: 'sorting', name: 'Sorting', topic: 'Algorithms', summary: 'Order', entries: 11 }
  ],
  tags: [{ id: 'divide-and-conquer', meaning: 'splits the problem', entries: 1 }]
};

test('a name starting with the words ranks first', () => {
  const [entries] = searchIndex(index, 'quick', ['entries']);
  assert.deepEqual(
    entries.items.map((e) => e.record.data.id),
    ['quick-sort', 'quickselect']
  );
});

test('each kind shows its best five and counts every match', () => {
  const [entries] = searchIndex(index, 'sort', ['entries']);
  assert.equal(entries.items.length, 5);
  assert.ok(entries.total > 5);
  assert.equal(entries.seeAll, '/docs/');
});

test('a scope searches only its kinds', () => {
  const opened = openScope('#divide', SCOPES);
  assert.equal(opened?.scope.label, 'Tags');
  assert.equal(opened?.rest, 'divide');
  const groups = searchIndex(index, opened!.rest, opened!.scope.kinds);
  assert.deepEqual(
    groups.map((g) => g.kind),
    ['tags']
  );
});

test('every record opens somewhere', () => {
  for (const g of searchIndex(index, 'sort', ['entries', 'paths', 'categories', 'tags']))
    for (const r of g.items) assert.ok(r.href?.startsWith('/'), r.id);
});
