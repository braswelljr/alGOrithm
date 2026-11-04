# web

The alGOrithm site: one Next.js app (static export) that runs the Go library, compiled to WebAssembly, in the browser. Visualizations, docs for every entry, learning paths, the quiz and Big O all live here; the library itself is the Go module at the repository root.

```text
web/
  src/app/                  routes
  src/components/ui/        the primitives (Base UI); everything else is built from them
  src/components/algorithm/ the app's own compositions
  src/visualizer/           the player, replay and the views
  test/                     every test, run by `pnpm test`
  public/                   algorithm.wasm, wasm_exec.js and the brand assets
```

From the repository root:

```bash
make web/install   # pnpm install
make dev           # build the wasm, start the site, rebuild on every Go change
make web/test      # wasm, then lint, type-check and the tests
make web/build     # wasm, then the static site in web/out, then the size budgets
```

See [AGENTS.md](AGENTS.md) for how it fits together.
