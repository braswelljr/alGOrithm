import type { ReactNode } from 'react';
import { cn } from '@/lib/utils';

/**
 * Template. The column every page is laid out in, so each one reads
 * like the next. The spacing standard lives here and in the layout, nowhere
 * else:
 *
 * - The layout's `<main>` owns the outer padding (`p-5 lg:p-6`), and the
 *   content runs the page's full width; a page adds no padding, margin or
 *   width cap of its own.
 * - Sections of a page sit `gap-6` apart, in this column.
 * - A panel (a card, a framed chart, a read-only summary) pads `p-5`.
 * - Tables are not put in cards unless necessary: the table's own border and
 *   header are its frame, toolbar above, pagination `gap-4` below (`PageTable`).
 * - Inputs are not put in cards: a form sits on the page (or in `DialogBody`),
 *   grouped by `FieldSet`, `FieldLegend` and headings; cards are for read-only
 *   summaries, stat tiles and choice tiles.
 */
export function PageContent({ children, className }: { children: ReactNode; className?: string }) {
  return <div className={cn('flex min-w-0 flex-1 flex-col gap-6', className)}>{children}</div>;
}

/** A register's table and its pagination, `gap-4` apart, straight on the page.
 *  It takes the height the page leaves, the table's frame growing to the
 *  pagination (`styles/globals.css`, `data-slot="page-table"`). */
export function PageTable({
  children,
  className,
  'data-tour': tour
}: {
  children: ReactNode;
  className?: string;
  /** Names it for a guided tour. */
  'data-tour'?: string;
}) {
  return (
    <div
      data-slot="page-table"
      data-tour={tour}
      className={cn('flex min-w-0 flex-1 flex-col gap-4', className)}
    >
      {children}
    </div>
  );
}
