import {
  BookOpen,
  Boxes,
  Database,
  Gamepad2,
  GitCompare,
  GraduationCap,
  LayoutGrid,
  ListChecks,
  Network,
  Play,
  ScrollText,
  Settings,
  SquareTerminal,
  TrendingUp
} from 'lucide-react';
import type { NavigationGroup, NavSection } from '@/types/navigation';

export type { NavItem, NavSection } from '@/types/navigation';

/**
 * The site's one menu, read by the sidebar, the top bar's title and the
 * breadcrumb trail. Every page has a fixed section and position, so it is
 * found in the same place on every visit.
 */
export const NAVIGATION: NavSection[] = [
  {
    label: 'Overview',
    items: [
      { label: 'Catalog', href: '/', icon: LayoutGrid },
      { label: 'Learn', href: '/learn/', icon: GraduationCap },
      { label: 'Quiz', href: '/quiz/', icon: ListChecks },
      { label: 'Cheat sheets', href: '/cheatsheets/', icon: ScrollText }
    ]
  },
  {
    label: 'Run',
    items: [
      { label: 'Visualize', href: '/visualize/', icon: Play },
      { label: 'Structures', href: '/structures/', icon: Boxes },
      { label: 'Editor', href: '/editor/', icon: SquareTerminal },
      { label: 'Compare', href: '/compare/', icon: GitCompare },
      { label: 'Your data', href: '/data/', icon: Database }
    ]
  },
  {
    label: 'Concepts',
    items: [
      { label: 'Big O', href: '/big-o/', icon: TrendingUp },
      { label: 'System design', href: '/system-design/', icon: Network }
    ]
  },
  {
    label: 'Reference',
    items: [
      { label: 'Docs', href: '/docs/', icon: BookOpen },
      { label: 'Arena', href: '/shortcuts/', icon: Gamepad2 }
    ]
  },
  {
    label: 'Settings',
    items: [
      // One entry: Appearance and Accessibility are tabs of the one Settings
      // layout, as in gsl-certificates.
      { label: 'Settings', href: '/settings/', icon: Settings }
    ]
  }
];

/** The path with the trailing slash the menu's hrefs carry. */
export function normalizePath(pathname: string): string {
  return pathname.endsWith('/') ? pathname : `${pathname}/`;
}

/** Whether `href` is the page at `pathname` or a page beneath it. */
export function isActivePath(href: string, pathname: string): boolean {
  const path = normalizePath(pathname);
  return href === '/' ? path === '/' : path.startsWith(href);
}

/** The menu as it shows on `pathname`, the open page marked. */
export function buildNavigation(pathname: string): NavigationGroup[] {
  return NAVIGATION.map((section) => ({
    label: section.label,
    items: section.items.map((item) => ({ ...item, isActive: isActivePath(item.href, pathname) }))
  }));
}
