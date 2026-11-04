/**
 * Use the component library (docs/rules/use-the-component-library.md): a card
 * is a Card, a heading a Heading, a button a Button, a field an Input or Field.
 * Every component is built on src/components/ui. This test scans the rest of
 * src/ (the primitives themselves and the visualizer's views are exempt) for
 * the hand-rolled versions and names each by file and line. An exception
 * takes a reason on the same or the previous line: `card-ok: why` or
 * `ui-ok: why`.
 */
import assert from 'node:assert/strict';
import { readdir, readFile } from 'node:fs/promises';
import { join, relative } from 'node:path';
import { test } from 'node:test';
import { fileURLToPath } from 'node:url';

const web = fileURLToPath(new URL('..', import.meta.url));
const site = join(web, 'src');
/** The primitives may render raw elements: they are what everything else uses.
 *  The visualizer's views draw shapes, not controls. */
const exempt = [
  join(site, 'components/ui'),
  join(site, 'components/data-table'),
  join(site, 'visualizer')
];

async function tsx(dir: string): Promise<string[]> {
  const out: string[] = [];
  for (const d of await readdir(dir, { withFileTypes: true })) {
    const p = join(dir, d.name);
    if (exempt.includes(p)) continue;
    if (d.isDirectory()) out.push(...(await tsx(p)));
    else if (d.name.endsWith('.tsx')) out.push(p);
  }
  return out;
}

const allowed = (lines: string[], i: number, tag: string) =>
  lines[i].includes(`${tag}:`) || (i > 0 && lines[i - 1].includes(`${tag}:`));

/** A class string that draws a card: rounded, outlined, and filled or shadowed. */
export function looksLikeCard(classes: string): boolean {
  const c = ` ${classes} `;
  const rounded = /\srounded(-[a-z0-9]+)?\s/.test(c);
  const outlined = /\s(border|ring-1)\s/.test(c);
  const surface = /\s(bg-card|bg-bg|shadow(-[a-z]+)?)\s/.test(c);
  return rounded && outlined && surface;
}

test('looksLikeCard spots a card and lets a control be', () => {
  assert.ok(looksLikeCard('rounded-xl border border-border bg-card p-4'));
  assert.ok(looksLikeCard('rounded-lg ring-1 ring-border shadow-sm'));
  assert.ok(!looksLikeCard('rounded-md px-2 py-1 hover:bg-subtle'));
  assert.ok(!looksLikeCard('rounded-full border border-border text-xs'));
});

test('cards are Cards', async () => {
  const bad: string[] = [];
  for (const f of await tsx(site)) {
    const lines = (await readFile(f, 'utf8')).split('\n');
    lines.forEach((line, i) => {
      for (const m of line.matchAll(/(['"`])([^'"`]*?)\1/g)) {
        if (looksLikeCard(m[2]) && !allowed(lines, i, 'card-ok')) {
          bad.push(`${relative(web, f)}:${i + 1}: use <Card> instead of "${m[2].slice(0, 60)}"`);
        }
      }
    });
  }
  assert.deepEqual(bad, [], `hand-drawn cards:\n${bad.join('\n')}`);
});

const raw: [RegExp, string][] = [
  [/<button(?=[\s>]|$)/, 'Button'],
  [/<input(?=[\s>/]|$)/, 'Input, Field or a control from the library'],
  [/<select(?=[\s>]|$)/, 'Select or NativeSelect'],
  [/<textarea(?=[\s>]|$)/, 'Textarea'],
  [/<h[1-6](?=[\s>]|$)/, 'Heading']
];

test('controls and headings come from the library', async () => {
  const bad: string[] = [];
  for (const f of await tsx(site)) {
    const lines = (await readFile(f, 'utf8')).split('\n');
    lines.forEach((line, i) => {
      for (const [re, use] of raw) {
        if (re.test(line) && !allowed(lines, i, 'ui-ok'))
          bad.push(`${relative(web, f)}:${i + 1}: use ${use}`);
      }
    });
  }
  assert.deepEqual(bad, [], `hand-rolled markup:\n${bad.join('\n')}`);
});

test('outline is rare and one primary leads each view', async () => {
  const bad: string[] = [];
  for (const f of await tsx(site)) {
    const text = await readFile(f, 'utf8');
    const lines = text.split('\n');
    lines.forEach((line, i) => {
      if (/variant(=|: )["']outline["']/.test(line) && !allowed(lines, i, 'ui-ok'))
        bad.push(
          `${relative(web, f)}:${i + 1}: use ghost (or primary for the main action), or say why with ui-ok`
        );
    });
    const primaries = (text.match(/variant=["']primary["']/g) ?? []).length;
    if (primaries > 1 && !text.includes('ui-ok: primaries'))
      console.warn(`${relative(web, f)}: ${primaries} primary buttons; a view has one main action`);
  }
  assert.deepEqual(bad, [], `outline variants:\n${bad.join('\n')}`);
});
