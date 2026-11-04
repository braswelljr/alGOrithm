/**
 * Every icon file: the generated sources are there at their sizes, not blank,
 * transparent where they may be and opaque where a platform requires it; and
 * after a build, out/ serves them, the manifest names them, and the home page
 * links them. The out/ checks are skipped when the site is not built.
 */
import assert from 'node:assert/strict';
import { existsSync, readFileSync } from 'node:fs';
import path from 'node:path';
import { test } from 'node:test';
import { inflateSync } from 'node:zlib';
import {
  allRasters,
  appIcons,
  appleIcons,
  brandPreviews,
  darkTabIcons,
  favicon,
  faviconSvg,
  ICONS,
  LOGO,
  maskableIcons,
  monochromeIcons,
  safariPinnedTab,
  shortcuts,
  startupImages,
  tabIcons,
  type RasterIcon
} from '../src/lib/brand.ts';

const site = path.join(import.meta.dirname, '..');
const out = path.join(site, 'out');
const built = existsSync(path.join(out, 'index.html'));

/** Width, height and straight RGBA pixels of an 8-bit RGBA PNG. */
function decodePng(buf: Buffer) {
  assert.equal(buf.subarray(1, 4).toString('ascii'), 'PNG', 'a PNG signature');
  const width = buf.readUInt32BE(16);
  const height = buf.readUInt32BE(20);
  const depth = buf[24];
  const color = buf[25];
  const idat: Buffer[] = [];
  for (let o = 8; o < buf.length;) {
    const len = buf.readUInt32BE(o);
    const type = buf.subarray(o + 4, o + 8).toString('ascii');
    if (type === 'IDAT') idat.push(buf.subarray(o + 8, o + 8 + len));
    o += 12 + len;
  }
  assert.equal(depth, 8, '8 bits per channel');
  assert.equal(color, 6, 'RGBA');
  const raw = inflateSync(Buffer.concat(idat));
  const stride = width * 4;
  const px = Buffer.alloc(stride * height);
  for (let y = 0; y < height; y++) {
    const f = raw[y * (stride + 1)];
    for (let x = 0; x < stride; x++) {
      const v = raw[y * (stride + 1) + 1 + x];
      const a = x >= 4 ? px[y * stride + x - 4] : 0;
      const b = y > 0 ? px[(y - 1) * stride + x] : 0;
      const c = x >= 4 && y > 0 ? px[(y - 1) * stride + x - 4] : 0;
      const p = a + b - c;
      const pa = Math.abs(p - a);
      const pb = Math.abs(p - b);
      const pc = Math.abs(p - c);
      const paeth = pa <= pb && pa <= pc ? a : pb <= pc ? b : c;
      const pred = [0, a, b, (a + b) >> 1, paeth][f];
      px[y * stride + x] = (v + pred) & 0xff;
    }
  }
  return { width, height, px };
}

/** Whether a look keeps transparent corners, must be fully opaque, or neither. */
function expectation(i: RasterIcon): 'transparent' | 'opaque' {
  if (i.look === 'apple' || i.look === 'maskable' || i.width > 512) return 'opaque';
  return 'transparent';
}

function checkRaster(file: string, i: RasterIcon) {
  const { width, height, px } = decodePng(readFileSync(file));
  assert.deepEqual([width, height], [i.width, i.height], `${file} size`);
  let opaque = 0;
  let clear = 0;
  const colors = new Set<number>();
  for (let k = 0; k < px.length; k += 4) {
    if (px[k + 3] === 255) opaque++;
    if (px[k + 3] === 0) clear++;
    if (px[k + 3] > 0) colors.add(px.readUInt32BE(k));
  }
  const n = width * height;
  assert.ok(colors.size > 1, `${file} is not blank`);
  if (expectation(i) === 'opaque') assert.equal(opaque, n, `${file} is fully opaque`);
  else {
    assert.equal(px[3], 0, `${file} has a transparent corner`);
    assert.ok(clear > 0 && clear < n, `${file} is partly transparent`);
  }
}

test('every generated PNG is there, at its size, not blank, with the right transparency', () => {
  for (const i of [...allRasters, ...brandPreviews]) checkRaster(path.join(site, i.file), i);
});

/** The sizes inside an ICO. */
function icoSizes(buf: Buffer): number[] {
  assert.equal(buf.readUInt16LE(2), 1, 'an icon resource');
  return Array.from({ length: buf.readUInt16LE(4) }, (_, i) => buf[6 + 16 * i] || 256);
}

test('favicon.ico holds 16, 32 and 48', () => {
  assert.deepEqual(icoSizes(readFileSync(path.join(site, favicon.file))), [16, 32, 48]);
});

test('the SVG icons are small, precise and one file each', () => {
  for (const f of [
    faviconSvg.file,
    safariPinnedTab.file,
    LOGO.mark.file,
    LOGO.markSmall.file,
    LOGO.mono.file,
    ...Object.values(LOGO.concepts).map((c) => c.file)
  ]) {
    const svg = readFileSync(path.join(site, f), 'utf8');
    assert.ok(svg.length < 1100, `${f} is ${svg.length} bytes`);
    assert.match(svg, /shape-rendering="geometricPrecision"/, `${f} renders precisely`);
    assert.doesNotMatch(svg, /<metadata|<!--/, `${f} carries no metadata`);
  }
  assert.match(readFileSync(path.join(site, faviconSvg.file), 'utf8'), /prefers-color-scheme:dark/);
});

