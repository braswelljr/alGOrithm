/**
 * A typed wrapper around the raw object the wasm build installs on
 * globalThis.alGOrithm. The raw object is unknown: each function is checked to
 * exist, and every string it returns is parsed with the schemas in schema.ts,
 * so a mismatch between the Go side and the page fails loudly here instead of
 * deep inside a view.
 */
import { z } from 'zod';
import { measureResponse, type Measurement } from './measure';
import { quizResponse, type Question } from './quiz';
import {
  catalogSchema,
  compareResponse,
  designsResponse,
  exampleResponse,
  pathsResponse,
  previewResponse,
  runResponse,
  snippetResponse,
  sourceResponse,
  textResponse,
  valuesResponse,
  type Catalog,
  type Designs,
  type ExampleProgram,
  type Failure,
  type Input,
  type LearningPath,
  type Preview,
  type RunResult,
  type Snippet,
  type Source
} from './schema';

/** A call that either succeeded with T or failed with the Go side's message. */
export type Outcome<T> = { ok: true; value: T } | { ok: false; failure: Failure };

export interface ParseOptions {
  /** The CSV column: a 1-based number or a header name. */
  column?: string;
  header?: boolean;
  skipEmpty?: boolean;
  round?: boolean;
  /** The JSON field holding the values. */
  field?: string;
}

export type DataFormat = 'text' | 'csv' | 'json';

export interface AlgorithmApi {
  version: string;
  catalog(): Catalog;
  run(id: string, input: Input, limit?: number): Outcome<RunResult>;
  compare(ids: string[], input: Input): Outcome<RunResult[]>;
  parse(format: DataFormat, text: string, options?: ParseOptions): Outcome<number[]>;
  preview(text: string, header: boolean): Outcome<Preview>;
  generate(kind: string, size: number, seed: number): Outcome<number[]>;
  generateGrid(kind: string, rows: number, cols: number, seed: number): Outcome<string>;
  source(id: string): Outcome<Source>;
  snippet(id: string, input: Input): Outcome<Snippet>;
  /** The Go source of a program in examples/ ("examples/leaderboard") and the command that runs it. */
  example(path: string): Outcome<ExampleProgram>;
  /** The learning paths, each step with its entry's name and command. */
  paths(): Outcome<LearningPath[]>;
  /** The system design case studies, and the pattern and foundation entries. */
  designs(): Outcome<Designs>;
  /** Questions about a run of id on input (its example when input is empty):
   * what it does next, checked against its events. */
  quiz(id: string, input: Input, count?: number, seed?: number): Outcome<Question[]>;
  /** Runs id at growing sizes (the defaults when sizes is empty) and fits the
   * growth of its operation count to O(1) ... O(2^n). */
  measure(id: string, sizes?: number[], order?: string, seed?: number): Outcome<Measurement>;
}

type Raw = (...args: unknown[]) => unknown;

function missing(name: string): Raw {
  return () =>
    JSON.stringify({ ok: false, error: `this build of the algorithms has no ${name}()` });
}

function method(raw: Record<string, unknown>, name: string, required: boolean): Raw {
  const f = raw[name];
  if (typeof f !== 'function') {
    if (required) throw new Error(`the wasm build has no ${name}()`);
    return missing(name);
  }
  return (...args: unknown[]) => Reflect.apply(f, raw, args);
}

function call<T extends z.ZodType>(f: Raw, schema: T, ...args: unknown[]): z.infer<T> {
  const text = f(...args);
  if (typeof text !== 'string') {
    throw new Error('the wasm build returned something other than JSON text');
  }
  const parsed = schema.safeParse(JSON.parse(text));
  if (!parsed.success) {
    throw new Error(
      `the wasm build returned an unexpected shape: ${z.prettifyError(parsed.error)}`
    );
  }
  return parsed.data;
}

function outcome<T, R extends { ok: boolean }>(
  r: R,
  pick: (ok: Extract<R, { ok: true }>) => T
): Outcome<T> {
  return r.ok
    ? { ok: true, value: pick(r as Extract<R, { ok: true }>) }
    : { ok: false, failure: r as unknown as Failure };
}

/** Wraps the raw wasm object. Throws when it lacks the core of the API. */
export function wrapApi(raw: unknown): AlgorithmApi {
  if (typeof raw !== 'object' || raw === null) {
    throw new Error('the wasm build did not install its API');
  }
  const obj: Record<string, unknown> = { ...raw };
  const version = typeof obj.version === 'string' ? obj.version : 'unknown';
  const catalog = method(obj, 'catalog', true);
  const run = method(obj, 'run', true);
  const compare = method(obj, 'compare', true);
  const parse = method(obj, 'parse', true);
  const preview = method(obj, 'preview', true);
  const generate = method(obj, 'generate', true);
  const generateGrid = method(obj, 'generateGrid', false);
  const source = method(obj, 'source', false);
  const snippet = method(obj, 'snippet', false);
  const example = method(obj, 'example', false);
  const paths = method(obj, 'paths', false);
  const designs = method(obj, 'designs', false);
  const quiz = method(obj, 'quiz', false);
  const measure = method(obj, 'measure', false);

  let cached: Catalog | undefined;
  return {
    version,
    catalog: () => (cached ??= call(catalog, catalogSchema)),
    run: (id, input, limit = 0) =>
      outcome(call(run, runResponse, id, JSON.stringify(input), limit), (r) => r.result),
    compare: (ids, input) =>
      outcome(
        call(compare, compareResponse, JSON.stringify(ids), JSON.stringify(input)),
        (r) => r.results
      ),
    parse: (format, text, options = {}) =>
      outcome(call(parse, valuesResponse, format, text, JSON.stringify(options)), (r) => r.values),
    preview: (text, header) => outcome(call(preview, previewResponse, text, header), (r) => r),
    generate: (kind, size, seed) =>
      outcome(call(generate, valuesResponse, kind, size, seed), (r) => r.values),
    generateGrid: (kind, rows, cols, seed) =>
      outcome(call(generateGrid, textResponse, kind, rows, cols, seed), (r) => r.text),
    source: (id) => outcome(call(source, sourceResponse, id), (r) => r),
    snippet: (id, input) =>
      outcome(call(snippet, snippetResponse, id, JSON.stringify(input)), (r) => r),
    example: (path) => outcome(call(example, exampleResponse, path), (r) => r),
    paths: () => outcome(call(paths, pathsResponse), (r) => r.paths),
    designs: () =>
      outcome(call(designs, designsResponse), (r) => ({
        studies: r.studies,
        patterns: r.patterns,
        foundations: r.foundations
      })),
    quiz: (id, input, count = 5, seed = 1) =>
      outcome(call(quiz, quizResponse, id, JSON.stringify(input), count, seed), (r) => r.questions),
    measure: (id, sizes = [], order = 'random', seed = 1) =>
      outcome(
        call(measure, measureResponse, id, sizes.length ? JSON.stringify(sizes) : '', order, seed),
        (r) => r.measurement
      )
  };
}
