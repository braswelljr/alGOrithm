'use client';

import { useMemo } from 'react';
import { GoDoc } from '@/components/docs/go-doc';
import { CodeSample } from '@/components/ui/code-block';
import { Heading } from '@/components/ui/heading';
import { repoUrl } from '@/config/site';
import type { AlgorithmApi, Entry } from '@/visualizer/core';

/** The entry's Go source, its doc comment, and where it lives in the repository. */
export function CodePanel({ api, entry }: { api: AlgorithmApi; entry: Entry }) {
  const source = useMemo(() => api.source(entry.id), [api, entry.id]);
  const file = source.ok ? source.value.file : entry.file;
  return (
    <div className="flex flex-col gap-3">
      {source.ok && source.value.doc ? (
        <div className="border-border bg-subtle rounded-md border p-3">
          <Heading
            level={3}
            className="text-sm font-semibold"
          >
            {entry.package && entry.func ? `${entry.package}.${entry.func}` : entry.name}
          </Heading>
          <GoDoc
            text={source.value.doc}
            className="mt-2 text-sm"
          />
        </div>
      ) : null}
      {entry.import ? (
        <div className="flex flex-col gap-2">
          <CodeSample
            title="Import"
            language="go"
            code={`import "${entry.import}"`}
            size="sm"
          />
          {file ? (
            <a
              className="w-fit text-sm underline"
              href={`${repoUrl}/blob/main/pkg/${file}`}
            >
              {file} on GitHub
            </a>
          ) : null}
        </div>
      ) : null}
      {source.ok ? (
        <CodeSample
          title={source.value.file}
          language="go"
          code={source.value.code}
          lineNumbers
        />
      ) : (
        <p className="text-muted text-sm">
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
