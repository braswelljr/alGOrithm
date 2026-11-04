/** The metadata and structured data every page describes itself with. */
import assert from 'node:assert/strict';
import { test } from 'node:test';
import { pageMetadata } from '../src/lib/metadata.ts';
import { absoluteUrl, author, siteUrl } from '../src/config/site.ts';
import { breadcrumbLd, siteLd } from '../src/lib/structured-data.ts';

test('a page has its own canonical and card', () => {
  const m = pageMetadata({ title: 'Compare', description: 'd', path: '/compare/' });
  assert.deepEqual(m.alternates, { canonical: '/compare/' });
  assert.equal((m.openGraph as { url: string }).url, '/compare/');
  assert.deepEqual((m.twitter as { images: string[] }).images, ['/og.png']);
});

test('absolute URLs are on the site, with one slash', () => {
  assert.equal(absoluteUrl('/docs/'), `${siteUrl}/docs/`);
  assert.equal(absoluteUrl('docs/'), `${siteUrl}/docs/`);
  assert.ok(!siteUrl.endsWith('/'));
});

test('the site credits braswelljr', () => {
  const graph = siteLd()['@graph'] as { '@type': string; url?: string; sameAs?: string[] }[];
  const person = graph.find((n) => n['@type'] === 'Person');
  assert.equal(person?.url, 'https://braswelljr.vercel.app');
  assert.deepEqual(person?.sameAs, ['https://github.com/braswelljr']);
  assert.equal(author.url, 'https://braswelljr.vercel.app');
});

test('breadcrumbs count from one and use absolute URLs', () => {
  const b = breadcrumbLd([
    { name: 'Home', path: '/' },
    { name: 'Docs', path: '/docs/' }
  ]) as { itemListElement: { position: number; item: string }[] };
  assert.deepEqual(
    b.itemListElement.map((i) => [i.position, i.item]),
    [
      [1, `${siteUrl}/`],
      [2, `${siteUrl}/docs/`]
    ]
  );
});
