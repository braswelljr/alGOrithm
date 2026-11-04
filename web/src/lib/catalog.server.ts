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
import type { SearchIndex } from '@/lib/search-index';
import {
  startAlgorithm,
  type AlgorithmApi,
  type Catalog,
  type LearningPath
} from '@/visualizer/core';

const publicDir = path.join(process.cwd(), 'public');

/** The started engine, kept on globalThis: the bundler gives each route its
 *  own copy of this module, and every copy in a build worker must share one. */
const STARTED = Symbol.for('alGOrithm.catalogApi');

/** The wasm build, started once per build process. */
function api(): Promise<AlgorithmApi> {
  const running = Reflect.get(globalThis, STARTED) as Promise<AlgorithmApi> | undefined;
  if (running) return running;
  const started = (async () => {
    if (typeof Reflect.get(globalThis, 'Go') !== 'function') {
      // wasm_exec.js defines globalThis.Go; run it as the browser would.
      runInThisContext(await readFile(path.join(publicDir, 'wasm_exec.js'), 'utf8'));
    }
    return startAlgorithm(await readFile(path.join(publicDir, 'algorithm.wasm')));
  })();
  Reflect.set(globalThis, STARTED, started);
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

/** The command palette's search index (src/lib/search-index.ts), from the
 *  catalog and the learning paths. */
export async function buildSearchIndex(): Promise<SearchIndex> {
  const [c, paths] = await Promise.all([buildCatalog(), buildPaths()]);
  const byId = new Map(c.categories.map((cat) => [cat.id, cat]));
  const counts = new Map<string, number>();
  for (const e of c.entries) counts.set(e.category, (counts.get(e.category) ?? 0) + 1);
  return {
    entries: c.entries.map((e) => ({
      id: e.id,
      name: e.name,
      category: e.category,
      categoryName: byId.get(e.category)?.name ?? e.category,
      topic: byId.get(e.category)?.topic ?? '',
      summary: e.summary,
      level: e.level,
      kind: e.kind,
      time: e.time,
      space: e.space,
      input: e.input,
      tags: e.tags
    })),
    paths: paths.map((p) => ({
      id: p.id,
      title: p.title,
      summary: p.summary,
      level: p.level,
      steps: p.steps.length
    })),
    categories: c.categories
      .filter((cat) => (counts.get(cat.id) ?? 0) > 0)
      .map((cat) => ({
        id: cat.id,
        name: cat.name,
        topic: cat.topic,
        summary: cat.summary,
        entries: counts.get(cat.id) ?? 0
      })),
    tags: c.tags.map((t) => ({ id: t.tag, meaning: t.meaning, entries: t.count }))
  };
}
