import { NAVIGATION, normalizePath } from '../config/navigation.ts';

/** One step of the trail: what it is called, and where it leads, if anywhere. */
export type Crumb = { label: string; href?: string };

/** Pages below the menu's own, by path. */
const EXTRA_LABELS: Record<string, string> = {
  '/docs/entry/': 'Entry',
  '/docs/tags/': 'Tags',
  '/learn/path/': 'Path'
};

/** A path segment written for a person: "my-page" reads "My page". */
function humanize(segment: string): string {
  const words = decodeURIComponent(segment).replace(/[-_]+/g, ' ').trim();
  return words ? words[0].toUpperCase() + words.slice(1) : segment;
}

/**
 * The trail to `pathname`: one crumb per folder, each a link but the last.
 * The menu names its pages; a record beneath one (`/docs/my-page/`) is named
 * from its path. The home page and the menu's top-level pages have no trail.
 */
export function crumbsFor(pathname: string): Crumb[] {
  const labels = new Map<string, string>(
    NAVIGATION.flatMap((s) => s.items.map((i) => [i.href, i.label] as const))
  );
  const segments = normalizePath(pathname).split('/').filter(Boolean);
  if (segments.length < 2) return [];
  const crumbs: Crumb[] = [];
  let href = '/';
  for (const segment of segments) {
    href += `${segment}/`;
    crumbs.push({ label: labels.get(href) ?? EXTRA_LABELS[href] ?? humanize(segment), href });
  }
  delete crumbs[crumbs.length - 1].href;
  return crumbs;
}
