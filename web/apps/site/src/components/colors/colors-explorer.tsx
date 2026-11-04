'use client';

import * as React from 'react';
import { CheckIcon, XIcon } from 'lucide-react';
import { useThemePrefs } from '@braswelljr/algorithm-ui/algorithm/theme-picker';
import { Badge } from '@braswelljr/algorithm-ui/components/badge';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@braswelljr/algorithm-ui/components/card';
import { Frame, FramePanel } from '@braswelljr/algorithm-ui/components/frame';
import { Field, FieldLabel } from '@braswelljr/algorithm-ui/components/field';
import { Heading } from '@braswelljr/algorithm-ui/components/heading';
import { NativeSelect } from '@braswelljr/algorithm-ui/components/native-select';
import { ToggleGroup, ToggleGroupItem } from '@braswelljr/algorithm-ui/components/toggle-group';
import {
  accentTokens,
  accents,
  contrast,
  css,
  hex,
  neutrals,
  pairs,
  status,
  surfaces,
  vizMarks,
  vizPalettes,
  type Mode,
  type Oklch,
  type VizRole
} from '@braswelljr/algorithm-ui/lib/colors';

const roleShape: Record<VizRole, { label: string; shape: string }> = {
  compare: { label: 'Compare', shape: '◇' },
  swap: { label: 'Swap', shape: '⇄' },
  pivot: { label: 'Pivot', shape: '▲' },
  found: { label: 'Found', shape: '●' },
  write: { label: 'Write', shape: '✎' },
  sorted: { label: 'Sorted', shape: '✓' },
  path: { label: 'Path', shape: '→' },
  frontier: { label: 'Frontier', shape: '◌' },
  visited: { label: 'Visited', shape: '·' },
  wall: { label: 'Wall', shape: '■' }
};

function Swatch({ color, name }: { color: Oklch; name: string }) {
  return (
    <figure className="flex flex-col gap-1.5">
      <span
        className="smooth-corners h-14 rounded-lg ring-1 ring-black/10 ring-inset"
        style={{ background: css(color) }}
      />
      <figcaption className="text-xs">
        <span className="block font-medium text-fg">{name}</span>
        <span className="font-mono text-muted">{hex(color)}</span>
      </figcaption>
    </figure>
  );
}

/** A theme drawn with its own colors, whatever the page's theme is. */
function ThemePreview({ mode, neutral, accent }: { mode: Mode; neutral: string; accent: string }) {
  const n = neutrals.find((x) => x.id === neutral) ?? neutrals[0];
  const a = accents.find((x) => x.id === accent) ?? accents[0];
  const s = surfaces(n, mode);
  const t = accentTokens(a, mode, n);
  return (
    <FramePanel
      className="mt-0! flex flex-col gap-3 p-4"
      style={{ background: css(s.bg), color: css(s.fg) }}
    >
      <p className="text-xs font-semibold tracking-wide uppercase" style={{ color: css(s.muted) }}>
        {mode === 'light' ? 'Light' : 'Dark'} · {n.name} · {a.name}
      </p>
      <div
        className="smooth-corners flex flex-col gap-2 rounded-lg p-3"
        style={{ background: css(s.card), boxShadow: `0 0 0 1px ${css(s.border)}` }}
      >
        <p className="font-semibold">Quick Sort</p>
        <p className="text-sm" style={{ color: css(s.muted) }}>
          Partition around a pivot, then sort each side.
        </p>
        <span className="flex flex-wrap items-center gap-2">
          <span
            className="rounded-md px-3 py-1.5 text-sm font-medium"
            style={{ background: css(t.accent), color: css(t['accent-fg']) }}
          >
            Run it
          </span>
          <span className="text-sm underline" style={{ color: css(t.accent) }}>
            Read the code
          </span>
          <span className="text-sm" style={{ color: css(status[mode].success) }}>
            ✓ Sorted
          </span>
        </span>
      </div>
    </FramePanel>
  );
}

