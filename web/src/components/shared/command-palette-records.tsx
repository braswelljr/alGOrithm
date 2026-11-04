'use client';

/**
 * How the command palette draws a found record: its icon, the kind label, the
 * facts beside its row, the preview pane and what can be done to it (after
 * gsl-special-courses' file of the same name). The kinds are this site's:
 * entries, learning paths, categories and tags.
 */
import {
  BookOpen,
  Clock,
  Copy,
  ExternalLink,
  FolderTree,
  GraduationCap,
  Hash,
  Layers,
  ListChecks,
  Pin,
  PinOff,
  Play,
  Signal
} from 'lucide-react';
import type { PaletteEntry } from '@/components/shared/command-palette-commands';
import { Badge } from '@/components/ui/badge';
import { Heading } from '@/components/ui/heading';
import type { RecordEntry, SearchRecord } from '@/hooks/use-command-search';
import type { SearchKind } from '@/lib/command-palette/scope';
import { docsHref, entryHref } from '@/lib/entries';
import type { SearchEntry } from '@/lib/search-index';
import { cn } from '@/lib/utils';

/** The icon of a kind of row: a record kind, or 'page'. */
export function KindIcon({ kind }: { readonly kind: string }) {
  switch (kind) {
    case 'entries':
      return <Layers />;
    case 'paths':
      return <GraduationCap />;
    case 'categories':
      return <FolderTree />;
    case 'tags':
      return <Hash />;
    default:
      return <BookOpen />;
  }
}

/** The label at the end of a row. */
export const KIND_HINT: Record<SearchKind, string> = {
  entries: 'Entry',
  paths: 'Path',
  categories: 'Category',
  tags: 'Tag'
};

/** Each kind's solid label colors, AA in both themes. */
export const KIND_TONE: Record<string, string> = {
  entries: 'bg-primary-100 text-primary-900 dark:bg-primary-900 dark:text-primary-100',
  paths: 'bg-secondary-100 text-secondary-900 dark:bg-secondary-800 dark:text-secondary-50',
  categories: 'bg-emerald-100 text-emerald-900 dark:bg-emerald-900 dark:text-emerald-100',
  tags: 'bg-violet-100 text-violet-900 dark:bg-violet-900 dark:text-violet-100',
  page: 'bg-subtle text-fg'
};

