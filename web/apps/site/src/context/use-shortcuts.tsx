'use client';

import * as React from 'react';
import dynamic from 'next/dynamic';
import { useRouter } from 'next/navigation';
import { isTypingTarget, type ShortcutScope } from '@braswelljr/algorithm-ui/lib/shortcuts';
import { useTheme } from 'next-themes';
import { useHotkeys } from 'react-hotkeys-hook';
import { toast } from 'sonner';
import { keysOf, shortcuts } from '@/config/shortcuts';

// The palette and the help load on first use, not with every page.
const SitePalette = dynamic(() => import('@/components/shared/command-palette').then((m) => m.SitePalette), {
  ssr: false
});
const ShortcutOverlay = dynamic(
  () =>
    import('@braswelljr/algorithm-ui/algorithm/shortcut-overlay').then((m) => m.ShortcutOverlay),
  { ssr: false }
);

type Shortcuts = {
  openPalette: () => void;
  openHelp: () => void;
  /** Pages say which scopes they add (the run page: player), for the help. */
  useScope: (scope: ShortcutScope) => void;
};

const Context = React.createContext<Shortcuts | null>(null);

/** The keyboard layer: open the palette or help, and register a page's scope. */
export function useShortcuts(): Shortcuts {
  const c = React.useContext(Context);
  if (!c) throw new Error('useShortcuts outside ShortcutsProvider');
  return c;
}

/** Never fire while the visitor is typing in a field. */
const ignoreTyping = {
  enableOnFormTags: false,
  ignoreEventWhen: (e: KeyboardEvent) => isTypingTarget(e.target)
};

const tipsKey = 'algorithm:shortcut-tips';
const MAX_TIPS = 3;

function readTips(): Record<string, number> {
  try {
    const v: unknown = JSON.parse(localStorage.getItem(tipsKey) ?? '{}');
    return v && typeof v === 'object' ? (v as Record<string, number>) : {};
  } catch {
    return {};
  }
}

/**
 * Global shortcuts from the registry, the command palette (⌘K), the help (?),
 * and shortcut tips: the first few times a visitor clicks something that has
 * a key (an element with data-shortcut="id"), a toast tells them the key,
 * until they dismiss it or have seen it three times.
 */
export function ShortcutsProvider({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const { resolvedTheme, setTheme } = useTheme();
  const [palette, setPalette] = React.useState(false);
  const [help, setHelp] = React.useState(false);
  const [scopes, setScopes] = React.useState<ShortcutScope[]>([]);

  const useScope = React.useCallback((scope: ShortcutScope) => {
    // A hook for pages: add the scope while mounted.
    // eslint-disable-next-line react-hooks/rules-of-hooks
    React.useEffect(() => {
      setScopes((s) => (s.includes(scope) ? s : [...s, scope]));
      return () => setScopes((s) => s.filter((x) => x !== scope));
    }, [scope]);
  }, []);

  const value = React.useMemo<Shortcuts>(
    () => ({ openPalette: () => setPalette(true), openHelp: () => setHelp(true), useScope }),
    [useScope]
  );

  useHotkeys(
    keysOf('palette.open'),
    (e) => {
      e.preventDefault();
      setPalette((o) => !o);
    },
    { enableOnFormTags: true, preventDefault: true }
  );
  useHotkeys(keysOf('help.open'), () => setHelp((o) => !o), { ...ignoreTyping, useKey: true });
  useHotkeys(
    keysOf('search.focus'),
    (e) => {
      const box = document.querySelector<HTMLInputElement>('[data-search], input[type=search]');
      if (box) {
        e.preventDefault();
        box.focus();
      } else setPalette(true);
    },
    ignoreTyping
  );
  useHotkeys(
    keysOf('theme.toggle'),
    () => setTheme(resolvedTheme === 'dark' ? 'light' : 'dark'),
    ignoreTyping
  );
  const go: [string, string][] = [
    ['go.home', '/'],
    ['go.learn', '/learn/'],
    ['go.quiz', '/quiz/'],
    ['go.docs', '/docs/'],
    ['go.bigo', '/big-o/'],
    ['go.cheatsheets', '/cheatsheets/']
  ];
  for (const [id, href] of go) {
    // The list is fixed, so the hooks are called in the same order every render.
    // eslint-disable-next-line react-hooks/rules-of-hooks
    useHotkeys(keysOf(id), () => router.push(href), ignoreTyping);
  }

  // Shortcut tips for pointer users.
  React.useEffect(() => {
    const onClick = (e: MouseEvent) => {
      if (e.detail === 0) return; // a keyboard "click" (Enter or Space)
      const el = (e.target as Element | null)?.closest?.('[data-shortcut]');
      const id = el?.getAttribute('data-shortcut');
      const s = id ? shortcuts.find((x) => x.id === id) : undefined;
      if (!s) return;
      const seen = readTips();
      if ((seen[s.id] ?? 0) >= MAX_TIPS) return;
      seen[s.id] = (seen[s.id] ?? 0) + 1;
      try {
        localStorage.setItem(tipsKey, JSON.stringify(seen));
      } catch {
        // Blocked storage: the tip may show again next visit.
      }
      toast(`Tip: ${s.description.toLowerCase()} with a key`, {
        id: `tip-${s.id}`,
        description: `Press ${s.keys.replace('mod', navigator.platform.includes('Mac') ? '⌘' : 'Ctrl')}. Press ? for every shortcut.`,
        action: {
          label: 'Got it',
          onClick: () => {
            const all = readTips();
            all[s.id] = MAX_TIPS;
            try {
              localStorage.setItem(tipsKey, JSON.stringify(all));
            } catch {
              // Nothing to do.
            }
          }
        }
      });
    };
    document.addEventListener('click', onClick);
    return () => document.removeEventListener('click', onClick);
  }, []);

  return (
    <Context.Provider value={value}>
      {children}
      {palette ? (
        <SitePalette
          open={palette}
          onOpenChange={setPalette}
          onHelp={() => setHelp(true)}
        />
      ) : null}
      {help ? (
        <ShortcutOverlay
          open={help}
          onOpenChange={setHelp}
          shortcuts={shortcuts}
          scopes={scopes}
          trainerHref="/shortcuts/"
        />
      ) : null}
    </Context.Provider>
  );
}