/** A file in out/ for a path the browser asks for. */
const served = (href: string) => path.join(out, href.split('?')[0]);

test('out/ serves every icon at its size', { skip: !built && 'site not built' }, () => {
  for (const i of [...allRasters, ...brandPreviews]) {
    const f = served(i.href);
    assert.ok(existsSync(f), `${i.href} in out/`);
    checkRaster(f, i);
  }
  assert.deepEqual(icoSizes(readFileSync(served('/favicon.ico'))), [16, 32, 48]);
  for (const f of [
    ICONS.svg,
    ICONS.safariPinnedTab,
    ICONS.browserconfig,
    LOGO.mark.href,
    LOGO.mono.href
  ])
    assert.ok(existsSync(served(f)), `${f} in out/`);
});

test(
  'the manifest has the required fields and real icons',
  { skip: !built && 'site not built' },
  () => {
    const m = JSON.parse(readFileSync(path.join(out, 'manifest.webmanifest'), 'utf8'));
    for (const k of [
      'id',
      'name',
      'short_name',
      'start_url',
      'scope',
      'display',
      'theme_color',
      'background_color',
      'icons'
    ])
      assert.ok(m[k], `manifest.${k}`);
    assert.equal(m.short_name, 'alGOrithm');
    assert.deepEqual(m.display_override, ['window-controls-overlay', 'standalone']);
    const by = (p: string) => m.icons.filter((i: { purpose: string }) => i.purpose === p).length;
    assert.equal(by('any'), appIcons.length);
    assert.equal(by('maskable'), maskableIcons.length);
    assert.equal(by('monochrome'), monochromeIcons.length);
    assert.ok(m.icons.some((i: { sizes: string }) => i.sizes === '192x192'));
    assert.ok(m.icons.some((i: { sizes: string }) => i.sizes === '512x512'));
    assert.equal(m.shortcuts.length, shortcuts.length);
    const images = [
      ...m.icons,
      ...m.shortcuts.flatMap((s: { icons: unknown[] }) => s.icons),
      ...(m.screenshots ?? [])
    ] as { src: string; sizes: string }[];
    for (const i of images) {
      const f = served(i.src);
      assert.ok(existsSync(f), `${i.src} in out/`);
      const buf = readFileSync(f);
      assert.equal(`${buf.readUInt32BE(16)}x${buf.readUInt32BE(20)}`, i.sizes, `${i.src} size`);
    }
  }
);

test('no brand file lives in src/app', () => {
  for (const f of [
    'favicon.ico',
    'icon.svg',
    'icon1.png',
    'apple-icon.png',
    'apple-icon1.png',
    'icon.tsx',
    'apple-icon.tsx'
  ])
    assert.ok(!existsSync(path.join(site, 'src/app', f)), `src/app/${f} should be in public/`);
});

test('the home page links every icon', { skip: !built && 'site not built' }, () => {
  const html = readFileSync(path.join(out, 'index.html'), 'utf8');
  const links = [...html.matchAll(/<link [^>]*>/g)].map((m) => m[0]);
  const esc = (v: string) => v.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
  /** The links with this rel and href, and every attribute given. */
  const find = (rel: string, href: string, attrs: Record<string, string> = {}) =>
    links.filter(
      (l) =>
        l.includes(`rel="${rel}"`) &&
        new RegExp(`href="${esc(href)}"`).test(l) &&
        Object.entries(attrs).every(([k, v]) => l.includes(`${k}="${v}"`))
    );
  const one = (rel: string, href: string, attrs?: Record<string, string>) =>
    assert.equal(find(rel, href, attrs).length, 1, `${rel} ${href} ${JSON.stringify(attrs ?? {})}`);
  one('icon', favicon.href, { sizes: '16x16 32x32 48x48' });
  one('icon', faviconSvg.href, { type: 'image/svg+xml' });
  for (const i of tabIcons)
    one('icon', i.href, { sizes: `${i.width}x${i.height}`, type: 'image/png' });
  for (const i of darkTabIcons)
    one('icon', i.href, { sizes: `${i.width}x${i.height}`, media: '(prefers-color-scheme: dark)' });
  for (const i of appleIcons) one('apple-touch-icon', i.href, { sizes: `${i.width}x${i.height}` });
  one('mask-icon', safariPinnedTab.href, { color: '#4b4fd8' });
  one('manifest', '/manifest.webmanifest');
  assert.equal(
    links.filter((l) => l.includes('rel="apple-touch-startup-image"')).length,
    startupImages.length,
    'launch screens'
  );
  assert.match(html, /name="msapplication-config" content="\/browserconfig\.xml"/);
  assert.match(html, /name="theme-color" content="#12141c" media="\(prefers-color-scheme: dark\)"/);
  // Every linked file exists.
  for (const l of links) {
    const href = /href="([^"]+)"/.exec(l)?.[1];
    if (href?.startsWith('/') && /\.(png|svg|ico)/.test(href))
      assert.ok(existsSync(served(href)), `${href} in out/`);
  }
});
