import { absoluteUrl, siteDescription, siteName, siteTagline } from '@/config/site';
import { buildCatalogSummary, buildPaths } from '@/lib/catalog.server';
import { categoriesWithEntries, categoryTitle, docsHref } from '@/lib/entries';
import { pathHref } from '@/lib/learn';

export const dynamic = 'force-static';

/**
 * /llms.txt (llmstxt.org): the site in plain Markdown for language models and
 * the tools built on them: what it is, its sections, every learning path, and
 * every entry's documentation with a line about it. Generated from the
 * catalog, like the sitemap, so it cannot list what the code does not have.
 */
export async function GET() {
  const [catalog, paths] = await Promise.all([buildCatalogSummary(), buildPaths()]);
  const lines: string[] = [
    `# ${siteName}`,
    '',
    `> ${siteTagline}. ${siteDescription}`,
    '',
    `${catalog.entries.length} entries, each the real Go code (github.com/braswelljr/alGOrithm, standard library only) compiled to WebAssembly and run in the browser; nothing a visitor loads leaves it.`,
    '',
    '## Sections',
    '',
    `- [Catalog](${absoluteUrl('/')}): every entry, filtered by kind, level and tag`,
    `- [Visualize](${absoluteUrl('/visualize/')}): run an entry on your own data and step through it`,
    `- [Docs](${absoluteUrl('/docs/')}): what each entry does, what it costs, its source and how to call it`,
    `- [Learn](${absoluteUrl('/learn/')}): ordered learning paths`,
    `- [Quiz](${absoluteUrl('/quiz/')}): predict the next step of a real run`,
    `- [Editor](${absoluteUrl('/editor/')}): edit the algorithms and your own Go, shell and Makefiles, and run them in a terminal with the gnux tools`,
    `- [Big O](${absoluteUrl('/big-o/')}): measure an entry's cost on growing inputs`,
    `- [System design](${absoluteUrl('/system-design/')}): case studies with live architectures`,
    `- [Cheat sheets](${absoluteUrl('/cheatsheets/')}): every cost and key idea on one table`,
    '',
    '## Learning paths',
    '',
    ...paths.map((p) => `- [${p.title}](${absoluteUrl(pathHref(p.id))}): ${p.summary}`)
  ];
  for (const { category, entries } of categoriesWithEntries(catalog)) {
    lines.push('', `## ${categoryTitle(category)}`, '');
    for (const e of entries) {
      lines.push(
        `- [${e.name}](${absoluteUrl(docsHref(e.id))}): ${e.summary} Time ${e.time}, space ${e.space}.`
      );
    }
  }
  return new Response(`${lines.join('\n')}\n`, {
    headers: { 'Content-Type': 'text/markdown; charset=utf-8' }
  });
}
