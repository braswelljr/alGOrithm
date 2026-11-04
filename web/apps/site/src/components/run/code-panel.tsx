'use client';

import { useMemo } from 'react';
import { Heading } from '@braswelljr/algorithm-ui/components/heading';
import { CodeView } from '@braswelljr/algorithm-visualizer';
import type { AlgorithmApi, Entry } from '@braswelljr/algorithm-visualizer/core';
import { repoUrl } from '@/config/site';

/** The entry's Go source, its doc comment, and where it lives in the repository. */
export function CodePanel({ api, entry }: { api: AlgorithmApi; entry: Entry }) {
  const source = useMemo(() => api.source(entry.id), [api, entry.id]);
  const file = source.ok ? source.value.file : entry.file;
  return (
    <div className="flex flex-col gap-3">
      {source.ok && source.value.doc ? (
        <div className="rounded-md border border-border bg-subtle p-3">
          <Heading
            level={3}
            className="text-sm font-semibold"
          >
            {entry.package && entry.func ? `${entry.package}.${entry.func}` : entry.name}
          </Heading>
          <p className="mt-1 text-sm whitespace-pre-wrap">{source.value.doc}</p>
        </div>
      ) : null}
      {entry.import ? (
        <p className="text-sm">
          Import: <code className="rounded bg-subtle px-1 font-mono">{entry.import}</code>
          {file ? (
            <>
              {' '}
              ·{' '}
              <a
                className="underline"
                href={`${repoUrl}/blob/main/pkg/${file}`}
              >
                {file} on GitHub
              </a>
            </>
          ) : null}
        </p>
      ) : null}
      {source.ok ? (
        <div className="av-root">
          <CodeView
            code={source.value.code}
            file={source.value.file}
          />
        </div>
      ) : (
        <p className="text-sm text-muted">
          The source is not in this build of the algorithms ({source.failure.error}).{' '}
          {entry.file ? (
            <a
              className="underline"
              href={`${repoUrl}/blob/main/pkg/${entry.file}`}
            >
              Read it on GitHub
            </a>
          ) : null}
        </p>
      )}
    </div>
  );
}
