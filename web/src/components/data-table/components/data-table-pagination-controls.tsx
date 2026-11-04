/**
 * Pagination for a table that pages on the server.
 *
 * Moved here from `components/tables` unchanged. It differs from
 * `DataTablePagination` in where it reads the counts: that one reads the
 * table, this one takes the query's own page/size/total, which is what a
 * table under `manualPagination` actually knows. Both are the library's;
 * a screen picks by whether the server or the table owns the count.
 */
import type { RowData } from '@tanstack/react-table';
import { LuChevronFirst, LuChevronLast } from 'react-icons/lu';
import { RxChevronLeft, RxChevronRight } from 'react-icons/rx';
import { NumberInput } from '@/components/shared/number-input';
import { Button } from '@/components/ui/button';
import { ButtonGroup } from '@/components/ui/button-group';
import { Select, SelectContent, SelectItem, SelectTrigger } from '@/components/ui/select';
import type { AppReactTable } from '@/lib/table';
import { cn } from '@/lib/utils';

// Shape used by the pagination controls. Pages pass a nuqs `useQueryStates`
// object whose fields are optional/nullable, so keep everything optional here.
type TablePagination = {
  page?: number | null;
  size?: number | null;
  items?: number | null;
  pages?: number | null;
  hasNextPage?: boolean | null;
  hasPreviousPage?: boolean | null;
};

// Only `isLoading` is read, so the prop is typed structurally rather than as
// react-query's `UseQueryResult`: a `useQuery` result still satisfies it, and
// the library keeps out of the data layer's dependency list.
const MAX_PAGE_SIZE = 200;
const DEFAULT_PAGE_SIZE = 20;
const PAGE_SIZE_OPTIONS = [5, 10, 20, 30, 40, 50, 100, MAX_PAGE_SIZE];

type TableQueryState = { isLoading?: boolean };

type PaginationControlsProps<TData extends RowData> = {
  table: AppReactTable<TData>;
  className?: string;
  pagination?: TablePagination;
  // Accepts a nuqs `setValues` updater as well as a React setState updater.
  setPagination?: (...args: any[]) => unknown;
  query?: TableQueryState;
};

