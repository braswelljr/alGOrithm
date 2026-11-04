import type { Metadata } from 'next';
import { siteName } from '../config/site.ts';

/** The social card of any page without its own (out/og.png). */
export const ogDefault = '/og.png';

/**
 * A page's metadata: its title, description, its one canonical URL, and the
 * social cards that repeat them. path is the page's own path ("/compare/"),
 * resolved against the root layout's metadataBase.
 */
export function pageMetadata({
  title,
  description,
  path,
  images
}: {
  title: string;
  description: string;
  path: string;
  images?: string[];
}): Metadata {
  const cards = images ?? [ogDefault];
  return {
    title,
    description,
    alternates: { canonical: path },
    openGraph: {
      type: 'website',
      siteName,
      title: `${title} · ${siteName}`,
      description,
      url: path,
      images: cards.map((url) => ({ url, width: 1200, height: 630, alt: title }))
    },
    twitter: {
      card: 'summary_large_image',
      title: `${title} · ${siteName}`,
      description,
      images: cards
    }
  };
}
