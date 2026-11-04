'use client';

import { useId, useState } from 'react';
import { Field } from '@/components/shared/field';
import { FileDropzone, formatBytes } from '@/components/shared/file-dropzone';
import type { AlgorithmApi as Api } from '@/visualizer/core';
import { formatOf, MAX_BYTES, type Format } from './formats';
import { ImportedText } from './imported-text';

const KIND: Record<Format, string> = { csv: 'CSV', json: 'JSON', text: 'Text' };

/** A file read in the browser; nothing is uploaded. */
export function FileTab({ api, onUse }: { api: Api; onUse: (v: number[], l: string) => void }) {
  const id = useId();
  const [chosen, setChosen] = useState<{ name: string; size: number; format: Format } | null>(null);
  const [text, setText] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  async function choose(f: File) {
    const format = formatOf(f.name, f.type);
    setChosen({ name: f.name, size: f.size, format });
    setText(null);
    setError(null);
    if (f.size > MAX_BYTES) {
      setError(`${formatBytes(f.size)}: the limit is 10 MB`);
      return;
    }
    setText(await f.text());
  }

  return (
    <div className="flex w-full flex-col gap-3">
      <Field
        label="File"
        htmlFor={id}
        hint="The file is read here, in your browser, and is not uploaded."
      >
        <FileDropzone
          id={id}
          accept=".csv,.json,.txt,.tsv,text/plain,text/csv,application/json"
          file={chosen ? { ...chosen, kind: KIND[chosen.format] } : null}
          error={error}
          hint="CSV, JSON or plain text, up to 10 MB"
          onChoose={(f) => void choose(f)}
          onClear={() => {
            setChosen(null);
            setText(null);
            setError(null);
          }}
        />
      </Field>
      {chosen && text !== null ? (
        <ImportedText
          key={chosen.name + text.length}
          api={api}
          text={text}
          format={chosen.format}
          source={chosen.name}
          onUse={onUse}
        />
      ) : null}
    </div>
  );
}
