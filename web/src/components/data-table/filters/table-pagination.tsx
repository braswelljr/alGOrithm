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

import React from 'react';
import type { RowData } from '@tanstack/react-table';
import { ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight } from 'lucide-react';
import { NumberInput } from '@/components/shared/number-input';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue
} from '@/components/ui/select';
import { Skeleton } from '@/components/ui/skeleton';
import type { AppReactTable } from '@/lib/table';

export interface TablePaginationProps<TData extends RowData> {
  table: AppReactTable<TData>;
  pageSizeOptions?: number[];
  defaultPageSize?: number;
  /**
   * External loading state (e.g., from API)
   */
  isLoading?: boolean;
  /**
   * External fetching state (e.g., from TanStack Query).
   * Disables nav while a request is in flight so users can't advance the
   * cursor before the next page resolves.
   */
  isFetching?: boolean;
  /**
   * Explicitly disable the next page button.
   * Useful when you want to prevent navigation during initial load but allow it during background fetching.
   */
  disableNextPage?: boolean;
  /**
   * Explicitly disable the previous page button.
   * Useful when you want to prevent navigation during initial load but allow it during background fetching.
   */
  disablePreviousPage?: boolean;
  /**
   * Total count of items from server (for server-side pagination).
   * If provided, this will be used instead of table.getFilteredRowModel().rows.length
   */
  totalCount?: number;
  onPageSizeChange?: (pageSize: number, pageIndex: number) => void;
  onPageChange?: (pageIndex: number) => void;
  onNextPage?: (pageIndex: number) => void;
  onPreviousPage?: (pageIndex: number) => void;
  /**
   * Callback when pagination initialization is complete
   */
  onPaginationReady?: () => void;
}
export function TablePagination<TData extends RowData>({
  table,
  pageSizeOptions = [10, 25, 50, 100, 200, 500],
  defaultPageSize = pageSizeOptions[0],
  isLoading,
  isFetching,
  disableNextPage,
  disablePreviousPage,
  totalCount,
  onPageSizeChange,
  onPageChange,
  onNextPage,
  onPreviousPage,
  onPaginationReady
}: TablePaginationProps<TData>) {
  const { pageIndex, pageSize } = table.state.pagination;

  // Use totalCount if provided (server-side), otherwise use filtered row model (client-side)
  const totalRows = totalCount ?? table.getFilteredRowModel().rows.length;
  const startItem = totalRows === 0 ? 0 : pageIndex * pageSize + 1;
  const endItem = Math.min((pageIndex + 1) * pageSize, totalRows);
  const totalPages = table.getPageCount();
  const currentPage = pageIndex + 1;

  // Disable nav during any in-flight load (initial OR background fetch) so
  // users can't advance the cursor while the next page is still resolving.
  const canNextPage = table.getCanNextPage();
  const isDisabled = isLoading || isFetching;
  const canGoNext = !disableNextPage && !isDisabled && canNextPage;
  const canGoPrevious = !disablePreviousPage && !isDisabled && table.getCanPreviousPage();

  // Set default page size on initial render
  React.useEffect(() => {
    if (pageSize !== defaultPageSize) {
      table.setPageSize(defaultPageSize);
    }
    onPaginationReady?.();
  }, []);

  const handlePageSizeChange = React.useCallback(
    // Base UI selects pass null on clear; Radix never does
    (value: string | null) => {
      if (!value) return;
      const newPageSize = Number(value);
      const newPageIndex = Math.floor((pageIndex * pageSize) / newPageSize);
      table.setPageSize(newPageSize);
      onPageSizeChange?.(newPageSize, newPageIndex);
    },
    [table, pageIndex, pageSize, onPageSizeChange]
  );

  const goToPage = React.useCallback(
    (page: number | null) => {
      if (page === null || !Number.isInteger(page) || page < 1 || page > totalPages) return;
      table.setPageIndex(page - 1);
      onPageChange?.(page - 1);
    },
    [totalPages, table, onPageChange]
  );

  const handlePreviousPage = React.useCallback(() => {
    const newPageIndex = pageIndex - 1;
    table.previousPage();
    onPreviousPage?.(newPageIndex);
  }, [table, pageIndex, onPreviousPage]);

  const handleNextPage = React.useCallback(() => {
    const newPageIndex = pageIndex + 1;
    table.nextPage();
    onNextPage?.(newPageIndex);
  }, [table, pageIndex, onNextPage]);

  // Both ends, because a long list is most often read from one or the other and
  // stepping there a page at a time is not navigation. They report through the
  // same callbacks as the step controls, so a table paginating on the server
  // hears about the jump the same way it hears about a step.
  const handleFirstPage = React.useCallback(() => {
    table.setPageIndex(0);
    onPageChange?.(0);
    onPreviousPage?.(0);
  }, [table, onPageChange, onPreviousPage]);

  const handleLastPage = React.useCallback(() => {
    const last = Math.max(0, totalPages - 1);
    table.setPageIndex(last);
    onPageChange?.(last);
    onNextPage?.(last);
  }, [table, totalPages, onPageChange, onNextPage]);

  // Show loading skeleton while initializing
  if (isLoading) {
    return (
      <div className="flex flex-wrap items-center justify-between gap-4 px-4 py-2">
        <div className="flex items-center space-x-2">
          <Skeleton className="h-8 w-24" />
          <Skeleton className="h-8 w-16" />
        </div>
        <Skeleton className="h-8 w-32" />
        <div className="flex items-center space-x-4">
          <div className="flex items-center space-x-2">
            <Skeleton className="h-8 w-12" />
            <Skeleton className="h-8 w-20" />
          </div>
          <div className="flex items-center space-x-1">
            <Skeleton className="h-8 w-8" />
            <Skeleton className="h-8 w-8" />
          </div>
        </div>
      </div>
    );
  }

  return (
    <nav
      className="flex flex-wrap items-center justify-between gap-x-6 gap-y-2 px-4 py-2"
      aria-label="Table pagination"
    >
      <div className="flex items-center space-x-2">
        <span
          className="text-sm whitespace-nowrap text-stone-600 dark:text-stone-400"
          id="pagination-page-size-label"
        >
          Items per page
        </span>
        <Select
          value={`${Number(pageSize) === 0 ? defaultPageSize : Number(pageSize)}`}
          onValueChange={handlePageSizeChange}
          disabled={isLoading}
        >
          <SelectTrigger
            size="sm"
            className="w-16 focus:ring-0"
            aria-label="Select page size"
            aria-labelledby="pagination-page-size-label"
          >
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {pageSizeOptions?.map((size) => (
              <SelectItem
                key={size}
                value={`${size}`}
              >
                {size}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      <div
        className="flex-1 text-end text-sm whitespace-nowrap text-stone-600 md:text-center dark:text-stone-400"
        role="status"
        aria-live="polite"
        aria-atomic="true"
      >
        {totalRows === 0 ? '0 items' : `${startItem}-${endItem} of ${totalRows} items`}
      </div>

      <div className="ms-auto flex items-center space-x-4">
        <div className="flex items-center space-x-2 text-sm text-stone-600 dark:text-stone-400">
          <Label
            htmlFor="page-number-input"
            className="sr-only"
          >
            Page number
          </Label>
          {/* Acts once the reader is done typing: on blur, Enter or a step. */}
          <NumberInput
            key={currentPage}
            id="page-number-input"
            size="sm"
            min={1}
            max={Math.max(totalPages, 1)}
            defaultValue={currentPage}
            onValueCommitted={goToPage}
            className="w-auto"
            groupClassName="w-28"
            disabled={totalPages < 1 || isLoading || isFetching}
            aria-label={`Page ${currentPage} of ${totalPages}`}
          />
          <span
            className="whitespace-nowrap"
            aria-hidden="true"
          >
            of {Math.max(1, totalPages)} pages
          </span>
        </div>

        <div className="flex items-center space-x-1">
          <Button
            variant="ghost"
            size="icon-lg"
            rounded="full"
            onClick={handleFirstPage}
            disabled={!canGoPrevious}
            aria-label="Go to first page"
          >
            <ChevronsLeft
              className="cn-rtl-flip h-4 w-4"
              aria-hidden="true"
            />
          </Button>
          <Button
            variant="ghost"
            size="icon-lg"
            rounded="full"
            onClick={handlePreviousPage}
            disabled={!canGoPrevious}
            aria-label={`Go to previous page, page ${pageIndex}`}
          >
            <ChevronLeft
              className="cn-rtl-flip h-4 w-4"
              aria-hidden="true"
            />
          </Button>
          <Button
            variant="ghost"
            size="icon-lg"
            rounded="full"
            onClick={handleNextPage}
            disabled={!canGoNext}
            aria-label={`Go to next page, page ${pageIndex + 2}`}
          >
            <ChevronRight
              className="cn-rtl-flip h-4 w-4"
              aria-hidden="true"
            />
          </Button>
          <Button
            variant="ghost"
            size="icon-lg"
            rounded="full"
            onClick={handleLastPage}
            disabled={!canGoNext}
            aria-label={`Go to last page, page ${Math.max(1, totalPages)}`}
          >
            <ChevronsRight
              className="cn-rtl-flip h-4 w-4"
              aria-hidden="true"
            />
          </Button>
        </div>
      </div>
    </nav>
  );
}

/**
 * @required displayName is required for auto feature detection
 * @see "feature-detection.ts"
 */

TablePagination.displayName = 'TablePagination';
