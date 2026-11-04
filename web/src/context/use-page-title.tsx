'use client';

import {
  createContext,
  useContext,
  useEffect,
  useMemo,
  useState,
  type Context,
  type ReactNode
} from 'react';

// The dashboard chrome owns the page's name in the navbar. A page announces it
// with `useSetPageTitle(...)`; a page that announces nothing leaves the navbar
// on the name its menu entry gives it. This is not the document title, which
// each route's `meta()` sets.

type PageTitleContextValue = {
  /** The title the active page announced, or null when it announced none. */
  title: string | null;
  setTitle: (title: string | null) => void;
};

export const PageTitleContext: Context<PageTitleContextValue> =
  createContext<PageTitleContextValue>({ title: null, setTitle: () => {} });

export function PageTitleProvider({ children }: { children?: ReactNode }) {
  const [title, setTitle] = useState<string | null>(null);
  const value = useMemo<PageTitleContextValue>(() => ({ title, setTitle }), [title]);
  return <PageTitleContext.Provider value={value}>{children}</PageTitleContext.Provider>;
}

/** The announced page title, for the chrome. */
export function usePageTitle(): PageTitleContextValue {
  return useContext(PageTitleContext);
}

/**
 * Announce the active page's title to the dashboard chrome. Pass a nullish
 * title while the real one is still loading, and the navbar keeps its
 * fallback until it resolves.
 */
export function useSetPageTitle(title?: string | null): void {
  const { setTitle } = usePageTitle();
  useEffect(() => {
    setTitle(title ?? null);
    return () => setTitle(null);
  }, [setTitle, title]);
}
