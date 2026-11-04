'use client';

import { useId, useMemo, type ReactNode } from 'react';
import { Card } from '@braswelljr/algorithm-ui/components/card';
import { GridView } from '@braswelljr/algorithm-visualizer';
import type {
  AlgorithmApi,
  Catalog,
  Input as RunInput
} from '@braswelljr/algorithm-visualizer/core';
import { parseAsInteger, parseAsString, useQueryStates } from 'nuqs';
import { Button } from '@/components/ui/button';
import { Field, Input, Textarea } from '@/components/ui/field';
import { Select } from '@/components/ui/select';
import {
  decodeGrid,
  encodeGrid,
  gridToText,
  parseGridText,
  SAMPLE_GRID,
  toggleWall
} from '@/lib/grid-text';

/** Grids up to this many cells go into the link; larger ones are regenerated from their seed. */
const URL_CELLS = 1600;

/**
 * The grid a pathfinder runs on: generated from a kind, a size and a seed, or
 * drawn: click a cell to toggle a wall, or paste a grid as text.
 */
export function GridInput({
  api,
  catalog,
  children
}: {
  api: AlgorithmApi;
  catalog: Catalog;
  children: (input: RunInput | null) => ReactNode;
}) {
  const id = useId();
  const kinds = catalog.gridKinds.length
    ? catalog.gridKinds
    : ['maze', 'walls', 'weighted', 'open'];
  const [q, setQ] = useQueryStates(
    {
      gk: parseAsString,
      gr: parseAsInteger,
      gc: parseAsInteger,
      gs: parseAsInteger,
      g: parseAsString
    },
    { history: 'replace' }
  );
  const kind = q.gk ?? 'maze';
  const rows = q.gr ?? 15;
  const cols = q.gc ?? 31;
  const seed = q.gs ?? 1;

  const generated = useMemo(() => {
    const r = api.generateGrid(kind, rows, cols, seed);
    return r.ok ? r.value : SAMPLE_GRID;
  }, [api, kind, rows, cols, seed]);
  const text = q.g ? decodeGrid(q.g) : generated;
  const grid = useMemo(() => parseGridText(text), [text]);

  const setText = (t: string) => {
    const g = parseGridText(t);
    void setQ({ g: g && g.cost.length <= URL_CELLS ? encodeGrid(t) : null });
  };

  return (
    <div className="flex flex-col gap-6">
      <Card
        size="md"
        render={<section />}
        aria-label="The grid"
        className="flex flex-col gap-4 px-(--card-spacing)"
      >
        <form
          className="grid gap-3 sm:grid-cols-[2fr_1fr_1fr_1fr_auto] sm:items-end"
          onSubmit={(e) => e.preventDefault()}
        >
          <Field
            label="Kind"
            htmlFor={`${id}-kind`}
          >
            <Select
              id={`${id}-kind`}
              value={kind}
              onChange={(e) => void setQ({ gk: e.target.value, g: null })}
            >
              {kinds.map((k) => (
                <option
                  key={k}
                  value={k}
                >
                  {k}
                </option>
              ))}
            </Select>
          </Field>
          <Field
            label="Rows"
            htmlFor={`${id}-rows`}
          >
            <Input
              id={`${id}-rows`}
              type="number"
              min={3}
              max={99}
              value={rows}
              onChange={(e) =>
                void setQ({ gr: Math.max(3, Math.min(99, Number(e.target.value) || 3)), g: null })
              }
            />
          </Field>
          <Field
            label="Columns"
            htmlFor={`${id}-cols`}
          >
            <Input
              id={`${id}-cols`}
              type="number"
              min={3}
              max={99}
              value={cols}
              onChange={(e) =>
                void setQ({ gc: Math.max(3, Math.min(99, Number(e.target.value) || 3)), g: null })
              }
            />
          </Field>
          <Field
            label="Seed"
            htmlFor={`${id}-seed`}
          >
            <Input
              id={`${id}-seed`}
              type="number"
              value={seed}
              onChange={(e) => void setQ({ gs: Number(e.target.value) || 0, g: null })}
            />
          </Field>
          <Button onClick={() => void setQ({ gs: seed + 1, g: null })}>New {kind}</Button>
        </form>
        {grid ? (
          <div className="flex flex-col gap-2">
            <p className="text-sm text-muted">
              Click a cell to add or remove a wall. S is the start, E the end; a number is the cost
              of entering that cell.
            </p>
            <div className="av-root mx-auto w-full max-w-3xl">
              <GridView
                grid={grid}
                label="Draw the grid"
                onToggle={(cell) => setText(gridToText(toggleWall(grid, cell)))}
              />
            </div>
          </div>
        ) : (
          <p className="text-sm text-danger">
            This grid cannot be read: use # . 1-9 S E, at most 100 by 100.
          </p>
        )}
        <details>
          <summary className="cursor-pointer text-sm font-medium">Edit as text</summary>
          <Field
            label="Grid"
            htmlFor={`${id}-text`}
            hint="One row per line: # wall, . open, 1-9 cost, S start, E end."
            className="mt-2"
          >
            <Textarea
              id={`${id}-text`}
              value={text}
              rows={Math.min(20, text.split('\n').length + 1)}
              onChange={(e) => setText(e.target.value)}
              spellCheck={false}
            />
          </Field>
          {q.g ? (
            <Button
              size="sm"
              className="mt-2"
              onClick={() => void setQ({ g: null })}
            >
              Back to the generated grid
            </Button>
          ) : null}
        </details>
      </Card>
      {children(grid ? { text } : null)}
    </div>
  );
}
