// Fails the build when a page gets heavier than its budget. It reads the
// static export in out/: for each route, the gzip size of the scripts and
// stylesheets its HTML loads (nomodule polyfills excluded, since modern
// browsers skip them), plus the wasm (brotli), the fonts and the logo mark.
// Run after `next build`: `node scripts/budget.mjs`. The numbers and why
// they are what they are: docs/workflows/performance.md.
import fs from 'node:fs';
import path from 'node:path';
import zlib from 'node:zlib';

const out = path.resolve(process.argv[2] ?? 'out');
const KB = 1024;

// First-load JS per route, gzip. React 19 and the Next.js runtime alone are
// about 125 KB; a page that lists entries adds its filters, a page that runs
// entries adds the visualizer and its views.
const routes = {
  '': 150, // home: the catalog from the build, no wasm
  learn: 150,
  docs: 150,
  'docs/tags': 150,
  'docs/quick-sort': 260, // an entry: the run view loads on demand
  visualize: 260,
  structures: 260,
  compare: 260,
  quiz: 300, // the game: motion and gsap
  'big-o': 260
};
const limits = {
  cssGzip: 50 * KB,
  wasmBrotli: 3.6 * 1024 * KB, // raised on purpose as entries land
  wasmGzip: 5.0 * 1024 * KB,
  fontsTotal: 100 * KB, // the text font and the mono font, latin woff2 (77 KB now)
  markSvg: 1 * KB
};

const gzip = (b) => zlib.gzipSync(b, { level: 9 }).length;
const brotli = (b) =>
  zlib.brotliCompressSync(b, { params: { [zlib.constants.BROTLI_PARAM_QUALITY]: 11 } }).length;
const file = (src) => path.join(out, decodeURIComponent(src.split('?')[0]));
const fmt = (n) => `${(n / KB).toFixed(1)} KB`;

const failures = [];
const rows = [];
const cssSeen = new Map();

for (const [route, budgetKB] of Object.entries(routes)) {
  const html = path.join(out, route, 'index.html');
  if (!fs.existsSync(html)) {
    rows.push([route || '/', 'missing', '']);
    continue;
  }
  const page = fs.readFileSync(html, 'utf8');
  const scripts = new Set();
  for (const m of page.matchAll(/<script\b([^>]*)>/g)) {
    const attrs = m[1];
    const src = /\bsrc="([^"]+)"/.exec(attrs)?.[1];
    if (src && !/\bnoModule\b|\bnomodule\b/i.test(attrs)) scripts.add(src);
  }
  let js = 0;
  for (const s of scripts) if (fs.existsSync(file(s))) js += gzip(fs.readFileSync(file(s)));
  for (const m of page.matchAll(/<link\b[^>]*rel="stylesheet"[^>]*href="([^"]+)"/g)) {
    const f = file(m[1]);
    if (fs.existsSync(f)) cssSeen.set(m[1], gzip(fs.readFileSync(f)));
  }
  rows.push([route || '/', fmt(js), `${budgetKB} KB`]);
  if (js > budgetKB * KB) failures.push(`/${route}: first-load JS ${fmt(js)} > ${budgetKB} KB`);
}

const cssMax = Math.max(0, ...cssSeen.values());
rows.push(['CSS (largest sheet)', fmt(cssMax), fmt(limits.cssGzip)]);
if (cssMax > limits.cssGzip) failures.push(`CSS ${fmt(cssMax)} > ${fmt(limits.cssGzip)}`);

const wasm = path.join(out, 'algorithm.wasm');
if (fs.existsSync(wasm)) {
  const b = fs.readFileSync(wasm);
  const br = brotli(b);
  const gz = gzip(b);
  rows.push(['wasm raw', fmt(b.length), '']);
  rows.push(['wasm gzip', fmt(gz), fmt(limits.wasmGzip)]);
  rows.push(['wasm brotli', fmt(br), fmt(limits.wasmBrotli)]);
  if (br > limits.wasmBrotli) failures.push(`wasm brotli ${fmt(br)} > ${fmt(limits.wasmBrotli)}`);
  if (gz > limits.wasmGzip) failures.push(`wasm gzip ${fmt(gz)} > ${fmt(limits.wasmGzip)}`);
}

// Fonts: only the faces a latin page downloads. next/font writes a face per
// unicode subset (cyrillic, greek, vietnamese, ...); the browser fetches a
// face only when the page has characters in its unicode-range, so the budget
// counts the faces whose range covers basic latin (U+0-FF, which minified CSS writes U+??), or that
// have no range at all.
const fonts = new Map();
const css = (dir) => {
  for (const e of fs.readdirSync(dir, { withFileTypes: true })) {
    const p = path.join(dir, e.name);
    if (e.isDirectory()) css(p);
    else if (e.name.endsWith('.css')) {
      const text = fs.readFileSync(p, 'utf8');
      for (const face of text.match(/@font-face\s*{[^}]*}/g) ?? []) {
        const range = /unicode-range:([^;}]*)/.exec(face)?.[1] ?? '';
        if (range && !/U\+(0+-0*FF\b|\?\?(?![?0-9A-F]))/i.test(range)) continue;
        const url = /url\(([^)]+\.woff2)\)/.exec(face)?.[1]?.replace(/["']/g, '');
        if (!url) continue;
        const f = url.startsWith('/') ? path.join(out, url) : path.resolve(path.dirname(p), url);
        if (fs.existsSync(f)) fonts.set(f, fs.statSync(f).size);
      }
    }
  }
};
css(path.join(out, '_next'));
const fontTotal = [...fonts.values()].reduce((a, b) => a + b, 0);
rows.push([`fonts (${fonts.size} latin woff2)`, fmt(fontTotal), fmt(limits.fontsTotal)]);
if (fontTotal > limits.fontsTotal)
  failures.push(`fonts ${fmt(fontTotal)} > ${fmt(limits.fontsTotal)}`);

for (const mark of ['logos/mark.svg', 'icon.svg']) {
  const f = path.join(out, mark);
  if (!fs.existsSync(f)) continue;
  const n = fs.statSync(f).size;
  rows.push([mark, `${n} B`, `${limits.markSvg} B`]);
  if (n > limits.markSvg) failures.push(`${mark} ${n} B > ${limits.markSvg} B`);
  break;
}

const w = Math.max(...rows.map((r) => r[0].length));
for (const [a, b, c] of rows)
  console.log(`${a.padEnd(w)}  ${b.padStart(10)}  ${c && `budget ${c}`}`);
if (failures.length) {
  console.error(`\nover budget:\n  ${failures.join('\n  ')}`);
  process.exit(1);
}
console.log('\nall within budget');
