/**
 * Loads the wasm build once. Go's loader (wasm_exec.js) and the binary are
 * static files the host serves; nothing is fetched from anywhere else.
 */
import { wrapApi, type AlgorithmApi } from './api';

export interface LoadOptions {
  /** Where algorithm.wasm is served. Default: /algorithm.wasm */
  wasmUrl?: string;
  /** Where Go's wasm_exec.js is served. Default: /wasm_exec.js */
  execUrl?: string;
  /**
   * Loads Go's runtime and returns the wasm's URL, for a host that chooses
   * where the builds come from (the site: a release's, else its own); when
   * set, wasmUrl and execUrl are not used.
   */
  resolve?: () => Promise<string>;
}

let loading: Promise<AlgorithmApi> | undefined;

function loadScript(src: string): Promise<void> {
  return new Promise((resolve, reject) => {
    if (typeof Reflect.get(globalThis, 'Go') === 'function') {
      resolve();
      return;
    }
    const existing = document.querySelector<HTMLScriptElement>('script[data-algorithm-exec]');
    const s = existing ?? document.createElement('script');
    s.addEventListener('load', () => resolve(), { once: true });
    s.addEventListener('error', () => reject(new Error(`could not load ${src}`)), { once: true });
    if (!existing) {
      s.src = src;
      s.async = true;
      s.dataset.algorithmExec = '';
      document.head.append(s);
    }
  });
}

/**
 * Starts Go on an already-compiled module or bytes and resolves with the API.
 * Used by the browser loader below and by Node (tests, scripts), which pass
 * the bytes they read themselves.
 */
/**
 * Starts are queued, one at a time: the Go side announces itself through one
 * global callback (`alGOrithmReady`), so two starts at once would hand the
 * first instance's announcement to the second and leave the first waiting
 * for ever. The queue lives on globalThis, shared by every copy of this
 * module a bundler makes.
 */
const QUEUE = Symbol.for('alGOrithm.startQueue');

export function startAlgorithm(
  module: BufferSource | WebAssembly.Module | Promise<Response>
): Promise<AlgorithmApi> {
  const previous =
    (Reflect.get(globalThis, QUEUE) as Promise<unknown> | undefined) ?? Promise.resolve();
  const next = previous.catch(() => undefined).then(() => startOne(module));
  Reflect.set(globalThis, QUEUE, next);
  return next;
}

async function startOne(
  module: BufferSource | WebAssembly.Module | Promise<Response>
): Promise<AlgorithmApi> {
  const GoCtor: unknown = Reflect.get(globalThis, 'Go');
  if (typeof GoCtor !== 'function') throw new Error('wasm_exec.js did not define Go');
  const go = Reflect.construct(GoCtor, []);
  const ready = new Promise<void>((resolve) => Reflect.set(globalThis, 'alGOrithmReady', resolve));
  let instance: WebAssembly.Instance;
  if (module instanceof Promise) {
    ({ instance } = await WebAssembly.instantiateStreaming(module, go.importObject));
  } else if (module instanceof WebAssembly.Module) {
    instance = await WebAssembly.instantiate(module, go.importObject);
  } else {
    ({ instance } = await WebAssembly.instantiate(module, go.importObject));
  }
  void go.run(instance);
  await ready;
  return wrapApi(Reflect.get(globalThis, 'alGOrithm'));
}

async function boot({
  wasmUrl: fixedUrl = '/algorithm.wasm',
  execUrl = '/wasm_exec.js',
  resolve
}: LoadOptions) {
  let wasmUrl = fixedUrl;
  if (resolve) wasmUrl = await resolve();
  else await loadScript(execUrl);
  try {
    return await startAlgorithm(fetch(wasmUrl));
  } catch {
    // A server that does not send application/wasm cannot stream it.
    const res = await fetch(wasmUrl);
    if (!res.ok) throw new Error(`could not fetch ${wasmUrl} (${res.status})`);
    return startAlgorithm(await res.arrayBuffer());
  }
}

/**
 * The algorithms, loading them on the first call. Later calls return the same
 * instance whatever options they pass; a failed load can be retried.
 */
export function loadAlgorithm(options: LoadOptions = {}): Promise<AlgorithmApi> {
  if (!loading) {
    loading = boot(options);
    loading.catch(() => {
      loading = undefined;
    });
  }
  return loading;
}
