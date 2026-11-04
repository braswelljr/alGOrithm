'use client';

import { useId, useState } from 'react';
import { Card } from '@braswelljr/algorithm-ui/components/card';
import { Heading } from '@braswelljr/algorithm-ui/components/heading';
import type { AlgorithmApi as Api } from '@braswelljr/algorithm-visualizer/core';
import { Lock } from 'lucide-react';
import { Tabs } from '@/components/ui/tabs';
import type { Dataset } from '@/hooks/use-dataset';
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
  title = 'Data'
}: {
  api: Api;
  kinds: string[];
  maxValues: number;
  dataset: Dataset;
  title?: string;
}) {
  const [tab, setTab] = useState('generate');
  const headingId = useId();
  return (
    <Card
      size="md"
      render={<section />}
      aria-labelledby={headingId}
      className="px-(--card-spacing)"
    >
      <div className="mb-2 flex flex-wrap items-baseline justify-between gap-2">
        <Heading
          level={2}
          id={headingId}
          className="text-base font-semibold"
        >
          {title}
        </Heading>
        <p className="text-sm text-muted">
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
      <p className="mt-3 flex items-center gap-1.5 text-xs text-muted">
        <Lock
          aria-hidden
          className="size-3.5"
        />{' '}
        Your data stays in this browser.
      </p>
    </Card>
  );
}
