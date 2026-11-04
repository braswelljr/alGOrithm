'use client';

import type { ReactNode } from 'react';
import { Heading } from '@/components/ui/heading';
import { OverflowRow } from '@/components/ui/overflow-row';
import { useNavigation } from '@/context/use-navigation';
import { usePageTitle, useSetPageTitle } from '@/context/use-page-title';
import { cn } from '@/lib/utils';

/** Whether two titles say the same thing, ignoring case and spacing. */
function sameTitle(a: string, b: string) {
  const norm = (s: string) => s.trim().replace(/\s+/g, ' ').toLowerCase();
  return norm(a) === norm(b);
}

/**
 * The heading every page opens with: what it is, what it is for,
 * and the actions it offers, on a row of their own under the title.
 *
 * The navbar already names the page, so a title that only repeats it is not
 * drawn again: it stays the page's `h1` for screen readers, and the header
 * shows the description and the actions. A title of its own, such as a
 * record's name with its status, is drawn. `hideTitle` says the title repeats
 * the navbar's in other words.
 */
export function PageHeader({
  title,
  description,
  actions,
  eyebrow,
  hideTitle = false,
  wrapActions = false,
  className
}: {
  title: ReactNode;
  description?: ReactNode;
  actions?: ReactNode;
  eyebrow?: ReactNode;
  hideTitle?: boolean;
  /** Buttons wrap onto more rows rather than scroll, for a page with many. */
  wrapActions?: boolean;
  className?: string;
}) {
  const { title: announced } = usePageTitle();
  const { active } = useNavigation();
  const named = announced || active?.label;
  const repeats = hideTitle || (typeof title === 'string' && !!named && sameTitle(title, named));

  const heading = (
    <Heading
      level={1}
      className={repeats ? 'sr-only' : 'font-heading text-2xl font-semibold text-balance'}
    >
      {title}
    </Heading>
  );

  // Nothing else to show: only the heading, for screen readers.
  if (repeats && !description && !actions && !eyebrow) return heading;

  return (
    <header
      className={cn(
        // The title and description first, the actions on a row of their own
        // under them: never squeezed beside the title.
        'border-border flex flex-col gap-4 border-b pb-5',
        className
      )}
    >
      <div className="min-w-0 space-y-1">
        {eyebrow ? (
          <p className="text-accent text-xs font-semibold tracking-wider uppercase">{eyebrow}</p>
        ) : null}
        {heading}
        {description ? (
          // The page's width: nothing sits beside it.
          <p className="text-muted text-sm text-pretty">{description}</p>
        ) : null}
      </div>
      {actions ? (
        wrapActions ? (
          // As many rows as they need, a label longer than the screen wrapping
          // inside its button too. The padding keeps focus rings in.
          <div className="-m-1 flex max-w-[calc(100%+0.5rem)] min-w-0 flex-wrap items-center gap-2 p-1 [&_a]:h-auto [&_a]:min-h-9 [&_a]:max-w-full [&_a]:whitespace-normal [&_button]:h-auto [&_button]:min-h-9 [&_button]:max-w-full [&_button]:whitespace-normal">
            {actions}
          </div>
        ) : (
          // One row, never wider than the page: what does not fit scrolls,
          // with arrows at the ends that have more, as the tabs have.
          <OverflowRow className="-m-1 max-w-[calc(100%+0.5rem)]">{actions}</OverflowRow>
        )
      ) : null}
    </header>
  );
}

/**
 * A page's opening: its title and the sentence that says what it is for. The
 * navbar shows the title (announced to it here), so it is not drawn again: it
 * stays the page's h1 for screen readers (PageHeader with hideTitle), and only the
 * description shows.
 */
export function PageHeading({ title, children }: { title: string; children?: ReactNode }) {
  // The navbar shows the title: a page about one entry names it there, not
  // only its menu section.
  useSetPageTitle(title);
  return (
    <PageHeader
      title={title}
      hideTitle
      description={children}
      className="mb-6"
    />
  );
}
