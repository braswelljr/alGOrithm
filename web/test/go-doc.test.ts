import assert from 'node:assert/strict';
import { test } from 'node:test';
import { parseGoDoc } from '../src/lib/go-doc.ts';

test('a Go doc comment becomes paragraphs, code, lists and headings', () => {
  const doc = [
    'Package sorting holds the sorting algorithms, one file each: bubble,',
    'cocktail shaker and comb.',
    '',
    'Every algorithm X has two functions:',
    '',
    '    sorted := sorting.Quick([]int{5, 3, 9, 1})',
    '',
    '    var log trace.Log',
    '    sorting.TraceQuick([]int{5, 3, 9, 1}, &log)',
    '',
    '# Stability',
    '',
    '  - insertion is stable',
    '  - quick is not,',
    '    it swaps across the pivot',
    '',
    'The end.'
  ].join('\n');
  assert.deepEqual(parseGoDoc(doc), [
    {
      kind: 'text',
      text: 'Package sorting holds the sorting algorithms, one file each: bubble, cocktail shaker and comb.'
    },
    { kind: 'text', text: 'Every algorithm X has two functions:' },
    {
      kind: 'code',
      code: 'sorted := sorting.Quick([]int{5, 3, 9, 1})\n\nvar log trace.Log\nsorting.TraceQuick([]int{5, 3, 9, 1}, &log)'
    },
    { kind: 'heading', text: 'Stability' },
    { kind: 'list', items: ['insertion is stable', 'quick is not, it swaps across the pivot'] },
    { kind: 'text', text: 'The end.' }
  ]);
});

test('tab-indented code keeps its own nesting', () => {
  assert.deepEqual(parseGoDoc('Use it:\n\n\tfor i := range s {\n\t\tf(i)\n\t}'), [
    { kind: 'text', text: 'Use it:' },
    { kind: 'code', code: 'for i := range s {\n\tf(i)\n}' }
  ]);
});
