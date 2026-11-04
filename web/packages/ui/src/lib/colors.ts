/**
 * The alGOrithm color system, as data: the neutrals, the accents, the status
 * colors and the visualization palettes, each in OKLCH for light and dark.
 *
 * This file is the one source. `scripts/colors.ts` writes it out as
 * `src/styles/colors.css` (CSS variables selected by `data-neutral`,
 * `data-accent` and `data-viz` on <html>, plus the `.dark` class), the theme
 * picker lists its names, the /colors page draws it, and `test/colors.test.ts`
 * checks every pair a reader depends on meets WCAG AA: 4.5:1 for text, 3:1 for
 * controls and marks. A color never carries meaning alone: every role is also
 * a shape, a pattern or a label (docs/rules/the-markup-says-what-the-thing-is.md).
 */

export type Mode = 'light' | 'dark';

/** An OKLCH color: lightness 0..1, chroma 0..0.4, hue in degrees. */
export type Oklch = { l: number; c: number; h: number };

export const oklch = (l: number, c: number, h: number): Oklch => ({ l, c, h });

export function css({ l, c, h }: Oklch): string {
  const r = (n: number, d: number) => Number(n.toFixed(d));
  return `oklch(${r(l, 3)} ${r(c, 3)} ${r(h, 1)})`;
}

// Neutrals ------------------------------------------------------------------

/** A neutral is a hue and how much of it the grays carry. */
export type Neutral = { id: string; name: string; h: number; c: number };

export const neutrals: readonly Neutral[] = [
  { id: 'slate', name: 'Slate', h: 257, c: 0.02 },
  { id: 'zinc', name: 'Zinc', h: 286, c: 0.006 },
  { id: 'stone', name: 'Stone', h: 56, c: 0.01 }
];

export const defaultNeutral = 'slate';

/** The surfaces and inks, as lightness and a share of the neutral's chroma. */
const surfaceSteps: Record<Mode, Record<SurfaceToken, [number, number]>> = {
  light: {
    bg: [0.985, 0.3],
    card: [1, 0],
    subtle: [0.955, 0.5],
    border: [0.9, 0.8],
    control: [0.6, 1],
    muted: [0.5, 1],
    fg: [0.22, 1.2]
  },
  dark: {
    bg: [0.17, 0.8],
    card: [0.21, 0.8],
    subtle: [0.25, 0.8],
    border: [0.32, 1],
    control: [0.56, 1],
    muted: [0.72, 1],
    fg: [0.95, 0.5]
  }
};

export type SurfaceToken = 'bg' | 'card' | 'subtle' | 'border' | 'control' | 'muted' | 'fg';

export function surfaces(n: Neutral, mode: Mode): Record<SurfaceToken, Oklch> {
  const out = {} as Record<SurfaceToken, Oklch>;
  for (const [token, [l, share]] of Object.entries(surfaceSteps[mode]) as [
    SurfaceToken,
    [number, number]
  ][]) {
    out[token] = oklch(l, n.c * share, n.h);
  }
  return out;
}

// Accents -------------------------------------------------------------------

/**
 * An accent: links, the primary action, focus rings, the selected state. In
 * light mode it is dark enough for white text on it and for text on the page;
 * in dark mode it is light, with the page's ink on it.
 */
export type Accent = {
  id: string;
  name: string;
  light: Oklch;
  dark: Oklch;
};

export const accents: readonly Accent[] = [
  { id: 'indigo', name: 'Indigo', light: oklch(0.5, 0.2, 265), dark: oklch(0.74, 0.14, 265) },
  { id: 'blue', name: 'Blue', light: oklch(0.5, 0.18, 250), dark: oklch(0.75, 0.13, 245) },
  { id: 'violet', name: 'Violet', light: oklch(0.5, 0.21, 295), dark: oklch(0.75, 0.15, 295) },
  { id: 'emerald', name: 'Emerald', light: oklch(0.48, 0.12, 162), dark: oklch(0.76, 0.14, 162) },
  { id: 'amber', name: 'Amber', light: oklch(0.48, 0.12, 60), dark: oklch(0.8, 0.14, 75) },
  { id: 'rose', name: 'Rose', light: oklch(0.5, 0.19, 15), dark: oklch(0.76, 0.13, 12) }
];

export const defaultAccent = 'indigo';

/** The keyboard colorway and ink derived from an accent. */
export function accentTokens(a: Accent, mode: Mode, n: Neutral) {
  const base = a[mode];
  const ink = surfaces(n, mode);
  return {
    accent: base,
    // On the accent: white in light mode, the page's dark ink in dark mode.
    'accent-fg': mode === 'light' ? oklch(0.99, 0, 0) : ink.bg,
    'lit-ink':
      mode === 'light'
        ? oklch(base.l - 0.02, base.c, base.h)
        : oklch(base.l + 0.04, base.c, base.h),
    'cap-accent-top': oklch(Math.min(base.l + 0.1, 0.95), base.c * 0.95, base.h),
    'cap-accent-face': base,
    'cap-accent-skirt': oklch(Math.max(base.l - 0.13, 0.25), base.c * 0.85, base.h)
  };
}