function RolesPreview({ palette, mode }: { palette: string; mode: Mode }) {
  const p = vizPalettes.find((x) => x.id === palette) ?? vizPalettes[0];
  const n = neutrals[0];
  const card = surfaces(n, mode).card;
  const heights = [40, 70, 55, 85, 30, 65, 50];
  return (
    <FramePanel
      className="mt-0! flex flex-col gap-3 p-4"
      style={{ background: css(card), color: css(surfaces(n, mode).fg) }}
    >
      <div className="flex h-28 items-end gap-2" role="img" aria-label={`Bars in each role, ${mode} mode`}>
        {vizMarks.map((role, i) => (
          <span key={role} className="flex flex-1 flex-col items-center gap-1">
            <span aria-hidden className="text-xs">{roleShape[role].shape}</span>
            <span
              className="w-full rounded-t-md"
              style={{ height: heights[i % heights.length], background: css(p.colors[mode][role]) }}
            />
          </span>
        ))}
      </div>
      <ul className="grid grid-cols-2 gap-x-4 gap-y-1 text-xs sm:grid-cols-4">
        {vizMarks.map((role) => {
          const ratio = contrast(p.colors[mode][role], card);
          return (
            <li key={role} className="flex items-center gap-1.5">
              <span aria-hidden className="size-3 rounded-sm" style={{ background: css(p.colors[mode][role]) }} />
              <span aria-hidden>{roleShape[role].shape}</span>
              {roleShape[role].label}
              <span className="font-mono opacity-70">{ratio.toFixed(1)}:1</span>
            </li>
          );
        })}
      </ul>
    </FramePanel>
  );
}