export function DataTablePaginationControls<TData extends RowData>({
  table,
  className,
  pagination,
  setPagination,
  query
}: PaginationControlsProps<TData>) {
  const currentPage = pagination?.page ?? table.state.pagination.pageIndex + 1;
  const totalPages = pagination?.pages ?? table.getPageCount();
  const pageSize = pagination?.size ?? table.state.pagination.pageSize;
  const hasPreviousPage = pagination?.hasPreviousPage ?? table.getCanPreviousPage();
  const hasNextPage = pagination?.hasNextPage ?? table.getCanNextPage();
  const totalItems = pagination?.items ?? table.getRowModel().rows.length;

  const handlePageSizeChange = (value: string) => {
    const requested = Number(value);
    const newSize = Number.isFinite(requested)
      ? Math.min(Math.max(Math.trunc(requested), 1), MAX_PAGE_SIZE)
      : DEFAULT_PAGE_SIZE;

    // Update both client and server pagination
    table.setPageSize(newSize);

    if (setPagination) {
      setPagination((prev: TablePagination) => ({
        ...prev,
        size: newSize,
        page: 1 // Reset to first page when changing page size
      }));
    }
  };

  const goToFirstPage = () => {
    if (setPagination) {
      setPagination((prev: TablePagination) => ({ ...prev, page: 1 }));
    } else {
      table.setPageIndex(0);
    }
  };

  const goToPreviousPage = () => {
    if (setPagination && hasPreviousPage) {
      setPagination((prev: TablePagination) => ({ ...prev, page: (prev.page ?? 1) - 1 }));
    } else if (hasPreviousPage) {
      table.previousPage();
    }
  };

  const goToNextPage = () => {
    if (setPagination && hasNextPage) {
      setPagination((prev: TablePagination) => ({ ...prev, page: (prev.page ?? 1) + 1 }));
    } else if (hasNextPage) {
      table.nextPage();
    }
  };

  const goToLastPage = () => {
    if (setPagination) {
      setPagination((prev: TablePagination) => ({ ...prev, page: prev.pages }));
    } else {
      table.setPageIndex(table.getPageCount() - 1);
    }
  };

  return (
    <div
      className={cn(
        'flex items-center justify-between px-2 max-sm:flex-col max-sm:items-start max-sm:space-y-3',
        className
      )}
    >
      <div className="flex-1 text-sm text-stone-600 dark:text-stone-400">
        {table.getFilteredSelectedRowModel().rows.length} of{' '}
        {table.getFilteredRowModel().rows.length} row(s) selected{' '}
        {totalItems > 0 && `of ${totalItems}`}.
      </div>

      <div className="relative flex items-center not-[@media_all_and_(min-width:_550px)]:flex-wrap not-[@media_all_and_(min-width:_550px)]:space-y-4 max-sm:w-full sm:space-x-6 lg:space-x-8">
        <div className="flex items-center space-x-2 not-[@media_all_and_(min-width:_550px)]:w-full max-sm:w-3/5">
          <p className="text-sm font-medium">Rows per page</p>
          <ButtonGroup className="h-8 w-max">
            <NumberInput
              key={pageSize}
              size="sm"
              min={1}
              max={MAX_PAGE_SIZE}
              defaultValue={pageSize}
              onValueCommitted={(n) => handlePageSizeChange(String(n ?? DEFAULT_PAGE_SIZE))}
              aria-label="Rows per page"
              className="w-auto shrink-0"
              groupClassName="w-28"
            />
            <Select
              value={`${pageSize}`}
              onValueChange={(val) => handlePageSizeChange(val ?? `${DEFAULT_PAGE_SIZE}`)}
              disabled={query?.isLoading}
            >
              <SelectTrigger
                size="sm"
                className="h-8! w-max shrink-0"
              />
              <SelectContent
                side="top"
                align="end"
              >
                {PAGE_SIZE_OPTIONS.map((_pageSize) => (
                  <SelectItem
                    key={_pageSize}
                    value={`${_pageSize}`}
                  >
                    {_pageSize}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </ButtonGroup>
        </div>

        <div className="flex w-32 items-center justify-center text-sm font-medium not-[@media_all_and_(min-width:_550px)]:w-1/2 not-[@media_all_and_(min-width:_550px)]:justify-start">
          Page {currentPage} of {totalPages}
        </div>

        <div className="relative flex items-center justify-end space-x-2 not-[@media_all_and_(min-width:_550px)]:w-1/2">
          <Button
            rounded="full"
            variant="ghost"
            className="size-8 border-current p-0 text-current disabled:text-stone-500!"
            onClick={goToFirstPage}
            disabled={!hasPreviousPage || query?.isLoading}
          >
            <span className="sr-only">Go to first page</span>
            <LuChevronFirst className="cn-rtl-flip size-4" />
          </Button>

          <Button
            rounded="full"
            variant="ghost"
            className="size-8 border-current p-0 text-current disabled:text-stone-500!"
            onClick={goToPreviousPage}
            disabled={!hasPreviousPage || query?.isLoading}
          >
            <span className="sr-only">Go to previous page</span>
            <RxChevronLeft className="cn-rtl-flip size-4" />
          </Button>

          <Button
            rounded="full"
            variant="ghost"
            className="size-8 border-current p-0 text-current disabled:text-stone-500!"
            onClick={goToNextPage}
            disabled={!hasNextPage || query?.isLoading}
          >
            <span className="sr-only">Go to next page</span>
            <RxChevronRight className="cn-rtl-flip size-4" />
          </Button>

          <Button
            rounded="full"
            variant="ghost"
            className="size-8 border-current p-0 text-current disabled:text-stone-500!"
            onClick={goToLastPage}
            disabled={!hasNextPage || query?.isLoading}
          >
            <span className="sr-only">Go to last page</span>
            <LuChevronLast className="cn-rtl-flip size-4" />
          </Button>
        </div>
      </div>
    </div>
  );
}
