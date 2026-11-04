'use client';

/**
 * One search box for the whole site, opened with ⌘K (ported from
 * gsl-special-courses' command palette: the same layout, keys, scopes,
 * memory, preview pane and footer; this site's data and no accounts).
 *
 * Opened empty, it offers what fits where the reader is (the entry they are
 * looking at), what they pinned, what they opened lately, then every page and
 * everything they can do. Typing narrows all of that and searches every entry,
 * learning path, category and tag; a leading `#`, `~`, `>` or `/` narrows the
 * search to tags, paths, commands or pages. Tab on a record lists what can be
 * done to it.
 */
import { useEffect, useMemo, useState } from 'react';
import { usePathname, useRouter } from 'next/navigation';
import {
  ArrowDown,
  ArrowLeft,
  ArrowRight,
  ArrowUp,
  Check,
  CornerDownLeft,
  History,
  Pin,
  RotateCw,
  Search,
  X
} from 'lucide-react';
import { matchSorter, rankings } from 'match-sorter';
import {
  copyText,
  usePaletteCommands,
  type PaletteEntry
} from '@/components/shared/command-palette-commands';
import {
  contextEntries,
  Highlight,
  KIND_HINT,
  KIND_TONE,
  KindIcon,
  recordActions,
  RecordFacts,
  RecordLead,
  RecordPreview
} from '@/components/shared/command-palette-records';
import { Button } from '@/components/ui/button';
import {
  Command,
  CommandCollection,
  CommandDialog,
  CommandDialogPopup,
  CommandEmpty,
  CommandFooter,
  CommandGroup,
  CommandGroupLabel,
  CommandInput,
  CommandItem,
  CommandList,
  CommandPanel,
  CommandSeparator,
  CommandShortcut
} from '@/components/ui/command';
import { Heading } from '@/components/ui/heading';
import { Skeleton } from '@/components/ui/skeleton';
import {
  KIND_LABELS,
  MIN_SEARCH,
  useCommandSearch,
  type RecordEntry
} from '@/hooks/use-command-search';
import { setCommandPaletteOpen, useCommandPaletteOpen } from '@/lib/command-palette-store';
import {
  clearRecent,
  LOCAL,
  rememberVisit,
  togglePin,
  usePaletteMemory,
  type RememberedEntry
} from '@/lib/command-palette/memory';
import { openScope, SCOPES, type PaletteScope } from '@/lib/command-palette/scope';
import { cn } from '@/lib/utils';

/** One key in the footer legend. */
function Key({ children, label }: { readonly children: React.ReactNode; readonly label?: string }) {
  return (
    <kbd
      aria-label={label}
      className="bg-subtle text-fg grid h-5 min-w-5 place-content-center rounded px-1 font-sans text-xs font-medium"
    >
      {children}
    </kbd>
  );
}

/** A labelled block of rows, as the list draws it. */
type ViewGroup = { value: string; label: string; items: PaletteEntry[] };

/**
 * Rank rows against the words typed. A title that starts with them beats one
 * that contains them, which beats a match only in the hidden keywords. Among
 * equally good matches the group's own order holds.
 */
function rank(entries: PaletteEntry[], term: string): PaletteEntry[] {
  if (!term) return entries;
  return matchSorter(entries, term, {
    keys: [
      'title',
      { key: 'keywords', threshold: rankings.CONTAINS },
      { key: 'subtitle', threshold: rankings.CONTAINS }
    ],
    threshold: rankings.CONTAINS,
    baseSort: () => 0
  });
}

const fromMemory = (entry: RememberedEntry, group: 'pinned' | 'recent'): PaletteEntry => ({
  id: `${group}:${entry.id}`,
  title: entry.title,
  subtitle: entry.subtitle,
  href: entry.href,
  kind: entry.kind,
  icon: <KindIcon kind={entry.kind} />,
  hint: entry.kind in KIND_HINT ? KIND_HINT[entry.kind as keyof typeof KIND_HINT] : undefined
});

/** The id a row is remembered under: a record's own, or its address. */
const memoryId = (entry: PaletteEntry) =>
  entry.record?.id ??
  (entry.id.startsWith('pinned:') || entry.id.startsWith('recent:')
    ? entry.id.replace(/^(pinned|recent):/, '')
    : (entry.href ?? entry.id));

