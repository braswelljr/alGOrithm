/**
 * Opening the editor from another page (an entry's docs, the visualizer, the
 * structures playground) with something to try: a program to write into the
 * visitor's home folder, and a library file to open beside it. Kept in
 * sessionStorage for the one navigation, so nothing goes in the address.
 */
const KEY = 'algorithm:editor-handoff';

export type Handoff = {
  /** The program's name: it is written to ~/<name>/main.go. */
  name: string;
  /** The program, a Go main package that uses the entry. */
  main: string;
  /** A file of the library to open too, relative to pkg/ (algorithms/sorting/quick.go). */
  file?: string;
};

export function handOff(h: Handoff): void {
  try {
    sessionStorage.setItem(KEY, JSON.stringify(h));
  } catch {
    // Blocked storage: the editor opens without the program.
  }
}

/** The waiting handoff, taken once. */
export function takeHandoff(): Handoff | undefined {
  try {
    const raw = sessionStorage.getItem(KEY);
    sessionStorage.removeItem(KEY);
    if (!raw) return undefined;
    const v: unknown = JSON.parse(raw);
    if (!v || typeof v !== 'object') return undefined;
    const { name, main, file } = v as Record<string, unknown>;
    if (typeof name !== 'string' || typeof main !== 'string' || !/^[\w.-]{1,60}$/.test(name))
      return undefined;
    return { name, main, file: typeof file === 'string' ? file : undefined };
  } catch {
    return undefined;
  }
}
