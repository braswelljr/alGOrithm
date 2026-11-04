/**
 * The one source of every alGOrithm mark: three original concepts drawn as
 * data on a 32 by 32 grid. The React logo, the social cards and the icon
 * generator (scripts/icons.mjs) all draw from here, so no copy can drift.
 * Plain TypeScript with no JSX and no imports, so Node runs it as is.
 *
 * Each shape has a tone. In one color every tone is the same color; in full
 * color, primary is the accent, secondary the highlight (the step being
 * looked at), muted a quieter accent.
 */

export type Tone = 'primary' | 'secondary' | 'muted';

export type Shape =
  | { kind: 'path'; d: string; tone: Tone; stroke?: number; fill?: boolean }
  | { kind: 'circle'; cx: number; cy: number; r: number; tone: Tone; stroke?: number }
  | {
      kind: 'rect';
      x: number;
      y: number;
      width: number;
      height: number;
      rx: number;
      tone: Tone;
      stroke?: number;
    };

export type Concept = {
  id: ConceptId;
  name: string;
  /** Why it looks like this, for the /brand page. */
  idea: string;
  /** On the 32 grid, strokes on whole or half units, so 16 and 32 px land on pixels. */
  shapes: Shape[];
  /**
   * A hinted variant for 24 px and below, where the full mark would muddy:
   * fewer parts, thicker strokes (at least 2 units, 1 px at 16 px).
   */
  small?: Shape[];
};

export type ConceptId = 'graph' | 'brackets' | 'cursor';

/**
 * (a) A graph that draws an "A": three nodes at the apex and the feet, joined
 * by edges, and a crossbar edge between two more. The apex is the node being
 * visited. It reads as a letter, a tree and a network at once.
 */
const graph: Concept = {
  id: 'graph',
  name: 'Graph A',
  idea: 'A graph whose nodes and edges draw the letter A: the apex is the node being visited, the crossbar an edge being relaxed. A letter, a tree and a network at once.',
  shapes: [
    { kind: 'path', d: 'M16 6 6 26M16 6 26 26', tone: 'muted', stroke: 3 },
    { kind: 'path', d: 'M10 18H22', tone: 'primary', stroke: 3 },
    { kind: 'circle', cx: 6, cy: 26, r: 3.5, tone: 'primary' },
    { kind: 'circle', cx: 26, cy: 26, r: 3.5, tone: 'primary' },
    { kind: 'circle', cx: 10, cy: 18, r: 2.5, tone: 'primary' },
    { kind: 'circle', cx: 22, cy: 18, r: 2.5, tone: 'primary' },
    { kind: 'circle', cx: 16, cy: 6, r: 4, tone: 'secondary' }
  ],
  // At 16 px: no crossbar nodes, a thicker A, nodes on whole pixels.
  small: [
    { kind: 'path', d: 'M16 6 6 26M16 6 26 26', tone: 'muted', stroke: 4 },
    { kind: 'path', d: 'M10 18H22', tone: 'primary', stroke: 4 },
    { kind: 'circle', cx: 6, cy: 26, r: 4, tone: 'primary' },
    { kind: 'circle', cx: 26, cy: 26, r: 4, tone: 'primary' },
    { kind: 'circle', cx: 16, cy: 6, r: 4, tone: 'secondary' }
  ]
};

/**
 * (b) "GO" between angle brackets, ‹GO›: code, and the Go in the name, without
 * borrowing anything from Go's own marks.
 */
const brackets: Concept = {
  id: 'brackets',
  name: '‹GO›',
  idea: 'GO between angle brackets, like a tag or a generic type: the code behind every view, and the Go the library is written in.',
  shapes: [
    { kind: 'path', d: 'M7 9.5 3 16l4 6.5M25 9.5l4 6.5-4 6.5', tone: 'muted', stroke: 2.5 },
    { kind: 'path', d: 'M14.5 13A4 4 0 1 0 15.5 16H12.5', tone: 'primary', stroke: 2.5 },
    { kind: 'circle', cx: 20.5, cy: 16, r: 4, tone: 'secondary', stroke: 2.5 }
  ]
};

/**
 * (c) A step-through cursor: three cells, the first done, the second the one
 * being looked at, and a pointer under it. What the player does at every step.
 */
const cursor: Concept = {
  id: 'cursor',
  name: 'Step cursor',
  idea: 'Three cells and a pointer under the middle one: the first is done, the second is the step being looked at. What the player does at every step.',
  shapes: [
    { kind: 'rect', x: 2, y: 6, width: 8, height: 11, rx: 2, tone: 'muted' },
    { kind: 'rect', x: 12, y: 4, width: 8, height: 13, rx: 2, tone: 'secondary' },
    { kind: 'rect', x: 23, y: 7, width: 7, height: 9, rx: 1.5, tone: 'primary', stroke: 2 },
    { kind: 'path', d: 'M16 20l6 8H10Z', tone: 'secondary', fill: true }
  ]
};

export const CONCEPTS: Record<ConceptId, Concept> = { graph, brackets, cursor };

/**
 * The mark the site uses everywhere: the header, the footer, the favicons,
 * the app icons, the social cards and the README. Change it here, rerun
 * `pnpm --dir web icons`, and every icon follows.
 */
export const ACTIVE_CONCEPT: ConceptId = 'graph';

export type Palette = Record<Tone, string>;

/** The site's accent tokens written out, for files that cannot read CSS variables. */
export const PALETTE = {
  light: { primary: '#4b4fd8', secondary: '#1f9d63', muted: '#8b8ee8' },
  dark: { primary: '#8f93ff', secondary: '#4fd08f', muted: '#5c60c4' },
  /** White on the accent: app icons and tiles. */
  onAccent: { primary: '#ffffff', secondary: '#c9f5dc', muted: '#c7c9ff' }
} satisfies Record<string, Palette>;

