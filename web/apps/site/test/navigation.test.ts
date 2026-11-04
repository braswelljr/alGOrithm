import assert from 'node:assert/strict';
import { readdirSync, statSync } from 'node:fs';
import { join } from 'node:path';
import { test } from 'node:test';
import { buildNavigation, isActivePath, NAVIGATION } from '../src/config/navigation.ts';
import { crumbsFor } from '../src/lib/breadcrumbs.ts';

test('every page in the menu exists, and every top-level page is in the menu', () => {
  const app = new URL('../src/app/', import.meta.url).pathname;
  const pages = readdirSync(app)
    .filter((d) => statSync(join(app, d)).isDirectory())
    .filter((d) => statSync(join(app, d, 'page.tsx'), { throwIfNoEntry: false }))
    .map((d) => `/${d}/`);
  const hrefs = NAVIGATION.flatMap((s) => s.items.map((i) => i.href));
  assert.deepEqual(new Set(hrefs).size, hrefs.length, 'a page is listed twice');
  for (const page of pages) assert.ok(hrefs.includes(page), `${page} is not in the menu`);
  for (const href of hrefs.filter((h) => h !== '/'))
    assert.ok(pages.includes(href), `${href} has no page`);
});

test('exactly one entry is active, the catalog only at the root', () => {
  const active = (p: string) =>
    buildNavigation(p)
      .flatMap((g) => g.items)
      .filter((i) => i.isActive)
      .map((i) => i.href);
  assert.deepEqual(active('/'), ['/']);
  assert.deepEqual(active('/docs/quick-sort/'), ['/docs/']);
  assert.deepEqual(active('/learn'), ['/learn/']);
  assert.equal(isActivePath('/', '/docs/'), false);
});

test('the trail names each step and links all but the last', () => {
  assert.deepEqual(crumbsFor('/'), []);
  assert.deepEqual(crumbsFor('/docs/'), []);
  assert.deepEqual(crumbsFor('/docs/tags/'), [
    { label: 'Docs', href: '/docs/' },
    { label: 'Tags' }
  ]);
  assert.deepEqual(crumbsFor('/docs/quick-sort'), [
    { label: 'Docs', href: '/docs/' },
    { label: 'Quick sort' }
  ]);
});
