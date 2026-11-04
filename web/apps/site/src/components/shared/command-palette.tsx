'use client';

import * as React from 'react';
import { useRouter } from 'next/navigation';
import {
  CommandPalette,
  type PaletteGroup
} from '@braswelljr/algorithm-ui/algorithm/command-palette';
import {
  BookOpenIcon,
  BoxIcon,
  CompassIcon,
  CopyIcon,
  FolderIcon,
  HashIcon,
  HistoryIcon,
  KeyboardIcon,
  MoonIcon,
  RouteIcon,
  SparklesIcon
} from 'lucide-react';
import { useTheme } from 'next-themes';
import { toast } from 'sonner';
import { useSiteAlgorithm } from '@/hooks/use-site-algorithm';
import { categoryTitle, docsHref, entryHref } from '@/lib/entries';
import { keysOf } from '@/config/shortcuts';

const recentKey = 'algorithm:recent';

type Recent = { id: string; label: string; href: string };

function readRecent(): Recent[] {
  try {
    const v: unknown = JSON.parse(localStorage.getItem(recentKey) ?? '[]');
    return Array.isArray(v)
      ? v
          .filter((x): x is Recent => !!x && typeof x.id === 'string' && typeof x.href === 'string')
          .slice(0, 6)
      : [];
  } catch {
    return [];
  }
}

function remember(r: Recent) {
  try {
    const next = [r, ...readRecent().filter((x) => x.id !== r.id)].slice(0, 6);
    localStorage.setItem(recentKey, JSON.stringify(next));
  } catch {
    // Blocked storage: no recent items.
  }
}

const pages: { label: string; href: string; hint: string; keys?: string }[] = [
  {
    label: 'Catalog',
    href: '/',
    hint: 'Every algorithm and data structure',
    keys: keysOf('go.home')
  },
  { label: 'Learn', href: '/learn/', hint: 'Learning paths, in order', keys: keysOf('go.learn') },
  { label: 'Quiz', href: '/quiz/', hint: 'Predict what happens next', keys: keysOf('go.quiz') },
  { label: 'Big O', href: '/big-o/', hint: 'How cost grows, measured', keys: keysOf('go.bigo') },
  { label: 'System design', href: '/system-design/', hint: 'Case studies you can run' },
  {
    label: 'Cheat sheets',
    href: '/cheatsheets/',
    hint: 'Costs and key ideas per category',
    keys: keysOf('go.cheatsheets')
  },
  { label: 'Docs', href: '/docs/', hint: 'Every entry documented', keys: keysOf('go.docs') },
  { label: 'Visualize', href: '/visualize/', hint: 'Run anything on your data' },
  { label: 'Compare', href: '/compare/', hint: 'Entries side by side' },
  { label: 'Colors', href: '/colors/', hint: 'Themes, palettes and contrast' },
  { label: 'Keyboard shortcuts', href: '/shortcuts/', hint: 'Learn the keys, with a trainer' },
  { label: 'Components', href: '/ui/', hint: 'The component gallery' }
];

