'use client';

import { useId, useMemo, useState } from 'react';
import { CircleAlert, Download, History, Upload } from 'lucide-react';
import { OpenInEditor } from '@/components/editor/open-in-editor';
import { Button } from '@/components/shared/button';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { buttonVariants } from '@/components/ui/button';
import { CodeSample } from '@/components/ui/code-block';
import { useRunHistory } from '@/hooks/use-run-history';
import { download } from '@/lib/share';
import type { Player } from '@/visualizer';
import {
  fallbackSnippet,
  makeSavedRun,
  readSavedRun,
  type AlgorithmApi,
  type Entry,
  type Input,
  type RunResult,
  type SavedRun
} from '@/visualizer/core';
import { CopyButton } from './copy-button';

/**
 * What can be done with a run besides watching it: replay it, share it as a
 * link, an iframe, a command or Go code, save it to a file, load a saved one,
 * or go back to an earlier run from this browser's history.
 */
export function RunToolbar({
  api,
  entry,
  input,
  result,
  player,
  onOpen
}: {
  api: AlgorithmApi;
  entry: Entry;
  input: Input;
  result: RunResult | null;
  player: Player;
  onOpen: (run: SavedRun) => void;
}) {
  const id = useId();
  const history = useRunHistory();
  const [importError, setImportError] = useState<string | null>(null);
  const [showEmbed, setShowEmbed] = useState(false);
  const snippet = useMemo(() => {
    const s = api.snippet(entry.id, input);
    return s.ok ? s.value : fallbackSnippet(entry, input);
  }, [api, entry, input]);
  const embedUrl = () => {
    const u = new URL(location.href);
    u.searchParams.set('embed', '1');
    return u.href;
  };
  const iframe = () =>
    `<iframe src="${embedUrl()}" title="${entry.name}, step by step" width="100%" height="520" style="border:0"></iframe>`;
  const names = new Map(api.catalog().entries.map((e) => [e.id, e.name]));

  return (
    <div className="flex flex-col gap-3">
      <div
        className="flex flex-wrap items-center gap-2"
        role="toolbar"
        aria-label="Share and save this run"
      >
        <Button
          size="sm"
          variant="primary"
          disabled={!player.can({ type: 'REPLAY' })}
          onClick={() => player.send({ type: 'REPLAY' })}
        >
          Replay
        </Button>
        <CopyButton text={() => location.href}>Copy link</CopyButton>
        <CopyButton text={() => snippet.cli}>Copy command</CopyButton>
        <CopyButton text={() => snippet.go}>Copy Go</CopyButton>
        <OpenInEditor
          id={entry.id}
          program={snippet.go}
          file={entry.file || undefined}
        />
        <Button
          size="sm"
          disabled={!result}
          onClick={() =>
            result &&
            download(
              `${entry.id}-run.json`,
              JSON.stringify(makeSavedRun(entry.id, input, result), null, 2)
            )
          }
        >
          <Download
            aria-hidden
            className="size-4"
          />
          Export run
        </Button>
        <label
          htmlFor={`${id}-import`}
          className={buttonVariants({
            variant: 'secondary',
            size: 'sm',
            className: 'focus-within:ring-accent cursor-pointer focus-within:ring-2'
          })}
        >
          <Upload
            aria-hidden
            className="size-4"
          />
          Import run
          {/* ui-ok: a file picker, hidden and driven by its styled label */}
          <input
            id={`${id}-import`}
            type="file"
            accept=".json,application/json"
            className="sr-only"
            onChange={async (e) => {
              const f = e.target.files?.[0];
              e.target.value = '';
              setImportError(null);
              if (!f) return;
              if (f.size > 50 * 1024 * 1024) {
                setImportError('The file is larger than 50 MB.');
                return;
              }
              const r = readSavedRun(await f.text());
              if (r.ok) onOpen(r.run);
              else setImportError(r.error);
            }}
          />
        </label>
        <Button
          size="sm"
          aria-expanded={showEmbed}
          aria-controls={`${id}-embed`}
          onClick={() => setShowEmbed((v) => !v)}
        >
          Embed
        </Button>
      </div>
      {importError ? (
        <Alert variant="destructive">
          <CircleAlert aria-hidden />
          <AlertDescription>{importError}</AlertDescription>
        </Alert>
      ) : null}
      {showEmbed ? (
        <div
          id={`${id}-embed`}
          className="border-border bg-subtle flex flex-col gap-2 rounded-md border p-3"
        >
          <p className="text-sm">
            Paste this into any page to show this run, without the site around it:
          </p>
          <CodeSample
            title="Embed"
            language="html"
            code={iframe()}
            wrap
            size="sm"
          />
        </div>
      ) : null}
      {/* The same run from a terminal, with the code block's copy button. */}
      <CodeSample
        title="Shell"
        language="bash"
        code={snippet.cli}
        size="sm"
      />
      {history.items.length > 0 ? (
        <details className="border-border rounded-md border p-3">
          <summary className="flex cursor-pointer items-center gap-2 text-sm font-medium">
            <History
              aria-hidden
              className="size-4"
            />
            Recent runs in this browser ({history.items.length})
          </summary>
          <ol className="mt-2 flex flex-col gap-1">
            {history.items.map((h, i) => (
              <li key={`${h.savedAt}-${i}`}>
                <Button
                  variant="link"
                  className="h-auto p-0 text-left whitespace-normal"
                  onClick={() => onOpen({ ...h })}
                >
                  {names.get(h.entry) ?? h.entry}: {describeInput(h.input)}
                </Button>{' '}
                <span className="text-muted text-xs">{new Date(h.savedAt).toLocaleString()}</span>
              </li>
            ))}
          </ol>
          <Button
            size="sm"
            variant="ghost"
            className="mt-2"
            onClick={history.clear}
          >
            Clear the history
          </Button>
        </details>
      ) : null}
    </div>
  );
}

function describeInput(i: Input): string {
  const parts: string[] = [];
  if (i.values?.length) parts.push(`${i.values.length} values`);
  if (i.target !== undefined) parts.push(`target ${i.target}`);
  if (i.text) parts.push(`"${i.text.length > 24 ? `${i.text.slice(0, 24)}...` : i.text}"`);
  if (i.other) parts.push(`"${i.other}"`);
  if (i.ops?.length) parts.push(`${i.ops.length} operations`);
  return parts.join(', ') || 'no input';
}
