/**
 * What the command palette searches, as one small file: every entry, learning
 * path, category and tag, with only what a row and its preview show.
 *
 * Built from the catalog at build time and served as `/search.json`
 * (src/app/search.json/route.ts), so the palette finds everything without
 * loading the wasm. The palette fetches it the first time it opens, the way
 * gsl-special-courses loads its page content.
 */

import { z } from 'zod';

const text = z.string().catch('');
const count = z.number().int().min(0).catch(0);

export const searchEntrySchema = z.object({
  id: z.string().min(1),
  name: z.string(),
  category: text,
  categoryName: text,
  topic: text,
  summary: text,
  level: text,
  kind: text,
  time: text,
  space: text,
  input: text,
  tags: z.array(z.string()).catch([])
});
const pathSchema = z.object({
  id: z.string().min(1),
  title: z.string(),
  summary: text,
  level: text,
  steps: count
});
const categorySchema = z.object({
  id: z.string().min(1),
  name: z.string(),
  topic: text,
  summary: text,
  entries: count
});
const tagSchema = z.object({ id: z.string().min(1), meaning: text, entries: count });

/** The index as the browser receives it: checked before it is searched
 *  (docs/rules/data-from-outside-is-unknown-until-checked.md). */
export const searchIndexSchema = z.object({
  entries: z.array(searchEntrySchema),
  paths: z.array(pathSchema).catch([]),
  categories: z.array(categorySchema).catch([]),
  tags: z.array(tagSchema).catch([])
});

export type SearchEntry = z.infer<typeof searchEntrySchema>;
export type SearchPath = z.infer<typeof pathSchema>;
export type SearchCategory = z.infer<typeof categorySchema>;
export type SearchTag = z.infer<typeof tagSchema>;
export type SearchIndex = z.infer<typeof searchIndexSchema>;

/** The address of the index, under the site's base path. */
export const searchIndexUrl = (basePath: string) => `${basePath}/search.json`;
