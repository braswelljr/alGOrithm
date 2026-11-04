'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { Heading } from '@braswelljr/algorithm-ui/components/heading';
import type { LearningPath } from '@braswelljr/algorithm-visualizer/core';
import { Badge } from '@/components/ui/badge';
import { Card } from '@/components/ui/card';
import { pathHref } from '@/lib/learn';
import { readDone, readPlacement } from '@/lib/learn-progress';
import { PlacementTest } from './placement-test';

function PathCard({
  path,
  titles,
  recommended,
  done
}: {
  path: LearningPath;
  titles: Map<string, string>;
  recommended: boolean;
  done: boolean;
}) {
  return (
    <Card
      variant={recommended ? 'selected' : 'default'}
      className="flex h-full flex-col gap-2"
    >
      {recommended || done ? (
        <p className="flex flex-wrap gap-2 text-xs font-medium">
          {recommended ? <span className="text-accent">Recommended for you</span> : null}
          {done ? <span className="text-success">✓ Done</span> : null}
        </p>
      ) : null}
      <Heading
        level={2}
        className="text-lg font-semibold"
      >
        <Link
          className="underline decoration-accent decoration-2 underline-offset-4 hover:text-accent"
          href={pathHref(path.id)}
        >
          {path.title}
        </Link>
      </Heading>
      <p className="text-sm text-muted">{path.summary}</p>
      <p className="mt-auto flex flex-wrap items-center gap-2 pt-2 text-xs text-muted">
        <Badge>{path.level}</Badge>
        <span>{path.steps.length} steps</span>
        <span>for {path.audience}</span>
      </p>
      {path.prerequisites.length > 0 ? (
        <p className="text-xs text-muted">
          First:{' '}
          {path.prerequisites.map((id, i) => (
            <span key={id}>
              {i > 0 ? ', ' : null}
              <Link
                className="underline"
                href={pathHref(id)}
              >
                {titles.get(id) ?? id}
              </Link>
            </span>
          ))}
        </p>
      ) : null}
    </Card>
  );
}

function Paths({ paths, level, done }: { paths: LearningPath[]; level?: string; done: string[] }) {
  const titles = new Map(paths.map((p) => [p.id, p.title]));
  return (
    <ul className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
      {paths.map((p) => (
        <li key={p.id}>
          <PathCard
            path={p}
            titles={titles}
            recommended={!!level && p.level === level}
            done={done.includes(p.id)}
          />
        </li>
      ))}
    </ul>
  );
}

/**
 * Every learning path, easiest first, as the Go side lists them. The paths
 * come from the build (buildPaths), so the page needs no wasm to show them.
 */
export function LearnIndex({ paths }: { paths: LearningPath[] }) {
  const [level, setLevel] = useState<string | undefined>();
  const [done, setDone] = useState<string[]>([]);
  useEffect(() => {
    setLevel(readPlacement()?.level);
    setDone(readDone());
  }, []);
  return (
    <div className="flex flex-col gap-6">
      <PlacementTest
        paths={paths}
        onPlaced={(p) => setLevel(p.level)}
      />
      <Paths
        paths={paths}
        level={level}
        done={done}
      />
    </div>
  );
}
