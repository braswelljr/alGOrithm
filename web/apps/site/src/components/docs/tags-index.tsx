'use client';

import Link from 'next/link';
import { Heading } from '@braswelljr/algorithm-ui/components/heading';
import type { Catalog } from '@braswelljr/algorithm-visualizer/core';
import { docsHref } from '@/lib/entries';

/** Every tag of the catalog: the vocabulary's meanings, or the tags the entries carry. */
function tagsOf(catalog: Catalog) {
  const used = new Map<string, string[]>();
  for (const e of catalog.entries)
    for (const t of e.tags) used.set(t, [...(used.get(t) ?? []), e.id]);
  const meaning = new Map(catalog.tags.map((t) => [t.tag, t.meaning]));
  const names = [...new Set([...catalog.tags.map((t) => t.tag), ...used.keys()])].sort();
  return names.map((tag) => ({ tag, meaning: meaning.get(tag) ?? '', ids: used.get(tag) ?? [] }));
}

/**
 * The tags page: each tag, what it means, and every entry that carries it,
 * linking to its documentation and to the catalog filtered by that tag.
 */
export function TagsIndex({ catalog }: { catalog: Catalog }) {
  return (
    <>
      {(() => {
        const name = new Map(catalog.entries.map((e) => [e.id, e.name]));
        const tags = tagsOf(catalog);
        return (
          <div className="flex flex-col gap-8">
            <nav aria-label="Tags">
              <ul className="flex flex-wrap gap-2">
                {tags.map((t) => (
                  <li key={t.tag}>
                    <a
                      href={`#tag-${t.tag}`}
                      className="rounded-full border border-border px-3 py-1 font-mono text-xs hover:border-accent"
                    >
                      #{t.tag} ({t.ids.length})
                    </a>
                  </li>
                ))}
              </ul>
            </nav>
            {tags.map((t) => (
              <section
                key={t.tag}
                id={`tag-${t.tag}`}
                aria-labelledby={`h-tag-${t.tag}`}
                className="scroll-mt-6"
              >
                <Heading
                  level={2}
                  id={`h-tag-${t.tag}`}
                  className="font-mono text-lg font-semibold"
                >
                  #{t.tag}
                </Heading>
                {t.meaning ? <p className="mt-1 max-w-3xl text-sm">{t.meaning}</p> : null}
                <p className="mt-1 text-xs text-muted">
                  {t.ids.length} {t.ids.length === 1 ? 'entry' : 'entries'} ·{' '}
                  <Link
                    className="underline"
                    href={`/?tag=${t.tag}`}
                  >
                    show them in the catalog
                  </Link>
                </p>
                {t.ids.length ? (
                  <ul className="mt-2 flex flex-wrap gap-x-4 gap-y-1 text-sm">
                    {t.ids.map((id) => (
                      <li key={id}>
                        <Link
                          className="underline decoration-border underline-offset-2 hover:decoration-accent"
                          href={docsHref(id)}
                        >
                          {name.get(id) ?? id}
                        </Link>
                      </li>
                    ))}
                  </ul>
                ) : null}
              </section>
            ))}
          </div>
        );
      })()}
    </>
  );
}