export const COLORS = {
  accent: '#4b4fd8',
  lightBg: '#f7f8fb',
  darkBg: '#12141c',
  ink: '#22252f',
  paper: '#eceef5'
} as const;

/** One color for every tone. */
export function mono(color: string): Palette {
  return { primary: color, secondary: color, muted: color };
}

function shapeSvg(s: Shape, color: string | undefined): string {
  // color undefined: classes instead of colors ("primary f" filled, "primary s"
  // stroked), painted by a <style> from toneCss.
  const stroked = s.kind !== 'path' ? Boolean(s.stroke) : Boolean(s.stroke) && !s.fill;
  const width = s.stroke ?? 0;
  const paint = stroked
    ? `fill="none" stroke-width="${width}" stroke-linecap="round" stroke-linejoin="round" ${
        color ? `stroke="${color}"` : `class="${s.tone} s"`
      }`
    : color
      ? `fill="${color}"`
      : `class="${s.tone} f"`;
  switch (s.kind) {
    case 'path':
      return `<path d="${s.d}" ${paint}/>`;
    case 'circle':
      return `<circle cx="${s.cx}" cy="${s.cy}" r="${s.r}" ${paint}/>`;
    case 'rect':
      return `<rect x="${s.x}" y="${s.y}" width="${s.width}" height="${s.height}" rx="${s.rx}" ${paint}/>`;
  }
}

/** CSS painting the classes markShapes writes when it is given no colors. */
export function toneCss(p: Palette): string {
  return (['primary', 'secondary', 'muted'] as const)
    .map((t) => `.${t}.f{fill:${p[t]}}.${t}.s{stroke:${p[t]}}`)
    .join('');
}

/** The shapes of a concept as SVG elements, without the document around them. */
export function markShapes(id: ConceptId, palette?: Palette, small = false): string {
  const c = CONCEPTS[id];
  const shapes = small && c.small ? c.small : c.shapes;
  return shapes.map((s) => shapeSvg(s, palette?.[s.tone])).join('');
}

export type MarkOptions = {
  /** Pixel size of the square document. */
  size?: number;
  /** A solid square behind the mark. */
  background?: string;
  /** Corner radius of the background, in grid units (32 is the side). */
  radius?: number;
  /** How much of the side the mark takes, 0 to 1; the rest is padding. */
  scale?: number;
  /** A <style> block inside the document, for prefers-color-scheme. */
  style?: string;
  /** Draw the hinted small variant (24 px and below). */
  small?: boolean;
};

/** A concept as a standalone square SVG document. */
export function markSvg(id: ConceptId, palette: Palette | undefined, o: MarkOptions = {}): string {
  const scale = o.scale ?? 1;
  const offset = (32 - 32 * scale) / 2;
  const size = o.size ? ` width="${o.size}" height="${o.size}"` : '';
  const bg = o.background
    ? `<rect width="32" height="32" rx="${o.radius ?? 0}" fill="${o.background}"/>`
    : '';
  const style = o.style ? `<style>${o.style}</style>` : '';
  const body = markShapes(id, palette, o.small);
  const g =
    scale === 1
      ? body
      : `<g transform="translate(${offset} ${offset}) scale(${scale})">${body}</g>`;
  return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32" shape-rendering="geometricPrecision"${size}>${style}${bg}${g}</svg>`;
}

/**
 * The scalable favicon: the full-color mark, lighter in a dark browser tab,
 * through prefers-color-scheme inside the file. small draws the hinted
 * variant, for the tab favicon, which is shown at 16 px.
 */
export function faviconSvg(id: ConceptId = ACTIVE_CONCEPT, small = false): string {
  const style = `${toneCss(PALETTE.light)}@media (prefers-color-scheme:dark){${toneCss(PALETTE.dark)}}`;
  return markSvg(id, undefined, { style, small });
}

/**
 * A superellipse ("squircle", |x/a|^n + |y/b|^n = 1, n = 5) filling the box,
 * as a closed path of cubic curves: continuous curvature at the corners, like
 * an iOS icon, where a rounded rectangle has a visible kink. The curve is
 * sampled and joined with Catmull-Rom tangents, so the path is smooth curves
 * rather than a polyline.
 */
export function squirclePath(
  x: number,
  y: number,
  w: number,
  h: number,
  n = 5,
  samples = 48
): string {
  const a = w / 2;
  const b = h / 2;
  const cx = x + a;
  const cy = y + b;
  const pts: [number, number][] = [];
  for (let i = 0; i < samples; i++) {
    const t = (i / samples) * 2 * Math.PI;
    const c = Math.cos(t);
    const s = Math.sin(t);
    pts.push([
      cx + a * Math.sign(c) * Math.abs(c) ** (2 / n),
      cy + b * Math.sign(s) * Math.abs(s) ** (2 / n)
    ]);
  }
  const r = (v: number) => Math.round(v * 100) / 100;
  let d = `M${r(pts[0][0])} ${r(pts[0][1])}`;
  for (let i = 0; i < samples; i++) {
    const p0 = pts[(i - 1 + samples) % samples];
    const p1 = pts[i];
    const p2 = pts[(i + 1) % samples];
    const p3 = pts[(i + 2) % samples];
    const c1 = [p1[0] + (p2[0] - p0[0]) / 6, p1[1] + (p2[1] - p0[1]) / 6];
    const c2 = [p2[0] - (p3[0] - p1[0]) / 6, p2[1] - (p3[1] - p1[1]) / 6];
    d += `C${r(c1[0])} ${r(c1[1])} ${r(c2[0])} ${r(c2[1])} ${r(p2[0])} ${r(p2[1])}`;
  }
  return `${d}Z`;
}
