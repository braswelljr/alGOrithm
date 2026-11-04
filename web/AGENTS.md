# AGENTS.md (web/)

Guidance for agents working on the visualizer. The repository rules in [`../docs/rules/`](../docs/rules/README.md) apply, the visualizer section above all; this file says where each one lives in the code.

## How it fits together

```text
+---------------------+   JSON strings   +------------------------------------+
| cmd/algorithm-wasm (Go)     | <--------------> | packages/visualizer/src/core       |
| core.Entries()  |  catalog, run,   |  api.ts      zod-checked wrapper   |
| Execute, Compare    |  source, ...     |  loader.ts   loadAlgorithm()       |
+---------------------+                  |  machine.ts  the player FSM        |
                                         |  replay/     events -> state at i  |
                                         |  layout.ts   snapshot -> positions |
                                         +-----------------+------------------+
                                                           |
                                         +-----------------v------------------+
                                         | packages/visualizer/src/components |
                                         |  views draw a state, never events  |
                                         |  RunView: view + player            |
                                         |  StructureFlow: React Flow         |
                                         +-----------------+------------------+
                                                           |  workspace:*
                                         +-----------------v------------------+
                                         | apps/site (Next.js, static export) |
                                         |  pages, inputs, URL state (nuqs),  |
                                         |  RunPanel, history, docs           |
                                         +------------------------------------+
```

## The package (`packages/visualizer`)

- `src/core/` has no React: it runs in Node and is the `./core` export. The tests import the built `dist/core.js`.
- `src/components/` are kebab-case, one component per file. Views (`bars-view`, `cells-view`, `grid-view`, `table-view`, `text-match-view`, `hanoi-view`, `structure-flow`, `timeline-view`, `track-view`, `lines-view`, `summary-list`) take small typed props and draw them. `run-frames` has one builder per kind of run (`runKind` picks it from the output: timeline, track, grid, table, text, structure, hanoi, bars, cells, lines); `run-view` owns the player and adds the extras (summary, a stats table, an explanation's lines); `visualizer` runs an entry and renders `RunView`.
- Styles are plain CSS in `src/styles.css` (classes `av-*`, tokens `--av-*` with dark mirrors), bundled with React Flow's base CSS into `dist/styles.css` by `scripts/build-css.mjs`. No Tailwind in the package.
- `assets/` holds `algorithm.wasm` and `wasm_exec.js`, written by `make web/wasm`, never committed.
- Build: `tsup` (one ESM build, two entry points sharing a chunk, so `.` and `./core` share one loader).

## The site (`apps/site`)

- Pages under `src/app/` are thin: a heading and a client app in `Suspense`. Static export only. Tool pages are query-string driven; entries (`/docs/[id]/`) and learning paths (`/learn/[id]/`) are static pages built from `src/lib/catalog.server.ts`, which runs the same `public/algorithm.wasm` in Node at build time, so they are readable and indexable before any JavaScript runs.
- Nothing in the root layout or `SiteChrome` may read the query (`useSearchParams`): it would make Next.js render every page in the browser only and leave the exported HTML empty. Embed mode is set by an inline script on `<html data-embed>` and hidden with Tailwind's `in-data-[embed=1]:` variants.
- SEO: every page sets its metadata with `pageMetadata()` from `src/lib/metadata.ts` (one canonical, Open Graph and Twitter cards); the domain is `siteUrl` in `src/lib/site.ts` (`https://algorithm.vercel.app`, overridable with `NEXT_PUBLIC_SITE_URL`). `sitemap.ts`, `robots.ts` and `manifest.ts` are static metadata routes in `src/app/`. Social cards and the manifest screenshots are PNG route handlers (`og.png`, `docs/[id]/og.png`, `screenshots/*.png`) drawn by `components/shared/og-card.tsx` in Ysabeau Office (its OFL .ttf files in `src/assets/fonts/`); not the `opengraph-image` convention, whose export has no file extension. Structured data comes from `src/lib/structured-data.ts` through `components/shared/json-ld.tsx`, with braswelljr (https://braswelljr.vercel.app) as author.
- Brand, icon only (no wordmark): every logo and icon file lives in `public/`, never in `src/app/`. `public/logos/` holds the mark SVGs (`mark.svg` full color with its own dark mode, `mark-small.svg` hinted for 24 px and below, `mark-mono.svg` for CSS masks, `concept-*.svg` for /brand, `safari-pinned-tab.svg`); `public/icons/` every raster (favicons 16 to 96 light and dark, `icon.svg`, Apple 120/152/167/180 and launch screens in `icons/splash/`, install icons 48 to 512, maskable, monochrome, Windows tiles, shortcut icons); `public/favicon.ico` (16, 32, 48) and `public/browserconfig.xml` at the root.
- Where each comes from: the concepts are drawn as data in `components/brand/marks.ts` (`ACTIVE_CONCEPT` picks the one in use, one line); `pnpm --dir web/apps/site icons` (`scripts/icons.mjs`) writes `public/logos/` from it, then draws every raster from those SVG files read back from disk (squircle plates from `squirclePath`, each size rendered directly and 4x then box-filtered, the sharper kept, an 8-bit RGBA PNG, recompressed by oxipng when installed). The outputs are committed; rerun the script after changing the mark or the list.
- Every path is named once in `src/lib/brand.ts` (`LOGO`, `ICONS` and the lists): the root layout's `metadata.icons`, `appleWebApp.startupImage` and msapplication tags, `manifest.ts`, the `Logo` component (`Logo` masks `/logos/mark-mono.svg` in `currentColor`; `LogoColor` draws the full-color mark inline from the tokens so it follows the site's theme), the /brand page and `test/icons.test.ts` all import from it. `test/icons.test.ts` checks every file's size, transparency and opacity, the ICO's sizes, the manifest, and the home page's link tags in `out/`.
- `src/components/` by feature: `layout/` (header, chrome with embed mode, wasm gate), `brand/` (logo), `shared/` (footer, JSON-LD, social card), `catalog/`, `data/` (one file per tab), `visualize/` (`input-form` builds every form from the entry's input kind, params, options and example; `array-input` and `grid-input` add the data panel and the grid editor), `structures/`, `compare/`, `docs/`, `run/` (`RunPanel`, its toolbar, code panel). `ui/` holds the atoms.
- `src/hooks/use-entry-input.ts` keeps an entry's input (`in`) and options (`opt`) in the URL as validated JSON, starting from `entry.example`.
- `src/hooks/use-dataset.ts` keeps data in the URL (at most 200 values; more stay in this tab); `use-run-history.ts` keeps the last 20 runs in localStorage; `use-dev-reload.ts` polls `/dev-build.json` in development.
- Tailwind for the site's own pages; `src/styles/main.css` maps the site's tokens onto the package's `--av-*` variables.

## Conventions

- The catalog is the only list of entries. Never hard-code an entry's id, name, cost or category; `apps/site/test/rules.test.ts` fails on an id in site code.
- Never cast wasm data, URL parameters or files. Add a schema in `core/schema.ts` and parse.
- Views take a state. A highlight carries a shape, outline, pattern or label as well as a color, and every step is described in words in a polite live region.
- Nothing autoplays; links open paused at their step. Above 50 steps a second the views drop per-step highlights. `prefers-reduced-motion` turns transitions off.
- US spelling, no em dashes, diagrams in plain text (checked by `apps/site/test/rules.test.ts`).

## Checks

```bash
pnpm lint          # ESLint and Prettier over both packages
pnpm typecheck     # builds the package, then tsc in each
pnpm test          # the machine, replay, layout, highlighting and every entry through the real wasm; the house rules
pnpm build         # the package, then the static site
```
