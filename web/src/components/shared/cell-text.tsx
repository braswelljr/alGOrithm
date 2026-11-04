'use client';

// Long text in a table cell.
//
// A description or a note is written for a page, not for a column, and a table
// has to hold both without letting either decide the row height: one paragraph
// of somebody's statement of interest can push every other row off the screen,
// and truncating to a single line hides the half of it that says what it is.
//
// So the cell keeps a fixed ceiling (a narrow column, two lines), and the rest
// is one hover away. The disclosure is only attached when the text is really
// clipped, measured rather than guessed from its length: a control that opens to
// show you what you can already read is noise, and how much fits depends on the
// column's width, which the reader can drag.
//
// It opens on hover and on tap, which is why it is a popover rather than a
// tooltip. Tooltips do not open on touch, and a phone is where the clipping bites
// hardest. The trigger is a real button, so it is reachable by keyboard and the
// table's row-click guard already knows to leave it alone.
import { useEffect, useRef, useState, type RefObject } from 'react';
import { Popover, PopoverPopup, PopoverTrigger } from '@/components/ui/popover';
import { cn } from '@/lib/utils';

/** Tailwind needs the whole class name, so the clamps are spelled out. */
const CLAMP = {
  1: 'line-clamp-1',
  2: 'line-clamp-2',
  3: 'line-clamp-3'
} as const;

export type CellTextProps = {
  children?: string | null;
  /** Lines to show before clipping. Two by default: enough to tell a note from
   *  a description, short enough that rows stay the same height. */
  lines?: keyof typeof CLAMP;
  /** Classes for the text itself, for the muted secondary lines cells use. */
  className?: string;
  /** Which side the full text opens on. */
  side?: 'top' | 'bottom' | 'left' | 'right';
};

/** Whether an element is showing less than it holds, kept in step with resizes:
 *  a column dragged wider stops clipping, and the disclosure has to go with it. */
function useIsClipped(ref: RefObject<HTMLElement | null>, text: string): boolean {
  const [clipped, setClipped] = useState(false);

  useEffect(() => {
    const element = ref.current;
    if (!element) return;

    const measure = () =>
      setClipped(
        element.scrollHeight > element.clientHeight + 1 ||
          element.scrollWidth > element.clientWidth + 1
      );

    measure();
    // Absent in some test environments, where one measurement is all there is
    // to make and nothing resizes afterwards.
    if (typeof ResizeObserver === 'undefined') return;

    const observer = new ResizeObserver(measure);
    observer.observe(element);
    return () => observer.disconnect();
  }, [ref, text]);

  return clipped;
}

export function CellText({ children, lines = 2, className, side = 'top' }: CellTextProps) {
  const ref = useRef<HTMLParagraphElement>(null);
  const text = children?.trim() ?? '';
  const clipped = useIsClipped(ref, text);

  if (!text) return null;

  // `break-words` rather than `truncate`: an unbroken run of characters, which a
  // reference number or a pasted URL often is, would otherwise widen the column
  // past the ceiling this exists to impose.
  const paragraph = (
    <p
      ref={ref}
      className={cn(CLAMP[lines], 'break-words', className)}
    >
      {text}
    </p>
  );

  if (!clipped) return <div className="w-full">{paragraph}</div>;

  return (
    <Popover>
      <PopoverTrigger
        openOnHover
        delay={200}
        className="focus-visible:ring-primary w-full cursor-pointer text-start focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:outline-none"
        // ui-ok: the clipped text is its own trigger, read as text, not a control
        render={<button type="button" />}
      >
        {paragraph}
      </PopoverTrigger>
      <PopoverPopup
        side={side}
        tooltipStyle
        className="max-w-sm"
      >
        <p className="text-xs break-words whitespace-pre-wrap">{text}</p>
      </PopoverPopup>
    </Popover>
  );
}
