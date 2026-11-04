import { type ComponentProps, type ReactNode } from 'react';
import { cn } from '@braswelljr/algorithm-ui/lib/utils';

/** The page's content column: one width and one gutter, everywhere. */
export function Container({ className, ...props }: ComponentProps<'div'>) {
  return (
    <div
      className={cn('mx-auto w-full max-w-6xl px-4 sm:px-6', className)}
      {...props}
    />
  );
}

/** Two numbers wide, as the drafting grid counts: 1 is "01". */
export function pad2(n: number) {
  return String(n).padStart(2, '0');
}

/**
 * The drafting frame: a hairline across the page at the top, and the content
 * column between two rails, with a tick where the rule meets each rail. Stack
 * frames and the rails run unbroken down the page.
 */
export function Frame({ className, children, ...props }: ComponentProps<'div'>) {
  return (
    <div
      // Clipped, not hidden: a tick at the screen edge must not widen the page,
      // and `clip` keeps the sticky header bars working.
      className={cn('relative overflow-x-clip', className)}
      {...props}
    >
      <span
        aria-hidden="true"
        className="absolute inset-x-0 top-0 h-px bg-line"
      />
      <div className="relative mx-auto w-full max-w-6xl border-x border-line">
        <span
          aria-hidden="true"
          className="tick -top-[5px] tick-half -left-[6px] z-30"
        />
        <span
          aria-hidden="true"
          className="tick -top-[5px] tick-half -right-[6px] z-30"
        />
        {children}
      </div>
    </div>
  );
}

type SectionProps = ComponentProps<'section'> & {
  /** The section's place on the page, shown as 01, 02, ... beside its label. */
  index?: number;
  /** The short name in the section's header bar. */
  label?: string;
  /** A note at the other end of the header bar. */
  aside?: string;
  /** Content laid edge to edge between the rails, after the padded body: grids of cells. */
  bleed?: ReactNode;
};

/**
 * A titled region of a page. Give it `aria-labelledby` pointing at its heading.
 * With a `label` it is drawn on the drafting grid: framed between the rails,
 * under a header bar that stays in view while the section scrolls, reading
 * "01 / Label".
 */
export function Section({
  className,
  children,
  index,
  label,
  aside,
  bleed,
  ...props
}: SectionProps) {
  if (label === undefined) {
    return (
      <section
        className={cn('py-12 sm:py-16', className)}
        {...props}
      >
        <Container>{children}</Container>
      </section>
    );
  }

  return (
    <section
      className={className}
      {...props}
    >
      <Frame>
        <div className="sticky top-14 z-20 border-b border-line bg-fd-background/70 backdrop-blur-md">
          <div className="flex items-center justify-between gap-4 px-4 py-3 sm:px-6">
            <p className="hud-label text-fd-muted-foreground">
              {index !== undefined && (
                <>
                  <span className="text-faint">{pad2(index)}</span>
                  <span
                    aria-hidden="true"
                    className="mx-2 text-faint"
                  >
                    /
                  </span>
                </>
              )}
              {label}
            </p>
            {aside && <p className="hidden hud-label text-faint sm:block">{aside}</p>}
          </div>
        </div>
        <div className="px-4 py-12 sm:px-6 sm:py-16">{children}</div>
        {bleed}
      </Frame>
    </section>
  );
}

type GridCellProps = ComponentProps<'div'> & {
  /** The cell's name, at the left of its header bar. */
  label: ReactNode;
  /** The cell's number within its grid, at the right: 1 is "01". */
  index: number;
};

/**
 * One cell of a drafting grid: a header bar with its name and number, then its
 * content. Lay cells in a `grid gap-px bg-line` so the gaps draw the lines.
 */
export function GridCell({ label, index, className, children, ...props }: GridCellProps) {
  return (
    <div
      className={cn('flex h-full flex-col bg-fd-background/70', className)}
      {...props}
    >
      <div className="flex items-center justify-between gap-4 border-b border-line px-4 py-2.5 sm:px-6">
        <p className="hud-label text-fd-muted-foreground">{label}</p>
        <p className="hud-label text-faint">
          <span className="sr-only">Number </span>
          {pad2(index)}
        </p>
      </div>
      <div className="flex-1">{children}</div>
    </div>
  );
}
