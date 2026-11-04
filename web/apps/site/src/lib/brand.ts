/**
 * Every brand file the site ships, and the one place their paths are written:
 * the root layout's metadata, manifest.ts, the Logo component, the /brand
 * page, scripts/icons.mjs and the tests all import from here.
 *
 * Everything lives in public/: the mark SVGs in public/logos/, every icon in
 * public/icons/, and favicon.ico and browserconfig.xml at the root, where
 * browsers ask for them. `pnpm --dir web/apps/site icons` writes them all.
 * Plain TypeScript with no imports, so Node runs it as is.
 */

/** How a raster icon is drawn. */
export type Look =
  /** The full-color mark on nothing, for browser tabs. */
  | 'tab-light'
  | 'tab-dark'
  /** White on a rounded accent tile: install icons, shortcuts. */
  | 'tile'
  /** White on a full accent square, no transparency: Apple home screens. */
  | 'apple'
  /** White on full-bleed accent, the mark inside the 80% safe zone. */
  | 'maskable'
  /** White on nothing: themed (monochrome) icons and Windows tiles. */
  | 'mono';

export type RasterIcon = {
  /** Path from the site directory (apps/site). */
  file: string;
  /** Path the browser asks for. */
  href: string;
  width: number;
  height: number;
  look: Look;
};

const pub = (name: string) => ({ file: `public/${name}`, href: `/${name}` });
const square = (n: number) => ({ width: n, height: n });

/** The ICO favicon at /favicon.ico, with these sizes inside. */
export const favicon = { ...pub('favicon.ico'), sizes: [16, 32, 48] };

/** The scalable favicon, light and dark inside one file (the hinted mark). */
export const faviconSvg = pub('icons/icon.svg');

/** The PNG favicons (48 for Windows, 96 for search results). */
export const tabIcons: RasterIcon[] = [16, 32, 48, 96].map((n) => ({
  ...pub(`icons/favicon-${n}x${n}.png`),
  ...square(n),
  look: 'tab-light' as const
}));

/** The same for dark browser chrome, linked with a prefers-color-scheme media query. */
export const darkTabIcons: RasterIcon[] = [16, 32, 48, 96].map((n) => ({
  ...pub(`icons/favicon-dark-${n}x${n}.png`),
  ...square(n),
  look: 'tab-dark' as const
}));

/** Apple touch icons. */
export const appleIcons: RasterIcon[] = [120, 152, 167, 180].map((n) => ({
  ...pub(`icons/apple-touch-icon-${n}x${n}.png`),
  ...square(n),
  look: 'apple' as const
}));

/** The installable app's icons (purpose "any"). */
export const appIcons: RasterIcon[] = [48, 72, 96, 144, 192, 256, 384, 512].map((n) => ({
  ...pub(`icons/icon-${n}.png`),
  ...square(n),
  look: 'tile' as const
}));

export const maskableIcons: RasterIcon[] = [192, 512].map((n) => ({
  ...pub(`icons/icon-maskable-${n}.png`),
  ...square(n),
  look: 'maskable' as const
}));

export const monochromeIcons: RasterIcon[] = [
  { ...pub('icons/icon-monochrome-512.png'), ...square(512), look: 'mono' }
];

/** Windows Start tiles, named in browserconfig.xml. */
export const tiles: RasterIcon[] = [
  { ...pub('icons/mstile-70x70.png'), ...square(70), look: 'tile' },
  { ...pub('icons/mstile-150x150.png'), ...square(150), look: 'tile' },
  { ...pub('icons/mstile-310x150.png'), width: 310, height: 150, look: 'tile' },
  { ...pub('icons/mstile-310x310.png'), ...square(310), look: 'tile' }
];

export const browserconfig = pub('browserconfig.xml');

/** Safari's pinned tab: one color, drawn by Safari in maskIconColor. */
export const safariPinnedTab = pub('logos/safari-pinned-tab.svg');

/**
 * The mark as SVG files, icon only. mark.svg is full color and follows the
 * visitor's light or dark mode by itself; mark-mono.svg is one flat color, for
 * CSS masks painted in currentColor; mark-small.svg is the hinted variant for
 * 24 px and below; concepts are the three candidates shown on /brand.
 */
export const LOGO = {
  mark: pub('logos/mark.svg'),
  markSmall: pub('logos/mark-small.svg'),
  mono: pub('logos/mark-mono.svg'),
  concepts: {
    graph: pub('logos/concept-graph.svg'),
    brackets: pub('logos/concept-brackets.svg'),
    cursor: pub('logos/concept-cursor.svg')
  },
  /** The README's copy of mark.svg, from the site directory. */
  readme: '../../../docs/assets/logo.svg'
};

