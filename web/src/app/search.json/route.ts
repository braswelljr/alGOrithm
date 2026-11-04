import { buildSearchIndex } from '@/lib/catalog.server';

/** The command palette's search index, written once into the static export. */
export const dynamic = 'force-static';

export async function GET(): Promise<Response> {
  return Response.json(await buildSearchIndex());
}
