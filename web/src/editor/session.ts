/**
 * The editor's terminal, booted once per page: Go's runtime and terminal.wasm
 * from the page's chosen base (the release's on GitHub Pages, else the site's
 * own; lib/wasm-source.ts), against an in-memory filesystem seeded with the
 * visitor's saved files.
 */
import { ensureGoRuntime, wasmUrl } from '@/lib/wasm-source';
import { restoreSaved } from './files';
import { createEditorFS, ENV } from './seed';
import { bootTerminal, type TerminalSession } from './terminal-runtime';

let booting: Promise<TerminalSession> | undefined;

async function instantiate(imports: WebAssembly.Imports): Promise<WebAssembly.Instance> {
  const url = await wasmUrl('terminal');
  try {
    // Compiles while it downloads when the host sends application/wasm.
    return (await WebAssembly.instantiateStreaming(fetch(url), imports)).instance;
  } catch {
    const res = await fetch(url);
    if (!res.ok) throw new Error(`could not download ${url} (${res.status})`);
    return (await WebAssembly.instantiate(await res.arrayBuffer(), imports)).instance;
  }
}

/** The page's terminal, starting it on the first call; a failed start can be retried. */
export function terminalSession(): Promise<TerminalSession> {
  booting ??= (async () => {
    const session = await bootTerminal({
      fs: createEditorFS(),
      env: ENV,
      loadGoRuntime: async () => {
        await ensureGoRuntime();
      },
      instantiate
    });
    // The visitor's own files, from this browser, over the starting ones.
    restoreSaved(session.fs);
    return session;
  })();
  booting.catch(() => {
    booting = undefined;
  });
  return booting;
}
