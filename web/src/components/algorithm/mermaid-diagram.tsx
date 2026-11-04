'use client';

import { useEffect, useId, useRef, useState } from 'react';
import { useTheme } from 'next-themes';
import { CodeSample } from '@/components/ui/code-block';
import { Skeleton } from '@/components/ui/skeleton';
import { cn } from '@/lib/utils';

/**
 * A Mermaid diagram, drawn in the browser. Mermaid is imported only when a
 * diagram mounts, so pages without one never load it. It renders in strict
 * security mode (no scripts, no raw HTML in labels) with its colors read from
 * the theme's tokens, so a diagram follows light and dark and the brand.
 *
 * The site only: the repository's markdown stays plain text
 * (docs/rules/diagrams-are-plain-text.md), which every reader can see.
 *
 *   <MermaidDiagram chart={'flowchart LR\n  a --> b'} label="What depends on what" />
 */
export function MermaidDiagram({
  chart,
  label,
  className
}: {
  /** The diagram, in Mermaid's syntax. */
  chart: string;
  /** What the diagram shows, for a screen reader (the drawing is hidden). */
  label: string;
  className?: string;
}) {
  const id = useId().replace(/:/g, '');
  const host = useRef<HTMLDivElement>(null);
  const { resolvedTheme } = useTheme();
  const [state, setState] = useState<'loading' | 'ready' | 'error'>('loading');

  useEffect(() => {
    let live = true;
    void (async () => {
      try {
        const { default: mermaid } = await import('mermaid');
        // Mermaid's color parser reads hex and rgb, not the theme's oklch(), so
        // each token is painted to one canvas pixel and read back as sRGB hex.
        const paint = document
          .createElement('canvas')
          .getContext('2d', { willReadFrequently: true });
        const css = getComputedStyle(document.documentElement);
        const token = (name: string, fallback: string) => {
          const value = css.getPropertyValue(name).trim();
          if (!value || !paint) return fallback;
          paint.clearRect(0, 0, 1, 1);
          paint.fillStyle = fallback;
          paint.fillStyle = value;
          paint.fillRect(0, 0, 1, 1);
          const [r, g, b] = paint.getImageData(0, 0, 1, 1).data;
          return `#${[r, g, b].map((n) => n.toString(16).padStart(2, '0')).join('')}`;
        };
        // The font Mermaid measures labels with: the resolved family, since
        // a CSS variable means nothing to its text measurement.
        const fontFamily = getComputedStyle(document.body).fontFamily || 'sans-serif';
        mermaid.initialize({
          startOnLoad: false,
          securityLevel: 'strict',
          theme: 'base',
          fontFamily,
          // Natural size: the chart scrolls sideways rather than shrinking
          // its labels to fit a narrow column.
          flowchart: {
            useMaxWidth: false,
            htmlLabels: false,
            padding: 12,
            nodeSpacing: 36,
            rankSpacing: 48
          },
          themeVariables: {
            darkMode: resolvedTheme === 'dark',
            fontFamily,
            fontSize: '14px',
            background: token('--bg', '#ffffff'),
            mainBkg: token('--subtle', '#f5f5f4'),
            primaryColor: token('--subtle', '#f5f5f4'),
            primaryTextColor: token('--fg', '#1c1917'),
            nodeTextColor: token('--fg', '#1c1917'),
            primaryBorderColor: token('--accent', '#066b86'),
            nodeBorder: token('--accent', '#066b86'),
            lineColor: token('--muted', '#78716c'),
            secondaryColor: token('--subtle', '#f5f5f4'),
            tertiaryColor: token('--bg', '#ffffff')
          }
        });
        const { svg } = await mermaid.render(`mermaid-${id}`, chart);
        if (!live || !host.current) return;
        host.current.innerHTML = svg;
        setState('ready');
      } catch {
        if (live) setState('error');
      }
    })();
    return () => {
      live = false;
    };
  }, [chart, id, resolvedTheme]);

  return (
    <figure
      className={cn('w-full', className)}
      aria-label={label}
    >
      {state === 'loading' ? <Skeleton className="h-48 w-full rounded-lg" /> : null}
      {state === 'error' ? (
        // Drawn as its source when Mermaid cannot draw it.
        <CodeSample
          title="Diagram source"
          language="text"
          code={chart}
          size="sm"
        />
      ) : null}
      <div
        ref={host}
        aria-hidden
        className={cn(
          'w-full overflow-x-auto [&_svg]:mx-auto [&_svg]:h-auto',
          state !== 'ready' && 'hidden'
        )}
      />
      <figcaption className="sr-only">{label}</figcaption>
    </figure>
  );
}
