'use client';

/**
 * Every entry's cost and key idea as one data-table (the data-table from
 * gsl-special-courses, set up the way its list pages are): search, facets for
 * domain, category and level, sortable columns, hidden columns, CSV export and
 * pages, all kept in the address so a filtered view can be shared.
 */
import { useCallback, useMemo } from 'react';
import Link from 'next/link';
import { useTable } from '@tanstack/react-table';
import { Info } from 'lucide-react';
import { parseAsArrayOf, parseAsInteger, parseAsString, useQueryStates } from 'nuqs';
import { HiMagnifyingGlass } from 'react-icons/hi2';
import { LuPrinter, LuScrollText } from 'react-icons/lu';
import { LevelBadge } from '@/components/algorithm/level-badge';
import {
  DataTableBody,
  DataTableColumnHeader,
  DataTableColumnSortMenu,
  DataTableColumnTitle,
  DataTable as DataTableContainer,
  DataTableEmptyBody,
  DataTableHeader,
  DataTablePaginationControls,
  DataTableRoot,
  DataTableToolbar,
  TableClearFilter,
  TableFacet,
  TableViewMenu
} from '@/components/data-table';
import { TableExportButton } from '@/components/data-table/filters/table-export-button';
import { CellText } from '@/components/shared/cell-text';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle
} from '@/components/ui/empty';
import { InputGroup, InputGroupAddon, InputGroupInput } from '@/components/ui/input-group';
import { docsHref } from '@/lib/entries';
import { appTableFeatures, type AppColumnDef } from '@/lib/table';

/** One row: an entry with what a cheat sheet shows. */
export type CheatRow = {
  id: string;
  name: string;
  topic: string;
  category: string;
  level: string;
  time: string;
  space: string;
  keyIdea: string;
};

const header = (title: string) =>
  function Header() {
    return (
      <DataTableColumnHeader>
        <DataTableColumnTitle className="text-sm uppercase">
          <span className="text-sm uppercase">{title}</span>
        </DataTableColumnTitle>
        <DataTableColumnSortMenu />
      </DataTableColumnHeader>
    );
  };

const columns: AppColumnDef<CheatRow>[] = [
  {
    id: 'name',
    accessorKey: 'name',
    header: header('Entry'),
    cell: ({ row }) => (
      <div className="flex min-w-0 flex-col gap-0.5">
        <Link
          href={docsHref(row.original.id)}
          className="text-fg font-medium hover:underline"
        >
          {row.original.name}
        </Link>
        {row.original.keyIdea ? (
          <CellText className="text-muted text-sm">{row.original.keyIdea}</CellText>
        ) : null}
      </div>
    )
  },
  {
    id: 'topic',
    accessorKey: 'topic',
    header: header('Domain'),
    filterFn: 'arrIncludesSome',
    cell: ({ getValue }) => <span className="text-sm">{String(getValue())}</span>
  },
  {
    id: 'category',
    accessorKey: 'category',
    header: header('Category'),
    filterFn: 'arrIncludesSome',
    cell: ({ getValue }) => <span className="text-sm">{String(getValue())}</span>
  },
  {
    id: 'level',
    accessorKey: 'level',
    header: header('Level'),
    filterFn: 'arrIncludesSome',
    cell: ({ getValue }) => <LevelBadge level={String(getValue())} />
  },
  {
    id: 'time',
    accessorKey: 'time',
    header: header('Time'),
    cell: ({ getValue }) => <span className="font-mono text-sm">{String(getValue())}</span>
  },
  {
    id: 'space',
    accessorKey: 'space',
    header: header('Space'),
    cell: ({ getValue }) => <span className="font-mono text-sm">{String(getValue())}</span>
  }
];

const options = (rows: CheatRow[], key: 'topic' | 'category' | 'level') =>
  [...new Set(rows.map((r) => r[key]))]
    .filter(Boolean)
    .map((value) => ({ value, label: value.charAt(0).toUpperCase() + value.slice(1) }));

