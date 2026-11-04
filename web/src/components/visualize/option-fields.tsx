'use client';

import { useId } from 'react';
import { Field } from '@/components/shared/field';
import { NumberInput } from '@/components/shared/number-input';
import type { Entry } from '@/visualizer/core';

/** A number field per option the entry declares (quantum, frames, head, ...). */
export function OptionFields({
  entry,
  values,
  onChange
}: {
  entry: Entry;
  values: Record<string, number>;
  onChange: (name: string, v: number) => void;
}) {
  const id = useId();
  if (entry.options.length < 1) return null;
  return (
    <>
      {entry.options.map((o) => (
        <Field
          key={o.name}
          label={o.label || o.name}
          htmlFor={`${id}-${o.name}`}
          hint={
            o.min !== undefined || o.max !== undefined
              ? `${o.min ?? ''} to ${o.max ?? ''}`
              : undefined
          }
        >
          <NumberInput
            id={`${id}-${o.name}`}
            min={o.min}
            max={o.max}
            value={values[o.name] ?? null}
            onValueChange={(n) => {
              if (n !== null && Number.isSafeInteger(n)) onChange(o.name, n);
            }}
          />
        </Field>
      ))}
    </>
  );
}
