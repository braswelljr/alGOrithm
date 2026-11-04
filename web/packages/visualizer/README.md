# @braswelljr/algorithm-visualizer

Watch real algorithms and data structures run, step by step, in any React app. The algorithms are the [alGOrithm](https://github.com/braswelljr/alGOrithm) Go library compiled to WebAssembly, so what you see is what the code does: every comparison, swap, write, visit, push and pop, replayed by a player you can pause, step, scrub and replay. Nothing leaves the browser.

## Install

```bash
pnpm add @braswelljr/algorithm-visualizer
```

React 19 is a peer dependency. Serve the two files in `assets/` from your site (for example, copy them into `public/`):

```bash
cp node_modules/@braswelljr/algorithm-visualizer/assets/{algorithm.wasm,wasm_exec.js} public/
```

and import the styles once:

```ts
import '@braswelljr/algorithm-visualizer/styles.css';
```

## Use

```tsx
import { Visualizer } from '@braswelljr/algorithm-visualizer';

export function Demo() {
  return (
    <Visualizer
      entryId="quick-sort"
      input={{ values: [5, 3, 9, 1, 7] }}
    />
  );
}
```

The files are loaded from `/algorithm.wasm` and `/wasm_exec.js`. Elsewhere, say where:

```tsx
import { AlgorithmProvider } from '@braswelljr/algorithm-visualizer';

<AlgorithmProvider
  wasmUrl="/static/algorithm.wasm"
  execUrl="/static/wasm_exec.js"
>
  <Visualizer
    entryId="a-star"
    input={{ text: 'S..#\n.#..\n...E' }}
  />
</AlgorithmProvider>;
```

Every entry, its input kind and its options come from the catalog:

```ts
import { loadAlgorithm } from '@braswelljr/algorithm-visualizer';

const api = await loadAlgorithm();
for (const e of api.catalog().entries) console.log(e.id, e.input, e.time);
```

### Inputs

| `entry.input`             | Input                                                                     |
| ------------------------- | ------------------------------------------------------------------------- |
| `ints`                    | `{ values }`                                                              |
| `sorted-ints`, `any-ints` | `{ values, target }`                                                      |
| `text`                    | `{ text, key?, shift? }`                                                  |
| `two-texts`, `pattern`    | `{ text, other }`                                                         |
| `number`, `two-numbers`   | `{ values: [a] }`, `{ values: [a, b] }`                                   |
| `amount`                  | `{ values: coins, target: amount }`                                       |
| `grid`                    | `{ text }`: `#` wall, `.` open, `1`-`9` cost, `S` start, `E` end          |
| `ops`                     | `{ values: start, ops: [{ name: 'push', value: '4' }, { name: 'pop' }] }` |
| `processes`               | `{ text }`: one process per line, `name arrival burst [priority]`         |
| `pages`                   | `{ values: references }`, with the `frames` option                        |
| `tracks`                  | `{ values: requests }`, with the `head`, `size` and `down` options        |
| `partitions`              | `{ values: partition sizes, text: process sizes }`                        |
| `banker`                  | `{ text }`, in the format of the entry's example                          |
| `simulation`              | options only                                                              |

Every entry has an `example` input to start from, and `options` lists its numeric settings (`quantum`, `frames`, `capacity`, ...), passed as `input.options`.

### Building blocks

`<Visualizer>` runs an entry and hands the result to `<RunView>`, which picks the view and adds the player. Use the pieces directly for your own layouts:

| Export                                                                                                                                                   | What it is                                                                     |
| -------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------ |
| `RunView`                                                                                                                                                | Draws any run result (even one saved earlier) with the right view and a player |
| `BarsView`, `CellsView`, `GridView`, `TableView`, `TextMatchView`, `HanoiView`, `StructureFlow`, `TimelineView`, `TrackView`, `LinesView`, `SummaryList` | The views; each draws a state, never an event                                  |
| `PlayerControls`, `usePlayer`                                                                                                                            | The player: play, pause, step, seek, speed, replay, keyboard                   |
| `playerReducer`, `createPlayerMachine`                                                                                                                   | The player's state machine, framework-free                                     |
| `ArrayReplayer`, `GridReplayer`, `TableReplayer`, `TextReplayer`                                                                                         | Events to the state after any step, with checkpoints for fast seeking          |
| `CodeView`, `highlightGo`                                                                                                                                | Go source with highlighting                                                    |
| `loadAlgorithm`, `startAlgorithm`, schemas                                                                                                               | The wasm API, every result validated with zod                                  |

`@braswelljr/algorithm-visualizer/core` has the framework-free half alone (the API, schemas, machine, replay and layout), for Node or another UI library.

### The player

A finite state machine, written as a pure reducer:

```text
         LOAD            LOADED
idle ----------> loading -------> ready --PLAY--> playing <--PLAY-- paused
  ^                 |                               |  ^              ^
  | RESET           | FAIL                    PAUSE |  | REPLAY       |
  |                 v                               v  |              |
  +-------------- error                          paused / finished ---+
```

It never starts on its own: only PLAY, TOGGLE and REPLAY play. An event that is not legal in the current state returns the same state. With `prefers-reduced-motion` nothing animates between states.

### Theming

Everything is plain CSS under `.av-root`. Override the custom properties (`--av-accent`, `--av-bg`, `--av-compare`, ...) to theme it; dark mode follows the system, or set `data-theme="dark"` or `"light"` on an element with class `av-root`.

## License

MIT. Includes styles from [React Flow](https://reactflow.dev) (MIT).
