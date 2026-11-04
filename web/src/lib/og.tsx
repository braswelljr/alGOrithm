/**
 * The social cards, drawn at build time by next/og from the shared OgCard.
 * They are PNG route handlers (og.png, docs/[id]/og.png) rather than the
 * opengraph-image convention: under `output: 'export'` that convention writes
 * files with no extension, which static hosts may serve without an image
 * content type, and crawlers then drop the card.
 */
import { readFileSync } from 'node:fs';
import path from 'node:path';
import { inflateSync } from 'node:zlib';
import { ImageResponse } from 'next/og';
import { OgCard } from '@/components/shared/og-card';
import { siteUrl } from '../config/site.ts';

export const ogSize = { width: 1200, height: 630 };

/**
 * A WOFF 1.0 font as the plain TrueType (sfnt) it wraps: the same tables,
 * each inflated with zlib where it was compressed, behind a rebuilt header.
 * The renderer gets TrueType, which it reads directly; handing it the WOFF
 * left next/og's build-time renderer waiting forever inside a build worker.
 */
export function woffToSfnt(woff: Buffer): Buffer {
  if (woff.readUInt32BE(0) !== 0x774f4646) throw new Error('not a WOFF 1.0 font');
  const flavor = woff.readUInt32BE(4);
  const count = woff.readUInt16BE(12);
  const tables: { tag: number; checksum: number; data: Buffer }[] = [];
  for (let i = 0; i < count; i++) {
    const at = 44 + i * 20;
    const tag = woff.readUInt32BE(at);
    const offset = woff.readUInt32BE(at + 4);
    const compressed = woff.readUInt32BE(at + 8);
    const original = woff.readUInt32BE(at + 12);
    const checksum = woff.readUInt32BE(at + 16);
    const raw = woff.subarray(offset, offset + compressed);
    tables.push({
      tag,
      checksum,
      data: compressed < original ? inflateSync(raw) : Buffer.from(raw)
    });
  }
  // The sfnt header's search fields, from the table count.
  let power = 1;
  let log = 0;
  while (power * 2 <= count) {
    power *= 2;
    log++;
  }
  const header = Buffer.alloc(12 + count * 16);
  header.writeUInt32BE(flavor, 0);
  header.writeUInt16BE(count, 4);
  header.writeUInt16BE(power * 16, 6);
  header.writeUInt16BE(log, 8);
  header.writeUInt16BE(count * 16 - power * 16, 10);
  const bodies: Buffer[] = [];
  let offset = header.length;
  tables.forEach((t, i) => {
    const at = 12 + i * 16;
    header.writeUInt32BE(t.tag, at);
    header.writeUInt32BE(t.checksum, at + 4);
    header.writeUInt32BE(offset, at + 8);
    header.writeUInt32BE(t.data.length, at + 12);
    const padded = Buffer.alloc((t.data.length + 3) & ~3);
    t.data.copy(padded);
    bodies.push(padded);
    offset += padded.length;
  });
  return Buffer.concat([header, ...bodies]);
}

/**
 * The site's faces for the cards: Roboto Serif for the title and Source Sans 3
 * for the rest, as on the pages, from @fontsource's WOFF files turned back
 * into TrueType (woffToSfnt), once per build.
 */
const fontDir = (pkg: string) => path.join(process.cwd(), 'node_modules/@fontsource', pkg, 'files');
const fonts = [
  { name: 'Roboto Serif', pkg: 'roboto-serif', weight: 700 as const },
  { name: 'Source Sans 3', pkg: 'source-sans-3', weight: 400 as const },
  { name: 'Source Sans 3', pkg: 'source-sans-3', weight: 600 as const }
].map((f) => ({
  name: f.name,
  weight: f.weight,
  style: 'normal' as const,
  data: woffToSfnt(
    readFileSync(path.join(fontDir(f.pkg), `${f.pkg}-latin-${f.weight}-normal.woff`))
  )
}));

const host = siteUrl.replace(/^https?:\/\//, '');

export function ogImage({
  eyebrow = 'Algorithms, step by step',
  title,
  subtitle,
  size = ogSize
}: {
  eyebrow?: string;
  title: string;
  subtitle: string;
  size?: { width: number; height: number };
}): ImageResponse {
  return new ImageResponse(
    <OgCard
      title={title}
      description={subtitle}
      eyebrow={eyebrow}
      site={host}
    />,
    { ...size, fonts }
  );
}
