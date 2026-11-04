'use client';

import { useMemo, type ReactNode } from 'react';
import Link from 'next/link';
import { parseAsString, useQueryStates } from 'nuqs';
import { LevelBadge } from '@/components/algorithm/level-badge';
import { TLDRCard } from '@/components/algorithm/tldr-card';
import { ListDetail, SearchList } from '@/components/shared/list-detail';
import { WasmGate } from '@/components/shared/wasm-gate';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Button, buttonVariants } from '@/components/ui/button';
import { Empty, EmptyDescription, EmptyHeader, EmptyTitle } from '@/components/ui/empty';
import { FramePanel } from '@/components/ui/frame';
import { Heading } from '@/components/ui/heading';
import { Item, ItemContent, ItemDescription, ItemGroup, ItemTitle } from '@/components/ui/item';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow
} from '@/components/ui/table';
import { docsHref } from '@/lib/entries';
import { RunView, usePlayer } from '@/visualizer';
import type { AlgorithmApi, Catalog, Designs, DesignStudy, Entry } from '@/visualizer/core';
import { Knobs } from './knobs';
import { Metrics } from './metrics';

/** Knob values in the link: "users:1000,reads:20". */
function readKnobs(s: string): Record<string, number> {
  const out: Record<string, number> = {};
  for (const pair of s.split(',')) {
    const [k, v] = pair.split(':');
    if (k && v !== undefined && Number.isFinite(Number(v))) out[k] = Number(v);
  }
  return out;
}

const writeKnobs = (o: Record<string, number>) =>
  Object.entries(o)
    .map(([k, v]) => `${k}:${v}`)
    .join(',');

