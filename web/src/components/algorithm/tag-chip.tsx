import type * as React from 'react';
import { HashIcon } from 'lucide-react';
import { Toggle } from '@/components/ui/toggle';
import { cn } from '@/lib/utils';

/**
 * A tag from the catalog's vocabulary. A link when given `href`, a toggle
 * when given `pressed` and `onClick`, else plain text. `title` is the tag's
 * meaning.
 *
 *   <TagChip tag="divide-and-conquer" href="/docs/tags/#divide-and-conquer" />
 */
export function TagChip({
  tag,
  href,
  pressed,
  onClick,
  title,
  className
}: {
  tag: string;
  href?: string;
  pressed?: boolean;
  onClick?: React.MouseEventHandler<HTMLButtonElement>;
  title?: string;
  className?: string;
}) {
  const cls = cn(
    'inline-flex items-center gap-0.5 rounded-full border px-2 py-0.5 text-xs transition-colors',
    pressed
      ? 'border-accent bg-primary text-accent-fg'
      : 'border-border bg-card text-muted hover:border-control hover:text-fg',
    className
  );
  const body = (
    <>
      <HashIcon
        aria-hidden
        className="size-3"
      />
      {tag}
    </>
  );
  if (href)
    return (
      <a
        href={href}
        title={title}
        className={cls}
      >
        {body}
      </a>
    );
  // A toggle is the library's Toggle (its chip variant): the pressed state,
  // the keys and the 44px touch target come from there.
  if (onClick)
    return (
      <Toggle
        variant="chip"
        size="sm"
        pressed={pressed ?? false}
        title={title}
        onClick={onClick}
        className={cn('gap-0.5', className)}
      >
        {body}
      </Toggle>
    );
  return (
    <span
      title={title}
      className={cls}
    >
      {body}
    </span>
  );
}
