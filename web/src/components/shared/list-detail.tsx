'use client';

/**
 * A list beside the thing chosen from it, the layout the system design and the
 * tags pages share.
 *
 * On a wide screen it is exactly the height under the top bar, so the page
 * itself does not move: the list stays where it is and scrolls under its
 * search, and everything else is one scroller beside it, the chosen item's
 * title pinned at its top while its content scrolls. On a narrow screen the
 * list is the page, and choosing an item opens it in a drawer; closing the
 * drawer comes back to the list.
 *
 * The chosen item lives in the link (each item is a link), so a view can be
 * shared and opens where it was left.
 */
import { useMemo, useState, type ReactNode } from 'react';
import Link from 'next/link';
import { HiMagnifyingGlass } from 'react-icons/hi2';
import {
  Drawer,
  DrawerDescription,
  DrawerHeader,
  DrawerPanel,
  DrawerPopup,
  DrawerTitle
} from '@/components/ui/drawer';
import { Empty, EmptyDescription, EmptyHeader, EmptyTitle } from '@/components/ui/empty';
import { Frame, FrameDescription, FrameHeader, FrameTitle } from '@/components/ui/frame';
import { Heading } from '@/components/ui/heading';
import { InputGroup, InputGroupAddon, InputGroupInput } from '@/components/ui/input-group';
import { Item, ItemActions, ItemContent, ItemDescription, ItemTitle } from '@/components/ui/item';
import { ScrollArea } from '@/components/ui/scroll-area';
import { useMediaQuery } from '@/hooks/use-media-query';
import { cn } from '@/lib/utils';

/** One row of the list. */
export type ListItem = {
  id: string;
  title: string;
  /** The second line: a level, a meaning. */
  subtitle?: string;
  /** Other words it is found by, never shown. */
  keywords?: string[];
  /** Shown at the end of the row: a count, a badge. */
  end?: ReactNode;
  href: string;
  /** Draw the title in the mono face (a tag, a code). */
  mono?: boolean;
};

/** A labelled block of rows. */
export type ListGroup = { label: string; items: ListItem[] };