/** The color system, to look at, compare and check. */
export function ColorsExplorer() {
  const [prefs, update] = useThemePrefs();
  const [mode, setMode] = React.useState<Mode>('light');
  const n = neutrals.find((x) => x.id === prefs.neutral) ?? neutrals[0];
  const a = accents.find((x) => x.id === prefs.accent) ?? accents[0];
  const s = surfaces(n, mode);
  const tokens: Record<string, Oklch> = { ...s, ...accentTokens(a, mode, n), ...status[mode] };
  const [fg, setFg] = React.useState('fg');
  const [bg, setBg] = React.useState('bg');
  const ratio = contrast(tokens[fg], tokens[bg]);
  const all = pairs(n, a, mode);

  return (
    <div className="flex flex-col gap-10">
      <section aria-labelledby="themes" className="flex flex-col gap-4">
        <Heading level={2} id="themes">
          The themes side by side
        </Heading>
        <p className="max-w-3xl text-muted">
          Pick an accent and a neutral: the site switches as you do, and the previews show both
          modes at once.
        </p>
        <div className="flex flex-wrap items-end gap-4">
          <Field>
            <FieldLabel htmlFor="accent">Accent</FieldLabel>
            <NativeSelect id="accent" value={prefs.accent} onChange={(e) => update({ accent: e.target.value })}>
              {accents.map((x) => (
                <option key={x.id} value={x.id}>{x.name}</option>
              ))}
            </NativeSelect>
          </Field>
          <Field>
            <FieldLabel htmlFor="neutral">Neutral</FieldLabel>
            <NativeSelect id="neutral" value={prefs.neutral} onChange={(e) => update({ neutral: e.target.value })}>
              {neutrals.map((x) => (
                <option key={x.id} value={x.id}>{x.name}</option>
              ))}
            </NativeSelect>
          </Field>
        </div>
        <Frame className="grid gap-1 md:grid-cols-2">
          <ThemePreview mode="light" neutral={prefs.neutral} accent={prefs.accent} />
          <ThemePreview mode="dark" neutral={prefs.neutral} accent={prefs.accent} />
        </Frame>
      </section>

      <section aria-labelledby="palette" className="flex flex-col gap-4">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <Heading level={2} id="palette">
            The tokens
          </Heading>
          <ToggleGroup
            variant="chip"
            aria-label="Mode"
            value={[mode]}
            onValueChange={(v) => {
              if (v[0] === 'light' || v[0] === 'dark') setMode(v[0]);
            }}
          >
            <ToggleGroupItem value="light">Light</ToggleGroupItem>
            <ToggleGroupItem value="dark">Dark</ToggleGroupItem>
          </ToggleGroup>
        </div>
        <div className="grid grid-cols-3 gap-3 sm:grid-cols-6 lg:grid-cols-8">
          {Object.entries(tokens).map(([name, color]) => (
            <Swatch key={name} name={name} color={color} />
          ))}
        </div>
      </section>

      <section aria-labelledby="checker" className="flex flex-col gap-4">
        <Heading level={2} id="checker">
          Contrast checker
        </Heading>
        <Card>
          <CardContent className="flex flex-col gap-4">
            <div className="flex flex-wrap items-end gap-4">
              <Field>
                <FieldLabel htmlFor="fg">Text</FieldLabel>
                <NativeSelect id="fg" value={fg} onChange={(e) => setFg(e.target.value)}>
                  {Object.keys(tokens).map((k) => (
                    <option key={k} value={k}>{k}</option>
                  ))}
                </NativeSelect>
              </Field>
              <Field>
                <FieldLabel htmlFor="bg">Background</FieldLabel>
                <NativeSelect id="bg" value={bg} onChange={(e) => setBg(e.target.value)}>
                  {Object.keys(tokens).map((k) => (
                    <option key={k} value={k}>{k}</option>
                  ))}
                </NativeSelect>
              </Field>
              <output className="text-3xl font-semibold" aria-live="polite">
                {ratio.toFixed(2)}:1
              </output>
              <span className="flex gap-2">
                {[
                  { name: 'AA text', min: 4.5 },
                  { name: 'AA large and UI', min: 3 },
                  { name: 'AAA text', min: 7 }
                ].map((t) => (
                  <Badge key={t.name} variant={ratio >= t.min ? 'beginner' : 'advanced'}>
                    {ratio >= t.min ? <CheckIcon aria-hidden /> : <XIcon aria-hidden />}
                    {t.name}
                    <span className="sr-only">{ratio >= t.min ? ' passes' : ' fails'}</span>
                  </Badge>
                ))}
              </span>
            </div>
            <p
              className="smooth-corners rounded-lg p-4 text-lg"
              style={{ background: css(tokens[bg]), color: css(tokens[fg]) }}
            >
              The quick brown fox sorts the lazy dogs in O(n log n).
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>Every pair the site depends on</CardTitle>
            <CardDescription>
              {n.name} with {a.name}, {mode}. A test (packages/ui/test/colors.test.ts) fails the
              build if any of these drops below its minimum, in any theme.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <table className="w-full text-sm">
              <thead>
                <tr className="text-left text-xs text-muted uppercase">
                  <th scope="col" className="py-1">Pair</th>
                  <th scope="col">Used for</th>
                  <th scope="col">Ratio</th>
                  <th scope="col">Minimum</th>
                </tr>
              </thead>
              <tbody>
                {all.map((p) => (
                  <tr key={`${p.fg}-${p.bg}`} className="border-t border-border">
                    <th scope="row" className="py-1.5 text-left font-mono font-normal">
                      {p.fg} on {p.bg}
                    </th>
                    <td className="text-muted">{p.what}</td>
                    <td className="font-mono">{p.ratio.toFixed(2)}</td>
                    <td className="font-mono">
                      {p.min}
                      {p.ratio >= p.min ? (
                        <CheckIcon aria-label="passes" className="ms-1 inline size-4 text-success" />
                      ) : (
                        <XIcon aria-label="fails" className="ms-1 inline size-4 text-danger" />
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </CardContent>
        </Card>
      </section>

      <section aria-labelledby="viz" className="flex flex-col gap-4">
        <Heading level={2} id="viz">
          The visualization palettes
        </Heading>
        <p className="max-w-3xl text-muted">
          Every role a view draws (compare, swap, pivot, found, write, sorted, path) has a color and
          also a shape and a label, so no one depends on telling hues apart. Choose the palette the
          visualizer uses:
        </p>
        <ToggleGroup
          variant="chip"
          aria-label="Visualization palette"
          value={[prefs.viz]}
          onValueChange={(v) => {
            if (v[0]) update({ viz: v[0] });
          }}
          className="flex-wrap"
        >
          {vizPalettes.map((p) => (
            <ToggleGroupItem key={p.id} value={p.id}>
              {p.name}
            </ToggleGroupItem>
          ))}
        </ToggleGroup>
        <p className="text-sm text-muted">{vizPalettes.find((p) => p.id === prefs.viz)?.description}</p>
        <Frame className="grid gap-1 md:grid-cols-2">
          <RolesPreview palette={prefs.viz} mode="light" />
          <RolesPreview palette={prefs.viz} mode="dark" />
        </Frame>
      </section>
    </div>
  );
}