/** The command palette's contents, from the wasm catalog. */
export function SitePalette({
  open,
  onOpenChange,
  onHelp
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onHelp: () => void;
}) {
  const router = useRouter();
  const { resolvedTheme, setTheme } = useTheme();
  const algorithm = useSiteAlgorithm();
  const [recent, setRecent] = React.useState<Recent[]>([]);
  React.useEffect(() => setRecent(readRecent()), [open]);

  const groups = React.useMemo<PaletteGroup[]>(() => {
    const go = (r: Recent) => () => {
      remember(r);
      router.push(r.href);
    };
    const out: PaletteGroup[] = [];
    if (recent.length)
      out.push({
        label: 'Recent',
        items: recent.map((r) => ({
          id: `recent:${r.id}`,
          label: r.label,
          icon: <HistoryIcon />,
          onSelect: go(r)
        }))
      });
    out.push({
      label: 'Actions',
      items: [
        {
          id: 'act:theme',
          label: resolvedTheme === 'dark' ? 'Switch to light' : 'Switch to dark',
          icon: <MoonIcon />,
          keys: keysOf('theme.toggle'),
          onSelect: () => setTheme(resolvedTheme === 'dark' ? 'light' : 'dark')
        },
        {
          id: 'act:help',
          label: 'Show keyboard shortcuts',
          icon: <KeyboardIcon />,
          keys: keysOf('help.open'),
          onSelect: onHelp
        },
        {
          id: 'act:link',
          label: 'Copy the link to this page',
          icon: <CopyIcon />,
          onSelect: () =>
            navigator.clipboard.writeText(location.href).then(
              () => toast.success('Link copied'),
              () => toast.error('Could not copy', { description: 'Copy it from the address bar.' })
            )
        },
        {
          id: 'act:quiz',
          label: 'Quiz me on this entry',
          hint: 'Five questions on what you are looking at',
          icon: <SparklesIcon />,
          onSelect: () => {
            const id = new URLSearchParams(location.search).get('id');
            router.push(id ? `/quiz/?id=${encodeURIComponent(id)}&count=5` : '/quiz/');
          }
        }
      ]
    });
    out.push({
      label: 'Pages',
      items: pages.map((p) => ({
        id: `page:${p.href}`,
        label: p.label,
        hint: p.hint,
        keys: p.keys,
        icon: <CompassIcon />,
        onSelect: go({ id: `page:${p.href}`, label: p.label, href: p.href })
      }))
    });
    if (algorithm.status === 'ready') {
      const { catalog, api } = algorithm;
      out.push({
        label: 'Entries',
        items: catalog.entries.map((e) => ({
          id: `entry:${e.id}`,
          label: e.name,
          hint: `${e.time} · ${e.summary}`,
          keywords: [e.id, e.category, ...e.tags],
          icon: e.input === 'ops' ? <BoxIcon /> : <BookOpenIcon />,
          onSelect: go({ id: `entry:${e.id}`, label: e.name, href: entryHref(e) })
        }))
      });
      const paths = api.paths();
      if (paths.ok)
        out.push({
          label: 'Learning paths',
          items: paths.value.map((p) => ({
            id: `path:${p.id}`,
            label: p.title,
            hint: p.summary,
            icon: <RouteIcon />,
            onSelect: go({ id: `path:${p.id}`, label: p.title, href: `/learn/path/?id=${p.id}` })
          }))
        });
      out.push({
        label: 'Categories',
        items: catalog.categories.map((c) => ({
          id: `cat:${c.id}`,
          label: categoryTitle(c),
          hint: c.summary,
          icon: <FolderIcon />,
          onSelect: go({
            id: `cat:${c.id}`,
            label: categoryTitle(c),
            href: `/cheatsheets/#${c.id}`
          })
        }))
      });
      out.push({
        label: 'Tags',
        items: catalog.tags.map((t) => ({
          id: `tag:${t.tag}`,
          label: `#${t.tag}`,
          hint: t.meaning,
          icon: <HashIcon />,
          onSelect: go({ id: `tag:${t.tag}`, label: `#${t.tag}`, href: `/docs/tags/#${t.tag}` })
        }))
      });
      out.push({
        label: 'Docs',
        items: catalog.entries.map((e) => ({
          id: `docs:${e.id}`,
          label: `${e.name} docs`,
          keywords: [e.id],
          icon: <BookOpenIcon />,
          onSelect: go({ id: `docs:${e.id}`, label: `${e.name} docs`, href: docsHref(e.id) })
        }))
      });
    }
    return out;
  }, [algorithm, recent, resolvedTheme, router, setTheme, onHelp]);

  return (
    <CommandPalette
      open={open}
      onOpenChange={onOpenChange}
      groups={groups}
    />
  );
}