/** The app's shortcuts, each a tile with a glyph (the glyphs are in scripts/icons.mjs). */
export const shortcuts = [
  { id: 'learn', name: 'Learn', description: 'Learning paths, step by step', url: '/learn/' },
  { id: 'quiz', name: 'Quiz', description: 'Predict the next step', url: '/quiz/' },
  {
    id: 'visualize',
    name: 'Visualize',
    description: 'Run an algorithm on your data',
    url: '/visualize/'
  },
  { id: 'big-o', name: 'Big O', description: 'How costs grow', url: '/big-o/' }
].map((s) => ({ ...s, icon: { ...pub(`icons/shortcut-${s.id}-96.png`), ...square(96) } }));

/** iPhone and iPad screens in portrait: CSS size and pixel ratio. */
export const devices = [
  { name: 'iPhone SE', width: 375, height: 667, ratio: 2 },
  { name: 'iPhone 8 Plus', width: 414, height: 736, ratio: 3 },
  { name: 'iPhone X, XS, 11 Pro, 12 mini, 13 mini', width: 375, height: 812, ratio: 3 },
  { name: 'iPhone XR, 11', width: 414, height: 896, ratio: 2 },
  { name: 'iPhone XS Max, 11 Pro Max', width: 414, height: 896, ratio: 3 },
  { name: 'iPhone 12, 13, 14', width: 390, height: 844, ratio: 3 },
  { name: 'iPhone 12, 13 Pro Max, 14 Plus', width: 428, height: 926, ratio: 3 },
  { name: 'iPhone 14 Pro, 15, 16', width: 393, height: 852, ratio: 3 },
  { name: 'iPhone 15 Pro Max, 16 Plus', width: 430, height: 932, ratio: 3 },
  { name: 'iPhone 16 Pro', width: 402, height: 874, ratio: 3 },
  { name: 'iPhone 16 Pro Max', width: 440, height: 956, ratio: 3 },
  { name: 'iPad mini', width: 768, height: 1024, ratio: 2 },
  { name: 'iPad Air', width: 820, height: 1180, ratio: 2 },
  { name: 'iPad Pro 11"', width: 834, height: 1194, ratio: 2 },
  { name: 'iPad Pro 12.9"', width: 1024, height: 1366, ratio: 2 }
] as const;

export type StartupImage = RasterIcon & { media: string; scheme: 'light' | 'dark' };

/** Apple launch screens: every device, light and dark. */
export const startupImages: StartupImage[] = devices.flatMap((d) =>
  (['light', 'dark'] as const).map((scheme) => {
    const w = d.width * d.ratio;
    const h = d.height * d.ratio;
    return {
      ...pub(`icons/splash/launch-${w}x${h}-${scheme}.png`),
      width: w,
      height: h,
      look: scheme === 'light' ? ('tab-light' as const) : ('tab-dark' as const),
      scheme,
      media: `(device-width: ${d.width}px) and (device-height: ${d.height}px) and (-webkit-device-pixel-ratio: ${d.ratio}) and (orientation: portrait) and (prefers-color-scheme: ${scheme})`
    };
  })
);

/** Every raster file the generator writes (the ICO apart). */
export const allRasters: RasterIcon[] = [
  ...tabIcons,
  ...darkTabIcons,
  ...appleIcons,
  ...appIcons,
  ...maskableIcons,
  ...monochromeIcons,
  ...tiles,
  ...shortcuts.map((s) => ({ ...s.icon, look: 'tile' as const })),
  ...startupImages
];

/**
 * The 16 px rendering of every concept, light and dark, for /brand's 4x zoom,
 * so edge quality can be judged on the real pixels.
 */
export const brandPreviews = (['graph', 'brackets', 'cursor'] as const).flatMap((concept) =>
  (['light', 'dark'] as const).map((scheme) => ({
    ...pub(`logos/preview-${concept}-16-${scheme}.png`),
    width: 16,
    height: 16,
    look: scheme === 'light' ? ('tab-light' as const) : ('tab-dark' as const),
    concept,
    scheme
  }))
);

/** The tile color for Windows and the mask color for Safari. */
export const tileColor = '#4b4fd8';

const byWidth = (list: RasterIcon[]) =>
  Object.fromEntries(list.map((i) => [i.width, i.href])) as Record<number, string>;

/** Single paths by name, for places that need one file rather than a list. */
export const ICONS = {
  favicon: favicon.href,
  svg: faviconSvg.href,
  tab: byWidth(tabIcons),
  tabDark: byWidth(darkTabIcons),
  apple: byWidth(appleIcons),
  apple180: byWidth(appleIcons)[180],
  app: byWidth(appIcons),
  maskable: byWidth(maskableIcons),
  monochrome512: monochromeIcons[0].href,
  safariPinnedTab: safariPinnedTab.href,
  browserconfig: browserconfig.href
};
