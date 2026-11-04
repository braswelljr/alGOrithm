'use client';

import { useId, useMemo, useState } from 'react';
import { Button } from '@/components/shared/button';
import { Field, Textarea } from '@/components/shared/field';
import type { AlgorithmApi as Api } from '@/visualizer/core';
import { Problems } from './problems';

/** Values typed or pasted. */
export function PasteTab({ api, onUse }: { api: Api; onUse: (v: number[], l: string) => void }) {
  const id = useId();
  const [text, setText] = useState('');
  const parsed = useMemo(() => (text.trim() ? api.parse('text', text) : null), [api, text]);
  return (
    <div className="flex flex-col gap-3">
      <Field
        label="Values"
        htmlFor={id}
        hint="Whole numbers separated by commas, spaces or new lines. A JSON array of numbers works too."
      >
        <Textarea
          id={id}
          value={text}
          onChange={(e) => setText(e.target.value)}
          placeholder="42, 7, 19, 3, 25"
        />
      </Field>
      {parsed?.ok ? (
        <div>
          <Button
            variant="primary"
            onClick={() => onUse(parsed.value, `${parsed.value.length} pasted values`)}
          >
            Use {parsed.value.length} values
          </Button>
        </div>
      ) : parsed ? (
        <Problems
          error={parsed.failure.error}
          problems={parsed.failure.problems}
        />
      ) : null}
    </div>
  );
}
