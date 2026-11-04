'use client';

import { useId } from 'react';
import type { Entry } from '@braswelljr/algorithm-visualizer/core';
import { Field, Input } from '@/components/ui/field';

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
  if (entry.options.length === 0) return null;
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
          <Input
            id={`${id}-${o.name}`}
            type="number"
            min={o.min}
            max={o.max}
            value={values[o.name]}
            onChange={(e) => {
              const n = Number(e.target.value);
              if (Number.isSafeInteger(n)) onChange(o.name, n);
            }}
          />
        </Field>
      ))}
    </>
  );
}
