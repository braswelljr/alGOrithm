'use client';

import * as React from 'react';
import {
  ArrowDownToLineIcon,
  ArrowLeftRightIcon,
  ArrowUpFromLineIcon,
  EyeIcon,
  FlagIcon,
  LinkIcon,
  PencilIcon,
  ScaleIcon,
  UnlinkIcon,
  type LucideIcon
} from 'lucide-react';
import { NumberTicker } from '@/components/algorithm/number-ticker';
import { Card } from '@/components/ui/card';
import { Toggle } from '@/components/ui/toggle';
import { Trace, TraceDescription, TraceIcon, TraceLine, TraceTitle } from '@/components/ui/trace';
import { cn } from '@/lib/utils';

/** One event of a run's trace, with its kind by name ("compare", "swap", ...). */
export type TraceEventItem = { kind: string; indices: number[]; values: number[]; note?: string };

const kindInfo: Record<string, { icon: LucideIcon; verb: string; tone: string }> = {
  compare: { icon: ScaleIcon, verb: 'Compare', tone: 'text-viz-compare' },
  swap: { icon: ArrowLeftRightIcon, verb: 'Swap', tone: 'text-viz-swap' },
  set: { icon: PencilIcon, verb: 'Set', tone: 'text-viz-write' },
  visit: { icon: EyeIcon, verb: 'Visit', tone: 'text-viz-path' },
  mark: { icon: FlagIcon, verb: 'Mark', tone: 'text-viz-pivot' },
  push: { icon: ArrowDownToLineIcon, verb: 'Push', tone: 'text-viz-found' },
  pop: { icon: ArrowUpFromLineIcon, verb: 'Pop', tone: 'text-viz-swap' },
  link: { icon: LinkIcon, verb: 'Link', tone: 'text-viz-sorted' },
  unlink: { icon: UnlinkIcon, verb: 'Unlink', tone: 'text-viz-swap' }
};

/** An event in words: "Compare [2] and [5] (7, 3)", "Mark [4]: pivot". */
export function describeEvent(e: TraceEventItem): string {
  const info = kindInfo[e.kind];
  const verb = info?.verb ?? e.kind;
  const at = e.indices.map((i) => `[${i}]`).join(' and ');
  const vals = e.values.length ? ` (${e.values.join(', ')})` : '';
  const note = e.note ? `: ${e.note}` : '';
  return `${verb}${at ? ` ${at}` : ''}${vals}${note}`;
}

const ROW = 32;

/**
 * A run's trace as a timeline beside the view: every compare, swap, set,
 * visit, mark, push and pop with its icon, indices, values and note. The
 * current event is highlighted and kept in view; clicking an event (or Enter
 * on it) seeks the player there; j and k (or the arrows) move through it;
 * chips filter by kind and show how many of each there are. Only the rows on
 * screen are rendered, so a trace of ten thousand events stays smooth. It is
 * a listbox: one tab stop, the active row announced.
 *
 *   <TraceView events={events} current={index - 1} onSeek={(i) => send({ type: 'SEEK', index: i + 1 })} />
 */
