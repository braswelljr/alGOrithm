'use client';

import { useId, useState } from 'react';
import { LuDownload, LuLink } from 'react-icons/lu';
import { Field } from '@/components/shared/field';
import {
  InputGroup,
  InputGroupAddon,
  InputGroupButton,
  InputGroupInput
} from '@/components/ui/input-group';
import type { AlgorithmApi as Api } from '@/visualizer/core';
import { formatOf, MAX_BYTES, type Format } from './formats';
import { ImportedText } from './imported-text';
import { Problems } from './problems';

/** A file fetched by this browser, straight from the server named. */
export function UrlTab({ api, onUse }: { api: Api; onUse: (v: number[], l: string) => void }) {
  const id = useId();
  const [url, setUrl] = useState('');
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [got, setGot] = useState<{ url: string; text: string; format: Format } | null>(null);

  async function load() {
    setError(null);
    setGot(null);
    let parsed: URL;
    try {
      parsed = new URL(url);
    } catch {
      setError('That is not a full address. It should start with https://.');
      return;
    }
    if (parsed.protocol !== 'https:' && parsed.protocol !== 'http:') {
      setError('Only http and https addresses can be loaded.');
      return;
    }
    setBusy(true);
    const ctl = new AbortController();
    const timer = setTimeout(() => ctl.abort(), 15_000);
    try {
      // Fetched by this browser, straight from that server, without cookies.
      const res = await fetch(parsed, { signal: ctl.signal, credentials: 'omit', mode: 'cors' });
      if (!res.ok) throw new Error(`The server answered ${res.status} ${res.statusText}.`);
      const length = Number(res.headers.get('content-length') ?? 0);
      if (length > MAX_BYTES) throw new Error('The file is larger than 10 MB.');
      const text = await res.text();
      if (text.length > MAX_BYTES) throw new Error('The file is larger than 10 MB.');
      setGot({
        url: parsed.href,
        text,
        format: formatOf(parsed.pathname, res.headers.get('content-type') ?? '')
      });
    } catch (err) {
      if (err instanceof DOMException && err.name === 'AbortError') {
        setError('The server took more than 15 seconds to answer.');
      } else if (err instanceof TypeError) {
        setError(
          'The file could not be read from here. Either the server does not allow other sites to read it (CORS), or it could not be reached. Download it and use the File tab instead.'
        );
      } else {
        setError(err instanceof Error ? err.message : String(err));
      }
    } finally {
      clearTimeout(timer);
      setBusy(false);
    }
  }

  return (
    <div className="flex flex-col gap-3">
      <form
        onSubmit={(e) => {
          e.preventDefault();
          void load();
        }}
      >
        <Field
          label="Address of a CSV, JSON or text file"
          htmlFor={id}
          hint="Your browser fetches it directly; this site never sees it. The server must allow cross-origin reads."
        >
          {/* The address and its Load button are one control: the button
              sits at the end of the field and spins while the file comes. */}
          <InputGroup>
            <InputGroupAddon>
              <LuLink
                aria-hidden
                className="text-muted size-4"
              />
            </InputGroupAddon>
            <InputGroupInput
              id={id}
              type="url"
              required
              value={url}
              onChange={(e) => setUrl(e.target.value)}
              placeholder="https://example.com/data.csv"
            />
            <InputGroupAddon align="inline-end">
              <InputGroupButton
                type="submit"
                variant="primary"
                size="sm"
                loading={busy}
                disabled={url.trim().length < 1}
              >
                <LuDownload aria-hidden />
                Load
              </InputGroupButton>
            </InputGroupAddon>
          </InputGroup>
        </Field>
      </form>
      {error ? (
        <Problems
          error={error}
          problems={[]}
        />
      ) : null}
      {got ? (
        <ImportedText
          key={got.url}
          api={api}
          text={got.text}
          format={got.format}
          source={new URL(got.url).hostname}
          onUse={onUse}
        />
      ) : null}
    </div>
  );
}
