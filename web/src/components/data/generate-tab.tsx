'use client';

import { useId, useState } from 'react';
import { CircleAlert } from 'lucide-react';
import { Button } from '@/components/shared/button';
import { Field } from '@/components/shared/field';
import { NumberInput } from '@/components/shared/number-input';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { CodeSample } from '@/components/ui/code-block';
import { FacetFilter } from '@/components/ui/facet-filter';
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
  const [size, setSize] = useState<number | null>(dataset.size);
  const [seed, setSeed] = useState<number | null>(dataset.seed);
  const valid =
    size !== null &&
    Number.isInteger(size) &&
    size >= 1 &&
    size <= maxValues &&
    seed !== null &&
    Number.isInteger(seed);
  // What the generator made, once the data in use is generated: to copy out.
  const generated = kinds.includes(dataset.kind) && dataset.values ? dataset.values : null;

  return (
    <div className="flex w-full flex-col gap-4">
      <form
        className="grid gap-3 sm:grid-cols-[2fr_1fr_1fr_auto] sm:items-end"
        onSubmit={(e) => {
          e.preventDefault();
          if (valid) dataset.generate(kind, size, seed);
        }}
      >
        <div className="flex flex-col gap-2">
          <span
            id={`${id}-kind`}
            className="text-sm font-medium"
          >
            Shape
          </span>
          {/* One shape of the data, picked as a facet is. */}
          <FacetFilter
            single
            title="Shape"
            size="default"
            searchable={kinds.length > 8}
            className="w-full justify-start"
            options={kinds.map((k) => ({ value: k, label: k.replace('-', ' ') }))}
            value={[kind]}
            onValueChange={(next) => setKind(next[0] ?? kind)}
          />
        </div>
        <Field
          label="Size"
          htmlFor={`${id}-size`}
        >
          <NumberInput
            id={`${id}-size`}
            min={1}
            max={maxValues}
            value={size}
            onValueChange={setSize}
          />
        </Field>
        <Field
          label="Seed"
          htmlFor={`${id}-seed`}
        >
          <NumberInput
            id={`${id}-seed`}
            value={seed}
            onValueChange={setSeed}
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
          <Alert
            variant="destructive"
            className="sm:col-span-4"
          >
            <CircleAlert aria-hidden />
            <AlertDescription>
              Size must be a whole number from 1 to {maxValues}, and the seed a whole number.
            </AlertDescription>
          </Alert>
        ) : null}
      </form>
      {generated ? (
        // The values themselves, with the code block's copy button: the same
        // data can be pasted into a test, a spreadsheet or the Paste tab.
        <CodeSample
          title={`${dataset.label} (comma separated)`}
          language="text"
          code={generated.join(', ')}
          wrap
          size="sm"
        />
      ) : null}
    </div>
  );
}
