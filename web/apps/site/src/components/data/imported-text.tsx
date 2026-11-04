'use client';

import { useId, useMemo, useState } from 'react';
import { Checkbox } from '@braswelljr/algorithm-ui/components/checkbox';
import type { AlgorithmApi as Api, Outcome, Preview } from '@braswelljr/algorithm-visualizer/core';
import { Button } from '@/components/ui/button';
import { Field, Input } from '@/components/ui/field';
import { Select } from '@/components/ui/select';
import { formatOf, type Format } from './formats';
import { Problems } from './problems';

/**
 * Reads text in one of the three formats: for CSV, a preview, a header
 * switch and a column picker; for JSON, the field holding the values.
 */
export function ImportedText({
  api,
  text,
  format: initialFormat,
  source,
  onUse
}: {
  api: Api;
  text: string;
  format: Format;
  source: string;
  onUse: (values: number[], label: string) => void;
}) {
  const id = useId();
  const [format, setFormat] = useState<Format>(initialFormat);
  const [header, setHeader] = useState(true);
  const [column, setColumn] = useState('1');
  const [field, setField] = useState('');
  const [round, setRound] = useState(false);
  const [skipEmpty, setSkipEmpty] = useState(false);

  const preview: Outcome<Preview> | null = useMemo(
    () => (format === 'csv' ? api.preview(text, header) : null),
    [api, format, text, header]
  );
  const parsed = useMemo(
    () => api.parse(format, text, { column, header, field, round, skipEmpty }),
    [api, format, text, column, header, field, round, skipEmpty]
  );

  return (
    <div className="flex flex-col gap-3">
      <div className="grid gap-3 sm:grid-cols-3">
        <Field
          label="Format"
          htmlFor={`${id}-format`}
        >
          <Select
            id={`${id}-format`}
            value={format}
            onChange={(e) => setFormat(formatOf(`.${e.target.value}`))}
          >
            <option value="text">Text: numbers separated by commas, spaces or lines</option>
            <option value="csv">CSV</option>
            <option value="json">JSON</option>
          </Select>
        </Field>
        {format === 'json' ? (
          <Field
            label="Field"
            htmlFor={`${id}-field`}
            hint="For an array of objects, or an object holding the array."
          >
            <Input
              id={`${id}-field`}
              value={field}
              onChange={(e) => setField(e.target.value)}
              placeholder="price"
            />
          </Field>
        ) : null}
        {format === 'csv' && preview?.ok ? (
          <Field
            label="Column"
            htmlFor={`${id}-column`}
          >
            <Select
              id={`${id}-column`}
              value={column}
              onChange={(e) => setColumn(e.target.value)}
            >
              {Array.from({ length: preview.value.width }, (_, i) => (
                <option
                  key={i}
                  value={String(i + 1)}
                >
                  {i + 1}. {preview.value.header[i] ?? `column ${i + 1}`}
                </option>
              ))}
            </Select>
          </Field>
        ) : null}
      </div>
      <div className="flex flex-wrap gap-4 text-sm">
        {format === 'csv' ? (
          <>
            <label className="flex items-center gap-2">
              <Checkbox
                checked={header}
                onCheckedChange={(on) => setHeader(on)}
              />{' '}
              First row is a header
            </label>
            <label className="flex items-center gap-2">
              <Checkbox
                checked={skipEmpty}
                onCheckedChange={(on) => setSkipEmpty(on)}
              />{' '}
              Skip empty cells
            </label>
          </>
        ) : null}
        <label className="flex items-center gap-2">
          <Checkbox
            checked={round}
            onCheckedChange={(on) => setRound(on)}
          />{' '}
          Round decimals
        </label>
      </div>
      {format === 'csv' && preview?.ok ? (
        <div className="overflow-x-auto rounded-md border border-border">
          <table className="w-full text-left font-mono text-xs">
            <caption className="sr-only">The first rows of the file</caption>
            {preview.value.header.length > 0 ? (
              <thead className="bg-subtle">
                <tr>
                  {preview.value.header.map((h, i) => (
                    <th
                      key={i}
                      scope="col"
                      className={
                        i + 1 === Number(column) ? 'px-2 py-1 text-accent underline' : 'px-2 py-1'
                      }
                    >
                      {h}
                    </th>
                  ))}
                </tr>
              </thead>
            ) : null}
            <tbody>
              {preview.value.rows.map((r, i) => (
                <tr
                  key={i}
                  className="border-t border-border"
                >
                  {r.map((c, j) => (
                    <td
                      key={j}
                      className="px-2 py-1"
                    >
                      {c}
                    </td>
                  ))}
                </tr>
              ))}
            </tbody>
          </table>
          <p className="px-2 py-1 text-xs text-muted">{preview.value.count} rows in all</p>
        </div>
      ) : null}
      {format === 'csv' && preview && !preview.ok ? (
        <Problems
          error={preview.failure.error}
          problems={preview.failure.problems}
        />
      ) : null}
      {parsed.ok ? (
        <div className="flex flex-wrap items-center gap-3">
          <Button
            variant="primary"
            onClick={() => onUse(parsed.value, `${parsed.value.length} values from ${source}`)}
          >
            Use {parsed.value.length} values
          </Button>
          <span className="truncate font-mono text-xs text-muted">
            {parsed.value.slice(0, 12).join(', ')}
            {parsed.value.length > 12 ? ', ...' : ''}
          </span>
        </div>
      ) : text.trim() !== '' && !(format === 'csv' && preview && !preview.ok) ? (
        <Problems
          error={parsed.failure.error}
          problems={parsed.failure.problems}
        />
      ) : null}
    </div>
  );
}
