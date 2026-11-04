'use client';

import { useId, useState } from 'react';
import { Lock } from 'lucide-react';
import { Tabs } from '@/components/shared/tabs';
import { Card } from '@/components/ui/card';
import { Heading } from '@/components/ui/heading';
import type { Dataset } from '@/hooks/use-dataset';
import { cn } from '@/lib/utils';
import type { AlgorithmApi as Api } from '@/visualizer/core';
import { FileTab } from './file-tab';
import { GenerateTab } from './generate-tab';
import { PasteTab } from './paste-tab';
import { UrlTab } from './url-tab';

/**
 * Where a page's data comes from: typed, a file, a URL or generated. Whatever
 * the source, the values are read and checked by the wasm build in this
 * browser (docs/rules/data-from-outside-is-unknown-until-checked.md).
 */
export function DataPanel({
  api,
  kinds,
  maxValues,
  dataset,
  bare = false,
  title = 'Data'
}: {
  api: Api;
  kinds: string[];
  maxValues: number;
  dataset: Dataset;
  title?: string;
  /** Without its card, for a surface that already is one (a dialog). */
  bare?: boolean;
}) {
  const [tab, setTab] = useState('generate');
  const headingId = useId();
  return (
    <Card
      size="md"
      render={<section />}
      aria-labelledby={headingId}
      className={cn(
        'px-(--card-spacing)',
        bare && 'gap-0 rounded-none bg-transparent p-0 shadow-none ring-0 before:hidden'
      )}
    >
      <div className="mb-2 flex flex-wrap items-baseline justify-between gap-2">
        <Heading
          level={2}
          id={headingId}
          className="text-base font-semibold"
        >
          {title}
        </Heading>
        <p className="text-muted text-sm">
          {dataset.error ? (
            <span className="text-danger">{dataset.error}</span>
          ) : dataset.label ? (
            <>Using {dataset.label}</>
          ) : (
            'Loading...'
          )}
        </p>
      </div>
      <Tabs
        label="Where the data comes from"
        value={tab}
        onChange={setTab}
        tabs={[
          { id: 'generate', label: 'Generate' },
          { id: 'paste', label: 'Paste' },
          { id: 'file', label: 'File' },
          { id: 'url', label: 'URL' }
        ]}
      >
        {tab === 'generate' ? (
          <GenerateTab
            kinds={kinds}
            dataset={dataset}
            maxValues={maxValues}
          />
        ) : null}
        {tab === 'paste' ? (
          <PasteTab
            api={api}
            onUse={dataset.use}
          />
        ) : null}
        {tab === 'file' ? (
          <FileTab
            api={api}
            onUse={dataset.use}
          />
        ) : null}
        {tab === 'url' ? (
          <UrlTab
            api={api}
            onUse={dataset.use}
          />
        ) : null}
      </Tabs>
      <p className="text-muted mt-3 flex items-center gap-1.5 text-xs">
        <Lock
          aria-hidden
          className="size-3.5"
        />{' '}
        Your data stays in this browser.
      </p>
    </Card>
  );
}
