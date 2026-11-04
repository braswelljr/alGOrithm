'use client';

import type * as React from 'react';
import { Tabs, TabsList, TabsTab } from '@/components/ui/tabs';

/**
 * A row of tabs over panels the caller renders: the library's Tabs, so the
 * tablist, the keys (arrows, Home, End), the ARIA wiring and the sliding
 * indicator come from one place. Each tab controls the panel with the id
 * `${idPrefix}-${value}`.
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
  return (
    <Tabs
      value={value}
      onValueChange={(v) => onValueChange(String(v))}
      className={className}
    >
      <TabsList
        aria-label={label}
        scrollable={false}
      >
        {tabs.map((t) => (
          <TabsTab
            key={t.value}
            value={t.value}
            id={idPrefix ? `${idPrefix}-tab-${t.value}` : undefined}
            aria-controls={idPrefix ? `${idPrefix}-${t.value}` : undefined}
          >
            {t.label}
          </TabsTab>
        ))}
      </TabsList>
    </Tabs>
  );
}
