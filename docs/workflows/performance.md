# Performance

How heavy the site is, what makes it heavy, and the budgets that keep it light. `node web/scripts/budget.mjs` checks the budgets against the static export in `web/out/`, and `make web/test` runs it after the build.

## What a visitor downloads

```text
every page     React + Next.js runtime (~125 KB gzip), the page's own code, one CSS file
pages that     + algorithm.wasm (the Go build: ~3.5 MB brotli), wasm_exec.js (4 KB gzip),
run entries      the visualizer views, React Flow only when a graph is drawn
pages that     + nothing: the catalog, paths and tags are in the HTML, from the build
list entries
```

The wasm dominates everything else by an order of magnitude, so the rule is: **a page loads the wasm only when the reader runs something.** Home, docs, tags and learn render their lists from build-time data (`buildCatalogSummary`, `buildPaths` in `src/lib/catalog.server.ts`) and never fetch it.

## Measurements

Gzip sizes of the first-load JavaScript per route (scripts in the HTML, `nomodule` polyfills excluded since modern browsers skip them), measured on the static export. The baseline is before the performance pass. "After" is the 08:28 build of 2026-11-04: it has the build-time catalog, the hooks-only entry and lazy React Flow, but it also has the new site shell (theme picker, shortcuts, toasts), which added about 75 KB to every page, and it predates the mobile menu and the lazy quiz screens. The list pages no longer download the 3.3 MB wasm at all, which is the change that matters most to a visitor. The JavaScript budgets are the targets for the final pass and fail today.

| Route | Baseline | After | Budget |
|-------|----------|-------|--------|
| `/` | 300.2 KB, plus the wasm | 262.5 KB, no wasm | 150 KB |
| `/learn` | 296.8 KB, plus the wasm | 259.7 KB, no wasm | 150 KB |
| `/docs` | 301 KB, plus the wasm | 262.9 KB, no wasm | 150 KB |
| `/docs/quick-sort` | 301.5 KB | 352.9 KB | 260 KB |
| `/visualize` | 311.8 KB | 411.9 KB | 260 KB |
| `/quiz` | 383.8 KB | 382.6 KB | 300 KB |
| `/big-o` | 301.4 KB | 271.0 KB | 260 KB |

| Asset | Baseline | Now | Budget |
|-------|----------|-----|--------|
| `algorithm.wasm` raw | 17.3 MB | 19.2 MB (more entries) | |
| `algorithm.wasm` gzip | 4.45 MB | 4.89 MB | 5.0 MB |
| `algorithm.wasm` brotli | 3.21 MB | 3.49 MB | 3.6 MB |
| `wasm_exec.js` | 17.0 KB (4.3 KB gzip) | same | |
| CSS (one sheet) | 41 KB gzip | 49.8 KB gzip | 50 KB gzip |
| Fonts | none yet | 77 KB (2 latin woff2) | 160 KB (text font + mono, latin woff2) |
| Logo mark SVG | 841 B | same | 1 KB |
| `algorithm` CLI | 20.1 MB | 14.8 MB with `-s -w` (what `make build` uses) | |

What the wasm is made of: the Go runtime and `encoding/json`, the code of every entry, and the embedded source of every package for the code view (2.0 MB raw, about 0.49 MB brotli, of which 0.37 MB raw is tests). It grows with every entry; the budget is raised on purpose, in this file, when a batch of entries lands.

The static export itself is large (118 MB) because every entry has its own page and its own social card PNG (477 PNGs, 54 MB). Visitors never download those unless they share a link; they only cost deploy time.

## What was done

