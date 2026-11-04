import { createContext, type Context } from 'react';

/**
 * A React context that survives hot module replacement.
 *
 * `createContext` at module scope makes a new context each time the module is
 * re-evaluated, so after an HMR update a provider from the old module and a
 * consumer from the new one no longer match and the consumer reads the default.
 * Keying the context in a registry on `globalThis` hands every evaluation the
 * same object.
 */
const REGISTRY = Symbol.for('gsl.react.contexts');

type Registry = Map<string, Context<unknown>>;

export function stableContext<T>(key: string, defaultValue: T): Context<T> {
  const host = globalThis as typeof globalThis & { [REGISTRY]?: Registry };
  const registry = (host[REGISTRY] ??= new Map());

  const existing = registry.get(key);
  if (existing) return existing as Context<T>;

  const created = createContext<T>(defaultValue);
  registry.set(key, created as Context<unknown>);
  return created;
}
