'use client';

/**
 * The command palette's record search (after gsl-special-courses' hook of the
 * same name). Where that one asks a server, this searches the site's own
 * index, /search.json, fetched once the first time the palette opens and kept
 * for the session. Entries, learning paths, categories and tags are ranked
 * with match-sorter, the best few of each kind shown with how many matched in
 * all.
 */
import { useCallback, useDeferredValue, useEffect, useMemo, useState } from 'react';
import { matchSorter, rankings } from 'match-sorter';
import { basePath } from '@/config/site';
import type { SearchKind } from '@/lib/command-palette/scope';
import { docsHref } from '@/lib/entries';
import { pathHref } from '@/lib/learn';
import {
  searchIndexSchema,
  searchIndexUrl,
  type SearchCategory,
  type SearchEntry,
  type SearchIndex,
  type SearchPath,
  type SearchTag
} from '@/lib/search-index';

/** Below this many characters records are not searched; nearly everything would match. */
export const MIN_SEARCH = 2;

/** How many rows of each kind are shown before "See all". */
const PER_KIND = 5;

/** One found record, with the kind it is so a row can be drawn for it. */
export type SearchRecord =
  | { kind: 'entries'; data: SearchEntry }
  | { kind: 'paths'; data: SearchPath }
  | { kind: 'categories'; data: SearchCategory }
  | { kind: 'tags'; data: SearchTag };

/** A found record as a palette row. */
export type RecordEntry = {
  id: string;
  kind: SearchKind;
  title: string;
  subtitle?: string;
  href?: string;
  record: SearchRecord;
};

/** One kind's rows, and where to see every match when there are more. */
export type RecordGroup = {
  kind: SearchKind;
  label: string;
  total: number;
  items: RecordEntry[];
  seeAll?: string;
};

/** What each kind is called, in the order the palette shows them. */
export const KIND_LABELS: Record<SearchKind, string> = {
  entries: 'Entries',
  paths: 'Learning paths',
  categories: 'Categories',
  tags: 'Tags'
};

/** Every kind the site can search: nothing here is behind an account. */
export const SEARCHABLE: readonly SearchKind[] = ['entries', 'paths', 'categories', 'tags'];

let loaded: Promise<SearchIndex> | undefined;

/** The index, fetched once per session; a failed fetch is forgotten so a retry asks again. */
function loadIndex(): Promise<SearchIndex> {
  loaded ??= fetch(searchIndexUrl(basePath))
    .then((r) => {
      if (!r.ok) throw new Error(`search index: ${r.status}`);
      return r.json();
    })
    .then((data: unknown) => {
      const parsed = searchIndexSchema.safeParse(data);
      if (!parsed.success) throw new Error('search index: unexpected shape');
      return parsed.data;
    })
    .catch((err: unknown) => {
      loaded = undefined;
      throw err;
    });
  return loaded;
}

/**
 * Fetches the index ahead of the first search: when the browser is idle after
 * a page loads, and when the pointer reaches the search button. A failure is
 * silent here; the palette asks again and says so.
 */
export function prefetchSearch(): void {
  void loadIndex().catch(() => {});
}

/**
 * Each record's searchable words, lowercased and joined once per index. Every
 * match at "contains" or better has the term somewhere in them, so a plain
 * substring test passes over the records that cannot match before match-sorter
 * ranks the rest.
 */
const haystacks = new WeakMap<object, string>();
function hay<T extends object>(row: T, keys: (keyof T)[]): string {
  let h = haystacks.get(row);
  if (h === undefined) {
    h = keys
      .map((k) => {
        const v = row[k];
        return Array.isArray(v) ? v.join(' ') : String(v ?? '');
      })
      .join('\u0000')
      .toLowerCase();
    haystacks.set(row, h);
  }
  return h;
}
const candidates = <T extends object>(rows: T[], term: string, keys: (keyof T)[]) => {
  const t = term.toLowerCase();
  return rows.filter((r) => hay(r, keys).includes(t));
};

