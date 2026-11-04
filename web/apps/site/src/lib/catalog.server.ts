/**
 * The catalog at build time, for what has to exist before any JavaScript runs:
 * one static page per entry, its metadata and social card, and the sitemap.
 * It runs the same algorithm.wasm the browser loads (in public/, written by
 * `make web/wasm`), so the build lists exactly the entries the site can run
 * (docs/rules/there-is-one-list-of-entries.md). Only server code imports this.
 */
import { readFile } from 'node:fs/promises';
import path from 'node:path';
import { runInThisContext } from 'node:vm';
import {
  startAlgorithm,
  type AlgorithmApi,
  type Catalog,
  type LearningPath
} from '@braswelljr/algorithm-visualizer/core';

const publicDir = path.join(process.cwd(), 'public');

let started: Promise<AlgorithmApi> | undefined;

/** The wasm build, started once per build process. */
function api(): Promise<AlgorithmApi> {
  started ??= (async () => {
    if (typeof Reflect.get(globalThis, 'Go') !== 'function') {
      // wasm_exec.js defines globalThis.Go; run it as the browser would.
      runInThisContext(await readFile(path.join(publicDir, 'wasm_exec.js'), 'utf8'));
    }
    return startAlgorithm(await readFile(path.join(publicDir, 'algorithm.wasm')));
  })();
  return started;
}

/** Every category and entry. */
export async function buildCatalog(): Promise<Catalog> {
  return (await api()).catalog();
}

/** Every learning path, or none when the build has no paths yet. */
export async function buildPaths(): Promise<LearningPath[]> {
  const a = await api();
  if (typeof a.paths !== 'function') return [];
  const r = a.paths();
  return r.ok ? r.value : [];
}

/**
 * The catalog with only what a list of entries shows (name, summary, cost,
 * level, tags), for pages that list entries without running any. The page
 * ships it in its HTML, so it renders, and filters, before any wasm loads;
 * the full entry comes with the wasm when the reader opens one.
 */
export async function buildCatalogSummary(): Promise<Catalog> {
  const c = await buildCatalog();
  return {
    ...c,
    entries: c.entries.map((e) => ({
      id: e.id,
      name: e.name,
      category: e.category,
      summary: e.summary,
      time: e.time,
      space: e.space,
      input: e.input,
      traced: e.traced,
      kind: e.kind,
      level: e.level,
      tags: e.tags,
      import: '',
      package: '',
      func: '',
      file: '',
      why: '',
      realWorld: [],
      params: [],
      options: [],
      ops: [],
      choices: [],
      example: {}
    }))
  };
}
