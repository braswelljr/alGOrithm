'use client';

/**
 * Everything the command palette offers before anything is typed: where the
 * reader can go, what they can start, and how the site looks (after
 * gsl-special-courses' file of the same name).
 *
 * Navigation is read from the sidebar's own list (src/config/navigation.ts)
 * rather than kept as a second one, so a page added to the sidebar is
 * searchable here without anybody remembering.
 */
import { useMemo } from 'react';
import { usePathname } from 'next/navigation';
import {
  Columns3,
  Laptop,
  Link2,
  ListChecks,
  Moon,
  Rows3,
  Rows4,
  Settings,
  Sun
} from 'lucide-react';
import { useTheme } from 'next-themes';
import { FiKey } from 'react-icons/fi';
import { toast } from 'sonner';
import { useThemeToggle } from '@/components/animated/theme-toggle';
import { GitHubIcon } from '@/components/brand/github-icon';
import { NAVIGATION } from '@/config/navigation';
import { repoUrl } from '@/config/site';
import type { RecordEntry } from '@/hooks/use-command-search';
import { usePreferencesStore, type Density } from '@/store/preferences';

/** One row in the palette. It either goes somewhere or does something. */
export type PaletteEntry = {
  id: string;
  title: string;
  /** The second line: where it is, what it is about. */
  subtitle?: string;
  icon: React.ReactNode;
  /** Other words it should be found by, never shown. */
  keywords?: string[];
  /** Shown at the end of the row: the kind of record, or a key to press. */
  hint?: string;
  /** For a choice, whether it is the current one. */
  checked?: boolean;
  /** Where selecting it goes, when it goes somewhere. */
  href?: string;
  /** What selecting it does, when it does something. */
  run?: () => void;
  /** For a found record, the record itself: its row, preview and actions. */
  record?: RecordEntry;
  /** What kind of thing it is, remembered with a recent or pinned row. */
  kind?: string;
};

/** One labelled block of rows. */
export type PaletteGroup = { value: string; items: PaletteEntry[] };

/** Every page in the sidebar, as palette rows. Pure, so the rule "the palette
 *  offers what the sidebar offers" can be tested without rendering either. */
export function navigationEntries(): PaletteEntry[] {
  return NAVIGATION.flatMap((section) =>
    section.items.map((item) => {
      const Icon = item.icon;
      return {
        id: `nav:${item.href}`,
        title: item.label,
        subtitle: section.label,
        href: item.href,
        kind: 'page',
        icon: <Icon aria-hidden />
      };
    })
  );
}

/** Copy text and say so, or say it could not be copied. */
export function copyText(text: string, message: string) {
  navigator.clipboard.writeText(text).then(
    () => toast.success(message),
    () => toast.error('Could not copy', { description: 'Select it and copy it by hand.' })
  );
}

const DENSITIES: { value: Density; label: string; icon: React.ReactNode }[] = [
  { value: 'compact', label: 'Compact', icon: <Rows4 /> },
  { value: 'standard', label: 'Standard', icon: <Rows3 /> },
  { value: 'comfortable', label: 'Comfortable', icon: <Columns3 /> }
];

/** The settings pages, as rows. */
const SETTINGS = [
  { to: 'appearance', label: 'Appearance', description: 'Theme, text size and density' },
  { to: 'accessibility', label: 'Accessibility', description: 'Contrast, motion, focus, spacing' }
];

/**
 * The palette's fixed groups, in the order they are shown: Navigation,
 * Actions, Settings, Appearance.
 */
export function usePaletteCommands({ onHelp }: { onHelp: () => void }): PaletteGroup[] {
  const pathname = usePathname() ?? '/';
  const { theme } = useTheme();
  const { switchTo } = useThemeToggle({ variant: 'diagonal' });
  const { density, setDensity } = usePreferencesStore();

  return useMemo(() => {
    const actions: PaletteEntry[] = [
      {
        id: 'action:copy-link',
        title: 'Copy link to this page',
        subtitle: pathname,
        icon: <Link2 />,
        keywords: ['share', 'url', 'address'],
        run: () => copyText(window.location.href, 'Link copied')
      },
      {
        id: 'action:quiz',
        title: 'Start a quiz',
        subtitle: 'Predict what happens next, on real runs',
        href: '/quiz/',
        icon: <ListChecks />,
        keywords: ['game', 'practice', 'test', 'predict']
      },
      {
        id: 'action:shortcuts',
        title: 'Keyboard shortcuts',
        subtitle: 'Every key, by page',
        icon: <FiKey />,
        keywords: ['keys', 'hotkeys', 'keyboard', 'help'],
        run: onHelp
      },
      {
        id: 'action:source',
        title: 'Source code on GitHub',
        subtitle: repoUrl.replace(/^https:\/\//, ''),
        icon: <GitHubIcon />,
        keywords: ['repository', 'code', 'contribute'],
        run: () => window.open(repoUrl, '_blank', 'noopener')
      }
    ];

    const current = theme ?? 'system';
    const appearance: PaletteEntry[] = [
      { value: 'light', label: 'Light theme', icon: <Sun /> },
      { value: 'dark', label: 'Dark theme', icon: <Moon /> },
      { value: 'system', label: 'Match the device', icon: <Laptop /> }
    ].map((option) => ({
      id: `theme:${option.value}`,
      title: option.label,
      icon: option.icon,
      keywords: ['theme', 'mode', 'color', option.value],
      checked: current === option.value,
      run: () => switchTo(option.value as 'light' | 'dark' | 'system')
    }));
    for (const option of DENSITIES) {
      appearance.push({
        id: `density:${option.value}`,
        title: `${option.label} density`,
        icon: option.icon,
        keywords: ['density', 'spacing', 'layout'],
        checked: density === option.value,
        run: () => setDensity(option.value)
      });
    }

    const settings: PaletteEntry[] = SETTINGS.map((tab) => ({
      id: `settings:${tab.to}`,
      title: tab.label,
      subtitle: tab.description,
      href: `/settings/${tab.to}/`,
      icon: <Settings />,
      keywords: ['settings', 'preferences']
    }));

    return [
      { value: 'Navigation', items: navigationEntries() },
      { value: 'Actions', items: actions },
      { value: 'Settings', items: settings },
      { value: 'Appearance', items: appearance }
    ];
  }, [pathname, theme, density, setDensity, onHelp, switchTo]);
}