const entryRow = (e: SearchEntry): RecordEntry => ({
  id: `entry:${e.id}`,
  kind: 'entries',
  title: e.name,
  subtitle: `${e.categoryName} · ${e.summary}`,
  href: docsHref(e.id),
  record: { kind: 'entries', data: e }
});
const pathRow = (p: SearchPath): RecordEntry => ({
  id: `path:${p.id}`,
  kind: 'paths',
  title: p.title,
  subtitle: p.summary,
  href: pathHref(p.id),
  record: { kind: 'paths', data: p }
});
const categoryRow = (c: SearchCategory): RecordEntry => ({
  id: `category:${c.id}`,
  kind: 'categories',
  title: c.name,
  subtitle: `${c.topic} · ${c.entries} entries`,
  href: `/docs/#${c.id}`,
  record: { kind: 'categories', data: c }
});
const tagRow = (t: SearchTag): RecordEntry => ({
  id: `tag:${t.id}`,
  kind: 'tags',
  title: `#${t.id}`,
  subtitle: t.meaning,
  href: `/docs/tags/#tag-${t.id}`,
  record: { kind: 'tags', data: t }
});

/** Rank one kind and cut it to the rows shown. Pure, so it can be tested. */
export function searchIndex(index: SearchIndex, term: string, kinds: readonly SearchKind[]) {
  const groups: RecordGroup[] = [];
  const push = (kind: SearchKind, rows: RecordEntry[], seeAll?: string) => {
    if (rows.length > 0)
      groups.push({
        kind,
        label: KIND_LABELS[kind],
        total: rows.length,
        items: rows.slice(0, PER_KIND),
        seeAll
      });
  };
  const opts = { threshold: rankings.CONTAINS } as const;
  if (kinds.includes('entries'))
    push(
      'entries',
      matchSorter(
        candidates(index.entries, term, ['name', 'id', 'categoryName', 'tags', 'summary']),
        term,
        {
          ...opts,
          keys: [
            'name',
            'id',
            'categoryName',
            { key: 'tags', threshold: rankings.EQUAL },
            'summary'
          ]
        }
      ).map(entryRow),
      '/docs/'
    );
  if (kinds.includes('paths'))
    push(
      'paths',
      matchSorter(candidates(index.paths, term, ['title', 'id', 'summary']), term, {
        ...opts,
        keys: ['title', 'id', 'summary']
      }).map(pathRow),
      '/learn/'
    );
  if (kinds.includes('categories'))
    push(
      'categories',
      matchSorter(candidates(index.categories, term, ['name', 'id', 'topic', 'summary']), term, {
        ...opts,
        keys: ['name', 'id', 'topic', 'summary']
      }).map(categoryRow),
      '/docs/'
    );
  if (kinds.includes('tags'))
    push(
      'tags',
      matchSorter(candidates(index.tags, term, ['id', 'meaning']), term, {
        ...opts,
        keys: ['id', 'meaning']
      }).map(tagRow),
      '/docs/tags/'
    );
  return groups;
}

/**
 * The records matching what was typed, for the kinds asked. Loads the index
 * when first asked for anything and reports whether it is still on its way.
 */
export function useCommandSearch(search: string, scopeKinds?: readonly SearchKind[]) {
  const [index, setIndex] = useState<SearchIndex | null>(null);
  const [error, setError] = useState(false);
  const [attempt, setAttempt] = useState(0);
  // The results follow the typing in a deferred render, so a keystroke never waits on them.
  const term = useDeferredValue(search.trim());
  const kinds = scopeKinds && scopeKinds.length > 0 ? scopeKinds : SEARCHABLE;

  useEffect(() => {
    let live = true;
    loadIndex().then(
      (i) => live && (setIndex(i), setError(false)),
      () => live && setError(true)
    );
    return () => {
      live = false;
    };
  }, [attempt]);

  const groups = useMemo(
    () => (index && term.length >= MIN_SEARCH ? searchIndex(index, term, kinds) : []),
    [index, term, kinds]
  );
  const retry = useCallback(() => setAttempt((n) => n + 1), []);

  return {
    groups,
    isLoading: !index && !error && term.length >= MIN_SEARCH,
    isError: error,
    retry,
    expected: term.length >= MIN_SEARCH ? kinds : [],
    searchable: SEARCHABLE,
    index
  };
}
