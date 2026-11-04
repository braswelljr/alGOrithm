import { type ComponentProps, type ReactNode } from 'react';
import { cn } from '@/lib/utils';

type TerminalWindowProps = Omit<ComponentProps<'div'>, 'title'> & {
  /** Shown centered in the title bar: a path, a file name, a shell. */
  title?: ReactNode;
  /** Controls at the right end of the title bar. */
  actions?: ReactNode;
  /** The element the window is. A live terminal is a labeled `section`. */
  as?: 'div' | 'section';
  /** Stand the window off the page, for the one on a screen that should draw the eye. */
  glow?: boolean;
};

/** The three window buttons. They do nothing, so they are hidden from assistive technology. */
function Lights() {
  return (
    <span
      aria-hidden="true"
      className="flex shrink-0 gap-1.5"
    >
      <span className="size-2.5 rounded-full bg-red-500/80 dark:bg-red-400/80" />
      <span className="size-2.5 rounded-full bg-amber-500/80 dark:bg-amber-400/80" />
      <span className="size-2.5 rounded-full bg-green-500/80 dark:bg-green-400/80" />
    </span>
  );
}

/**
 * The frame of a terminal window: a title bar with window lights, and a body.
 * It draws the chrome only; what runs inside is the caller's.
 */
export function TerminalWindow({
  title,
  actions,
  as: Tag = 'div',
  glow = false,
  className,
  children,
  ...props
}: TerminalWindowProps) {
  return (
    <Tag
      className={cn(
        'bg-fd-card/90 text-fd-card-foreground flex min-w-0 flex-col overflow-hidden rounded-xl border backdrop-blur-sm',
        glow ? 'lifted' : 'shadow-sm',
        className
      )}
      {...props}
    >
      <div className="bg-fd-secondary/40 grid h-9 shrink-0 grid-cols-[1fr_auto_1fr] items-center gap-2 border-b px-3">
        <Lights />
        <span className="text-fd-muted-foreground truncate text-center font-mono text-xs">
          {title}
        </span>
        <span className="flex justify-end">{actions}</span>
      </div>
      {children}
    </Tag>
  );
}
