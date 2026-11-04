'use client';

import { useMemo } from 'react';
import Link from 'next/link';
import { Alert, AlertDescription, AlertTitle } from '@braswelljr/algorithm-ui/components/alert';
import { Badge } from '@braswelljr/algorithm-ui/components/badge';
import { Button, buttonVariants } from '@braswelljr/algorithm-ui/components/button';
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyTitle
} from '@braswelljr/algorithm-ui/components/empty';
import {
  Frame,
  FrameDescription,
  FrameHeader,
  FramePanel,
  FrameTitle
} from '@braswelljr/algorithm-ui/components/frame';
import { Heading } from '@braswelljr/algorithm-ui/components/heading';
import {
  Item,
  ItemContent,
  ItemDescription,
  ItemGroup,
  ItemTitle
} from '@braswelljr/algorithm-ui/components/item';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow
} from '@braswelljr/algorithm-ui/components/table';
import { RunView, usePlayer } from '@braswelljr/algorithm-visualizer';
import type {
  AlgorithmApi,
  Catalog,
  Designs,
  DesignStudy,
  Entry
} from '@braswelljr/algorithm-visualizer/core';
import { parseAsString, useQueryStates } from 'nuqs';
import { WasmGate } from '@/components/shared/wasm-gate';
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

/** The list of case studies; the one in the link is marked current. */
function StudyList({ studies, current }: { studies: DesignStudy[]; current: string }) {
  return (
    <nav aria-label="Case studies">
      <ItemGroup className="gap-1">
        {studies.map((c) => (
          <Item
            key={c.id}
            size="sm"
            variant={c.id === current ? 'muted' : 'default'}
            render={
              <Link
                href={`/system-design/?study=${c.id}`}
                aria-current={c.id === current ? 'page' : undefined}
              />
            }
          >
            <ItemContent>
              <ItemTitle>{c.title}</ItemTitle>
              <ItemDescription className="capitalize">{c.level}</ItemDescription>
            </ItemContent>
          </Item>
        ))}
      </ItemGroup>
    </nav>
  );
}

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
          <p className="mb-3 text-sm text-muted">{estimate.value.output.lines[2]}</p>
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
        <p className="mb-4 text-sm text-muted">
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
            disabled={Object.keys(knobs).length === 0}
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

/** A grid of entries by id, each linking to its run. */
function EntryLinks({ ids, catalog, label }: { ids: string[]; catalog: Catalog; label: string }) {
  return (
    <ItemGroup
      aria-label={label}
      className="grid gap-2 sm:grid-cols-2 lg:grid-cols-3"
    >
      {ids.map((id) => {
        const e = catalog.entries.find((x) => x.id === id);
        if (!e) return null;
        return (
          <Item
            key={id}
            size="sm"
            variant="muted"
            render={<Link href={`/visualize/?id=${id}`} />}
          >
            <ItemContent>
              <ItemTitle>{e.name}</ItemTitle>
              <ItemDescription>{e.summary}</ItemDescription>
            </ItemContent>
          </Item>
        );
      })}
    </ItemGroup>
  );
}

/** The hub once the case studies have loaded from the wasm build. */
function Hub({ api, catalog, designs }: { api: AlgorithmApi; catalog: Catalog; designs: Designs }) {
  const first = designs.studies[0]?.id ?? '';
  const [q, setQ] = useQueryStates(
    { study: parseAsString.withDefault(first), knobs: parseAsString.withDefault('') },
    { history: 'replace' }
  );
  const study = designs.studies.find((c) => c.id === q.study) ?? designs.studies[0];
  if (!study) {
    return (
      <Empty>
        <EmptyHeader>
          <EmptyTitle>No case studies</EmptyTitle>
          <EmptyDescription>This build carries no system design case studies.</EmptyDescription>
        </EmptyHeader>
      </Empty>
    );
  }
  const knobs = readKnobs(q.knobs);
  return (
    <div className="flex flex-col gap-10">
      <div className="grid gap-6 lg:grid-cols-[14rem_minmax(0,1fr)]">
        <StudyList
          studies={designs.studies}
          current={study.id}
        />
        <Frame>
          <FrameHeader>
            <FrameTitle className="flex flex-wrap items-center gap-2 text-base">
              <Heading
                level={2}
                className="text-lg font-semibold"
              >
                {study.title}
              </Heading>
              <Badge
                variant="secondary"
                className="capitalize"
              >
                {study.level}
              </Badge>
            </FrameTitle>
            <FrameDescription>{study.pitch}</FrameDescription>
          </FrameHeader>
          <StudyBrief study={study} />
          <StudyRun
            key={study.id}
            api={api}
            catalog={catalog}
            study={study}
            knobs={knobs}
            setKnobs={(next) => void setQ({ knobs: writeKnobs(next) || null })}
          />
        </Frame>
      </div>
      <section aria-labelledby="patterns">
        <Heading
          level={2}
          id="patterns"
          className="mb-1 text-lg font-semibold"
        >
          Data patterns
        </Heading>
        <p className="mb-4 max-w-3xl text-sm text-muted">
          The building blocks the case studies are made of: replication, sharding, consistency,
          events and the patterns that keep services up under load.
        </p>
        <EntryLinks
          ids={designs.patterns}
          catalog={catalog}
          label="Data patterns"
        />
      </section>
      <section aria-labelledby="foundations">
        <Heading
          level={2}
          id="foundations"
          className="mb-1 text-lg font-semibold"
        >
          Sizing and reliability
        </Heading>
        <p className="mb-4 max-w-3xl text-sm text-muted">
          Estimate before you design, give every record an ID, and know what your availability and
          your queues will do.
        </p>
        <EntryLinks
          ids={designs.foundations}
          catalog={catalog}
          label="Sizing and reliability"
        />
      </section>
    </div>
  );
}

/**
 * The system design hub: the case studies, each with its requirements,
 * estimate, trade-offs and failure modes, and a live architecture that runs
 * the real simulation with knobs; then the data patterns and the sizing and
 * reliability tools. Everything comes from the wasm build (pkg/systems/design
 * Studies); the study and the knobs live in the link.
 */
export function SystemDesignHub() {
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
          />
        );
      }}
    </WasmGate>
  );
}
