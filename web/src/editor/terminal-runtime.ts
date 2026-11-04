/**
 * Boots the editor's terminal (web/runtime/cmd/terminal, gnux's shell and the
 * editor's commands) against an in-memory filesystem, after gnux's own
 * playground runtime (lib/playground/runtime.ts there).
 *
 * Go's js/wasm runtime reads the Node-style `fs`, `process` and `path`
 * globals while it starts. The terminal writes its starting files (every
 * package of the repository) before it publishes its API, and those writes
 * finish on the event loop, so the API arrives through the
 * `algorithmTerminalReady` callback rather than at once.
 *
 * Nothing here touches the DOM or the network: the caller supplies how to load
 * `wasm_exec.js` and the module, so the same code runs in a test.
 */
import { arrayOf, isFunction, isNumber, isRecord, isString, prop } from './guards';
import type { MemFS } from './memfs';

/** Receives what a shell writes to standard output (1) or standard error (2). */
export type OutputHandler = (fd: 1 | 2, chunk: Uint8Array, terminal: number) => void;

/** One problem in a file, from the terminal's checks (web/runtime/tooling). */
export interface Diagnostic {
  line: number;
  col: number;
  endLine?: number;
  endCol?: number;
  severity: 'error' | 'warning' | 'info';
  message: string;
  /** The check: go, gofmt, json, make, shell, style. */
  source: string;
}

/** A file formatted, or why it could not be (it does not parse). */
export interface Formatted {
  text: string;
  error?: string;
}

interface GoInstance {
  argv: string[];
  env: Record<string, string>;
  exit: (code: number) => void;
  importObject: WebAssembly.Imports;
  run(instance: WebAssembly.Instance): Promise<void>;
  _scheduledTimeouts?: unknown;
}

type Globals = {
  Go?: new () => GoInstance;
  fs?: unknown;
  path?: unknown;
  process?: unknown;
  algorithmTerminal?: unknown;
  algorithmTerminalReady?: unknown;
};

export interface BootOptions {
  fs: MemFS;
  env?: Record<string, string>;
  /** Evaluates `wasm_exec.js` so `globalThis.Go` is defined. */
  loadGoRuntime: () => Promise<void>;
  /** Instantiates `terminal.wasm` with Go's import object. */
  instantiate: (imports: WebAssembly.Imports) => Promise<WebAssembly.Instance>;
  /** How long to wait for the terminal to seed its files and start. */
  timeoutMs?: number;
}

/** A running terminal and the filesystem it sees. */
export interface TerminalSession {
  /** Every command the shell knows: gnux's tools and the editor's. */
  commands: string[];
  home: string;
  /** Where the repository's packages are. */
  library: string;
  themes: string[];
  fs: MemFS;
  /**
   * Runs one line in a terminal (0 when left out): each is a shell of its own,
   * whose variables and directory carry over between its lines. Resolves the
   * line's status.
   */
  run(line: string, terminal?: number): Promise<number>;
  cwd(terminal?: number): string;
  /** Forgets a terminal: its shell, variables and directory. */
  close(terminal: number): void;
  setenv(name: string, value: string): void;
  /** Receives every terminal's output. */
  onOutput(handler: OutputHandler | undefined): void;
  /** Receives one terminal's output; returns how to stop. */
  listen(terminal: number, handler: (fd: 1 | 2, chunk: Uint8Array) => void): () => void;
  /** The file formatted: gofmt for Go, and tidy whitespace for every file. */
  format(path: string, text: string): Formatted;
  /** The problems in a file, for the editor's flags and problems panel. */
  lint(path: string, text: string): Diagnostic[];
  /** Receives the paths `code FILE` asks the editor to open. */
  onOpen(handler: ((path: string) => void) | undefined): void;
  dispose(): void;
}