export function TraceView({
  events,
  current = -1,
  onSeek,
  height = 320,
  truncated,
  className
}: {
  events: TraceEventItem[];
  /** The index of the event the player is on, or -1 before the first. */
  current?: number;
  onSeek?: (index: number) => void;
  height?: number;
  /** The trace was cut at a limit (its counts still cover everything). */
  truncated?: boolean;
  className?: string;
}) {
  const id = React.useId();
  const [hidden, setHidden] = React.useState<Set<string>>(new Set());
  const counts = React.useMemo(() => {
    const m = new Map<string, number>();
    for (const e of events) m.set(e.kind, (m.get(e.kind) ?? 0) + 1);
    return m;
  }, [events]);
  const shown = React.useMemo(
    () => events.map((e, i) => ({ e, i })).filter(({ e }) => !hidden.has(e.kind)),
    [events, hidden]
  );
  const scroller = React.useRef<HTMLDivElement>(null);
  const [top, setTop] = React.useState(0);
  const pos = shown.findIndex((r) => r.i === current);
  const nearest = pos >= 0 ? pos : shown.findIndex((r) => r.i > current);

  // Keep the current event in view as the player moves.
  React.useEffect(() => {
    const el = scroller.current;
    if (!el || pos < 0) return;
    const y = pos * ROW;
    if (y < el.scrollTop || y + ROW > el.scrollTop + el.clientHeight) {
      el.scrollTop = Math.max(0, y - el.clientHeight / 2 + ROW / 2);
    }
  }, [pos]);

  const first = Math.max(0, Math.floor(top / ROW) - 8);
  const last = Math.min(shown.length, Math.ceil((top + height) / ROW) + 8);

  const move = (by: number) => {
    if (!shown.length) return;
    const from = pos >= 0 ? pos : by > 0 ? nearest - 1 : nearest;
    const to = Math.min(shown.length - 1, Math.max(0, from + by));
    onSeek?.(shown[to].i);
  };
  const onKeyDown = (e: React.KeyboardEvent) => {
    const map: Record<string, number> = {
      j: 1,
      ArrowDown: 1,
      k: -1,
      ArrowUp: -1,
      PageDown: 10,
      PageUp: -10
    };
    if (e.key in map) {
      e.preventDefault();
      move(map[e.key]);
    } else if (e.key === 'Home') {
      e.preventDefault();
      if (shown.length) onSeek?.(shown[0].i);
    } else if (e.key === 'End') {
      e.preventDefault();
      if (shown.length) onSeek?.(shown[shown.length - 1].i);
    }
  };

  const toggle = (kind: string) =>
    setHidden((h) => {
      const next = new Set(h);
      if (next.has(kind)) next.delete(kind);
      else next.add(kind);
      return next;
    });

  return (
    <Card
      size="sm"
      aria-label="Trace"
      className={cn('gap-2 px-(--card-spacing)', className)}
      render={<section />}
    >
      <div
        role="group"
        aria-label="Show events of kind"
        className="flex flex-wrap gap-1.5"
      >
        {[...counts.entries()].map(([kind, n]) => {
          const info = kindInfo[kind];
          const Icon = info?.icon ?? FlagIcon;
          const on = !hidden.has(kind);
          return (
            <Toggle
              key={kind}
              variant="chip"
              size="sm"
              pressed={on}
              onPressedChange={() => toggle(kind)}
              className={cn('gap-1', !on && 'text-muted line-through')}
            >
              {/* The icon and the count take the chip's own text color: on the
                  primary fill an event's tone, or a muted count, loses its contrast. */}
              <Icon
                aria-hidden
                className="size-3.5 opacity-100!"
              />
              {info?.verb ?? kind}
              <NumberTicker
                value={n}
                className="font-mono font-semibold tabular-nums"
              />
            </Toggle>
          );
        })}
        {truncated ? (
          <span className="text-muted self-center text-xs">
            The trace was cut short; the counts cover every event.
          </span>
        ) : null}
      </div>
      <div
        ref={scroller}
        role="listbox"
        tabIndex={0}
        aria-label={`${shown.length} events. j and k move, Enter seeks.`}
        aria-activedescendant={pos >= 0 ? `${id}-e${shown[pos].i}` : undefined}
        onKeyDown={onKeyDown}
        onScroll={(e) => setTop(e.currentTarget.scrollTop)}
        style={{ height }}
        className="scroll-quiet focus-visible:ring-accent relative overflow-y-auto rounded-lg outline-none focus-visible:ring-2"
      >
        <Trace
          className="relative gap-0"
          style={{ height: shown.length * ROW }}
        >
          {shown.slice(first, last).map(({ e, i }, k) => {
            const info = kindInfo[e.kind];
            const Icon = info?.icon ?? FlagIcon;
            const active = i === current;
            const done = i < current;
            return (
              <div
                key={i}
                id={`${id}-e${i}`}
                role="option"
                aria-selected={active}
                onClick={() => onSeek?.(i)}
                style={{ top: (first + k) * ROW, height: ROW }}
                className={cn(
                  'absolute inset-x-0 flex cursor-pointer items-center gap-2 rounded-md px-2',
                  active ? 'bg-accent/10 ring-accent/40 ring-1' : 'hover:bg-subtle'
                )}
              >
                <TraceLine
                  status={done || active ? undefined : 'pending'}
                  className="w-full"
                >
                  <span className="text-muted w-12 shrink-0 text-right font-mono text-xs tabular-nums">
                    {i + 1}
                  </span>
                  <TraceIcon className={info?.tone}>
                    <Icon />
                  </TraceIcon>
                  <TraceTitle className={cn(active && 'text-fg font-medium')}>
                    {describeEvent(e)}
                  </TraceTitle>
                  {active ? <TraceDescription className="ms-auto">now</TraceDescription> : null}
                </TraceLine>
              </div>
            );
          })}
        </Trace>
      </div>
      <p
        aria-live="polite"
        className="sr-only"
      >
        {pos >= 0 ? `Event ${current + 1}: ${describeEvent(shown[pos].e)}` : ''}
      </p>
    </Card>
  );
}
