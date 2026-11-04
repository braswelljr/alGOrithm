'use client';

import { prefetchSearch } from '@/hooks/use-command-search';

/**
 * The command palette's code and the search index, fetched before they are
 * needed: when the page is idle, and when the pointer reaches the search
 * button. Both are cached, so asking twice costs nothing.
 */
export function prefetchPalette(): void {
  void import('@/components/shared/command-palette').catch(() => {});
  prefetchSearch();
}