/** What the design must do and how well, and its trade-offs and failures. */
function StudyBrief({ study }: { study: DesignStudy }) {
  return (
    <>
      <FramePanel>
        <div className="grid gap-6 md:grid-cols-2">
          <section aria-labelledby="functional">
            <Heading
              level={3}
              id="functional"
              className="mb-2 text-sm font-semibold"
            >
              Functional requirements
            </Heading>
            <ul className="list-disc space-y-1 pl-5 text-sm">
              {study.functional.map((f) => (
                <li key={f}>{f}</li>
              ))}
            </ul>
          </section>
          <section aria-labelledby="nonfunctional">
            <Heading
              level={3}
              id="nonfunctional"
              className="mb-2 text-sm font-semibold"
            >
              Non-functional requirements
            </Heading>
            <ul className="list-disc space-y-1 pl-5 text-sm">
              {study.nonFunctional.map((f) => (
                <li key={f}>{f}</li>
              ))}
            </ul>
          </section>
        </div>
      </FramePanel>
      <FramePanel>
        <Heading
          level={3}
          className="mb-2 text-sm font-semibold"
        >
          Trade-offs
        </Heading>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead scope="col">Choice</TableHead>
              <TableHead scope="col">Gains</TableHead>
              <TableHead scope="col">Costs</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {study.tradeoffs.map((t) => (
              <TableRow key={t.choice}>
                <TableCell className="font-medium">{t.choice}</TableCell>
                <TableCell>{t.gain}</TableCell>
                <TableCell>{t.cost}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </FramePanel>
      <FramePanel>
        <Heading
          level={3}
          className="mb-2 text-sm font-semibold"
        >
          Failure modes
        </Heading>
        <ItemGroup className="gap-2">
          {study.failures.map((f) => (
            <Item
              key={f.what}
              size="sm"
              variant="muted"
              role="listitem"
            >
              <ItemContent>
                <ItemTitle>{f.what}</ItemTitle>
                <ItemDescription>{f.handling}</ItemDescription>
              </ItemContent>
            </Item>
          ))}
        </ItemGroup>
      </FramePanel>
    </>
  );
}

/** The estimate, the live architecture with its knobs, and the run's numbers. */
function StudyRun({
  api,
  catalog,
  study,
  knobs,
  setKnobs
}: {
  api: AlgorithmApi;
  catalog: Catalog;
  study: DesignStudy;
  knobs: Record<string, number>;
  setKnobs: (next: Record<string, number>) => void;
}) {
  const entry = catalog.entries.find((e) => e.id === study.id);
  const player = usePlayer({ speed: 4 });
  const options = useMemo(() => {
    const o: Record<string, number> = {};
    for (const opt of entry?.options ?? []) o[opt.name] = knobs[opt.name] ?? opt.default;
    return o;
  }, [entry, knobs]);
  const key = JSON.stringify([study.id, options]);
  const run = useMemo(() => (entry ? api.run(entry.id, { options }, 5000) : null), [api, key]);
  const estimate = useMemo(
    () => api.run('back-of-the-envelope', { options: study.estimate }, 0),
    [api, study.id]
  );
  if (!entry) {
    return (
      <Empty>
        <EmptyHeader>
          <EmptyTitle>Not in this build</EmptyTitle>
          <EmptyDescription>The simulation {study.id} is not in the catalog yet.</EmptyDescription>
        </EmptyHeader>
      </Empty>
    );
  }
  const related = study.related
    .map((id) => catalog.entries.find((e) => e.id === id))
    .filter((e): e is Entry => Boolean(e));
  return (
    <>
      {estimate.ok ? (
        <FramePanel>
          <Heading
            level={3}
            className="mb-1 text-sm font-semibold"
          >
            The estimate
          </Heading>
          <p className="text-muted mb-3 text-sm">{estimate.value.output.lines[2]}</p>
          <Metrics
            label="Estimate"
            summary={estimate.value.output.summary}
          />
        </FramePanel>
      ) : null}
      <FramePanel>
        <Heading
          level={3}
          className="mb-1 text-sm font-semibold"
        >
          Try it
        </Heading>
        <p className="text-muted mb-4 text-sm">
          Change a knob and the design runs again in your browser. Each request is drawn by its
          status: a ring for a cache hit, a hollow dot for a miss, a square while queued.
        </p>
        <Knobs
          options={entry.options}
          values={options}
          onChange={(name, value) => setKnobs({ ...knobs, [name]: value })}
        />
        <div className="mt-4 flex flex-wrap items-center gap-2">
          <Link
            href={`/visualize/?id=${entry.id}`}
            className={buttonVariants({ variant: 'primary', size: 'sm' })}
          >
            Open in the visualizer
          </Link>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => setKnobs({})}
            disabled={Object.keys(knobs).length < 1}
          >
            Reset the knobs
          </Button>
        </div>
      </FramePanel>
      <FramePanel>
        {run && run.ok ? (
          <div className="flex flex-col gap-4">
            <RunView
              key={key}
              result={run.value}
              input={{ options }}
              entry={entry}
              player={player}
              label={`${study.title}: the architecture`}
            />
            <Metrics
              label="The run"
              summary={run.value.output.summary}
            />
            <ul className="list-disc space-y-1 pl-5 text-sm">
              {run.value.output.lines.map((l) => (
                <li key={l}>{l}</li>
              ))}
            </ul>
          </div>
        ) : (
          <Alert variant="destructive">
            <AlertTitle>The simulation rejected these settings</AlertTitle>
            <AlertDescription>{run && !run.ok ? run.failure.error : 'No run.'}</AlertDescription>
          </Alert>
        )}
      </FramePanel>
      {related.length ? (
        <FramePanel>
          <Heading
            level={3}
            className="mb-2 text-sm font-semibold"
          >
            Run next
          </Heading>
          <ul className="flex flex-wrap gap-2">
            {related.map((e) => (
              <li key={e.id}>
                <Link
                  href={`/visualize/?id=${e.id}`}
                  className="text-sm underline underline-offset-4"
                >
                  {e.name}
                </Link>
              </li>
            ))}
          </ul>
        </FramePanel>
      ) : null}
    </>
  );
}

/**
 * A data pattern or sizing tool, opened from the list: what it is and costs,
 * its knobs, and a live run of its example.
 */
function PatternRun({
  api,
  entry,
  knobs,
  setKnobs
}: {
  api: AlgorithmApi;
  entry: Entry;
  knobs: Record<string, number>;
  setKnobs: (next: Record<string, number>) => void;
}) {
  const player = usePlayer({ speed: 4 });
  const options = useMemo(() => {
    const o: Record<string, number> = {};
    for (const opt of entry.options) o[opt.name] = knobs[opt.name] ?? opt.default;
    return o;
  }, [entry, knobs]);
  const input = useMemo(
    () => ({ ...entry.example, ...(entry.options.length ? { options } : {}) }),
    [entry, options]
  );
  const key = JSON.stringify([entry.id, options]);
  const run = useMemo(() => api.run(entry.id, input, 5000), [api, key]);
  return (
    <>
      <FramePanel>
        <dl className="grid gap-4 text-sm sm:grid-cols-2">
          <div>
            <dt className="text-muted text-xs font-semibold uppercase">Time</dt>
            <dd className="font-mono">{entry.time}</dd>
          </div>
          <div>
            <dt className="text-muted text-xs font-semibold uppercase">Space</dt>
            <dd className="font-mono">{entry.space}</dd>
          </div>
        </dl>
        {entry.why ? <p className="mt-4 text-sm">{entry.why}</p> : null}
        {entry.notes ? (
          <TLDRCard
            notes={entry.notes}
            className="mt-4 w-full"
          />
        ) : null}
      </FramePanel>
      {entry.options.length ? (
        <FramePanel>
          <Heading
            level={3}
            className="mb-1 text-sm font-semibold"
          >
            Try it
          </Heading>
          <p className="text-muted mb-4 text-sm">
            Change a knob and it runs again in your browser.
          </p>
          <Knobs
            options={entry.options}
            values={options}
            onChange={(name, value) => setKnobs({ ...knobs, [name]: value })}
          />
          <Button
            variant="ghost"
            size="sm"
            className="mt-4"
            onClick={() => setKnobs({})}
            disabled={Object.keys(knobs).length < 1}
          >
            Reset the knobs
          </Button>
        </FramePanel>
      ) : null}
      <FramePanel>
        {run.ok ? (
          <div className="flex flex-col gap-4">
            <RunView
              key={key}
              result={run.value}
              input={input}
              entry={entry}
              player={player}
              label={`${entry.name}: the run`}
            />
            {/* The run view prints its own numbers and lines. */}
          </div>
        ) : (
          <Alert variant="destructive">
            <AlertTitle>The run was rejected</AlertTitle>
            <AlertDescription>{run.failure.error}</AlertDescription>
          </Alert>
        )}
      </FramePanel>
      <FramePanel className="flex flex-wrap gap-2">
        <Link
          href={`/visualize/?id=${entry.id}`}
          className={buttonVariants({ variant: 'primary', size: 'sm' })}
        >
          Open in the visualizer
        </Link>
        <Link
          href={docsHref(entry.id)}
          className={buttonVariants({ variant: 'ghost', size: 'sm' })}
        >
          Read the docs
        </Link>
      </FramePanel>
    </>
  );
}

/** The hub once the case studies have loaded from the wasm build. */
function Hub({
  api,
  catalog,
  designs,
  intro
}: {
  api: AlgorithmApi;
  catalog: Catalog;
  designs: Designs;
  intro?: ReactNode;
}) {
  const [q, setQ] = useQueryStates(
    { study: parseAsString, pattern: parseAsString, knobs: parseAsString.withDefault('') },
    { history: 'replace' }
  );
  const byId = useMemo(() => new Map(catalog.entries.map((e) => [e.id, e])), [catalog.entries]);
  const entriesOf = (ids: string[]) => ids.flatMap((id) => byId.get(id) ?? []);
  const patterns = entriesOf(designs.patterns);
  const foundations = entriesOf(designs.foundations);

  const chosenStudy = designs.studies.find((c) => c.id === q.study);
  const chosenPattern = chosenStudy
    ? undefined
    : [...patterns, ...foundations].find((e) => e.id === q.pattern);
  // Nothing chosen: a wide screen shows the first study, a phone the list.
  const study = chosenStudy ?? (chosenPattern ? undefined : designs.studies[0]);
  const knobs = readKnobs(q.knobs);
  const setKnobs = (next: Record<string, number>) => void setQ({ knobs: writeKnobs(next) || null });

  if (!study && !chosenPattern) {
    return (
      <Empty>
        <EmptyHeader>
          <EmptyTitle>No case studies</EmptyTitle>
          <EmptyDescription>This build carries no system design case studies.</EmptyDescription>
        </EmptyHeader>
      </Empty>
    );
  }

  const asItem = (e: Entry) => ({
    id: e.id,
    title: e.name,
    subtitle: e.summary,
    keywords: [e.level, ...e.tags],
    href: `/system-design/?pattern=${e.id}`
  });
  const list = (
    <SearchList
      label="System design"
      placeholder="Search studies and patterns..."
      emptyHint="Try a shorter word, such as cache, shard or queue."
      current={chosenPattern?.id ?? study?.id}
      groups={[
        {
          label: 'Case studies',
          items: designs.studies.map((c) => ({
            id: c.id,
            title: c.title,
            subtitle: c.level,
            keywords: [c.pitch],
            href: `/system-design/?study=${c.id}`
          }))
        },
        { label: 'Data patterns', items: patterns.map(asItem) },
        { label: 'Sizing and reliability', items: foundations.map(asItem) }
      ]}
    />
  );

  return (
    <ListDetail
      list={list}
      intro={intro}
      open={!!(chosenStudy || chosenPattern)}
      onClose={() => void setQ({ study: null, pattern: null, knobs: null })}
      detailKey={chosenPattern?.id ?? study?.id ?? ''}
      title={
        <>
          <span className="font-heading font-semibold">{chosenPattern?.name ?? study?.title}</span>
          {(chosenPattern?.level ?? study?.level) ? (
            <LevelBadge level={(chosenPattern?.level ?? study?.level) as string} />
          ) : null}
        </>
      }
      description={chosenPattern?.summary ?? study?.pitch}
    >
      <div className="flex flex-col gap-1">
        {chosenPattern ? (
          <PatternRun
            key={chosenPattern.id}
            api={api}
            entry={chosenPattern}
            knobs={knobs}
            setKnobs={setKnobs}
          />
        ) : study ? (
          <>
            <StudyBrief study={study} />
            <StudyRun
              key={study.id}
              api={api}
              catalog={catalog}
              study={study}
              knobs={knobs}
              setKnobs={setKnobs}
            />
          </>
        ) : null}
      </div>
    </ListDetail>
  );
}

/**
 * The system design hub: the case studies, each with its requirements,
 * estimate, trade-offs and failure modes, and a live architecture that runs
 * the real simulation with knobs; then the data patterns and the sizing and
 * reliability tools. Everything comes from the wasm build (pkg/systems/design
 * Studies); the study and the knobs live in the link.
 */
export function SystemDesignHub({ intro }: { intro?: ReactNode }) {
  return (
    <WasmGate>
      {(api, catalog) => {
        const designs = api.designs();
        if (!designs.ok) {
          return (
            <Alert variant="destructive">
              <AlertTitle>The case studies could not be loaded</AlertTitle>
              <AlertDescription>{designs.failure.error}</AlertDescription>
            </Alert>
          );
        }
        return (
          <Hub
            api={api}
            catalog={catalog}
            designs={designs.value}
            intro={intro}
          />
        );
      }}
    </WasmGate>
  );
}
