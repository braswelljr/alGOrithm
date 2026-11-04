'use client';

import { usePlayerShortcuts } from '@/hooks/use-player-shortcuts';
import { kinds } from '@braswelljr/algorithm-visualizer/core';
import { TraceView } from '@braswelljr/algorithm-ui/algorithm/trace-view';
import { Alert } from '@braswelljr/algorithm-ui/components/alert';
import { useEffect, useMemo, useRef, useState } from 'react';
import Link from 'next/link';
import { buttonVariants } from '@braswelljr/algorithm-ui/components/button';
import { Card } from '@braswelljr/algorithm-ui/components/card';
import { RunView, usePlayer } from '@braswelljr/algorithm-visualizer';
import {
  makeSavedRun,
  type AlgorithmApi,
  type Entry,
  type Input,
  type SavedRun
} from '@braswelljr/algorithm-visualizer/core';
import { parseAsInteger, parseAsStringLiteral, useQueryStates } from 'nuqs';
import { RealWorldCode } from '@/components/docs/real-world-code';
import { Button } from '@/components/ui/button';
import { Tabs } from '@/components/ui/tabs';
import { useEntryInput } from '@/hooks/use-entry-input';
import { useRunHistory } from '@/hooks/use-run-history';
import { CodePanel } from './code-panel';
import { RunToolbar } from './run-toolbar';

/** The most events kept for an animation. */
export const EVENT_LIMIT = 200_000;

const params = {
  step: parseAsInteger,
  speed: parseAsInteger,
  tab: parseAsStringLiteral(['run', 'code', 'real'] as const)
};

/**
 * One run of an entry: the view and player, the code, and the toolbar. The
 * step and speed live in the link and the player opens paused there
 * (docs/rules/state-a-reader-can-arrive-at-lives-in-the-url.md). A run opened
 * from a file or the history replaces the page's own until it is closed.
 */
export function RunPanel({
  api,
  entry,
  input,
  followEnd = false
}: {
  api: AlgorithmApi;
  entry: Entry;
  input: Input;
  /** Open each new run at its last step (the structures playground). */
  followEnd?: boolean;
}) {
  const [q, setQ] = useQueryStates(params, { history: 'replace' });
  const [opened, setOpened] = useState<SavedRun | null>(null);
  const history = useRunHistory();
  const player = usePlayer({ speed: q.speed ?? 10 });
  usePlayerShortcuts(player, (q.tab ?? 'run') === 'run');
  // Views that edit their own input (a circuit's switches) write it to the
  // link, which reruns the entry; the options stay in their own parameter.
  const [, setLinkInput] = useEntryInput(entry);
  const onInputChange = useMemo(
    () => (next: Input) => setLinkInput({ ...next, options: undefined }),
    [setLinkInput]
  );
  // The step in the link when the page opened; later changes come from the player.
  const [initialStep] = useState(() => q.step ?? 0);

  const shownEntry = (opened && api.catalog().entries.find((e) => e.id === opened.entry)) || entry;
  const shownInput = opened?.input ?? input;
  const key = JSON.stringify([shownEntry.id, shownInput]);
  const outcome = useMemo(() => {
    if (opened?.result) return { ok: true as const, value: opened.result };
    return api.run(shownEntry.id, shownInput, EVENT_LIMIT);
  }, [api, key, opened]);

  // Remember a fresh run in this browser's history.
  const recorded = useRef<string | null>(null);
  useEffect(() => {
    if (!outcome.ok || opened || recorded.current === key) return;
    recorded.current = key;
    history.add(makeSavedRun(entry.id, input));
  }, [outcome, opened, key]);

  // Keep the step and the speed in the link whenever the player comes to rest.
  const { status, index, speed } = player.state;
  useEffect(() => {
    if (status === 'playing' || status === 'idle' || status === 'loading') return;
    void setQ({ step: index || null, speed: speed === 10 ? null : speed });
  }, [status, index, speed]);

  const tab = q.tab ?? 'run';
  return (
    <Card
      size="md"
      render={<section />}
      aria-label={`${shownEntry.name}: the run`}
      className="flex flex-col gap-4 px-(--card-spacing)"
    >
      {opened ? (
        <Alert
          variant="info"
          role="status"
          className="flex flex-wrap items-center justify-between gap-2"
        >
          <span>
            Showing a {opened.result ? 'saved' : 'recent'} run of {shownEntry.name} from{' '}
            {new Date(opened.savedAt).toLocaleString()}.
          </span>
          <Button
            size="sm"
            onClick={() => setOpened(null)}
          >
            Back to this page&apos;s run
          </Button>
        </Alert>
      ) : null}
      <Tabs
        label="Run or code"
        value={tab}
        onChange={(t) => void setQ({ tab: t === 'code' || t === 'real' ? t : null })}
        tabs={[
          { id: 'run', label: entry.traced || entry.input === 'ops' ? 'Visualize' : 'Result' },
          { id: 'code', label: 'Code' },
          ...(shownEntry.realWorld.length ? [{ id: 'real', label: 'Real world' }] : [])
        ]}
      >
        {tab === 'real' && shownEntry.realWorld.length ? (
          <RealWorldCode
            api={api}
            entry={shownEntry}
          />
        ) : tab === 'code' ? (
          <CodePanel
            api={api}
            entry={shownEntry}
          />
        ) : outcome.ok ? (
          <div className="grid gap-4 xl:grid-cols-[minmax(0,1fr)_20rem]">
            <RunView
              result={outcome.value}
              input={shownInput}
              entry={shownEntry}
              player={player}
              initialStep={opened ? 0 : followEnd ? Infinity : initialStep}
              onInputChange={opened ? undefined : onInputChange}
            />
            {outcome.value.events.length ? (
              <TraceView
                events={outcome.value.events.map((e) => ({ ...e, kind: kinds[e.kind] ?? String(e.kind) }))}
                current={player.state.index - 1}
                onSeek={(i) => player.send({ type: 'SEEK', index: i + 1 })}
                truncated={outcome.value.truncated}
                height={360}
              />
            ) : null}
          </div>
        ) : (
          <Alert
            variant="destructive"
            role="alert"
          >
            <p className="font-medium text-danger">{outcome.failure.error}</p>
            {outcome.failure.problems.length ? (
              <ul className="mt-1 list-disc pl-5 font-mono text-xs">
                {outcome.failure.problems.map((p, i) => (
                  <li key={i}>
                    {p.line ? `line ${p.line}` : `item ${p.column ?? '?'}`}: &quot;{p.value}&quot;
                    is {p.reason}
                  </li>
                ))}
              </ul>
            ) : null}
          </Alert>
        )}
      </Tabs>
      <p className="text-sm">
        <Link
          href={`/quiz/?entry=${shownEntry.id}`}
          className={buttonVariants({ variant: 'secondary', size: 'sm' })}
        >
          Quiz me on {shownEntry.name}
        </Link>
      </p>
      <RunToolbar
        api={api}
        entry={shownEntry}
        input={shownInput}
        result={outcome.ok ? outcome.value : null}
        player={player}
        onOpen={(r) => {
          setOpened(r);
          void setQ({ tab: null, step: null });
        }}
      />
    </Card>
  );
}
