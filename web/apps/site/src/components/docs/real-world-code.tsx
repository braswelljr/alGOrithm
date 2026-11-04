'use client';

import { useMemo, useState } from 'react';
import { CodeSample } from '@braswelljr/algorithm-ui/components/code-block';
import { Heading } from '@braswelljr/algorithm-ui/components/heading';
import { Toggle } from '@braswelljr/algorithm-ui/components/toggle';
import type { AlgorithmApi, Entry } from '@braswelljr/algorithm-visualizer/core';
import { examplePrograms, programName, runCommand } from '@/lib/real-world';
import { repoUrl } from '@/config/site';

/** One example program: the command to run it, then its Go source. */
function Program({ api, path }: { api: AlgorithmApi; path: string }) {
  const program = useMemo(() => api.example(path), [api, path]);
  const name = programName(path);
  return (
    <div className="flex min-w-0 flex-col gap-2">
      <CodeSample
        title="Shell"
        language="bash"
        code={runCommand(path)}
      />
      {program.ok ? (
        <CodeSample
          title={`${path}/main.go`}
          language="go"
          code={program.value.code}
          lineNumbers
        />
      ) : (
        <p className="text-sm text-muted">
          The source of {name} is not in this build.{' '}
          <a
            className="underline underline-offset-2"
            href={`${repoUrl}/tree/main/${path}`}
          >
            Read it on GitHub
          </a>
          .
        </p>
      )}
    </div>
  );
}

/**
 * Where an entry is used in real software, and the runnable programs in
 * examples/ that use it the way production code would. Renders nothing for
 * an entry without uses.
 */
export function RealWorldCode({ api, entry }: { api: AlgorithmApi; entry: Entry }) {
  const programs = examplePrograms(entry.realWorld);
  const [shown, setShown] = useState(programs[0] ?? '');
  if (!entry.realWorld.length) return null;
  return (
    <section
      aria-labelledby="real-world"
      className="flex flex-col gap-4"
    >
      <Heading
        level={2}
        id="real-world"
        className="text-lg font-semibold"
      >
        Real-world code
      </Heading>
      <ul className="grid gap-3 sm:grid-cols-2">
        {entry.realWorld.map((u) => (
          <li
            key={u.title + u.where}
            className="rounded-lg border border-border bg-subtle p-3 text-sm"
          >
            <p className="font-medium">{u.title}</p>
            <p className="mt-1 text-muted">{u.where}</p>
          </li>
        ))}
      </ul>
      {programs.length > 1 ? (
        <div
          role="group"
          aria-label="Example programs"
          className="flex flex-wrap gap-2"
        >
          {programs.map((p) => (
            <Toggle
              key={p}
              variant="chip"
              size="sm"
              pressed={p === shown}
              onPressedChange={() => setShown(p)}
              className="font-mono text-xs"
            >
              {programName(p)}
            </Toggle>
          ))}
        </div>
      ) : null}
      {shown ? (
        <Program
          api={api}
          path={shown}
        />
      ) : null}
    </section>
  );
}
