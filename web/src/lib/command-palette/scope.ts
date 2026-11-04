/**
 * Scoped search in the command palette: a character typed first narrows what
 * is searched, the way a mail client's "from:" does (after
 * gsl-special-courses).
 *
 *   #  tags: the vocabulary, and the entries that carry each tag
 *   ~  learning paths
 *   >  commands: actions, settings, appearance
 *   /  pages: everywhere on the site
 *
 * Entries (algorithms, structures and simulations) are searched in every
 * scope but the command and page ones.
 */

export type SearchKind = 'entries' | 'paths' | 'categories' | 'tags';

export type PaletteScope = {
  /** The character that opens it. */
  prefix: '#' | '~' | '>' | '/';
  /** What the chip in the input says. */
  label: string;
  /** What the hint beside the prefix says. */
  hint: string;
  /** The kinds of record searched; empty for a scope of fixed rows only. */
  kinds: readonly SearchKind[];
  /** Which fixed groups stay: navigation and pages, or the commands. */
  fixed: 'none' | 'commands' | 'pages';
};

export const SCOPES: readonly PaletteScope[] = [
  { prefix: '#', label: 'Tags', hint: 'tags', kinds: ['tags'], fixed: 'none' },
  { prefix: '~', label: 'Paths', hint: 'learning paths', kinds: ['paths'], fixed: 'none' },
  { prefix: '>', label: 'Commands', hint: 'commands', kinds: [], fixed: 'commands' },
  { prefix: '/', label: 'Pages', hint: 'pages', kinds: [], fixed: 'pages' }
];

/**
 * The scope a typed value opens, and what remains of the value once the prefix
 * is taken off, or null when it opens none.
 */
export function openScope(
  value: string,
  available: readonly PaletteScope[]
): { scope: PaletteScope; rest: string } | null {
  const scope = available.find((s) => value.startsWith(s.prefix));
  if (!scope) return null;
  return { scope, rest: value.slice(scope.prefix.length).replace(/^\s+/, '') };
}
