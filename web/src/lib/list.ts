/**
 * Formats an array of strings into a localized human-readable list
 * using `Intl.ListFormat`.
 *
 * Example:
 *   formatter(["apples", "bananas", "oranges"])
 *   → "apples, bananas, and oranges"
 *
 * @param {string[]} list
 *   The list of strings to format.
 *
 * @param {Intl.ListFormatOptions & { locales?: Intl.LocalesArgument }} [options]
 *   Options for configuring the list formatter.
 *   - `locales`: Overrides the locale (default: "en")
 *   - `style`: "long" | "short" | "narrow" (default: "short")
 *   - `type`: "conjunction" | "disjunction" | "unit" (default: "conjunction")
 *
 * @returns {string}
 *   The formatted list string.
 */
export function formatter(
  list: string[],
  options: Intl.ListFormatOptions & { locales?: string | string[] } = {
    locales: 'en',
    style: 'short',
    type: 'conjunction'
  }
): string {
  const { locales = 'en', ...formatterOptions } = options;

  const listFormatter = new Intl.ListFormat(locales, formatterOptions);

  return listFormatter.format(list);
}

/**
 * The array at `value`, or an empty one.
 *
 * `Array.isArray(value) ? value : []`, named, for an expression that must not
 * be evaluated twice, such as `table.getHeaderGroups()`. `?? []` is not the same
 * check: it lets a non-array through.
 */
export function asArray<T>(value: readonly T[] | null | undefined): T[];
export function asArray<T>(value: unknown): T[];
export function asArray<T>(value: unknown): T[] {
  return Array.isArray(value) ? (value as T[]) : [];
}