/** The words typed, marked where they appear in a title or a second line. */
export function Highlight({ text, term }: { readonly text: string; readonly term: string }) {
  const words = term
    .trim()
    .split(/\s+/)
    .filter((word) => word.length > 1)
    .map((word) => word.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'));
  if (words.length < 1 || !text) return <>{text}</>;
  const parts = text.split(new RegExp(`(${words.join('|')})`, 'gi'));
  return (
    <>
      {parts.map((part, index) =>
        index % 2 === 1 ? (
          <mark
            key={index}
            className="rounded-sm bg-amber-100 px-px text-stone-950 dark:bg-amber-900 dark:text-stone-50"
          >
            {part}
          </mark>
        ) : (
          part
        )
      )}
    </>
  );
}

/** What leads a record's row: its kind's icon. */
export function RecordLead({ record }: { readonly record: SearchRecord }) {
  return <KindIcon kind={record.kind} />;
}

/** A small solid chip: the library's Badge at its sm size. */
function Chip({ className, children }: { className?: string; children: React.ReactNode }) {
  return (
    <Badge
      variant="secondary"
      size="sm"
      case="normal"
      className={cn('font-mono', className)}
    >
      {children}
    </Badge>
  );
}

/** The facts beside a record's row: an entry's cost and level, a path's length. */
export function RecordFacts({ record }: { readonly record: SearchRecord }) {
  switch (record.kind) {
    case 'entries':
      return (
        <span className="flex items-center gap-1.5">
          <Chip>{record.data.time}</Chip>
          <Chip className="font-sans capitalize">{record.data.level}</Chip>
        </span>
      );
    case 'paths':
      return <Chip className="font-sans">{record.data.steps} steps</Chip>;
    case 'categories':
      return <Chip className="font-sans">{record.data.entries} entries</Chip>;
    case 'tags':
      return <Chip className="font-sans">{record.data.entries} entries</Chip>;
  }
}

/** One labelled fact in the preview. */
function Fact({
  icon,
  label,
  children
}: {
  icon: React.ReactNode;
  label: string;
  children: React.ReactNode;
}) {
  if (children === undefined || children === null || children === '') return null;
  return (
    <div className="flex items-start gap-2.5">
      <span className="text-muted mt-0.5 shrink-0 [&>svg]:size-4">{icon}</span>
      <div className="min-w-0">
        <dt className="text-muted text-xs font-medium">{label}</dt>
        <dd className="text-fg text-sm break-words">{children}</dd>
      </div>
    </div>
  );
}

/** The pane beside the list: everything a record's row could not fit. */
export function RecordPreview({ entry }: { readonly entry: RecordEntry }) {
  const { record } = entry;
  return (
    <div className="flex flex-col gap-4">
      <header className="flex items-start gap-3">
        <span className="bg-subtle text-fg grid size-10 shrink-0 place-content-center rounded-lg [&>svg]:size-5">
          <KindIcon kind={record.kind} />
        </span>
        <div className="min-w-0">
          <p className="text-muted text-xs font-medium">{KIND_HINT[record.kind]}</p>
          <Heading
            level={3}
            className="text-fg font-sans text-sm font-semibold break-words"
          >
            {entry.title}
          </Heading>
        </div>
      </header>
      <dl className="flex flex-col gap-3">
        {record.kind === 'entries' ? (
          <>
            <Fact
              icon={<BookOpen />}
              label="About"
            >
              {record.data.summary}
            </Fact>
            <Fact
              icon={<FolderTree />}
              label="Category"
            >
              {record.data.categoryName}
            </Fact>
            <Fact
              icon={<Clock />}
              label="Cost"
            >
              <span className="font-mono">
                time {record.data.time}
                {record.data.space ? `, space ${record.data.space}` : ''}
              </span>
            </Fact>
            <Fact
              icon={<Signal />}
              label="Level"
            >
              <span className="capitalize">{record.data.level}</span>
            </Fact>
            {record.data.tags.length > 0 ? (
              <Fact
                icon={<Hash />}
                label="Tags"
              >
                <span className="flex flex-wrap gap-1">
                  {record.data.tags.map((t) => (
                    <Chip
                      key={t}
                      className="font-sans"
                    >
                      #{t}
                    </Chip>
                  ))}
                </span>
              </Fact>
            ) : null}
          </>
        ) : record.kind === 'paths' ? (
          <>
            <Fact
              icon={<BookOpen />}
              label="About"
            >
              {record.data.summary}
            </Fact>
            <Fact
              icon={<Signal />}
              label="Level"
            >
              <span className="capitalize">{record.data.level}</span>
            </Fact>
            <Fact
              icon={<ListChecks />}
              label="Steps"
            >
              {record.data.steps}
            </Fact>
          </>
        ) : record.kind === 'categories' ? (
          <>
            <Fact
              icon={<BookOpen />}
              label="About"
            >
              {record.data.summary}
            </Fact>
            <Fact
              icon={<Layers />}
              label="Domain"
            >
              {record.data.topic}
            </Fact>
            <Fact
              icon={<ListChecks />}
              label="Entries"
            >
              {record.data.entries}
            </Fact>
          </>
        ) : (
          <>
            <Fact
              icon={<BookOpen />}
              label="Means"
            >
              {record.data.meaning}
            </Fact>
            <Fact
              icon={<ListChecks />}
              label="Entries"
            >
              {record.data.entries}
            </Fact>
          </>
        )}
      </dl>
    </div>
  );
}

/** What the actions need from the palette. */
export type RecordActionContext = {
  isPinned: (id: string) => boolean;
  togglePin: (entry: RecordEntry) => void;
  copy: (text: string, message: string) => void;
  openInNewTab: (href: string) => void;
};

/** The command that runs an entry's example in a terminal. */
export const runCommand = (id: string) => `algorithm run ${id} --example`;

/** What can be done to one record, listed with Tab or →. */
export function recordActions(entry: RecordEntry, ctx: RecordActionContext): PaletteEntry[] {
  const href = entry.href;
  const out: PaletteEntry[] = href
    ? [
        { id: `${entry.id}:open`, title: 'Open', href, icon: <ExternalLink /> },
        {
          id: `${entry.id}:tab`,
          title: 'Open in a new tab',
          icon: <ExternalLink />,
          hint: '⌘↵',
          run: () => ctx.openInNewTab(href)
        }
      ]
    : [];
  const record = entry.record;
  if (record.kind === 'entries') {
    const e = record.data;
    out.push(
      {
        id: `${entry.id}:run`,
        title: 'Visualize it',
        subtitle: 'Run it step by step on your data',
        href: entryHref(e),
        icon: <Play />
      },
      {
        id: `${entry.id}:quiz`,
        title: 'Quiz me on it',
        subtitle: 'Five questions on what it does next',
        href: `/quiz/?entry=${e.id}`,
        icon: <ListChecks />
      },
      {
        id: `${entry.id}:command`,
        title: 'Copy the command',
        subtitle: runCommand(e.id),
        icon: <Copy />,
        run: () => ctx.copy(runCommand(e.id), 'Command copied')
      }
    );
  }
  if (href) {
    out.push({
      id: `${entry.id}:link`,
      title: 'Copy the link',
      icon: <Copy />,
      run: () => ctx.copy(new URL(href, window.location.origin).toString(), 'Link copied')
    });
  }
  const pinned = ctx.isPinned(entry.id);
  out.push({
    id: `${entry.id}:pin`,
    title: pinned ? 'Unpin' : 'Pin to the palette',
    icon: pinned ? <PinOff /> : <Pin />,
    run: () => ctx.togglePin(entry)
  });
  return out;
}

/** The entry an address is about, when it is about one. */
function entryOn(pathname: string, search: string): string | null {
  const docs = /^\/docs\/([^/]+)\/?$/.exec(pathname);
  if (docs && docs[1] !== 'tags') return docs[1];
  if (/^\/(visualize|structures)\/?$/.test(pathname)) return new URLSearchParams(search).get('id');
  return null;
}

/** What fits the page the reader is on: the entry they are looking at. */
export function contextEntries(
  pathname: string,
  search: string,
  entries: ReadonlyMap<string, SearchEntry>
): PaletteEntry[] {
  const id = entryOn(pathname, search);
  const e = id ? entries.get(id) : undefined;
  if (!e) return [];
  const onDocs = pathname.startsWith('/docs/');
  return [
    onDocs
      ? { id: `here:${e.id}:run`, title: `Visualize ${e.name}`, href: entryHref(e), icon: <Play /> }
      : {
          id: `here:${e.id}:docs`,
          title: `Read about ${e.name}`,
          href: docsHref(e.id),
          icon: <BookOpen />
        },
    {
      id: `here:${e.id}:quiz`,
      title: `Quiz me on ${e.name}`,
      href: `/quiz/?entry=${e.id}`,
      icon: <ListChecks />
    },
    {
      id: `here:${e.id}:command`,
      title: 'Copy the command',
      subtitle: runCommand(e.id),
      icon: <Copy />,
      run: () =>
        navigator.clipboard.writeText(runCommand(e.id)).catch(() => {
          // The command is in the row; it can be copied by hand.
        })
    }
  ];
}