// Status --------------------------------------------------------------------

export type StatusToken = 'danger' | 'success' | 'warning';

export const status: Record<Mode, Record<StatusToken, Oklch>> = {
  light: {
    danger: oklch(0.52, 0.2, 25),
    success: oklch(0.5, 0.13, 150),
    warning: oklch(0.5, 0.12, 65)
  },
  dark: {
    danger: oklch(0.74, 0.16, 25),
    success: oklch(0.76, 0.14, 150),
    warning: oklch(0.82, 0.14, 80)
  }
};

// Visualization -------------------------------------------------------------

/** The roles the views draw. Each is also a shape or a label in the views. */
export const vizRoles = [
  'compare',
  'swap',
  'pivot',
  'found',
  'write',
  'sorted',
  'path',
  'frontier',
  'visited',
  'wall'
] as const;
export type VizRole = (typeof vizRoles)[number];

/** The roles drawn as marks (bars, outlines, lines), which need 3:1 on a card.
 *  `frontier` and `visited` are fills behind a label, and `wall` is a block. */
export const vizMarks: readonly VizRole[] = [
  'compare',
  'swap',
  'pivot',
  'found',
  'write',
  'sorted',
  'path'
];

export type VizPalette = {
  id: string;
  name: string;
  description: string;
  colors: Record<Mode, Record<VizRole, Oklch>>;
};

export const vizPalettes: readonly VizPalette[] = [
  {
    id: 'default',
    name: 'Default',
    description: 'Amber, rose, violet, sky and green, tuned to the site.',
    colors: {
      light: {
        compare: oklch(0.6, 0.15, 65),
        swap: oklch(0.56, 0.21, 15),
        write: oklch(0.55, 0.2, 300),
        sorted: oklch(0.57, 0.14, 150),
        pivot: oklch(0.56, 0.15, 235),
        found: oklch(0.57, 0.14, 150),
        path: oklch(0.56, 0.15, 150),
        frontier: oklch(0.8, 0.13, 75),
        visited: oklch(0.85, 0.05, 230),
        wall: oklch(0.35, 0.02, 260)
      },
      dark: {
        compare: oklch(0.8, 0.15, 75),
        swap: oklch(0.72, 0.18, 15),
        write: oklch(0.72, 0.16, 300),
        sorted: oklch(0.72, 0.14, 150),
        pivot: oklch(0.72, 0.13, 230),
        found: oklch(0.72, 0.14, 150),
        path: oklch(0.72, 0.15, 150),
        frontier: oklch(0.55, 0.12, 75),
        visited: oklch(0.42, 0.06, 230),
        wall: oklch(0.08, 0.01, 260)
      }
    }
  },
  {
    id: 'colorblind',
    name: 'Colorblind-safe',
    description:
      'The Okabe-Ito hues, distinguishable with every common color vision deficiency, darkened in light mode to keep 3:1.',
    colors: {
      light: {
        compare: oklch(0.53, 0.14, 245), // blue
        swap: oklch(0.58, 0.18, 45), // vermillion
        pivot: oklch(0.55, 0.13, 350), // reddish purple
        found: oklch(0.55, 0.12, 165), // bluish green
        write: oklch(0.6, 0.14, 75), // orange
        sorted: oklch(0.55, 0.12, 165),
        path: oklch(0.6, 0.14, 75),
        frontier: oklch(0.88, 0.15, 105), // yellow
        visited: oklch(0.85, 0.07, 235), // sky blue
        wall: oklch(0.25, 0, 0)
      },
      dark: {
        compare: oklch(0.73, 0.12, 235),
        swap: oklch(0.68, 0.18, 45),
        pivot: oklch(0.72, 0.12, 350),
        found: oklch(0.68, 0.13, 165),
        write: oklch(0.78, 0.16, 75),
        sorted: oklch(0.68, 0.13, 165),
        path: oklch(0.78, 0.16, 75),
        frontier: oklch(0.5, 0.1, 105),
        visited: oklch(0.42, 0.07, 235),
        wall: oklch(0.05, 0, 0)
      }
    }
  },
  {
    id: 'contrast',
    name: 'High contrast',
    description: 'Fewer, stronger colors at 4.5:1 or more against the card.',
    colors: {
      light: {
        compare: oklch(0.42, 0.18, 255),
        swap: oklch(0.45, 0.2, 28),
        pivot: oklch(0.4, 0.17, 320),
        found: oklch(0.42, 0.12, 150),
        write: oklch(0.45, 0.12, 60),
        sorted: oklch(0.42, 0.12, 150),
        path: oklch(0.42, 0.18, 255),
        frontier: oklch(0.85, 0.14, 95),
        visited: oklch(0.88, 0.04, 255),
        wall: oklch(0.1, 0, 0)
      },
      dark: {
        compare: oklch(0.82, 0.12, 245),
        swap: oklch(0.78, 0.16, 30),
        pivot: oklch(0.82, 0.12, 330),
        found: oklch(0.85, 0.15, 150),
        write: oklch(0.88, 0.14, 85),
        sorted: oklch(0.85, 0.15, 150),
        path: oklch(0.82, 0.12, 245),
        frontier: oklch(0.45, 0.1, 95),
        visited: oklch(0.38, 0.05, 245),
        wall: oklch(0, 0, 0)
      }
    }
  }
];

