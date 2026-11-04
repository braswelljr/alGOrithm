'use client';

import Link from 'next/link';
import { tagsOf } from '@/components/docs/tags-explorer';
import { Heading } from '@/components/ui/heading';
import { docsHref } from '@/lib/entries';
import type { Catalog } from '@/visualizer/core';

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
                      className="border-border hover:border-accent rounded-full border px-3 py-1 font-mono text-xs"
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
                {t.meaning ? <p className="mt-1 text-sm">{t.meaning}</p> : null}
                <p className="text-muted mt-1 text-xs">
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
                          className="decoration-border hover:decoration-accent underline underline-offset-2"
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