export function CheatSheetTable({ rows }: { rows: CheatRow[] }) {
  const [url, setUrl] = useQueryStates({
    query: parseAsString.withDefault(''),
    page: parseAsInteger.withDefault(1),
    size: parseAsInteger.withDefault(25),
    domain: parseAsArrayOf(parseAsString).withDefault([]),
    category: parseAsArrayOf(parseAsString).withDefault([]),
    level: parseAsArrayOf(parseAsString).withDefault([])
  });
  const columnFilters = useMemo(
    () =>
      [
        { id: 'topic', value: url.domain },
        { id: 'category', value: url.category },
        { id: 'level', value: url.level }
      ].filter((f) => f.value.length > 0),
    [url.domain, url.category, url.level]
  );

  const table = useTable({
    features: appTableFeatures,
    data: rows,
    columns,
    getRowId: (row) => row.id,
    onGlobalFilterChange: (value) =>
      void setUrl({ query: typeof value === 'string' ? value : '', page: 1 }),
    onColumnFiltersChange: (update) => {
      const next = typeof update === 'function' ? update(columnFilters) : update;
      const of = (id: string) =>
        (next.find((f) => f.id === id)?.value as string[] | undefined) ?? [];
      void setUrl({ domain: of('topic'), category: of('category'), level: of('level'), page: 1 });
    },
    onPaginationChange: (update) => {
      const current = { pageIndex: url.page - 1, pageSize: url.size };
      const next = typeof update === 'function' ? update(current) : update;
      void setUrl({ page: next.pageIndex + 1, size: next.pageSize });
    },
    globalFilterFn: (row, _id, value) => {
      const q = String(value).toLowerCase();
      const r = row.original;
      return [r.name, r.id, r.category, r.topic, r.keyIdea, r.time].some((s) =>
        s.toLowerCase().includes(q)
      );
    },
    state: {
      globalFilter: url.query,
      columnFilters,
      pagination: { pageIndex: url.page - 1, pageSize: url.size }
    }
  });

  // Print every row, not only the page on screen: show them all, print, then
  // put the page size back once the print dialog closes.
  const print = useCallback(() => {
    const size = table.state.pagination.pageSize;
    table.setPageSize(rows.length);
    window.addEventListener('afterprint', () => table.setPageSize(size), { once: true });
    requestAnimationFrame(() => window.print());
  }, [table, rows.length]);

  const filtered = url.query !== '' || columnFilters.length > 0;
  const shown = table.getFilteredRowModel().rows.length;

  return (
    <div className="flex w-full min-w-0 flex-col gap-4">
      <DataTableToolbar
        search={
          <InputGroup
            size="sm"
            className="w-full max-w-lg overflow-hidden"
          >
            <InputGroupAddon>
              <HiMagnifyingGlass className="text-muted size-4" />
            </InputGroupAddon>
            <InputGroupInput
              placeholder="Search entries, costs, ideas..."
              aria-label="Search the cheat sheets"
              value={url.query}
              onChange={(e) => table.setGlobalFilter(e.target.value)}
            />
          </InputGroup>
        }
      >
        <TableFacet
          column={table.getColumn('topic')}
          title="Domain"
          options={options(rows, 'topic')}
        />
        <TableFacet
          column={table.getColumn('category')}
          title="Category"
          options={options(rows, 'category')}
        />
        <TableFacet
          column={table.getColumn('level')}
          title="Level"
          options={options(rows, 'level')}
        />
        <TableClearFilter table={table} />
        <TableViewMenu table={table} />
        <TableExportButton
          table={table}
          filename="algorithm-cheat-sheet"
          useHeaderLabels
          variant="secondary"
          size="sm"
        />
        <Button
          variant="secondary"
          size="sm"
          onClick={print}
          className="print:hidden"
        >
          <LuPrinter aria-hidden />
          Print
        </Button>
      </DataTableToolbar>

      {filtered ? (
        <Alert
          variant="info"
          className="print:hidden"
        >
          <Info aria-hidden />
          <AlertDescription>
            Showing {shown} of {rows.length} entries. Export and Print take what the filters show;
            clear them for the whole catalog.
          </AlertDescription>
        </Alert>
      ) : null}

      <DataTableRoot
        table={table}
        columns={columns}
      >
        {/* A table keeps its columns: under 52rem it scrolls sideways rather than
            squeezing them. */}
        <DataTableContainer className="[&_table]:min-w-[52rem]">
          <DataTableHeader sticky />
          <DataTableBody>
            <DataTableEmptyBody>
              <Empty>
                <EmptyHeader>
                  <EmptyMedia variant="icon">
                    <LuScrollText />
                  </EmptyMedia>
                  <EmptyTitle>Nothing matches</EmptyTitle>
                  <EmptyDescription>Try fewer words, or clear the filters.</EmptyDescription>
                </EmptyHeader>
              </Empty>
            </DataTableEmptyBody>
          </DataTableBody>
        </DataTableContainer>
      </DataTableRoot>

      <DataTablePaginationControls table={table} />
    </div>
  );
}
