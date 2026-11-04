import { existsSync, readdirSync } from 'node:fs';
import path from 'node:path';
import type { MetadataRoute } from 'next';
import { absoluteUrl } from '@/config/site';
import { buildCatalog, buildPaths } from '@/lib/catalog.server';
import { docsHref } from '@/lib/entries';
import { pathHref } from '@/lib/learn';

export const dynamic = 'force-static';

/** The pages every visitor can arrive at, plus one per entry and per path. */
const pages: { path: string; priority: number }[] = [
  { path: '/', priority: 1 },
  { path: '/learn/', priority: 0.9 },
  { path: '/docs/', priority: 0.9 },
  { path: '/visualize/', priority: 0.8 },
  { path: '/structures/', priority: 0.8 },
  { path: '/big-o/', priority: 0.8 },
  { path: '/system-design/', priority: 0.8 },
  { path: '/quiz/', priority: 0.7 },
  { path: '/editor/', priority: 0.7 },
  { path: '/compare/', priority: 0.6 },
  { path: '/data/', priority: 0.6 },
  { path: '/docs/tags/', priority: 0.5 },
  { path: '/cheatsheets/', priority: 0.7 },
  { path: '/shortcuts/', priority: 0.4 }
];

const app = path.join(process.cwd(), 'src/app');
/** The route groups, (site) and (docs): folders that add nothing to the address. */
const groups = readdirSync(app).filter((d) => /^\(.+\)$/.test(d));

/**
 * Whether the site has a page at route, so a page still being built is left
 * out. A page sits in src/app or inside one of its route groups; looking in
 * src/app alone left every page but the entries out of the sitemap.
 */
function exists(route: string): boolean {
  return ['', ...groups].some((g) => existsSync(path.join(app, g, route, 'page.tsx')));
}

export default async function sitemap(): Promise<MetadataRoute.Sitemap> {
  const [catalog, paths] = await Promise.all([buildCatalog(), buildPaths()]);
  const lastModified = new Date();
  return [
    ...pages
      .filter((p) => exists(p.path))
      .map((p) => ({ url: absoluteUrl(p.path), lastModified, priority: p.priority })),
    ...paths.map((p) => ({
      url: absoluteUrl(pathHref(p.id)),
      lastModified,
      priority: 0.7
    })),
    ...catalog.entries.map((e) => ({
      url: absoluteUrl(docsHref(e.id)),
      lastModified,
      priority: 0.6
    }))
  ];
}
