import * as React from 'react';
import { cn } from '@/lib/utils';

/**
 * The table element on its own, without the scroll container `Table` wraps it
 * in. For callers that provide their own scrolling viewport, which the data
 * table does: nesting two overflow containers breaks sticky headers and column
 * pinning, both of which position against the nearest scrolling ancestor.
 */
const TableComponent = React.forwardRef<HTMLTableElement, React.HTMLAttributes<HTMLTableElement>>(
  ({ className, ...props }, ref) => (
    <table
      ref={ref}
      data-slot="table"
      className={cn('w-full caption-bottom text-sm', className)}
      {...props}
    />
  )
);
TableComponent.displayName = 'TableComponent';

const Table = React.forwardRef<HTMLTableElement, React.HTMLAttributes<HTMLTableElement>>(
  ({ className, ...props }, ref) => (
    <div
      data-slot="table-container"
      className="relative w-full overflow-auto overscroll-x-contain"
    >
      <TableComponent
        ref={ref}
        className={className}
        {...props}
      />
    </div>
  )
);
Table.displayName = 'Table';

const TableHeader = React.forwardRef<
  HTMLTableSectionElement,
  React.HTMLAttributes<HTMLTableSectionElement>
>(({ className, ...props }, ref) => (
  <thead
    ref={ref}
    // The rule under the header, and every row's, is named rather than left to
    // Tailwind's default: v4 defaults an uncoloured border to currentColor, so
    // a bare `border-b` drew each divider in the text colour.
    className={cn(
      'bg-bg [&_tr]:border-border [&_tr]:border-b',
      // A header row is a `TableRow` like any other, so it inherits the row
      // hover. Hover marks a row you can act on; a head is not one.
      '[&_tr:hover]:bg-transparent dark:[&_tr:hover]:bg-transparent',
      className
    )}
    {...props}
  />
));
TableHeader.displayName = 'TableHeader';

const TableBody = React.forwardRef<
  HTMLTableSectionElement,
  React.HTMLAttributes<HTMLTableSectionElement>
>(({ className, ...props }, ref) => (
  <tbody
    ref={ref}
    className={cn('[&_tr:last-child]:border-0', className)}
    {...props}
  />
));
TableBody.displayName = 'TableBody';

const TableFooter = React.forwardRef<
  HTMLTableSectionElement,
  React.HTMLAttributes<HTMLTableSectionElement>
>(({ className, ...props }, ref) => (
  <tfoot
    ref={ref}
    className={cn(
      'border-border bg-subtle/50 border-t font-medium [&>tr]:last:border-b-0',
      className
    )}
    {...props}
  />
));
TableFooter.displayName = 'TableFooter';

const TableRow = React.forwardRef<HTMLTableRowElement, React.HTMLAttributes<HTMLTableRowElement>>(
  ({ className, ...props }, ref) => (
    <tr
      ref={ref}
      className={cn(
        'border-border hover:bg-bg data-[state=selected]:bg-subtle border-b transition-colors',
        className
      )}
      {...props}
    />
  )
);
TableRow.displayName = 'TableRow';

const TableHead = React.forwardRef<
  HTMLTableCellElement,
  React.ThHTMLAttributes<HTMLTableCellElement>
>(({ className, scope = 'col', ...props }, ref) => (
  <th
    ref={ref}
    scope={scope}
    className={cn(
      'text-muted sticky top-0 z-10 h-10 bg-inherit px-4 text-left align-middle text-sm font-semibold tracking-wide uppercase has-[[role=checkbox]]:pr-0',
      className
    )}
    {...props}
  />
));
TableHead.displayName = 'TableHead';

const TableCell = React.forwardRef<
  HTMLTableCellElement,
  React.TdHTMLAttributes<HTMLTableCellElement>
>(({ className, ...props }, ref) => (
  <td
    ref={ref}
    className={cn('p-4 align-middle has-[[role=checkbox]]:pr-0', className)}
    {...props}
  />
));
TableCell.displayName = 'TableCell';

const TableCaption = React.forwardRef<
  HTMLTableCaptionElement,
  React.HTMLAttributes<HTMLTableCaptionElement>
>(({ className, ...props }, ref) => (
  <caption
    ref={ref}
    className={cn('text-muted mt-4 text-sm', className)}
    {...props}
  />
));
TableCaption.displayName = 'TableCaption';

export {
  Table,
  TableComponent,
  TableHeader,
  TableBody,
  TableFooter,
  TableHead,
  TableRow,
  TableCell,
  TableCaption
};
