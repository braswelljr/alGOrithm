/**
 * The social cards, drawn at build time by next/og from the shared OgCard.
 * They are PNG route handlers (og.png, docs/[id]/og.png) rather than the
 * opengraph-image convention: under `output: 'export'` that convention writes
 * files with no extension, which static hosts may serve without an image
 * content type, and crawlers then drop the card.
 */
import { readFileSync } from 'node:fs';
import path from 'node:path';
import { ImageResponse } from 'next/og';
import { OgCard } from '@/components/shared/og-card';
import { siteUrl } from '../config/site.ts';

export const ogSize = { width: 1200, height: 630 };

/**
 * The site's text face, Ysabeau Office, from its OFL .ttf files in
 * src/assets/fonts (Satori needs TrueType, not the WOFF2 next/font serves).
 * Read once per build.
 */
const fontDir = path.join(process.cwd(), 'src/assets/fonts');
const fonts = [
  { name: 'Ysabeau Office', weight: 500 as const, file: 'YsabeauOffice-Medium.ttf' },
  { name: 'Ysabeau Office', weight: 700 as const, file: 'YsabeauOffice-Bold.ttf' }
].map((f) => ({
  name: f.name,
  weight: f.weight,
  style: 'normal' as const,
  data: readFileSync(path.join(fontDir, f.file))
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