export async function bootTerminal(options: BootOptions): Promise<TerminalSession> {
  const g = globalThis as unknown as Globals;
  const memfs = options.fs;
  g.fs = memfs.nodeFs();
  g.path = memfs.nodePath();

  if (!g.Go) {
    const hadProcess = g.process !== undefined;
    await options.loadGoRuntime();
    if (!hadProcess) delete g.process;
    if (!g.Go) throw new Error('wasm_exec.js did not define globalThis.Go');
  }

  const go = new g.Go();
  go.argv = ['terminal'];
  go.env = { ...options.env };
  let exited = false;
  go.exit = () => {
    exited = true;
  };
  const instance = await options.instantiate(go.importObject);

  const ready = new Promise<unknown>((resolve, reject) => {
    const timer = setTimeout(
      () => reject(new Error('The terminal did not start in time.')),
      options.timeoutMs ?? 30_000
    );
    g.algorithmTerminalReady = (api: unknown) => {
      clearTimeout(timer);
      resolve(api);
    };
  });

  // Go reads `process` during its start, which runs synchronously inside
  // go.run until main first waits; the page's own `process` (if any) is put
  // back straight after.
  const realProcess = g.process;
  g.fs = memfs.nodeFs();
  g.path = memfs.nodePath();
  g.process = memfs.nodeProcess();
  try {
    void go.run(instance);
  } finally {
    if (realProcess === undefined) delete g.process;
    else g.process = realProcess;
  }
  if (exited) throw new Error('terminal.wasm exited while starting');

  const published = await ready;
  delete g.algorithmTerminalReady;
  delete g.algorithmTerminal;
  const api = readApi(published);

  // One handler on the program, handing each chunk to its terminal.
  let all: OutputHandler | undefined;
  const listeners = new Map<number, (fd: 1 | 2, chunk: Uint8Array) => void>();
  api.object.onOutput = (fd: unknown, chunk: unknown, id: unknown) => {
    if ((fd !== 1 && fd !== 2) || !(chunk instanceof Uint8Array)) return;
    const t = isNumber(id) ? id : 0;
    all?.(fd, chunk, t);
    listeners.get(t)?.(fd, chunk);
  };

  return {
    commands: api.commands,
    home: api.home,
    library: api.library,
    themes: api.themes,
    fs: memfs,
    async run(line, terminal = 0) {
      const status: unknown = await api.run(line, terminal);
      return isNumber(status) ? status : 1;
    },
    cwd(terminal = 0) {
      const dir = api.cwd(terminal);
      return isString(dir) ? dir : api.home;
    },
    close(terminal) {
      listeners.delete(terminal);
      api.close?.(terminal);
    },
    setenv(name, value) {
      api.setenv(name, value);
    },
    onOutput(handler) {
      all = handler;
    },
    listen(terminal, handler) {
      listeners.set(terminal, handler);
      return () => {
        if (listeners.get(terminal) === handler) listeners.delete(terminal);
      };
    },
    format(path, text) {
      if (!api.format) return { text };
      const r = api.format(path, text);
      if (!isRecord(r)) return { text };
      const out = prop(r, 'text');
      const error = prop(r, 'error');
      return { text: isString(out) ? out : text, error: isString(error) ? error : undefined };
    },
    lint(path, text) {
      if (!api.lint) return [];
      const raw = api.lint(path, text);
      if (!isString(raw)) return [];
      try {
        const parsed: unknown = JSON.parse(raw);
        return arrayOf(parsed, isDiagnostic);
      } catch {
        return [];
      }
    },
    onOpen(handler) {
      api.object.onOpen =
        handler &&
        ((path: unknown) => {
          if (isString(path)) handler(path);
        });
    },
    dispose() {
      all = undefined;
      listeners.clear();
      api.object.onOutput = undefined;
      api.object.onOpen = undefined;
      const timeouts = go._scheduledTimeouts;
      if (timeouts instanceof Map) {
        for (const id of timeouts.values()) clearTimeout(id);
        timeouts.clear();
      }
    }
  };
}

/**
 * Checks what the program published against the contract in
 * web/runtime/cmd/terminal/main.go: it comes from another program, so
 * nothing about it is taken on trust.
 */
function readApi(value: unknown) {
  if (!isRecord(value)) throw new Error('terminal.wasm started but did not publish its API');
  const run = prop(value, 'run');
  const cwd = prop(value, 'cwd');
  const setenv = prop(value, 'setenv');
  if (!isFunction(run) || !isFunction(cwd) || !isFunction(setenv)) {
    throw new Error('terminal.wasm published an API without run(), cwd() and setenv()');
  }
  const home = prop(value, 'home');
  const library = prop(value, 'library');
  // Added later than the rest: an older terminal.wasm lacks them.
  const close = prop(value, 'close');
  const format = prop(value, 'format');
  const lint = prop(value, 'lint');
  return {
    object: value,
    commands: arrayOf(prop(value, 'commands'), isString),
    themes: arrayOf(prop(value, 'themes'), isString),
    home: isString(home) ? home : '/home/guest',
    library: isString(library) ? library : '/home/guest/go/src',
    run: (line: string, id: number): unknown => run.call(value, line, id),
    cwd: (id: number): unknown => cwd.call(value, id),
    close: isFunction(close) ? (id: number): unknown => close.call(value, id) : undefined,
    format: isFunction(format)
      ? (path: string, text: string): unknown => format.call(value, path, text)
      : undefined,
    lint: isFunction(lint)
      ? (path: string, text: string): unknown => lint.call(value, path, text)
      : undefined,
    setenv: (name: string, v: string): unknown => setenv.call(value, name, v)
  };
}

function isDiagnostic(v: unknown): v is Diagnostic {
  if (!isRecord(v)) return false;
  const severity = prop(v, 'severity');
  return (
    isNumber(prop(v, 'line')) &&
    isNumber(prop(v, 'col')) &&
    isString(prop(v, 'message')) &&
    isString(prop(v, 'source')) &&
    (severity === 'error' || severity === 'warning' || severity === 'info')
  );
}
