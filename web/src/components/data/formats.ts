/** Reading files and addresses: the formats and the size limit. */

/** The most text read from a file or a URL: 10 MB. */
export const MAX_BYTES = 10 * 1024 * 1024;

export type Format = 'text' | 'csv' | 'json';

/** The format a file name or content type suggests. */
export function formatOf(name: string, type = ''): Format {
  const n = name.toLowerCase();
  if (n.endsWith('.csv') || type.includes('csv')) return 'csv';
  if (n.endsWith('.json') || type.includes('json')) return 'json';
  return 'text';
}
