/**
 * Guards for values that crossed a boundary: the WebAssembly program's API, a
 * URL parameter, anything read off `globalThis`. Such a value is `unknown`
 * until one of these has narrowed it; a cast would only tell the compiler to
 * stop asking.
 */

/** A plain object: not null, not an array. */
export function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value);
}

export function isString(value: unknown): value is string {
  return typeof value === 'string';
}

/** A real number. `typeof NaN` is "number", which is why this is not `typeof`. */
export function isNumber(value: unknown): value is number {
  return typeof value === 'number' && !Number.isNaN(value);
}

export function isFunction(value: unknown): value is (...args: unknown[]) => unknown {
  return typeof value === 'function';
}

/** Reads a property without vouching for its type, so the next check stays honest. */
export function prop(value: unknown, key: string): unknown {
  return isRecord(value) ? value[key] : undefined;
}

/** The elements of `value` that pass `guard`; an empty array if it is not an array at all. */
export function arrayOf<T>(value: unknown, guard: (item: unknown) => item is T): T[] {
  return Array.isArray(value) ? value.filter(guard) : [];
}

/** The docs-versions route's answer: each version's page paths. */
export function isVersionPaths(value: unknown): value is Record<string, string[]> {
  return (
    isRecord(value) &&
    Object.values(value).every((paths) => Array.isArray(paths) && paths.every(isString))
  );
}
