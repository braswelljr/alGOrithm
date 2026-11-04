# web/

The alGOrithm visualizer: an npm package of React components over the Go library compiled to WebAssembly, and the site built on it. A pnpm workspace.

```text
web/
  packages/visualizer/   @braswelljr/algorithm-visualizer: the components, the player, replay, the wasm loader
  apps/site/             the Next.js static site: catalog, visualize, structures, compare, data, docs
```

From the repository root:

```bash
make web/install   # pnpm install
make web/dev       # build the wasm, then the package (watching) and the site
make web/test      # wasm, then lint, type-check and every test against the real wasm
make web/build     # wasm, then the package and the static site in apps/site/out
```

`make dev` at the root also rebuilds the wasm whenever Go code changes and reloads the page.

See [AGENTS.md](AGENTS.md) for how it fits together, and [the package README](packages/visualizer/README.md) to use the components in another app.
