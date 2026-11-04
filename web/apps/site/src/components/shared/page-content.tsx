import type { ReactNode } from 'react';
import { cn } from '@/utils/utils';

/**
 * Template. The column a page is laid out in, so each one reads like the
 * next. The spacing standard lives here and in the shell, nowhere else:
 *
 * - The shell's `<main>` owns the outer padding (`p-5 lg:p-6`), and the
 *   content runs the page's full width; a page adds no padding of its own.
 * - Sections of a page sit `gap-6` apart, in this column.
 * - A panel (a card, a framed chart, a read-only summary) pads `p-5`.
 */
export function PageContent({ children, className }: { children: ReactNode; className?: string }) {
  return <div className={cn('flex min-w-0 flex-1 flex-col gap-6', className)}>{children}</div>;
}
