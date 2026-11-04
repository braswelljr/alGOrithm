import { buildCatalog } from '@/lib/catalog.server';
import { categoryTitle } from '@/lib/entries';
import { ogImage } from '@/lib/og';

// One card per entry, written at build time as out/docs/<id>/og.png.
export const dynamic = 'force-static';
export const dynamicParams = false;

export async function generateStaticParams() {
  return (await buildCatalog()).entries.map((e) => ({ id: e.id }));
}

export async function GET(_req: Request, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const catalog = await buildCatalog();
  const entry = catalog.entries.find((e) => e.id === id);
  if (!entry) return new Response('no such entry', { status: 404 });
  const c = catalog.categories.find((x) => x.id === entry.category);
  return ogImage({
    eyebrow: `${c ? categoryTitle(c) : entry.category} · time ${entry.time}`,
    title: entry.name,
    subtitle: entry.summary
  });
}
