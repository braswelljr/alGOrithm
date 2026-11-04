/**
 * Filtering the catalog by kind, level and tags. Within kind and level any
 * chosen value matches (they are exclusive per entry); every chosen tag must
 * be present (AND), so adding a chip only ever narrows the list.
 */
import type { Catalog, Entry } from '@/visualizer/core';

export interface Filters {
  kind: string[];
  level: string[];
  tag: string[];
}

export const noFilters: Filters = { kind: [], level: [], tag: [] };

export const isFiltered = (f: Filters) => f.kind.length + f.level.length + f.tag.length > 0;

export function matches(e: Entry, f: Filters): boolean {
  return (
    (f.kind.length < 1 || f.kind.includes(e.kind)) &&
    (f.level.length < 1 || f.level.includes(e.level)) &&
    f.tag.every((t) => e.tags.includes(t))
  );
}

/** The catalog with only the matching entries. */
export function filterCatalog(catalog: Catalog, f: Filters): Catalog {
  return isFiltered(f)
    ? { ...catalog, entries: catalog.entries.filter((e) => matches(e, f)) }
    : catalog;
}

/** The levels in teaching order, then any other value. */
const levelOrder = ['beginner', 'intermediate', 'advanced'];

export interface Facet {
  value: string;
  /** Entries that would match if this value were toggled on (or, when it is on, that match now). */
  count: number;
  active: boolean;
  meaning?: string;
}

/** The choices for one filter, each with how many entries it would leave. */
export function facets(catalog: Catalog, f: Filters, key: keyof Filters): Facet[] {
  const values = new Set<string>();
  for (const e of catalog.entries) {
    if (key === 'tag') e.tags.forEach((t) => values.add(t));
    else if (e[key]) values.add(e[key]);
  }
  const meaning = new Map(catalog.tags.map((t) => [t.tag, t.meaning]));
  const sorted = [...values].sort((a, b) =>
    key === 'level'
      ? (levelOrder.indexOf(a) + 1 || 99) - (levelOrder.indexOf(b) + 1 || 99)
      : a.localeCompare(b)
  );
  return sorted.map((value) => {
    const active = f[key].includes(value);
    const next: Filters = {
      ...f,
      [key]: active ? f[key] : key === 'tag' ? [...f.tag, value] : [value]
    };
    return {
      value,
      active,
      count: catalog.entries.filter((e) => matches(e, next)).length,
      meaning: key === 'tag' ? meaning.get(value) : undefined
    };
  });
}

/** A kind or level as words: "data-structure" reads "data structure". */
export const words = (s: string) => s.replaceAll('-', ' ');
