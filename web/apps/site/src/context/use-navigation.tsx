'use client';

import { createContext, useContext, useMemo, type ReactNode } from 'react';
import { usePathname } from 'next/navigation';
import { buildNavigation } from '@/config/navigation';
import type { NavigationGroup, NavigationItem } from '@/types/navigation';

// The menu on the current address. Built once per address and read by the
// sidebar, the top bar's title and the trail, so they never disagree.

type NavigationContextValue = {
  /** The sections above the footer. */
  groups: NavigationGroup[];
  /** The Design section, kept apart for the sidebar's footer. */
  design: NavigationItem[];
  /** The page the reader is on, when the menu lists it. */
  active: NavigationItem | null;
};

const NavigationContext = createContext<NavigationContextValue>({
  groups: [],
  design: [],
  active: null
});

export function NavigationProvider({ children }: { children?: ReactNode }) {
  const pathname = usePathname() ?? '/';
  const value = useMemo<NavigationContextValue>(() => {
    const all = buildNavigation(pathname);
    const groups = all.filter((g) => g.label !== 'Design');
    const design = all.find((g) => g.label === 'Design')?.items ?? [];
    const active = all.flatMap((g) => g.items).find((i) => i.isActive) ?? null;
    return { groups, design, active };
  }, [pathname]);
  return <NavigationContext.Provider value={value}>{children}</NavigationContext.Provider>;
}

/** The menu on the current address. */
export function useNavigation(): NavigationContextValue {
  return useContext(NavigationContext);
}
