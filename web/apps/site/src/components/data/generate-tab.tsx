'use client';

import { useId, useState } from 'react';
import { Button } from '@/components/ui/button';
import { Field, Input } from '@/components/ui/field';
import { Select } from '@/components/ui/select';
import type { Dataset } from '@/hooks/use-dataset';

/** A generated dataset: a shape, a size and a seed, all kept in the link. */
export function GenerateTab({
  kinds,
  dataset,
  maxValues
}: {
  kinds: string[];
  dataset: Dataset;
  maxValues: number;
}) {
  const id = useId();
  const [kind, setKind] = useState(
    kinds.includes(dataset.kind) ? dataset.kind : (kinds[0] ?? 'random')
  );
  const [size, setSize] = useState(String(dataset.size));
  const [seed, setSeed] = useState(String(dataset.seed));
  const n = Number(size);
  const s = Number(seed);
  const valid = Number.isInteger(n) && n >= 1 && n <= maxValues && Number.isInteger(s);
  return (
    <form
      className="grid gap-3 sm:grid-cols-[2fr_1fr_1fr_auto] sm:items-end"
      onSubmit={(e) => {
        e.preventDefault();
        if (valid) dataset.generate(kind, n, s);
      }}
    >
      <Field
        label="Shape"
        htmlFor={`${id}-kind`}
      >
        <Select
          id={`${id}-kind`}
          value={kind}
          onChange={(e) => setKind(e.target.value)}
        >
          {kinds.map((k) => (
            <option
              key={k}
              value={k}
            >
              {k.replace('-', ' ')}
            </option>
          ))}
        </Select>
      </Field>
      <Field
        label="Size"
        htmlFor={`${id}-size`}
      >
        <Input
          id={`${id}-size`}
          type="number"
          min={1}
          max={maxValues}
          value={size}
          onChange={(e) => setSize(e.target.value)}
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
          onChange={(e) => setSeed(e.target.value)}
        />
      </Field>
      <Button
        type="submit"
        variant="primary"
        disabled={!valid}
      >
        Generate
      </Button>
      {!valid ? (
        <p className="text-sm text-danger sm:col-span-4">
          Size must be a whole number from 1 to {maxValues}, and the seed a whole number.
        </p>
      ) : null}
    </form>
  );
}
