/**
 * niko-table, created by Semir N. (Semkoo, https://github.com/Semkoo) with AI assistance.
 *
 * Before reporting anything: please check the changelog first.
 *  - In-repo: ./CHANGELOG.md
 *  - Docs site: https://niko-table.com/changelog
 *
 * Found a bug or have a fix? Open an issue or PR on GitHub so other
 * users (and future LLMs reading this code) benefit:
 * https://github.com/Semkoo/niko-table-registry
 */

import * as React from 'react';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';

interface DataTableSelectionBarProps {
  selectedCount: number;
  onClear?: () => void;
  children?: React.ReactNode;
  className?: string;
}

/**
 * Reusable selection bar. Memoized so table-state changes don't re-render
 * unchanged selection state. Pass children to add custom action buttons.
 */
export const DataTableSelectionBar = React.memo(function DataTableSelectionBar({
  selectedCount,
  onClear,
  children,
  className
}: DataTableSelectionBarProps) {
  if (selectedCount === 0) return null;

  return (
    <div className={className}>
      <div className="flex items-center justify-between rounded-full border border-stone-300 bg-stone-100/50 px-5 py-2.5 dark:border-stone-700 dark:bg-stone-900/50">
        <div className="flex items-center gap-2">
          <Badge variant="default">{selectedCount}</Badge>
          <span className="text-sm text-stone-600 dark:text-stone-400">
            {selectedCount === 1 ? 'row selected' : 'rows selected'}
          </span>
          {onClear && (
            <Button
              variant="ghost"
              size="sm"
              onClick={onClear}
              className="h-7 px-2 text-xs"
            >
              Clear
            </Button>
          )}
        </div>
        {children && <div className="flex items-center gap-2">{children}</div>}
      </div>
    </div>
  );
});