/** The list: a search over every group, then each group's rows. */
export function SearchList({
  label,
  groups,
  current,
  placeholder,
  emptyHint
}: {
  /** The list's accessible name: "Case studies". */
  label: string;
  groups: ListGroup[];
  /** The chosen row's id; it is filled with the primary. */
  current?: string;
  placeholder: string;
  /** What to try when nothing matches. */
  emptyHint: string;
}) {
  const [query, setQuery] = useState('');
  const total = groups.reduce((n, g) => n + g.items.length, 0);
  const shown = useMemo(() => {
    const q = query.trim().toLowerCase().replace(/^#/, '');
    if (!q) return groups;
    return groups
      .map((g) => ({
        ...g,
        items: g.items.filter((i) =>
          [i.title, i.subtitle ?? '', ...(i.keywords ?? [])].some((v) =>
            v.toLowerCase().includes(q)
          )
        )
      }))
      .filter((g) => g.items.length > 0);
  }, [groups, query]);
  const count = shown.reduce((n, g) => n + g.items.length, 0);

  return (
    <nav
      aria-label={label}
      data-tour="item-list"
      className="flex min-h-0 flex-col gap-3"
    >
      <InputGroup
        size="sm"
        className="w-full shrink-0"
      >
        <InputGroupAddon>
          <HiMagnifyingGlass className="text-muted size-4" />
        </InputGroupAddon>
        <InputGroupInput
          placeholder={placeholder}
          aria-label={`Search ${label.toLowerCase()}`}
          value={query}
          onChange={(e) => setQuery(e.target.value)}
        />
      </InputGroup>
      <p
        role="status"
        className="text-muted shrink-0 text-xs"
      >
        {count === total ? `${total} in all` : `${count} of ${total} match`}
      </p>
      <ScrollArea
        orientation="vertical"
        className="min-h-0 flex-1"
      >
        {count < 1 ? (
          <Empty>
            <EmptyHeader>
              <EmptyTitle>Nothing matches</EmptyTitle>
              <EmptyDescription>{emptyHint}</EmptyDescription>
            </EmptyHeader>
          </Empty>
        ) : (
          <div className="flex flex-col gap-4 pe-2">
            {shown.map((g) => (
              <section
                key={g.label}
                aria-label={g.label}
                className="flex flex-col gap-1"
              >
                {groups.length > 1 ? (
                  <Heading
                    level={2}
                    className="text-muted px-2 font-sans text-xs font-semibold tracking-wide uppercase"
                  >
                    {g.label}
                  </Heading>
                ) : null}
                <ul className="flex flex-col gap-1">
                  {g.items.map((i) => {
                    const active = i.id === current;
                    return (
                      <li key={i.id}>
                        <Item
                          size="sm"
                          className={cn(
                            'hover:bg-subtle',
                            // The chosen row is a keycap, the primary cap, as in the sidebar.
                            active && 'keycap keycap-primary hover:bg-transparent'
                          )}
                          render={
                            <Link
                              href={i.href}
                              scroll={false}
                              aria-current={active ? 'page' : undefined}
                            />
                          }
                        >
                          <ItemContent>
                            <ItemTitle className={cn(i.mono && 'font-mono')}>{i.title}</ItemTitle>
                            {i.subtitle ? (
                              <ItemDescription
                                className={cn('line-clamp-1', active && 'text-stone-800!')}
                              >
                                {i.subtitle}
                              </ItemDescription>
                            ) : null}
                          </ItemContent>
                          {i.end ? <ItemActions>{i.end}</ItemActions> : null}
                        </Item>
                      </li>
                    );
                  })}
                </ul>
              </section>
            ))}
          </div>
        )}
      </ScrollArea>
    </nav>
  );
}

/** The list and the chosen item, wide or narrow. */
export function ListDetail({
  list,
  intro,
  open,
  onClose,
  detailKey,
  title,
  description,
  children
}: {
  /** The SearchList. */
  list: ReactNode;
  /** A line above the detail on a wide screen, above the list on a narrow one. */
  intro?: ReactNode;
  /** Whether an item was chosen: on a narrow screen, opens the drawer. */
  open: boolean;
  /** The drawer was closed: forget the choice, back to the list. */
  onClose: () => void;
  /** Changes with the item, so its scroll starts at the top. */
  detailKey: string;
  title: ReactNode;
  description?: ReactNode;
  /** The item's content. */
  children: ReactNode;
}) {
  const wide = useMediaQuery('(min-width: 64rem)');

  if (!wide) {
    return (
      <div className="flex flex-col gap-6">
        {intro ? <div className="text-muted text-sm">{intro}</div> : null}
        {list}
        <Drawer
          open={open}
          onOpenChange={(next) => {
            if (!next) onClose();
          }}
        >
          <DrawerPopup
            showBar
            showCloseButton
          >
            {open ? (
              <>
                <DrawerHeader>
                  <DrawerTitle className="flex flex-wrap items-center gap-2">{title}</DrawerTitle>
                  {description ? <DrawerDescription>{description}</DrawerDescription> : null}
                </DrawerHeader>
                <DrawerPanel className="px-3">{children}</DrawerPanel>
              </>
            ) : null}
          </DrawerPopup>
        </Drawer>
      </div>
    );
  }

  return (
    <div className="grid h-[calc(100svh-var(--header-height,3rem)-3rem)] min-h-0 grid-cols-[16rem_minmax(0,1fr)] gap-6 xl:grid-cols-[18rem_minmax(0,1fr)]">
      {list}
      <ScrollArea
        key={detailKey}
        orientation="vertical"
        className="min-h-0"
      >
        <div className="flex flex-col gap-6 pe-3">
          {intro ? <div className="text-muted text-sm">{intro}</div> : null}
          <Frame data-tour="item-detail">
            <FrameHeader className="bg-subtle sticky top-0 z-10 rounded-t-2xl">
              <FrameTitle className="flex flex-wrap items-center gap-2 text-lg">{title}</FrameTitle>
              {description ? <FrameDescription>{description}</FrameDescription> : null}
            </FrameHeader>
            {children}
          </Frame>
        </div>
      </ScrollArea>
    </div>
  );
}
