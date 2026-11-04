'use client';

import * as React from 'react';
import { CodeSample } from '@braswelljr/algorithm-ui/components/code-block';
import { SpringTabs } from '@braswelljr/algorithm-ui/algorithm/spring-tabs';

export type CodeTab = { value: string; label: string; language: string; code: string; title?: string };

/**
 * The same run as Go, as a command and as JSON, one tab each, with the spring
 * indicator. Tabs with no code are left out.
 *
 *   <CodeTabs tabs={[{ value: 'go', label: 'Go', language: 'go', code: snippet.go },
 *                    { value: 'cli', label: 'CLI', language: 'bash', code: snippet.cli }]} />
 */
export function CodeTabs({ tabs, className }: { tabs: CodeTab[]; className?: string }) {
  const present = tabs.filter((t) => t.code);
  const [value, setValue] = React.useState(present[0]?.value ?? '');
  const id = React.useId();
  const tab = present.find((t) => t.value === value) ?? present[0];
  if (!tab) return null;
  return (
    <div className={className}>
      <SpringTabs
        label="Code"
        idPrefix={id}
        value={tab.value}
        onValueChange={setValue}
        tabs={present.map((t) => ({ value: t.value, label: t.label }))}
      />
      <div
        role="tabpanel"
        id={`${id}-${tab.value}`}
        aria-labelledby={`${id}-tab-${tab.value}`}
        className="mt-2"
      >
        <CodeSample
          code={tab.code}
          language={tab.language}
          title={tab.title ?? tab.label}
        />
      </div>
    </div>
  );
}
