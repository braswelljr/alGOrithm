'use client';

import { useId, useMemo, type ReactNode } from 'react';
import { parseAsInteger, parseAsString, useQueryStates } from 'nuqs';
import type { UseFormReturn } from 'react-hook-form';
import { Button } from '@/components/shared/button';
import { Field, Textarea } from '@/components/shared/field';
import { NumberInput } from '@/components/shared/number-input';
import { FacetFilter } from '@/components/ui/facet-filter';
import {
  decodeGrid,
  encodeGrid,
  gridToText,
  parseGridText,
  SAMPLE_GRID,
  toggleWall
} from '@/lib/grid-text';
import { GridView } from '@/visualizer';
import type { AlgorithmApi, Catalog, Entry, Input as RunInput } from '@/visualizer/core';
import { InputDialog } from './input-dialog';

/** Grids up to this many cells go into the link; larger ones are regenerated from their seed. */
const URL_CELLS = 1600;

/** The grid's settings, as the form holds them: `text` is a drawn grid, or null for the generated one. */
type GridValues = { kind: string; rows: number; cols: number; seed: number; text: string | null };

/** The grid the settings describe: the drawn one, else the generated one. */
function useGridText(api: AlgorithmApi, v: GridValues) {
  const generated = useMemo(() => {
    const r = api.generateGrid(v.kind, v.rows, v.cols, v.seed);
    return r.ok ? r.value : SAMPLE_GRID;
  }, [api, v.kind, v.rows, v.cols, v.seed]);
  return v.text ?? generated;
}

/** The dialog's fields: kind, size and seed, the grid itself to draw on, and its text. */
function GridForm({
  api,
  kinds,
  form
}: {
  api: AlgorithmApi;
  kinds: string[];
  form: UseFormReturn<GridValues>;
}) {
  const id = useId();
  const v = form.watch();
  const text = useGridText(api, v);
  const grid = useMemo(() => parseGridText(text), [text]);
  // A new kind, size or seed generates again, dropping a drawn grid.
  const regenerate = (patch: Partial<GridValues>) => {
    for (const [k, val] of Object.entries(patch))
      form.setValue(k as keyof GridValues, val as never, { shouldDirty: true });
    form.setValue('text', null, { shouldDirty: true });
  };
  return (
    <section
      aria-label="The grid"
      className="flex flex-col gap-4"
    >
      <div className="grid gap-3 sm:grid-cols-[auto_1fr_1fr_1fr_auto] sm:items-end">
        <div className="flex flex-col gap-2">
          <span className="text-sm font-medium">Kind</span>
          <FacetFilter
            single
            title="Kind"
            size="default"
            searchable={false}
            options={kinds.map((k) => ({ value: k, label: k }))}
            value={[v.kind]}
            onValueChange={(next) => regenerate({ kind: next[0] ?? v.kind })}
          />
        </div>
        <Field
          label="Rows"
          htmlFor={`${id}-rows`}
        >
          <NumberInput
            id={`${id}-rows`}
            min={3}
            max={99}
            value={v.rows}
            onValueChange={(n) => regenerate({ rows: Math.max(3, Math.min(99, n ?? 3)) })}
          />
        </Field>
        <Field
          label="Columns"
          htmlFor={`${id}-cols`}
        >
          <NumberInput
            id={`${id}-cols`}
            min={3}
            max={99}
            value={v.cols}
            onValueChange={(n) => regenerate({ cols: Math.max(3, Math.min(99, n ?? 3)) })}
          />
        </Field>
        <Field
          label="Seed"
          htmlFor={`${id}-seed`}
        >
          <NumberInput
            id={`${id}-seed`}
            value={v.seed}
            onValueChange={(n) => regenerate({ seed: n ?? 0 })}
          />
        </Field>
        <Button
          type="button"
          onClick={() => regenerate({ seed: v.seed + 1 })}
        >
          New {v.kind}
        </Button>
      </div>
      {grid ? (
        <div className="flex flex-col gap-2">
          <p className="text-muted text-sm">
            Click a cell to add or remove a wall. S is the start, E the end; a number is the cost of
            entering that cell.
          </p>
          <div className="av-root w-full">
            <GridView
              grid={grid}
              label="Draw the grid"
              onToggle={(cell) =>
                form.setValue('text', gridToText(toggleWall(grid, cell)), { shouldDirty: true })
              }
            />
          </div>
        </div>
      ) : (
        <p className="text-danger text-sm">
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
            onChange={(e) => form.setValue('text', e.target.value, { shouldDirty: true })}
            spellCheck={false}
            className="font-mono text-sm"
          />
        </Field>
        {v.text !== null ? (
          <Button
            type="button"
            size="sm"
            className="mt-2"
            onClick={() => form.setValue('text', null, { shouldDirty: true })}
          >
            Back to the generated grid
          </Button>
        ) : null}
      </details>
    </section>
  );
}

/**
 * The grid a pathfinder runs on: generated from a kind, a size and a seed, or
 * drawn: click a cell to toggle a wall, or paste a grid as text. Set in the
 * configure dialog; the link keeps what the run uses.
 */
export function GridInput({
  api,
  catalog,
  entry,
  children
}: {
  api: AlgorithmApi;
  catalog: Catalog;
  entry: Entry;
  children: (input: RunInput | null) => ReactNode;
}) {
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
  const current: GridValues = {
    kind: q.gk ?? 'maze',
    rows: q.gr ?? 15,
    cols: q.gc ?? 31,
    seed: q.gs ?? 1,
    text: q.g ? decodeGrid(q.g) : null
  };
  const text = useGridText(api, current);
  const grid = useMemo(() => parseGridText(text), [text]);

  return (
    <div className="flex flex-col gap-6">
      <InputDialog<GridValues>
        entry={entry}
        wide
        values={current}
        summary={
          current.text !== null
            ? `your own grid${grid ? `, ${grid.rows} by ${grid.cols}` : ''}`
            : `a ${current.kind}, ${current.rows} by ${current.cols}, seed ${current.seed}`
        }
        onApply={(v) => {
          const drawn = v.text === null ? null : parseGridText(v.text);
          void setQ({
            gk: v.kind,
            gr: v.rows,
            gc: v.cols,
            gs: v.seed,
            g:
              v.text !== null && drawn && drawn.cost.length <= URL_CELLS ? encodeGrid(v.text) : null
          });
        }}
      >
        {(form) => (
          <GridForm
            api={api}
            kinds={kinds}
            form={form}
          />
        )}
      </InputDialog>
      {children(grid ? { text } : null)}
    </div>
  );
}
