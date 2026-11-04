/**
 * House rules a machine can check, over everything under web/: no em dashes,
 * nothing that sends data off the page, no cast of what the wasm build or a
 * file returns, no Mermaid, and no algorithm hard-coded in the site (the
 * catalog is the only list: docs/rules/there-is-one-list-of-entries.md).
 */
import assert from 'node:assert/strict';
import { readdir, readFile } from 'node:fs/promises';
import { join } from 'node:path';
import { test } from 'node:test';
import { fileURLToPath } from 'node:url';

const web = fileURLToPath(new URL('../../..', import.meta.url));
const EM_DASH = String.fromCharCode(0x2014);
const skip = new Set(['node_modules', '.next', 'out', 'dist', 'public', 'assets']);

async function files(dir: string): Promise<string[]> {
  const out: string[] = [];
  for (const d of await readdir(dir, { withFileTypes: true })) {
    if (skip.has(d.name)) continue;
    const p = join(dir, d.name);
    if (d.isDirectory()) out.push(...(await files(p)));
    else if (/\.(ts|tsx|css|md|mjs|json|yaml)$/.test(d.name) && d.name !== 'pnpm-lock.yaml')
      out.push(p);
  }
  return out;
}

const code = async () =>
  (await files(web)).filter((f) => /\.(ts|tsx)$/.test(f) && !f.includes('/test/'));

test('no em dashes in anything we write', async () => {
  const bad: string[] = [];
  for (const f of await files(web)) {
    if ((await readFile(f, 'utf8')).includes(EM_DASH)) bad.push(f.slice(web.length));
  }
  assert.deepEqual(bad, [], 'replace each em dash with a colon, a comma or parentheses');
});

test('diagrams are plain text, never Mermaid', async () => {
  for (const f of await files(web)) {
    if (!f.endsWith('.md')) continue;
    assert.ok(
      !/```mermaid/.test(await readFile(f, 'utf8')),
      `${f.slice(web.length)} has a Mermaid block`
    );
  }
});

test('the user data never leaves the browser', async () => {
  const forbidden = [
    /sendBeacon/,
    /\bgtag\(/,
    /google-analytics|googletagmanager|plausible|posthog|segment\.com/i,
    /XMLHttpRequest/,
    /method:\s*['"]POST/i
  ];
  for (const f of await code()) {
    const text = await readFile(f, 'utf8');
    for (const re of forbidden) assert.ok(!re.test(text), `${f.slice(web.length)} matches ${re}`);
  }
});

test('outside data is parsed, never cast', async () => {
  for (const f of await code()) {
    const t = await readFile(f, 'utf8');
    assert.ok(
      !/JSON\.parse\([^)]*\)\s+as\s/.test(t),
      `${f.slice(web.length)} casts a JSON.parse result`
    );
    assert.ok(
      !/Reflect\.get\(globalThis, 'alGOrithm'\)\s+as/.test(t),
      `${f.slice(web.length)} casts the wasm API`
    );
  }
});

test('the site names no algorithm: everything comes from the catalog', async () => {
  // Ids an entry could have. The site may mention them in prose, not in code.
  const ids =
    /['"`](quick-sort|merge-sort|bubble-sort|binary-search|linear-search|dijkstra|a-star|stack|queue|min-heap|avl-tree|trie|edit-distance|knuth-morris-pratt)['"`]/;
  const site = (await code()).filter((f) => f.includes('/apps/site/src/'));
  for (const f of site) {
    const t = await readFile(f, 'utf8');
    assert.ok(!ids.test(t), `${f.slice(web.length)} hard-codes an entry id`);
  }
});
