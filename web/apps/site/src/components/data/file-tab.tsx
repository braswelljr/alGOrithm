'use client';

import { useId, useState } from 'react';
import type { AlgorithmApi as Api } from '@braswelljr/algorithm-visualizer/core';
import { Field, Input } from '@/components/ui/field';
import { formatOf, MAX_BYTES, type Format } from './formats';
import { ImportedText } from './imported-text';
import { Problems } from './problems';

/** A file read in the browser; nothing is uploaded. */
export function FileTab({ api, onUse }: { api: Api; onUse: (v: number[], l: string) => void }) {
  const id = useId();
  const [file, setFile] = useState<{ name: string; text: string; format: Format } | null>(null);
  const [error, setError] = useState<string | null>(null);
  return (
    <div className="flex flex-col gap-3">
      <Field
        label="File"
        htmlFor={id}
        hint="CSV, JSON or plain text. The file is read here, in your browser, and is not uploaded."
      >
        <Input
          id={id}
          type="file"
          accept=".csv,.json,.txt,.tsv,text/plain,text/csv,application/json"
          className="py-1.5"
          onChange={async (e) => {
            const f = e.target.files?.[0];
            setError(null);
            setFile(null);
            if (!f) return;
            if (f.size > MAX_BYTES) {
              setError(`The file is ${(f.size / 1048576).toFixed(1)} MB; the limit is 10 MB.`);
              return;
            }
            setFile({ name: f.name, text: await f.text(), format: formatOf(f.name, f.type) });
          }}
        />
      </Field>
      {error ? (
        <Problems
          error={error}
          problems={[]}
        />
      ) : null}
      {file ? (
        <ImportedText
          key={file.name + file.text.length}
          api={api}
          text={file.text}
          format={file.format}
          source={file.name}
          onUse={onUse}
        />
      ) : null}
    </div>
  );
}
