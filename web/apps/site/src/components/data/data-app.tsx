'use client';

import Link from 'next/link';
import { useSearchParams } from 'next/navigation';
import { Card } from '@braswelljr/algorithm-ui/components/card';
import { Heading } from '@braswelljr/algorithm-ui/components/heading';
import type { AlgorithmApi as Api, Catalog } from '@braswelljr/algorithm-visualizer/core';
import { DataPanel } from '@/components/data/data-panel';
import { WasmGate } from '@/components/shared/wasm-gate';
import { useDataset } from '@/hooks/use-dataset';

function Data({ api, catalog }: { api: Api; catalog: Catalog }) {
  const dataset = useDataset(api, 32);
  const params = useSearchParams().toString();
  const query = params ? `?${params}` : '';
  // The first sort and search in the catalog: no entry is named here.
  const sort = catalog.entries.find((e) => e.input === 'ints' && e.traced);
  const search = catalog.entries.find((e) => e.input === 'sorted-ints');
  return (
    <div className="flex flex-col gap-6">
      <DataPanel
        api={api}
        kinds={catalog.datasets}
        maxValues={catalog.maxValues}
        dataset={dataset}
        title="Load data"
      />
      {dataset.values ? (
        <Card
          size="md"
          render={<section />}
          aria-labelledby="cur"
          className="px-(--card-spacing)"
        >
          <Heading
            level={2}
            id="cur"
            className="font-semibold"
          >
            {dataset.label}
          </Heading>
          <p className="mt-2 font-mono text-sm break-all">
            {dataset.values.slice(0, 200).join(', ')}
            {dataset.values.length > 200 ? `, ... (${dataset.values.length - 200} more)` : ''}
          </p>
          <p className="mt-4 flex flex-wrap gap-4 text-sm">
            <Link
              className="underline"
              href={`/visualize/${query}${query ? '&' : '?'}id=${sort?.id ?? ''}`}
            >
              Watch {sort?.name ?? 'a sort'} on it
            </Link>
            <Link
              className="underline"
              href={`/visualize/${query}${query ? '&' : '?'}id=${search?.id ?? ''}`}
            >
              Search it
            </Link>
            <Link
              className="underline"
              href={`/compare/${query}`}
            >
              Compare sorts on it
            </Link>
          </p>
        </Card>
      ) : null}
    </div>
  );
}

export function DataApp() {
  return (
    <WasmGate>
      {(api, catalog) => (
        <Data
          api={api}
          catalog={catalog}
        />
      )}
    </WasmGate>
  );
}
