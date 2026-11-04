import type { ReactNode } from 'react';
import { DocsLayout, type DocsNav } from '@/components/layouts/docs-layout';
import { buildCatalogSummary } from '@/lib/catalog.server';
import { categoriesWithEntries, categoryTitle } from '@/lib/entries';

/** The docs, in their own layout: the list of every entry beside the page. */
export default async function DocsRouteLayout({ children }: { children: ReactNode }) {
  const catalog = await buildCatalogSummary();
  const nav: DocsNav = [];
  for (const { category, entries } of categoriesWithEntries(catalog)) {
    let group = nav.find((g) => g.topic === category.topic);
    if (!group) nav.push((group = { topic: category.topic, categories: [] }));
    group.categories.push({
      id: category.id,
      name: categoryTitle(category),
      entries: entries.map((e) => ({ id: e.id, name: e.name }))
    });
  }
  return <DocsLayout nav={nav}>{children}</DocsLayout>;
}
