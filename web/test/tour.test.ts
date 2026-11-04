/**
 * The tours point at pages only through data-tour names
 * (docs/rules/a-page-with-a-tour-marks-its-parts-with-data-tour.md): every
 * tour file is in the index, every target is a [data-tour="..."] selector, and
 * every name is carried by some page.
 */
import assert from 'node:assert/strict';
import { readdirSync, readFileSync, statSync } from 'node:fs';
import { join } from 'node:path';
import { test } from 'node:test';
import { matchRoute } from '../src/lib/tour/define.ts';
import { TOURS } from '../src/lib/tour/tours/index.ts';

const src = new URL('../src/', import.meta.url).pathname;
const files = (dir: string): string[] =>
  readdirSync(dir).flatMap((f) => {
    const p = join(dir, f);
    return statSync(p).isDirectory() ? files(p) : /\.tsx?$/.test(f) ? [p] : [];
  });

test('every tour file is listed in the index', () => {
  const dir = join(src, 'lib/tour/tours');
  const listed = new Set(TOURS.map((t) => t.id));
  const index = readFileSync(join(dir, 'index.ts'), 'utf8');
  for (const f of readdirSync(dir).filter((f) => f !== 'index.ts')) {
    assert.ok(index.includes(`./${f.replace(/\.ts$/, '')}'`), `${f} is not in tours/index.ts`);
  }
  assert.equal(listed.size, TOURS.length, 'two tours share an id');
});

test('every target is a data-tour name a page carries', () => {
  const code = [...files(join(src, 'components')), ...files(join(src, 'visualizer'))]
    .map((f) => readFileSync(f, 'utf8'))
    .join('\n');
  const carried = new Set<string>();
  for (const m of code.matchAll(/(?:data-tour|\btour)=(?:"([a-z0-9-]+)"|\{'([a-z0-9-]+)'\})/g))
    carried.add(m[1] ?? m[2]);
  // The configure dialog's default name.
  carried.add('configure');
  for (const tour of TOURS) {
    for (const step of tour.steps) {
      const m = /^\[data-tour="([a-z0-9-]+)"\]$/.exec(step.target);
      assert.ok(m, `${tour.id}: "${step.target}" is not a [data-tour="..."] selector`);
      assert.ok(carried.has(m[1]), `${tour.id}: no page carries data-tour="${m[1]}"`);
    }
  }
});

test('routes match their pages, with and without the trailing slash', () => {
  assert.ok(matchRoute('/visualize', '/visualize/'));
  assert.ok(matchRoute('/docs/:id', '/docs/quick-sort/'));
  assert.ok(!matchRoute('/docs/:id', '/docs/'));
  assert.ok(matchRoute('/', '/'));
  assert.ok(!matchRoute('/', '/learn/'));
});