- **The wasm only where it runs.** Home, docs, tags and learn take their data from the build. The structures principles panel on `/docs`, the one part of those pages that runs entries, loads the wasm and its views on its own (`next/dynamic`).
- **No views in the hook.** The site's `useSiteAlgorithm` imports `@/visualizer/react`, a module with only the hooks, instead of the package root, which pulled every view (and React Flow) into every page.
- **React Flow on demand.** The four views drawn with React Flow (graph, structure, state machine, circuit) are `React.lazy` wrappers (`web/src/visualizer/components/lazy-flow.tsx`), with a skeleton holding their space; React Flow downloads only when a graph is shown.
- **Development code out of production.** `DevReload` (and its zod import) renders only in development.
- **The quiz game on demand.** The play and results screens (GSAP, the live view, the burst) load while the start screen is up.
- **Caching that cannot go stale.** The wasm and `wasm_exec.js` URLs carry the wasm's content hash (`?v=`, from `next.config.ts`), so `vercel.json` caches them for a year, immutably, and a new build is a new URL. `public/sw.js` keeps the current version in the Cache API for repeat and offline visits, keyed by that URL, and deletes older ones.
- **Streaming compile.** The loader uses `WebAssembly.instantiateStreaming`, so the wasm compiles while it downloads.
- **Smaller wasm, when the tools are there.** `make web/wasm` builds with `-trimpath -ldflags "-s -w"`, runs `wasm-opt -Oz` when binaryen is installed (`brew install binaryen`; usually 10 to 20% smaller), and writes `.wasm.gz` and `.wasm.br` beside it for hosts that serve precompressed files (Vercel compresses on its own).
- **Mobile.** The header's navigation becomes a 44px menu button and a sheet (`components/layout/mobile-nav.tsx`, on the ui package's `Sheet` and `Button`) below the `md` breakpoint; the GitHub link is a 44px target on touch screens; wide tables use the ui package's `Table`, which scrolls inside itself; `pre` blocks scroll inside themselves; the page never scrolls sideways (`overflow-x: clip`); fields are at least 16px on touch screens so iOS does not zoom; the installed app keeps clear of the notch with `env(safe-area-inset-*)`; `[data-decorative]` elements are hidden under `prefers-reduced-data`.

## Targets

| Metric | Target | How it is kept |
|--------|--------|----------------|
| LCP | under 2.5 s on 4G | list pages are static HTML with their content; the wasm is never on the critical path of a page that does not run anything |
| CLS | under 0.05 | skeletons hold the space of lazy views; fonts through `next/font` with a size-adjusted fallback |
| INP | under 200 ms | see the Web Worker item below: a large traced run can block the main thread for longer |
| First-load JS | the budgets above | `scripts/budget.mjs` |

## Quality, not traded for size

| Check | How |
|-------|-----|
| No scaled raster image where an SVG would do | logos, the mark and the icons in the UI are SVG; PNGs only where a platform needs one (favicons, touch icons, social cards) |
| Icons crisp at device pixel ratios 1, 2 and 3 | SVG icons everywhere in the page; the raster favicons and touch icons are generated at every size they are shown at, never scaled |
| Animations at 60 fps, no layout thrash | animate only `transform` and `opacity`; no animation reads layout in a loop |
| No visible CLS | skeletons for lazy views, reserved heights for the player and the views, fonts with a size-adjusted fallback |
| No flash of invisible text | `next/font` with `display: swap` and a metric-matched fallback |
| Social cards and images sharp when small | OG cards rendered at 1200x630 (2x of their usual display), icons at their exact pixel sizes |
| Fonts | one open variable text font and one mono font, both through `next/font/google`, latin subset; the budget counts both woff2 files |

## Not done yet

- **The shared shell.** Every page now loads zod (31 KB gzip, through the command palette's `useSiteAlgorithm` and the package loader's schemas), sonner and motion (22 KB) and react-hotkeys-hook (with nuqs, 25 KB) before anything is shown. Load the palette, the toaster and the shortcuts on first use (idle or first key press), and give the loader a schema-free fast path for the catalog, to bring the list pages back under 150 KB.

- **Runs in a Web Worker.** Measured in Node on a laptop: starting the wasm takes about 95 ms and `catalog()` 70 ms; a traced quick sort of 2,000 values takes 57 ms, merge sort of 5,000 values 125 ms, bubble sort of 1,000 values 156 ms, most of it building and serializing a 1 to 4 MB result. A mid-range phone is roughly four times slower, so large inputs block the main thread well past the 50 ms that keeps input responsive. Moving `run` into a worker makes the package's API asynchronous; it is the next performance change, after the views being built settle.
- **Smaller results.** Most of a large run's time is the JSON of its events; a compact encoding (arrays instead of objects per event) would cut it.
- **TinyGo.** Not installed here, so not measured; it usually makes a much smaller wasm but supports less of the standard library, and every entry's tests would have to pass under it.
- **`motion` through `LazyMotion`.** The game uses the full `motion` component; `LazyMotion` with `domAnimation` would load less.
