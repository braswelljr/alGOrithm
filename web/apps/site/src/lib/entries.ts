/**
 * Where each entry opens, and names for categories. Everything else about an
 * entry (its name, cost, input) comes from the wasm catalog, never from here
 * (docs/rules/there-is-one-list-of-entries.md).
 */
import type { Catalog, Category, Entry } from '@braswelljr/algorithm-visualizer/core';

/** Whether an entry is a data structure operated by hand. */
export function isStructure(e: Pick<Entry, 'category' | 'input'>): boolean {
  return e.input === 'ops';
}

/** The page an entry opens in. */
export function entryHref(e: Pick<Entry, 'id' | 'category' | 'input'>): string {
  return isStructure(e) ? `/structures/?id=${e.id}` : `/visualize/?id=${e.id}`;
}

/** The documentation page of an entry. */
export function docsHref(id: string): string {
  return `/docs/${id}/`;
}

/** A readable title for a category, when the catalog gives only an id. */
export function categoryTitle(c: Category): string {
  return c.name && c.name !== c.id ? c.name : c.id.charAt(0).toUpperCase() + c.id.slice(1);
}

/** The categories that have entries, in the catalog's order. */
export function categoriesWithEntries(
  catalog: Catalog
): { category: Category; entries: Entry[] }[] {
  return catalog.categories
    .map((category) => ({
      category,
      entries: catalog.entries.filter((e) => e.category === category.id)
    }))
    .filter((g) => g.entries.length > 0);
}

/** The catalog by topic, then by category, both in the catalog's order. */
export function topicsWithCategories(catalog: Catalog) {
  const groups = categoriesWithEntries(catalog);
  const topics: { topic: string; groups: typeof groups }[] = [];
  for (const g of groups) {
    const t = topics.find((x) => x.topic === g.category.topic);
    if (t) t.groups.push(g);
    else topics.push({ topic: g.category.topic, groups: [g] });
  }
  return topics;
}
