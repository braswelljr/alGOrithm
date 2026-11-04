# AGENTS.md (web/)

Guidance for agents working on the visualizer. The repository rules in [`../docs/rules/`](../docs/rules/README.md) apply, the visualizer section above all; this file says where each one lives in the code.

## How it fits together

One Next.js app (static export), no workspace and no npm package: the Go module
is the library, and this app is the site.

```text
+--------------------------+   JSON strings   +-----------------------------------+
| cmd/algorithm-wasm (Go)  | <--------------> | src/visualizer/core               |
| core.Entries(), Execute, |  catalog, run,   |  api.ts      zod-checked wrapper  |
| Compare, MakeSnippet     |  source, ...     |  loader.ts   loadAlgorithm()      |
+--------------------------+                  |  machine.ts  the player FSM       |
                                              |  replay/     events -> state at i |
                                              +----------------+------------------+
                                                               |
                                              +----------------v------------------+
                                              | src/visualizer/components         |
                                              |  views draw a state, never events |
                                              +----------------+------------------+
                                                               |
+--------------------------+                  +----------------v------------------+
| src/components/ui        | ---------------> | src/app, src/components/<feature> |
|  the primitives (Base UI)|  everything is   |  pages, inputs, URL state (nuqs)  |
| src/components/algorithm |  built from them |                                   |
+--------------------------+                  +-----------------------------------+
```

```text
web/
  src/app/                  routes (static export)
  src/components/ui/        the primitives: Base UI, shadcn base-vega, synced from gsl-certificates
                            and gnux (SYNC.md); everything else is built from these
  src/components/algorithm/ the app's own compositions (EntryCard, TraceView, ...), from ui/
  src/components/<feature>/ one folder per feature (learn, quiz, docs, run, ...)
  src/visualizer/           core/ (no React: api, loader, player machine, replay, schemas),
                            components/ (one view per file), hooks/, styles.css, theme.css
  src/lib/ src/hooks/       shared helpers; lib/colors.ts and lib/control.ts are the color and
                            control-size systems
  src/styles/               theme.css (tokens), colors.css (generated), main.css (entry)
  test/                     node --test via tsx: test/ui, test/visualizer and the site's tests
  scripts/                  icons, colors, budget
  public/                   algorithm.wasm and wasm_exec.js (from `make web/wasm`), brand assets
```

## The visualizer (`src/visualizer`)

- `core/` has no React: it runs in Node too (the tests import it directly, and `src/lib/catalog.server.ts` uses it at build time).
- `components/` are kebab-case, one view per file. `run-frames` has one builder per kind of run (`runKind` picks it from the output); `run-view` owns the player.
- Styles are Tailwind utilities on the app's theme; `theme.css` adds the roles' colors (compare, swap, sorted, the logic levels, the float fields) and keyframes, and `styles.css` the `--av-*` defaults the theme overrides on `.av-root`.

## The site (`src/app`, `src/components`)

