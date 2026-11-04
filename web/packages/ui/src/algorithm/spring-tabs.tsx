'use client';

import * as React from 'react';
import { motion } from 'motion/react';
import { useReducedMotion } from '@braswelljr/algorithm-ui/hooks/use-reduced-motion';
import { cn } from '@braswelljr/algorithm-ui/lib/utils';

/**
 * A row of tabs whose indicator springs to the selected one (a shared
 * layoutId, as in beUI's tabs). It is a real tablist: arrow keys move between
 * tabs, Home and End jump, and each tab controls the panel with its id.
 * Under reduced motion the indicator moves at once.
 *
 *   <SpringTabs label="Code" value={tab} onValueChange={setTab}
 *     tabs={[{ value: 'go', label: 'Go' }, { value: 'cli', label: 'CLI' }]} />
 */
export function SpringTabs({
  tabs,
  value,
  onValueChange,
  label,
  idPrefix,
  className
}: {
  tabs: { value: string; label: React.ReactNode }[];
  value: string;
  onValueChange: (value: string) => void;
  /** The tablist's accessible name. */
  label: string;
  /** Prefix of the panel ids: a tab controls `${idPrefix}-${value}`. */
  idPrefix?: string;
  className?: string;
}) {
  const reduce = useReducedMotion();
  const id = React.useId();
  const prefix = idPrefix ?? id;
  const refs = React.useRef<(HTMLButtonElement | null)[]>([]);
  const onKey = (e: React.KeyboardEvent, i: number) => {
    const n = tabs.length;
    const to =
      e.key === 'ArrowRight'
        ? (i + 1) % n
        : e.key === 'ArrowLeft'
          ? (i - 1 + n) % n
          : e.key === 'Home'
            ? 0
            : e.key === 'End'
              ? n - 1
              : -1;
    if (to < 0) return;
    e.preventDefault();
    onValueChange(tabs[to].value);
    refs.current[to]?.focus();
  };
  return (
    <div
      role="tablist"
      aria-label={label}
      className={cn('inline-flex items-center gap-1 rounded-lg bg-subtle p-1', className)}
    >
      {tabs.map((t, i) => {
        const selected = t.value === value;
        return (
          <button
            key={t.value}
            ref={(el) => {
              refs.current[i] = el;
            }}
            type="button"
            role="tab"
            id={`${prefix}-tab-${t.value}`}
            aria-selected={selected}
            aria-controls={`${prefix}-${t.value}`}
            tabIndex={selected ? 0 : -1}
            onClick={() => onValueChange(t.value)}
            onKeyDown={(e) => onKey(e, i)}
            className={cn(
              'relative rounded-md px-3 py-1.5 text-sm font-medium transition-colors',
              selected ? 'text-fg' : 'text-muted hover:text-fg'
            )}
          >
            {selected ? (
              <motion.span
                layoutId={`${prefix}-indicator`}
                aria-hidden
                // card-ok: the sliding indicator under the selected tab, not a card
                className="absolute inset-0 rounded-md bg-card shadow-sm ring-1 ring-border"
                transition={reduce ? { duration: 0 } : { type: 'spring', stiffness: 500, damping: 38 }}
              />
            ) : null}
            <span className="relative">{t.label}</span>
          </button>
        );
      })}
    </div>
  );
}