export const defaultViz = 'default';

// Contrast ------------------------------------------------------------------

/** OKLCH to linear sRGB, clamped into the gamut. */
export function toLinearRgb({ l, c, h }: Oklch): [number, number, number] {
  const a = c * Math.cos((h * Math.PI) / 180);
  const b = c * Math.sin((h * Math.PI) / 180);
  const l_ = l + 0.3963377774 * a + 0.2158037573 * b;
  const m_ = l - 0.1055613458 * a - 0.0638541728 * b;
  const s_ = l - 0.0894841775 * a - 1.291485548 * b;
  const L = l_ ** 3;
  const M = m_ ** 3;
  const S = s_ ** 3;
  const clamp = (x: number) => Math.min(1, Math.max(0, x));
  return [
    clamp(4.0767416621 * L - 3.3077115913 * M + 0.2309699292 * S),
    clamp(-1.2684380046 * L + 2.6097574011 * M - 0.3413193965 * S),
    clamp(-0.0041960863 * L - 0.7034186147 * M + 1.707614701 * S)
  ];
}

/** WCAG relative luminance. Linear sRGB is what the formula wants. */
export function luminance(color: Oklch): number {
  const [r, g, b] = toLinearRgb(color);
  return 0.2126 * r + 0.7152 * g + 0.0722 * b;
}

/** The WCAG contrast ratio of two colors, 1 to 21. */
export function contrast(a: Oklch, b: Oklch): number {
  const [hi, lo] = [luminance(a), luminance(b)].sort((x, y) => y - x);
  return (hi + 0.05) / (lo + 0.05);
}

/** sRGB hex, for showing a color's value to a reader. */
export function hex(color: Oklch): string {
  const enc = (x: number) => (x <= 0.0031308 ? 12.92 * x : 1.055 * x ** (1 / 2.4) - 0.055);
  return `#${toLinearRgb(color)
    .map((x) =>
      Math.round(enc(x) * 255)
        .toString(16)
        .padStart(2, '0')
    )
    .join('')}`;
}

/** A pair a reader depends on, and the ratio it must reach. */
export type Pair = { fg: string; bg: string; min: number; what: string };

/** Every pair the site relies on, for one neutral, accent and mode. */
export function pairs(n: Neutral, a: Accent, mode: Mode): (Pair & { ratio: number })[] {
  const s = surfaces(n, mode);
  const acc = accentTokens(a, mode, n);
  const st = status[mode];
  const all: Record<string, Oklch> = { ...s, ...acc, ...st };
  const check: Pair[] = [];
  for (const bg of ['bg', 'card', 'subtle']) {
    check.push({ fg: 'fg', bg, min: 4.5, what: 'body text' });
    check.push({ fg: 'muted', bg, min: 4.5, what: 'secondary text' });
    check.push({ fg: 'control', bg, min: 3, what: 'a control border' });
  }
  for (const bg of ['bg', 'card']) {
    check.push({ fg: 'accent', bg, min: 4.5, what: 'a link' });
    for (const t of ['danger', 'success', 'warning']) {
      check.push({ fg: t, bg, min: 4.5, what: `${t} text` });
    }
  }
  check.push({ fg: 'accent-fg', bg: 'accent', min: 4.5, what: 'text on the accent' });
  check.push({ fg: 'accent', bg: 'subtle', min: 3, what: 'a focus ring or selected state' });
  return check.map((p) => ({ ...p, ratio: contrast(all[p.fg], all[p.bg]) }));
}

/** Every visualization mark against the card it is drawn on. */
export function vizPairs(p: VizPalette, n: Neutral, mode: Mode) {
  const card = surfaces(n, mode).card;
  const min = p.id === 'contrast' ? 4.5 : 3;
  return vizMarks.map((role) => ({
    role,
    min,
    ratio: contrast(p.colors[mode][role], card)
  }));
}
