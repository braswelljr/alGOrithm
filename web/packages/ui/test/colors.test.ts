/**
 * The color system meets WCAG AA in every theme: each neutral with each
 * accent, light and dark, for every pair the site depends on, and every
 * visualization palette's marks against the card. And colors.css is current.
 */
import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import { test } from 'node:test';
import { fileURLToPath } from 'node:url';
import {
  accents,
  contrast,
  hex,
  neutrals,
  oklch,
  pairs,
  vizPairs,
  vizPalettes,
  type Mode
} from '../src/lib/colors.ts';
import { renderColorsCss } from '../scripts/render-colors.ts';

const modes: Mode[] = ['light', 'dark'];

test('the contrast math matches known values', () => {
  assert.equal(contrast(oklch(1, 0, 0), oklch(0, 0, 0)).toFixed(1), '21.0');
  assert.equal(hex(oklch(1, 0, 0)), '#ffffff');
  // #777777 on white is 4.48:1.
  assert.ok(Math.abs(contrast(oklch(0.5693, 0, 0), oklch(1, 0, 0)) - 4.48) < 0.05);
});

test('every token pair meets AA in every theme', () => {
  const bad: string[] = [];
  for (const n of neutrals)
    for (const a of accents)
      for (const m of modes)
        for (const p of pairs(n, a, m))
          if (p.ratio < p.min)
            bad.push(`${n.id}/${a.id}/${m}: ${p.fg} on ${p.bg} is ${p.ratio.toFixed(2)}, needs ${p.min} (${p.what})`);
  assert.deepEqual(bad, []);
});

test('every visualization mark stands out from the card', () => {
  const bad: string[] = [];
  for (const v of vizPalettes)
    for (const n of neutrals)
      for (const m of modes)
        for (const p of vizPairs(v, n, m))
          if (p.ratio < p.min) bad.push(`${v.id}/${n.id}/${m}: ${p.role} is ${p.ratio.toFixed(2)}, needs ${p.min}`);
  assert.deepEqual(bad, []);
});

test('colors.css is generated from colors.ts', async () => {
  const file = await readFile(fileURLToPath(new URL('../src/styles/colors.css', import.meta.url)), 'utf8');
  assert.equal(file, renderColorsCss(), 'stale: run `pnpm --dir web/packages/ui colors`');
});
