import assert from 'node:assert/strict';
import { test } from 'node:test';
import { filterPalette, fuzzyScore } from '../src/algorithm/command-palette-filter.ts';
import {
  conflicts,
  formatKeys,
  isTypingTarget,
  matchesChord,
  parseKeys,
  speakKeys,
  type Shortcut
} from '../src/lib/shortcuts.ts';

test('keys parse into alternatives, sequences and chords', () => {
  assert.deepEqual(parseKeys('mod+k'), [[['mod', 'k']]]);
  assert.deepEqual(parseKeys('g>l'), [[['g'], ['l']]]);
  assert.deepEqual(parseKeys('1,2'), [[['1']], [['2']]]);
});

test('keys read per platform and aloud', () => {
  assert.equal(formatKeys('mod+k', 'apple'), '⌘K');
  assert.equal(formatKeys('mod+k', 'pc'), 'Ctrl+K');
  assert.equal(formatKeys('g>l', 'pc'), 'G then L');
  assert.equal(formatKeys('1,2,3,4', 'pc'), '1, 2, 3 or 4');
  assert.equal(speakKeys('mod+k'), 'Command or Control plus K');
});

test('typing in a field is never a shortcut', () => {
  const el = (tag: string, type?: string) => ({
    closest: (sel: string) =>
      sel.includes(tag) ? { getAttribute: (a: string) => (a === 'type' ? (type ?? null) : null) } : null
  });
  assert.equal(isTypingTarget(el('input') as unknown as EventTarget), true);
  assert.equal(isTypingTarget(el('textarea') as unknown as EventTarget), true);
  assert.equal(isTypingTarget(el('input', 'checkbox') as unknown as EventTarget), false);
  assert.equal(isTypingTarget(null), false);
});

test('chords match key events per platform', () => {
  const ev = (key: string, o: Partial<{ meta: boolean; ctrl: boolean; shift: boolean }> = {}) => ({
    key,
    metaKey: !!o.meta,
    ctrlKey: !!o.ctrl,
    altKey: false,
    shiftKey: !!o.shift
  });
  assert.ok(matchesChord(ev('k', { meta: true }), ['mod', 'k'], 'apple'));
  assert.ok(!matchesChord(ev('k', { meta: true }), ['mod', 'k'], 'pc'));
  assert.ok(matchesChord(ev('k', { ctrl: true }), ['mod', 'k'], 'pc'));
  assert.ok(matchesChord(ev('?', { shift: true }), ['shift', 'slash'], 'pc') || matchesChord(ev('/', { shift: true }), ['shift', 'slash'], 'pc'));
});

test('conflicts finds keys bound twice in one scope', () => {
  const s = (id: string, keys: string, scope: Shortcut['scope']): Shortcut => ({ id, keys, scope, description: id, category: 'x' });
  assert.equal(conflicts([s('a', 'r', 'player'), s('b', 'r', 'quiz')]).length, 0);
  assert.equal(conflicts([s('a', 'r', 'player'), s('b', 'r', 'player')]).length, 1);
  assert.equal(conflicts([s('a', 'shift+k', 'list'), s('b', 'k+shift', 'global')]).length, 1);
});

test('the palette matches fuzzily and ranks better matches first', () => {
  assert.ok(fuzzyScore('Quick Sort', 'qs') > 0);
  assert.equal(fuzzyScore('Quick Sort', 'xyz'), 0);
  assert.ok(fuzzyScore('Quick Sort', 'quick') > fuzzyScore('Quicksilver Sorting', 'qcks'));
  const groups = filterPalette(
    [
      {
        label: 'Entries',
        items: [
          { id: 'a', label: 'Merge Sort', onSelect: () => {} },
          { id: 'b', label: 'Quick Sort', keywords: ['partition'], onSelect: () => {} },
          { id: 'c', label: 'Dijkstra', onSelect: () => {} }
        ]
      }
    ],
    'partition'
  );
  assert.deepEqual(
    groups.flatMap((g) => g.items.map((i) => i.id)),
    ['b']
  );
});
