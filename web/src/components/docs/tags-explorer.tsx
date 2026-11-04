'use client';

/**
 * The tags page, in the layout the system design page has: every tag with
 * what it means in a searchable list, and the chosen tag's entries beside it
 * (a drawer on a phone).
 *
 * The chosen tag is in the address (?tag=greedy). An old #tag-greedy link, and
 * a board link with several tags (?tag=array,greedy), open the first tag.
 */
import { useEffect, useMemo } from 'react';
import Link from 'next/link';
import { parseAsString, useQueryState } from 'nuqs';
import { LuLayoutGrid } from 'react-icons/lu';
import { LevelBadge } from '@/components/algorithm/level-badge';
import { ListDetail, SearchList } from '@/components/shared/list-detail';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { FramePanel } from '@/components/ui/frame';
import { Item, ItemContent, ItemDescription, ItemTitle } from '@/components/ui/item';
import { docsHref } from '@/lib/entries';
import type { Catalog } from '@/visualizer/core';

type Tag = { tag: string; meaning: string; ids: string[] };

/** Every tag of the catalog: the vocabulary's meanings, and the entries carrying each. */
export function tagsOf(catalog: Catalog): Tag[] {
  const used = new Map<string, string[]>();
  for (const e of catalog.entries)
    for (const t of e.tags) used.set(t, [...(used.get(t) ?? []), e.id]);
  const meaning = new Map(catalog.tags.map((t) => [t.tag, t.meaning]));
  const names = [...new Set([...catalog.tags.map((t) => t.tag), ...used.keys()])].sort();
  return names.map((tag) => ({ tag, meaning: meaning.get(tag) ?? '', ids: used.get(tag) ?? [] }));
}

export function TagsExplorer({ catalog, intro }: { catalog: Catalog; intro?: React.ReactNode }) {
  const tags = useMemo(() => tagsOf(catalog), [catalog]);
  const byId = useMemo(() => new Map(catalog.entries.map((e) => [e.id, e])), [catalog.entries]);
  const [param, setParam] = useQueryState('tag', parseAsString);

  // A board link (?tag=a,b) keeps its first tag.
  const name = param?.split(',')[0] ?? null;
  const chosen = tags.find((t) => t.tag === name);
  // Nothing chosen: a wide screen shows the first tag, a phone the list.
  const shown = chosen ?? tags[0];

  // An old #tag-x link (the palette, other pages) opens that tag.
  // The anchor is then dropped and the page put back at the top: the browser
  // had jumped to the anchor in the server-rendered list.
  useEffect(() => {
    if (!window.location.hash.startsWith('#tag-')) return;
    const hash = window.location.hash.slice('#tag-'.length);
    if (!param && tags.some((t) => t.tag === hash)) void setParam(hash);
    history.replaceState(history.state, '', window.location.pathname + window.location.search);
    window.scrollTo(0, 0);
  }, [param, setParam, tags]);

  if (!shown) return null;

  const list = (
    <SearchList
      label="Tags"
      placeholder="Search tags and what they mean..."
      emptyHint="Try a shorter word, such as graph or sort."
      current={shown.tag}
      groups={[
        {
          label: 'Tags',
          items: tags.map((t) => ({
            id: t.tag,
            title: `#${t.tag}`,
            subtitle: t.meaning,
            mono: true,
            href: `/docs/tags/?tag=${t.tag}`,
            end: (
              <Badge
                variant={t.tag === shown.tag ? 'default' : 'secondary'}
                size="sm"
                case="normal"
                className="tabular-nums"
              >
                {t.ids.length}
              </Badge>
            )
          }))
        }
      ]}
    />
  );

  return (
    <ListDetail
      list={list}
      intro={intro}
      open={!!chosen}
      onClose={() => void setParam(null)}
      detailKey={shown.tag}
      title={
        <>
          <span className="font-mono font-semibold">#{shown.tag}</span>
          <Badge
            variant="secondary"
            size="sm"
            case="normal"
          >
            {shown.ids.length} {shown.ids.length === 1 ? 'entry' : 'entries'}
          </Badge>
        </>
      }
      description={shown.meaning}
    >
      <FramePanel className="flex flex-col gap-3">
        <Button
          variant="ghost"
          size="sm"
          className="w-fit"
          render={<Link href={`/?tag=${shown.tag}`} />}
        >
          <LuLayoutGrid aria-hidden />
          Show them in the catalog
        </Button>
        <ul className="grid gap-1.5 xl:grid-cols-2">
          {shown.ids.map((id) => {
            const e = byId.get(id);
            return (
              <li key={id}>
                <Item
                  size="sm"
                  variant="muted"
                  render={<Link href={docsHref(id)} />}
                  className="hover:bg-subtle h-full"
                >
                  <ItemContent>
                    <ItemTitle>{e?.name ?? id}</ItemTitle>
                    {e?.summary ? (
                      <ItemDescription className="line-clamp-2">{e.summary}</ItemDescription>
                    ) : null}
                  </ItemContent>
                  {e?.level ? <LevelBadge level={e.level} /> : null}
                </Item>
              </li>
            );
          })}
        </ul>
      </FramePanel>
    </ListDetail>
  );
}
