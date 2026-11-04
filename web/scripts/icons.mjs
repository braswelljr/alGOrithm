// Writes every brand file the site ships, all into public/. First the mark
// SVGs in public/logos/ (drawn from the concepts in
// src/components/brand/marks.ts); then every raster, the ICO and the Windows
// config, drawn from those SVG files as read back from disk, so the files
// in public/logos/ are the one source every icon is made from. The paths are
// in src/lib/brand.ts. Run it after changing either:
//
//   pnpm --dir web icons
//
// The outputs are committed, so a build needs no rasterizer. Social cards are
// not made here: next/og draws them at build time (src/lib/og.tsx).
import { execFileSync } from 'node:child_process';
import { mkdirSync, readFileSync, writeFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { crc32, deflateSync } from 'node:zlib';
import { Resvg } from '@resvg/resvg-js';

const site = join(dirname(fileURLToPath(import.meta.url)), '..');
const marks = await import(join(site, 'src/components/brand/marks.ts'));
const set = await import(join(site, 'src/lib/brand.ts'));

const { ACTIVE_CONCEPT: id, COLORS, PALETTE, mono, markSvg, squirclePath, toneCss } = marks;
const white = mono('#ffffff');
/** A deeper accent, behind the squircle where a platform needs no transparency. */
const accentDeep = '#3a3db0';

function write(file, data) {
  const path = join(site, file);
  mkdirSync(dirname(path), { recursive: true });
  writeFileSync(path, data);
}

// 1. The mark SVGs, icon only.
const L = set.LOGO;
write(L.mark.file, marks.faviconSvg(id));
write(L.markSmall.file, marks.faviconSvg(id, true));
write(L.mono.file, markSvg(id, mono('#000000')));
for (const [concept, f] of Object.entries(L.concepts)) write(f.file, marks.faviconSvg(concept));
write(set.safariPinnedTab.file, markSvg(id, mono('#000000')));
write(set.faviconSvg.file, marks.faviconSvg(id, true));
// The README shows its logo with Markdown, which cannot size an image, so the
// file says its own size: 96 px, rather than as wide as the page.
write(L.readme, marks.faviconSvg(id).replace('<svg ', '<svg width="96" height="96" '));

/** The drawing inside an SVG file in public/: its shapes, without the document or its style. */
function shapesOf(file) {
  const svg = readFileSync(join(site, file), 'utf8');
  return svg
    .replace(/^.*?<\/style>/s, '')
    .replace(/<rect width="32" height="32"[^>]*\/>/, '')
    .replace(/<\/svg>\s*$/, '');
}

const drawings = {
  full: shapesOf(L.mark.file),
  small: shapesOf(L.markSmall.file),
  ...Object.fromEntries(Object.entries(L.concepts).map(([c, f]) => [c, shapesOf(f.file)]))
};

/** The shapes of the mark for a size: the hinted file at 24 px and below. */
function drawing(concept, short) {
  if (concept !== id) return drawings[concept];
  return short <= 24 ? drawings.small : drawings.full;
}

// 2. Every raster, from those files.

/**
 * A document of width by height with the mark centered at `share` of the
 * short side, over an optional solid ground and an optional squircle plate
 * (`plate`, the share of the box it fills).
 */
function canvas(width, height, { ground, plate, palette, share, concept = id }) {
  const short = Math.min(width, height);
  const side = short * share;
  const x = (width - side) / 2;
  const y = (height - side) / 2;
  let bg = ground ? `<rect width="${width}" height="${height}" fill="${ground}"/>` : '';
  if (plate) {
    const pw = width * plate;
    const ph = height * plate;
    bg += `<path d="${squirclePath((width - pw) / 2, (height - ph) / 2, pw, ph)}" fill="${COLORS.accent}"/>`;
  }
  return `<svg xmlns="http://www.w3.org/2000/svg" width="${width}" height="${height}" viewBox="0 0 ${width} ${height}" shape-rendering="geometricPrecision"><style>${toneCss(palette)}</style>${bg}<g transform="translate(${x} ${y}) scale(${side / 32})">${drawing(concept, short)}</g></svg>`;
}

/** How each look is drawn (see Look in src/lib/brand.ts). */
function looks(icon, concept = id) {
  const { width: w, height: h } = icon;
  const splash = w > 512;
  const c = { concept };
  switch (icon.look) {
    case 'tab-light':
      return splash
        ? canvas(w, h, { ...c, ground: COLORS.lightBg, palette: PALETTE.light, share: 0.28 })
        : canvas(w, h, { ...c, palette: PALETTE.light, share: 1 });
    case 'tab-dark':
      return splash
        ? canvas(w, h, { ...c, ground: COLORS.darkBg, palette: PALETTE.dark, share: 0.28 })
        : canvas(w, h, { ...c, palette: PALETTE.dark, share: 1 });
    case 'tile':
      // A squircle plate on nothing: install icons, Windows tiles, shortcuts.
      return canvas(w, h, { ...c, plate: 1, palette: PALETTE.onAccent, share: 0.6 });
    case 'apple':
      // Apple wants no transparency and masks the corners itself: the squircle
      // fills the square, and the corners outside it, which iOS cuts away,
      // are a deeper accent rather than clear (iOS would paint them black).
      return canvas(w, h, {
        ...c,
        ground: accentDeep,
        plate: 1,
        palette: PALETTE.onAccent,
        share: 0.6
      });
    case 'maskable':
      // Full bleed, with the squircle and the mark inside the 80% safe zone,
      // so any mask a launcher applies keeps the whole plate's look.
      return canvas(w, h, {
        ...c,
        ground: accentDeep,
        plate: 0.8,
        palette: PALETTE.onAccent,
        share: 0.48
      });
    case 'mono':
      // A silhouette for themed icons: alpha is all that counts.
      return canvas(w, h, { ...c, palette: white, share: 0.62 });
    default:
      throw new Error(`unknown look ${icon.look}`);
  }
}

/** Straight-alpha RGBA from resvg's premultiplied pixels. */
function unpremultiply(px) {
  const out = Buffer.alloc(px.length);
  for (let i = 0; i < px.length; i += 4) {
    const a = px[i + 3];
    out[i + 3] = a;
    if (a === 0) continue;
    for (let k = 0; k < 3; k++) out[i + k] = Math.min(255, Math.round((px[i + k] * 255) / a));
  }
  return out;
}

/** A box filter from a k-times render down to width by height, in premultiplied space. */
function downsample(px, width, height, k) {
  const out = Buffer.alloc(width * height * 4);
  const big = width * k;
  for (let y = 0; y < height; y++)
    for (let x = 0; x < width; x++)
      for (let c = 0; c < 4; c++) {
        let sum = 0;
        for (let dy = 0; dy < k; dy++)
          for (let dx = 0; dx < k; dx++) sum += px[((y * k + dy) * big + x * k + dx) * 4 + c];
        out[(y * width + x) * 4 + c] = Math.round(sum / (k * k));
      }
  return out;
}

/** Edge energy: how sharp the edges are (higher is crisper), over the alpha channel. */
function sharpness(px, width, height) {
  let e = 0;
  for (let y = 0; y < height - 1; y++)
    for (let x = 0; x < width - 1; x++) {
      const i = (y * width + x) * 4 + 3;
      e += Math.abs(px[i] - px[i + 4]) + Math.abs(px[i] - px[i + width * 4]);
    }
  return e;
}

/** An 8-bit RGBA PNG, written by hand: IHDR, one IDAT, IEND. */
function encodePng(rgba, width, height) {
  const chunk = (type, data) => {
    const len = Buffer.alloc(4);
    len.writeUInt32BE(data.length);
    const td = Buffer.concat([Buffer.from(type, 'ascii'), data]);
    const crc = Buffer.alloc(4);
    crc.writeUInt32BE(crc32(td) >>> 0);
    return Buffer.concat([len, td, crc]);
  };
  const ihdr = Buffer.alloc(13);
  ihdr.writeUInt32BE(width, 0);
  ihdr.writeUInt32BE(height, 4);
  ihdr[8] = 8; // bit depth
  ihdr[9] = 6; // RGBA
  const rows = Buffer.alloc((width * 4 + 1) * height);
  for (let y = 0; y < height; y++) {
    rows[y * (width * 4 + 1)] = 0; // no filter; deflate does the work
    rgba.copy(rows, y * (width * 4 + 1) + 1, y * width * 4, (y + 1) * width * 4);
  }
  return Buffer.concat([
    Buffer.from([0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a]),
    chunk('IHDR', ihdr),
    chunk('IDAT', deflateSync(rows, { level: 9 })),
    chunk('IEND', Buffer.alloc(0))
  ]);
}

const stats = { direct: 0, supersampled: 0 };

/**
 * Renders svg at width by height two ways, directly (resvg's analytic
 * anti-aliasing) and at 4x then box-filtered down, and keeps the sharper.
 * Launch screens are large and flat, so they are rendered directly.
 */
function png(svg, width, height = width) {
  // .pixels is a getter that copies, so read it once.
  const direct = new Resvg(svg, { fitTo: { mode: 'width', value: width } }).render().pixels;
  let best = direct;
  if (width * height <= 512 * 512) {
    const big = new Resvg(svg, { fitTo: { mode: 'width', value: width * 4 } }).render();
    const down = downsample(big.pixels, width, height, 4);
    if (sharpness(down, width, height) > sharpness(best, width, height)) best = down;
  }
  stats[best === direct ? 'direct' : 'supersampled']++;
  return encodePng(unpremultiply(best), width, height);
}

/** An ICO holding PNG images, one per size (Vista and later read PNG entries). */
function ico(images) {
  const header = Buffer.alloc(6 + 16 * images.length);
  header.writeUInt16LE(0, 0);
  header.writeUInt16LE(1, 2);
  header.writeUInt16LE(images.length, 4);
  let offset = header.length;
  images.forEach(({ size, data }, i) => {
    const e = 6 + 16 * i;
    header.writeUInt8(size >= 256 ? 0 : size, e);
    header.writeUInt8(size >= 256 ? 0 : size, e + 1);
    header.writeUInt8(0, e + 2);
    header.writeUInt8(0, e + 3);
    header.writeUInt16LE(1, e + 4);
    header.writeUInt16LE(32, e + 6);
    header.writeUInt32LE(data.length, e + 8);
    header.writeUInt32LE(offset, e + 12);
    offset += data.length;
  });
  return Buffer.concat([header, ...images.map((i) => i.data)]);
}

/** The small glyphs on the shortcut tiles, on a 32 grid, drawn white. */
const glyphs = {
  learn:
    '<path d="M4 9C9 7 13 8 16 10C19 8 23 7 28 9V24C23 22 19 23 16 25C13 23 9 22 4 24Z" fill="none" stroke="#fff" stroke-width="2.4" stroke-linejoin="round"/><path d="M16 10V25" stroke="#fff" stroke-width="2.4"/>',
  quiz: '<path d="M11 11.5A5 5 0 1 1 18.5 15.8C16.8 16.8 16 17.8 16 19.8" fill="none" stroke="#fff" stroke-width="3.2" stroke-linecap="round"/><circle cx="16" cy="25" r="2.2" fill="#fff"/>',
  visualize:
    '<path d="M11 7 25 16 11 25Z" fill="#fff" stroke="#fff" stroke-width="2" stroke-linejoin="round"/>',
  'big-o':
    '<path d="M5 6V26H27" fill="none" stroke="#fff" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"/><path d="M7 25C15 25 20 20 25 7" fill="none" stroke="#fff" stroke-width="3" stroke-linecap="round"/>'
};

/** A shortcut: the glyph large on the accent tile, the mark small in a corner badge. */
function shortcutSvg(glyph, n) {
  return `<svg xmlns="http://www.w3.org/2000/svg" width="${n}" height="${n}" viewBox="0 0 96 96" shape-rendering="geometricPrecision"><path d="${squirclePath(0, 0, 96, 96)}" fill="${COLORS.accent}"/><g transform="translate(18 16) scale(1.6)">${glyph}</g><circle cx="74" cy="74" r="17" fill="${PALETTE.light.secondary}"/><style>${toneCss(white)}</style><g transform="translate(62 62) scale(0.75)">${drawings.full}</g></svg>`;
}

let count = 0;
for (const icon of set.allRasters) {
  const shortcut = set.shortcuts.find((s) => s.icon.file === icon.file);
  const svg = shortcut ? shortcutSvg(glyphs[shortcut.id], icon.width) : looks(icon);
  write(icon.file, png(svg, icon.width, icon.height));
  count++;
}

write(
  set.favicon.file,
  ico(
    set.favicon.sizes.map((size) => ({
      size,
      data: png(looks({ width: size, height: size, look: 'tab-light' }), size)
    }))
  )
);
for (const p of set.brandPreviews) write(p.file, png(looks(p, p.concept), 16));

const tile = (icon, tag) => `      <${tag} src="${icon.href}"/>`;
const [t70, t150, t310w, t310] = set.tiles;
write(
  set.browserconfig.file,
  `<?xml version="1.0" encoding="utf-8"?>
<browserconfig>
  <msapplication>
    <tile>
${tile(t70, 'square70x70logo')}
${tile(t150, 'square150x150logo')}
${tile(t310w, 'wide310x150logo')}
${tile(t310, 'square310x310logo')}
      <TileColor>${set.tileColor}</TileColor>
    </tile>
  </msapplication>
</browserconfig>
`
);

// Lossless recompression when oxipng is installed; skipped otherwise.
try {
  execFileSync('oxipng', [
    '-o',
    '4',
    '--strip',
    'safe',
    '-q',
    '-r',
    join(site, 'public'),
    join(site, 'src/app')
  ]);
  console.log('icons: oxipng recompressed the PNGs');
} catch {
  console.log('icons: oxipng not found, PNGs left as deflate -9');
}

console.log(
  `icons: ${stats.direct} rendered directly, ${stats.supersampled} supersampled 4x (the sharper kept)`
);
console.log(
  `icons: wrote ${count} PNGs, favicon.ico, the logos/ SVGs, icons/icon.svg and browserconfig.xml (mark: ${id})`
);
