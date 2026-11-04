/** The quiz game's settings, seeded picks, kept bests, and its built page. */
import assert from 'node:assert/strict';
import { existsSync, readFileSync } from 'node:fs';
import { test } from 'node:test';
import { fileURLToPath } from 'node:url';
import {
  candidates,
  domainsOf,
  loadBests,
  poolSize,
  saveBest,
  seeded,
  shuffled,
  type GameConfig
} from '../src/lib/quiz-seed.ts';

const entry = (id: string, category: string, level: string) => ({ id, category, level }) as never;
const catalog = {
  categories: [
    { id: 'sorting', topic: 'Algorithms' },
    { id: 'scheduling', topic: 'Operating systems' }
  ],
  entries: [
    entry('a', 'sorting', 'beginner'),
    entry('b', 'sorting', 'advanced'),
    entry('c', 'scheduling', 'beginner'),
    entry('d', 'scheduling', 'intermediate')
  ]
} as never;
const config = (over: Partial<GameConfig> = {}): GameConfig => ({
  mode: 'classic',
  domain: 'mixed',
  level: 'any',
  seed: 42,
  entry: null,
  ...over
});
const ids = (es: { id: string }[]) => es.map((e) => e.id);

test('the same seed gives the same order; another seed may not', () => {
  const xs = Array.from({ length: 20 }, (_, i) => i);
  assert.deepEqual(shuffled(xs, seeded(7)), shuffled(xs, seeded(7)));
  assert.notDeepEqual(shuffled(xs, seeded(7)), shuffled(xs, seeded(8)));
  assert.deepEqual(
    [...shuffled(xs, seeded(7))].sort((a, b) => a - b),
    xs,
    'a shuffle keeps every item'
  );
  const r = seeded(1);
  for (let i = 0; i < 100; i++) {
    const v = r();
    assert.ok(v >= 0 && v < 1);
  }
});

test('candidates follow the topic and level, in a seeded order', () => {
  assert.deepEqual(ids(candidates(catalog, config())).sort(), ['a', 'b', 'c', 'd']);
  assert.deepEqual(ids(candidates(catalog, config({ domain: 'Operating systems' }))).sort(), [
    'c',
    'd'
  ]);
  assert.deepEqual(ids(candidates(catalog, config({ level: 'beginner' }))).sort(), ['a', 'c']);
  assert.deepEqual(ids(candidates(catalog, config({ entry: 'b' }))), ['b']);
  assert.deepEqual(ids(candidates(catalog, config())), ids(candidates(catalog, config())));
  assert.deepEqual(domainsOf(catalog), ['Algorithms', 'Operating systems']);
});

test('pool sizes: 10 for classic, more for the endless modes, 5 for one entry', () => {
  assert.equal(poolSize({ mode: 'classic', entry: null }), 10);
  assert.equal(poolSize({ mode: 'time', entry: null }), 40);
  assert.equal(poolSize({ mode: 'streak', entry: 'x' }), 5);
});

test('bests are kept per mode, only when beaten, and bad storage is survived', () => {
  const store = new Map<string, string>();
  const storage = {
    getItem: (k: string) => store.get(k) ?? null,
    setItem: (k: string, v: string) => void store.set(k, v)
  };
  assert.deepEqual(loadBests(storage), {});
  assert.equal(saveBest(storage, 'classic', 500, 3).newBest, true);
  assert.equal(saveBest(storage, 'classic', 400, 5).newBest, false);
  assert.deepEqual(loadBests(storage), { classic: { score: 500, streak: 5 } });
  store.set('algorithm.quiz.best', 'not json');
  assert.deepEqual(loadBests(storage), {});
  const broken = {
    getItem: () => {
      throw new Error('blocked');
    },
    setItem: () => {
      throw new Error('blocked');
    }
  };
  assert.deepEqual(loadBests(broken), {});
  assert.equal(saveBest(broken, 'time', 1, 1).newBest, true);
  assert.deepEqual(loadBests(undefined), {});
});

test('the /quiz page renders in the static export', (t) => {
  const html = fileURLToPath(new URL('../out/quiz/index.html', import.meta.url));
  if (!existsSync(html)) {
    t.skip('build the site first (pnpm --dir web -r build)');
    return;
  }
  const page = readFileSync(html, 'utf8');
  assert.match(page, /<h1[^>]*>Quiz<\/h1>/);
  assert.match(page, /Predict what an algorithm does next/);
});
