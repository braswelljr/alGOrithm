import { useMemo, useState } from 'react';
import { highlightGo } from '../core/highlight';
import { CopyIcon } from './icons';

/**
 * Go source with line numbers, light highlighting and a copy button. The code
 * is the real source of the entry, as the wasm build returns it.
 */
export function CodeView({
  code,
  file,
  highlight = [],
  label
}: {
  code: string;
  file?: string;
  /** 1-based line numbers to mark. */
  highlight?: number[];
  label?: string;
}) {
  const lines = useMemo(() => highlightGo(code), [code]);
  const marked = new Set(highlight);
  const [copied, setCopied] = useState<'idle' | 'done' | 'failed'>('idle');
  const copy = async () => {
    try {
      await navigator.clipboard.writeText(code);
      setCopied('done');
    } catch {
      setCopied('failed');
    }
    setTimeout(() => setCopied('idle'), 2000);
  };
  return (
    <figure
      className="av-code"
      style={{ margin: 0 }}
    >
      <figcaption className="av-code-head">
        <span>{label ?? file ?? 'Go source'}</span>
        <button
          type="button"
          className="av-btn"
          onClick={copy}
        >
          <CopyIcon />
          {copied === 'done' ? 'Copied' : copied === 'failed' ? 'Copy failed' : 'Copy'}
        </button>
      </figcaption>
      <pre tabIndex={0}>
        <code>
          {lines.map((line, n) => (
            <span
              key={n}
              className="av-code-line"
              data-highlight={marked.has(n + 1) || undefined}
            >
              <span
                className="av-code-num"
                aria-hidden
              >
                {n + 1}
              </span>
              {line.map((t, k) =>
                t.kind === 'plain' ? (
                  t.text
                ) : (
                  <span
                    key={k}
                    className={`av-tok-${t.kind}`}
                  >
                    {t.text}
                  </span>
                )
              )}
              {'\n'}
            </span>
          ))}
        </code>
      </pre>
      <span
        className="av-sr-only"
        aria-live="polite"
      >
        {copied === 'done' ? 'Code copied' : copied === 'failed' ? 'Could not copy the code' : ''}
      </span>
    </figure>
  );
}
