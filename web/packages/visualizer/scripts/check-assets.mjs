// Refuses to publish without the wasm build beside the code: run
// `make web/wasm` at the repository root first.
import { stat } from 'node:fs/promises';

for (const f of ['assets/algorithm.wasm', 'assets/wasm_exec.js']) {
  try {
    await stat(new URL(`../${f}`, import.meta.url));
  } catch {
    console.error(`missing ${f}: run \`make web/wasm\` at the repository root first`);
    process.exit(1);
  }
}