export function CommandPalette({ onHelp }: { onHelp: () => void }) {
  const router = useRouter();
  const pathname = usePathname() ?? '/';
  const commands = usePaletteCommands({ onHelp });
  const memory = usePaletteMemory(LOCAL);

  // Open state lives outside this component: the shortcut is one way in and
  // the navbar's search control is another, in a different subtree.
  const open = useCommandPaletteOpen();
  const [search, setSearch] = useState('');
  const [scope, setScope] = useState<PaletteScope | null>(null);
  /** The record whose actions are listed, when they are. */
  const [actionsFor, setActionsFor] = useState<RecordEntry | null>(null);
  const [highlighted, setHighlighted] = useState<PaletteEntry | undefined>();
  const term = search.trim();

  const records = useCommandSearch(actionsFor ? '' : search, scope ? scope.kinds : undefined);
  const entriesById = useMemo(
    () => new Map((records.index?.entries ?? []).map((e) => [e.id, e])),
    [records.index]
  );

  // Each opening starts from nothing.
  useEffect(() => {
    if (open) return;
    // eslint-disable-next-line react-hooks/set-state-in-effect -- resetting on close is the point
    setSearch('');
    setScope(null);
    setActionsFor(null);
  }, [open]);

  const remember = (entry: PaletteEntry): RememberedEntry | null =>
    entry.href
      ? {
          id: memoryId(entry),
          title: entry.title,
          subtitle: entry.subtitle,
          href: entry.href,
          kind: entry.record?.kind ?? entry.kind ?? 'page'
        }
      : null;

  const openInNewTab = (href: string) => {
    setCommandPaletteOpen(false);
    window.open(href, '_blank', 'noopener');
  };

  const groups = useMemo<ViewGroup[]>(() => {
    // Listing what can be done to one record.
    if (actionsFor) {
      const actions = recordActions(actionsFor, {
        isPinned: (id) => memory.pinned.some((entry) => entry.id === id),
        togglePin: (entry) =>
          togglePin(LOCAL, {
            id: entry.id,
            title: entry.title,
            subtitle: entry.subtitle,
            href: entry.href ?? '',
            kind: entry.kind
          }),
        copy: copyText,
        openInNewTab
      });
      return [{ value: 'actions', label: actionsFor.title, items: rank(actions, term) }];
    }

    const out: ViewGroup[] = [];
    const fixed = (names: string[]) =>
      commands
        .filter((group) => names.includes(group.value))
        .map((group) => ({
          value: group.value,
          label: group.value,
          items: rank(group.items, term)
        }))
        .filter((group) => group.items.length > 0);

    // Opened empty: what fits here, what was pinned, what was opened lately.
    if (!scope && !term) {
      const here = contextEntries(pathname, window.location.search, entriesById);
      if (here.length > 0) out.push({ value: 'context', label: 'On this page', items: here });
      if (memory.pinned.length > 0) {
        out.push({
          value: 'pinned',
          label: 'Pinned',
          items: memory.pinned.map((entry) => fromMemory(entry, 'pinned'))
        });
      }
      if (memory.recent.length > 0) {
        out.push({
          value: 'recent',
          label: 'Recent',
          items: [
            ...memory.recent.map((entry) => fromMemory(entry, 'recent')),
            {
              id: 'recent:clear',
              title: 'Clear recent',
              icon: <X />,
              run: () => clearRecent(LOCAL)
            }
          ]
        });
      }
    }

    if (!scope) out.push(...fixed(commands.map((group) => group.value)));
    else if (scope.fixed === 'commands') out.push(...fixed(['Actions', 'Settings', 'Appearance']));
    else if (scope.fixed === 'pages') out.push(...fixed(['Navigation']));

    // Records follow the fixed rows, so what is already on screen does not
    // move when the index arrives.
    for (const group of records.groups) {
      const items: PaletteEntry[] = group.items.map((entry) => ({
        id: entry.id,
        title: entry.title,
        subtitle: entry.subtitle,
        href: entry.href,
        icon: <RecordLead record={entry.record} />,
        hint: KIND_HINT[entry.kind],
        record: entry
      }));
      if (group.seeAll && group.total > group.items.length) {
        items.push({
          id: `see-all:${group.kind}`,
          title: `See all ${group.total} ${group.label.toLowerCase()}`,
          href: group.seeAll,
          icon: <ArrowRight />
        });
      }
      out.push({ value: group.kind, label: group.label, items });
    }
    return out;
    // openInNewTab only closes the palette and opens a window; it changes nothing the list reads.
  }, [actionsFor, memory, term, commands, scope, pathname, records.groups, entriesById]);

  const flat = useMemo(() => groups.flatMap((group) => group.items), [groups]);
  const lastGroup = groups.at(-1)?.value;

  const showActions = (entry: PaletteEntry | undefined) => {
    if (!entry?.record) return false;
    setActionsFor(entry.record);
    setSearch('');
    return true;
  };

  const select = (entry: PaletteEntry) => {
    if (entry.run) {
      setCommandPaletteOpen(false);
      entry.run();
      return;
    }
    if (!entry.href) {
      if (entry.record) showActions(entry);
      return;
    }
    const remembered = remember(entry);
    if (remembered && !entry.id.startsWith('see-all:')) rememberVisit(LOCAL, remembered);
    setCommandPaletteOpen(false);
    router.push(entry.href);
  };

  const onValueChange = (value: string) => {
    if (!actionsFor && !scope) {
      const opened = openScope(value, SCOPES);
      if (opened) {
        setScope(opened.scope);
        setSearch(opened.rest);
        return;
      }
    }
    setSearch(value);
  };

  const onKeyDown = (
    event: React.KeyboardEvent<HTMLInputElement> & { preventBaseUIHandler?: () => void }
  ) => {
    const input = event.currentTarget;
    const atEnd = input.selectionStart === input.value.length;
    const mod = event.metaKey || event.ctrlKey;

    // Backspace on an empty box steps back out: from actions to the list, or
    // out of a scope.
    if (event.key === 'Backspace' && input.value === '') {
      if (actionsFor) setActionsFor(null);
      else if (scope) setScope(null);
      return;
    }
    if (event.key === 'ArrowLeft' && input.value === '' && actionsFor) {
      event.preventDefault();
      setActionsFor(null);
      return;
    }
    // Tab, or → at the end of the words, opens a record's actions.
    if (
      !actionsFor &&
      (event.key === 'Tab' || (event.key === 'ArrowRight' && atEnd)) &&
      highlighted?.record
    ) {
      event.preventDefault();
      event.preventBaseUIHandler?.();
      showActions(highlighted);
      return;
    }
    // ⌘↵ opens the highlighted row in a new tab instead of this one.
    if (event.key === 'Enter' && mod && highlighted?.href) {
      event.preventDefault();
      event.preventBaseUIHandler?.();
      openInNewTab(highlighted.href);
      return;
    }
    // ⌘1 to ⌘9 jump straight to a row.
    if (mod && /^[1-9]$/.test(event.key)) {
      const entry = flat[Number(event.key) - 1];
      if (entry) {
        event.preventDefault();
        event.preventBaseUIHandler?.();
        select(entry);
      }
    }
  };

  // The kinds expected back but not yet here, each held in place while the
  // index loads.
  const pending = records.isLoading
    ? records.expected.filter((kind) => !records.groups.some((group) => group.kind === kind))
    : [];
  const preview = actionsFor ?? highlighted?.record;
  const placeholder = actionsFor
    ? `Actions for ${actionsFor.title}…`
    : scope
      ? `Search ${scope.hint}…`
      : 'Type a command or search…';

  return (
    <CommandDialog
      open={open}
      onOpenChange={(next, details) => {
        // Escape inside a record's actions goes back to the list rather than
        // closing the palette.
        if (!next && details?.reason === 'escape-key' && actionsFor) {
          setActionsFor(null);
          return;
        }
        setCommandPaletteOpen(next);
      }}
    >
      <CommandDialogPopup
        viewportClassName="max-sm:p-0"
        className="max-h-[min(44rem,80vh)] max-w-4xl overflow-hidden before:rounded-none max-sm:h-dvh max-sm:max-h-none max-sm:max-w-none max-sm:rounded-none max-sm:border-0"
      >
        <Command
          items={groups}
          value={search}
          onValueChange={onValueChange}
          // Matching happens above, across fixed rows and records; the list
          // shows exactly what it is given.
          filter={null}
          onItemHighlighted={(item) => {
            const next = item as PaletteEntry | undefined;
            setHighlighted((prev) => (prev?.id === next?.id ? prev : next));
          }}
        >
          <CommandInput
            placeholder={placeholder}
            size="default"
            onKeyDown={onKeyDown}
            className="text-fg placeholder:text-muted h-13 border-0! border-none bg-transparent! text-base shadow-none sm:text-sm"
            containerClassName="border-border [&_[data-slot=autocomplete-start-addon]_svg]:text-muted flex items-center gap-2 border-b px-3 py-0"
            leading={
              actionsFor ? (
                <Button
                  variant="secondary"
                  size="xs"
                  onClick={() => setActionsFor(null)}
                  className="max-w-48 shrink-0 gap-1 text-xs"
                >
                  <ArrowLeft className="size-3 shrink-0" />
                  <span className="truncate">{actionsFor.title}</span>
                </Button>
              ) : scope ? (
                <Button
                  variant="secondary"
                  size="xs"
                  aria-label={`Stop searching only ${scope.hint}`}
                  onClick={() => setScope(null)}
                  className="shrink-0 gap-1 text-xs"
                >
                  <span className="font-mono">{scope.prefix}</span>
                  {scope.label}
                  <X className="size-3" />
                </Button>
              ) : null
            }
            trailing={
              <span className="flex shrink-0 items-center gap-1.5">
                {search ? (
                  <Button
                    variant="ghost"
                    size="xs"
                    aria-label="Clear the query"
                    onClick={() => setSearch('')}
                    className="text-xs"
                  >
                    Clear
                  </Button>
                ) : null}
                <Button
                  variant="ghost"
                  size="xs"
                  aria-label="Close"
                  onClick={() => setCommandPaletteOpen(false)}
                  className="h-6 px-1.5 text-xs"
                >
                  Esc
                </Button>
              </span>
            }
          />
          <CommandPanel className="flex min-h-0 flex-1 rounded-none border-0">
            <div className="flex min-h-0 min-w-0 flex-1 flex-col">
              {groups.length < 1 && pending.length < 1 && !records.isError ? (
                <CommandEmpty>
                  <div className="px-6 py-12 text-center">
                    <Search className="text-muted mx-auto size-6" />
                    <p className="text-fg mt-3 text-sm font-medium">
                      {term ? `No results for "${term}"` : `Search ${scope?.hint ?? 'everything'}`}
                    </p>
                    <p className="text-muted mt-1 text-xs">
                      {term && term.length < MIN_SEARCH
                        ? 'Keep typing: records are searched from two characters.'
                        : scope?.prefix === '#'
                          ? 'A tag such as divide-and-conquer, greedy or graph.'
                          : scope?.prefix === '~'
                            ? 'A learning path such as sorting or how an OS works.'
                            : 'Try an algorithm, a data structure, a tag or a page.'}
                    </p>
                  </div>
                </CommandEmpty>
              ) : (
                <CommandList className="max-h-[min(34rem,62vh)] max-sm:max-h-none max-sm:flex-1">
                  {(group: ViewGroup) => (
                    <div key={group.value}>
                      <CommandGroup items={group.items}>
                        <CommandGroupLabel className="text-muted flex items-center gap-1.5 px-3 pt-2 pb-1.5 text-xs font-semibold">
                          {group.value === 'pinned' ? <Pin className="size-3" /> : null}
                          {group.value === 'recent' ? <History className="size-3" /> : null}
                          {group.value === 'actions' ? `Actions for ${group.label}` : group.label}
                        </CommandGroupLabel>
                        <CommandCollection>
                          {(entry: PaletteEntry) => {
                            const position = flat.indexOf(entry);
                            return (
                              <CommandItem
                                key={entry.id}
                                value={entry}
                                onClick={() => select(entry)}
                                className={cn(
                                  'text-fg data-highlighted:bg-subtle flex cursor-pointer items-center gap-3 rounded-md px-3 py-3 text-sm sm:py-2',
                                  entry.id.startsWith('see-all:') && 'text-accent font-medium'
                                )}
                              >
                                <span className="text-muted flex size-5 shrink-0 items-center justify-center [&>svg]:size-4">
                                  {entry.icon}
                                </span>
                                <span className="flex min-w-0 flex-1 flex-col">
                                  <span className="truncate font-medium">
                                    <Highlight
                                      text={entry.title}
                                      term={term}
                                    />
                                  </span>
                                  {entry.subtitle ? (
                                    <span className="text-muted truncate text-xs">
                                      <Highlight
                                        text={entry.subtitle}
                                        term={term}
                                      />
                                    </span>
                                  ) : null}
                                </span>
                                {entry.record ? (
                                  <span className="hidden shrink-0 sm:flex">
                                    <RecordFacts record={entry.record.record} />
                                  </span>
                                ) : null}
                                {entry.checked ? (
                                  <Check
                                    aria-label="Current"
                                    className="text-accent size-4 shrink-0"
                                  />
                                ) : null}
                                {entry.hint ? (
                                  <CommandShortcut
                                    className={cn(
                                      'hidden rounded px-1.5 py-0.5 text-xs font-semibold tracking-normal sm:inline',
                                      KIND_TONE[entry.record?.kind ?? entry.kind ?? 'page'] ??
                                        KIND_TONE.page
                                    )}
                                  >
                                    {entry.hint}
                                  </CommandShortcut>
                                ) : null}
                                {position >= 0 && position < 9 ? (
                                  <kbd className="text-muted hidden w-7 shrink-0 text-right font-sans text-xs font-medium md:inline">
                                    ⌘{position + 1}
                                  </kbd>
                                ) : null}
                              </CommandItem>
                            );
                          }}
                        </CommandCollection>
                      </CommandGroup>
                      {group.value !== lastGroup ? (
                        <CommandSeparator className="bg-border mx-1 my-1.5 h-px" />
                      ) : null}
                    </div>
                  )}
                </CommandList>
              )}

              {pending.length > 0 ? (
                <div
                  aria-live="polite"
                  className="border-border border-t px-2 py-2"
                >
                  {pending.map((kind) => (
                    <div key={kind}>
                      <p className="text-muted px-3 pt-1 pb-1.5 text-xs font-semibold">
                        Searching {KIND_LABELS[kind].toLowerCase()}…
                      </p>
                      <div className="flex items-center gap-3 px-3 py-2">
                        <Skeleton className="size-4 shrink-0 rounded" />
                        <Skeleton className="h-3.5 w-1/2" />
                      </div>
                    </div>
                  ))}
                </div>
              ) : null}

              {records.isError ? (
                <div
                  role="alert"
                  className="border-border flex items-center justify-between gap-3 border-t px-5 py-3 text-sm"
                >
                  <span className="text-fg">
                    The search index could not be loaded. Pages and commands still work.
                  </span>
                  <Button
                    variant="secondary"
                    size="xs"
                    onClick={records.retry}
                    className="shrink-0 gap-1.5 text-xs"
                  >
                    <RotateCw className="size-3" />
                    Try again
                  </Button>
                </div>
              ) : null}
            </div>

            <aside
              aria-label="Preview"
              className="border-border hidden w-80 shrink-0 overflow-y-auto border-l p-5 lg:block"
            >
              {preview ? (
                <RecordPreview entry={preview} />
              ) : highlighted ? (
                <div className="flex flex-col gap-3">
                  <span className="bg-subtle text-fg grid size-10 place-content-center rounded-lg [&>svg]:size-5">
                    {highlighted.icon}
                  </span>
                  <Heading
                    level={3}
                    className="text-fg font-sans text-sm font-semibold"
                  >
                    {highlighted.title}
                  </Heading>
                  {highlighted.subtitle ? (
                    <p className="text-fg text-sm">{highlighted.subtitle}</p>
                  ) : null}
                  <p className="text-muted text-xs">
                    {highlighted.run
                      ? 'Press Enter to do this.'
                      : 'Press Enter to open, or ⌘Enter for a new tab.'}
                  </p>
                </div>
              ) : null}
              {highlighted?.record && !actionsFor ? (
                <p className="border-border text-muted mt-5 flex items-center gap-1.5 border-t pt-4 text-xs">
                  <Key label="Tab key">Tab</Key>
                  for what you can do with it
                </p>
              ) : null}
            </aside>
          </CommandPanel>

          <CommandFooter className="hidden rounded-none sm:flex">
            <ul
              aria-label="Narrow the search"
              className="text-muted flex items-center gap-3 text-xs"
            >
              {SCOPES.map((s) => (
                <li
                  key={s.prefix}
                  className="flex items-center gap-1"
                >
                  <Key>{s.prefix}</Key>
                  <span>{s.hint}</span>
                </li>
              ))}
            </ul>

            <ul className="text-muted flex items-center gap-3 text-xs">
              <li className="flex items-center gap-1">
                <Key label="Arrow up">
                  <ArrowUp className="size-3" />
                </Key>
                <Key label="Arrow down">
                  <ArrowDown className="size-3" />
                </Key>
                <span>Move</span>
              </li>
              <li className="flex items-center gap-1">
                <Key label="Enter key">
                  <CornerDownLeft className="size-3" />
                </Key>
                <span>Open</span>
              </li>
              <li className="hidden items-center gap-1 lg:flex">
                <Key label="Command Enter">⌘↵</Key>
                <span>New tab</span>
              </li>
              <li className="flex items-center gap-1">
                <Key label="Tab key">Tab</Key>
                <span>Actions</span>
              </li>
            </ul>
          </CommandFooter>
        </Command>
      </CommandDialogPopup>
    </CommandDialog>
  );
}
