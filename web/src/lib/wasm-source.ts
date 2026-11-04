/**
 * Where the WebAssembly builds are loaded from, chosen once per page.
 *
 * On the deployed site the release's files come from GitHub Pages
 * (NEXT_PUBLIC_WASM_BASE, e.g. https://braswelljr.github.io/alGOrithm/wasm/v1.2.0),
 * which the release workflow publishes: Vercel builds only the Next.js site,
 * and Pages serves the files to other origins and from a CDN. When that base is
 * not set (local development) or cannot be reached, the site's own copies are
 * used. Go's wasm_exec.js must come from the same toolchain as the wasm, and a
 * page loads one, so the base that served it serves every build on the page.
 */
import { basePath, wasmVersion } from '@/config/site';

/** A place the builds may come from: wasm_exec.js and each .wasm beside it. */
export type WasmBase = { name: 'release' | 'site'; url: string; query: string };

const CHOSEN = Symbol.for('alGOrithm.wasmBase');

/** The bases to try, the release's first. */
export function wasmBases(): WasmBase[] {
  const release = (process.env.NEXT_PUBLIC_WASM_BASE ?? '').replace(/\/+$/, '');
  return [
    ...(release ? [{ name: 'release' as const, url: release, query: '' }] : []),
    // The site's copies change with each build, so they carry its hash.
    { name: 'site', url: basePath, query: `?v=${wasmVersion}` }
  ];
}

/** The base this page settled on, if it has. */
export function chosenBase(): WasmBase | undefined {
  return Reflect.get(globalThis, CHOSEN) as WasmBase | undefined;
}

function loadScript(src: string): Promise<void> {
  return new Promise((resolve, reject) => {
    const s = document.createElement('script');
    s.src = src;
    s.async = true;
    s.dataset.algorithmExec = '';
    s.addEventListener('load', () => resolve(), { once: true });
    s.addEventListener(
      'error',
      () => {
        s.remove();
        reject(new Error(`could not load ${src}`));
      },
      { once: true }
    );
    document.head.append(s);
  });
}

let choosing: Promise<WasmBase> | undefined;

/**
 * Loads Go's wasm_exec.js from the first base that serves it and remembers
 * that base for the page. Resolves at once when a base is already chosen.
 */
export function ensureGoRuntime(): Promise<WasmBase> {
  const known = chosenBase();
  if (known && typeof Reflect.get(globalThis, 'Go') === 'function') return Promise.resolve(known);
  choosing ??= (async () => {
    let last: unknown;
    for (const base of wasmBases()) {
      try {
        await loadScript(`${base.url}/wasm_exec.js${base.query}`);
        if (typeof Reflect.get(globalThis, 'Go') !== 'function') throw new Error('no Go runtime');
        Reflect.set(globalThis, CHOSEN, base);
        return base;
      } catch (err) {
        last = err;
      }
    }
    choosing = undefined;
    throw last instanceof Error ? last : new Error('Go runtime could not be loaded');
  })();
  return choosing;
}

/** The URL of a build (`algorithm` or `terminal`) on the page's base. */
export async function wasmUrl(name: 'algorithm' | 'terminal'): Promise<string> {
  const base = await ensureGoRuntime();
  return `${base.url}/${name}.wasm${base.query}`;
}
