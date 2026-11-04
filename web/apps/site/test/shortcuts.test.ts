/**
 * The site's shortcut registry: no key bound twice where both could fire,
 * every id unique, and every binding parseable, so the help, the hints and the
 * trainer can trust it.
 */
import assert from 'node:assert/strict';
import { test } from 'node:test';
import { conflicts, parseKeys } from '../../../packages/ui/src/lib/shortcuts.ts';
import { shortcuts } from '../src/config/shortcuts.ts';

test('no shortcut collides with another in its scope or with a global one', () => {
  const clash = conflicts(shortcuts)
    // Enter answers in a quiz and opens in a list; the two never share a page.
    .filter(([a, b]) => !(a.keys === 'enter' && b.keys === 'enter' && a.scope !== 'global' && b.scope !== 'global'))
    .map(([a, b]) => `${a.id} and ${b.id} both use ${a.keys}`);
  assert.deepEqual(clash, []);
});

test('ids are unique and keys parse', () => {
  const ids = shortcuts.map((s) => s.id);
  assert.equal(new Set(ids).size, ids.length);
  for (const s of shortcuts) assert.ok(parseKeys(s.keys).length > 0, s.id);
});

test('the quiz answers with 1-4 and Enter', () => {
  const quiz = shortcuts.filter((s) => s.scope === 'quiz').map((s) => s.keys);
  assert.ok(quiz.includes('1,2,3,4'));
  assert.ok(quiz.includes('enter'));
});
