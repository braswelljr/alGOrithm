/**
 * The button fills are solid, and every one keeps its text readable: WCAG AA,
 * 4.5:1, in light and dark. The colors are the brand scales in src/lib/colors.ts
 * (Gopher Blue primary, chocolate-brown secondary) and the neutrals the ghost
 * fill sits on.
 */
import assert from 'node:assert/strict';
import { test } from 'node:test';
import { primary, secondary } from '../../src/lib/colors.ts';

const lin = (c: number) => {
  const v = c / 255;
  return v <= 0.04045 ? v / 12.92 : ((v + 0.055) / 1.055) ** 2.4;
};
const lum = (hex: string) => {
  const [r, g, b] = [1, 3, 5].map((i) => lin(parseInt(hex.slice(i, i + 2), 16)));
  return 0.2126 * r + 0.7152 * g + 0.0722 * b;
};
const ratio = (a: string, b: string) => {
  const [x, y] = [lum(a), lum(b)].sort((p, q) => q - p);
  return (x + 0.05) / (y + 0.05);
};

const ink = '#09090b';
const pairs: [string, string, string][] = [
  ['primary, light: ink on 500', ink, primary[500]],
  ['primary, dark: ink on 400', ink, primary[400]],
  ['secondary, light: 800 on 100', secondary[800], secondary[100]],
  ['secondary, light hover: 800 on 200', secondary[800], secondary[200]],
  ['secondary, dark: 50 on 800', secondary[50], secondary[800]],
  ['sidebar: white on secondary 950', '#ffffff', secondary[950]],
  ['sidebar, open page: ink on primary 500', ink, primary[500]],
  ['link, light: primary 700 on white', primary[700], '#ffffff'],
  ['link, dark: primary 300 on ink', primary[300], ink]
];

test('every button fill keeps its text AA in both modes', () => {
  const bad = pairs
    .map(([what, fg, bg]) => ({ what, r: ratio(fg, bg) }))
    .filter((p) => p.r < 4.5)
    .map((p) => `${p.what}: ${p.r.toFixed(2)}`);
  assert.deepEqual(bad, []);
});