- Pages under `src/app/` are thin: a heading and a client app in `Suspense`. Static export only. Tool pages are query-string driven; entries (`/docs/[id]/`) and learning paths (`/learn/[id]/`) are static pages built from `src/lib/catalog.server.ts`, which runs the same `public/algorithm.wasm` in Node at build time, so they are readable and indexable before any JavaScript runs.
- Nothing in the root layout or `SiteChrome` may read the query (`useSearchParams`): it would make Next.js render every page in the browser only and leave the exported HTML empty. Embed mode is set by an inline script on `<html data-embed>` and hidden with Tailwind's `in-data-[embed=1]:` variants.
- SEO: every page sets its metadata with `pageMetadata()` from `src/lib/metadata.ts` (one canonical, Open Graph and Twitter cards); the domain is `siteUrl` in `src/lib/site.ts` (`https://algorithm-beige.vercel.app`, overridable with `NEXT_PUBLIC_SITE_URL`). `sitemap.ts`, `robots.ts` and `manifest.ts` are static metadata routes in `src/app/`. Social cards and the manifest screenshots are PNG route handlers (`og.png`, `docs/[id]/og.png`, `screenshots/*.png`) drawn by `components/shared/og-card.tsx` in Roboto Serif and Source Sans 3 (the `@fontsource` WOFF files, which Satori can read); not the `opengraph-image` convention, whose export has no file extension. Structured data comes from `src/lib/structured-data.ts` through `components/shared/json-ld.tsx`, with braswelljr (https://braswelljr.vercel.app) as author.
- Brand, icon only (no wordmark): every logo and icon file lives in `public/`, never in `src/app/`. `public/logos/` holds the mark SVGs (`mark.svg` full color with its own dark mode, `mark-small.svg` hinted for 24 px and below, `mark-mono.svg` for CSS masks, `concept-*.svg` for /brand, `safari-pinned-tab.svg`); `public/icons/` every raster (favicons 16 to 96 light and dark, `icon.svg`, Apple 120/152/167/180 and launch screens in `icons/splash/`, install icons 48 to 512, maskable, monochrome, Windows tiles, shortcut icons); `public/favicon.ico` (16, 32, 48) and `public/browserconfig.xml` at the root.
- Where each comes from: the concepts are drawn as data in `components/brand/marks.ts` (`ACTIVE_CONCEPT` picks the one in use, one line); `pnpm --dir web icons` (`scripts/icons.mjs`) writes `public/logos/` from it, then draws every raster from those SVG files read back from disk (squircle plates from `squirclePath`, each size rendered directly and 4x then box-filtered, the sharper kept, an 8-bit RGBA PNG, recompressed by oxipng when installed). The outputs are committed; rerun the script after changing the mark or the list.
- Every path is named once in `src/lib/brand.ts` (`LOGO`, `ICONS` and the lists): the root layout's `metadata.icons`, `appleWebApp.startupImage` and msapplication tags, `manifest.ts`, the `Logo` component (`Logo` masks `/logos/mark-mono.svg` in `currentColor`; `LogoColor` draws the full-color mark inline from the tokens so it follows the site's theme), the /brand page and `test/icons.test.ts` all import from it. `test/icons.test.ts` checks every file's size, transparency and opacity, the ICO's sizes, the manifest, and the home page's link tags in `out/`.
- `src/components/` by feature: `layout/` (header, chrome with embed mode, wasm gate), `brand/` (logo), `shared/` (footer, JSON-LD, social card), `catalog/`, `data/` (one file per tab), `visualize/` (`input-form` builds every form from the entry's input kind, params, options and example; `array-input` and `grid-input` add the data panel and the grid editor), `structures/`, `compare/`, `docs/`, `run/` (`RunPanel`, its toolbar, code panel). `ui/` holds the atoms.
- `src/hooks/use-entry-input.ts` keeps an entry's input (`in`) and options (`opt`) in the URL as validated JSON, starting from `entry.example`.
- `src/hooks/use-dataset.ts` keeps data in the URL (at most 200 values; more stay in this tab); `use-run-history.ts` keeps the last 20 runs in localStorage; `use-dev-reload.ts` polls `/dev-build.json` in development.
- Tailwind for everything; `src/styles/main.css` maps the theme's tokens onto the visualizer's `--av-*` variables.

## Conventions

- The catalog is the only list of entries. Never hard-code an entry's id, name, cost or category; `test/rules.test.ts` fails on an id in site code.
- Never cast wasm data, URL parameters or files. Add a schema in `core/schema.ts` and parse.
- Views take a state. A highlight carries a shape, outline, pattern or label as well as a color, and every step is described in words in a polite live region.
- Nothing autoplays; links open paused at their step. Above 50 steps a second the views drop per-step highlights. `prefers-reduced-motion` turns transitions off.
- US spelling, no em dashes, diagrams in plain text (checked by `test/rules.test.ts`).

## Checks

```bash
pnpm lint          # ESLint and Prettier
pnpm typecheck     # next typegen, then tsc
pnpm test          # ui, visualizer (every entry through the real wasm) and the house rules
pnpm build         # the static site into out/
```
