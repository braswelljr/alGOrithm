'use client';

import { Fragment } from 'react';
import { ChevronRight } from 'lucide-react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { crumbsFor } from '@/lib/breadcrumbs';
import { cn } from '@/utils/utils';

/**
 * Molecule. Where the page sits in the site: a `nav` landmark with an ordered
 * list, every step a link but the last, which is the page itself. On a narrow
 * screen the steps between the first and the last two fold into an ellipsis.
 * Top-level pages have no trail.
 */
export function Breadcrumbs({ className }: { className?: string }) {
  const crumbs = crumbsFor(usePathname() ?? '/');
  if (crumbs.length < 1) return null;

  // Steps that fold on a narrow screen: all but the first and the last two.
  const folds = (i: number) => i > 0 && i < crumbs.length - 2;
  const folded = crumbs.some((_, i) => folds(i));

  return (
    <nav
      aria-label="Breadcrumb"
      className={cn('min-w-0', className)}
    >
      <ol className="flex min-w-0 items-center gap-1.5 text-sm text-muted">
        {crumbs.map((crumb, i) => {
          const last = i === crumbs.length - 1;
          return (
            <Fragment key={`${crumb.href ?? 'current'}-${i}`}>
              {i === 1 && folded ? (
                <li
                  aria-hidden
                  className="flex items-center gap-1.5 sm:hidden"
                >
                  <span>…</span>
                  <ChevronRight className="size-3.5 shrink-0 opacity-60" />
                </li>
              ) : null}
              <li
                className={cn(
                  'flex min-w-0 items-center gap-1.5',
                  folds(i) && 'hidden sm:flex',
                  last && 'shrink'
                )}
              >
                {crumb.href && !last ? (
                  <Link
                    href={crumb.href}
                    className="truncate rounded-sm hover:text-fg hover:underline focus-visible:ring-2 focus-visible:ring-accent focus-visible:outline-none"
                  >
                    {crumb.label}
                  </Link>
                ) : (
                  <span
                    aria-current="page"
                    className="truncate font-medium text-fg"
                  >
                    {crumb.label}
                  </span>
                )}
                {last ? null : (
                  <ChevronRight
                    aria-hidden
                    className="size-3.5 shrink-0 opacity-60"
                  />
                )}
              </li>
            </Fragment>
          );
        })}
      </ol>
    </nav>
  );
}
