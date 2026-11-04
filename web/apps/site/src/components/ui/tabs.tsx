'use client';

import { useId, type ReactNode } from 'react';
import { SpringTabs } from '@braswelljr/algorithm-ui/algorithm/spring-tabs';

export interface Tab {
  id: string;
  label: string;
}

/**
 * Tabs with the ARIA tabs pattern (the library's SpringTabs): arrow keys,
 * Home and End move between tabs, and only the selected panel is rendered.
 */
export function Tabs({
  tabs,
  value,
  onChange,
  children,
  label
}: {
  tabs: Tab[];
  value: string;
  onChange: (id: string) => void;
  children: ReactNode;
  label: string;
}) {
  const base = useId();
  return (
    <div>
      <SpringTabs
        label={label}
        idPrefix={base}
        value={value}
        onValueChange={onChange}
        tabs={tabs.map((t) => ({ value: t.id, label: t.label }))}
      />
      <div
        role="tabpanel"
        id={`${base}-${value}`}
        aria-labelledby={`${base}-tab-${value}`}
        className="pt-4"
      >
        {children}
      </div>
    </div>
  );
}
