import { ogImage } from '@/lib/og';
import { siteDescription, siteTagline } from '@/config/site';

// Written once at build time as out/og.png, the card every page shares by default.
export const dynamic = 'force-static';

export function GET() {
  return ogImage({ title: siteTagline, subtitle: siteDescription });
}
