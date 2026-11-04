/**
 * Runs saved to a file or kept in the run history. A saved run is read back
 * like anything else from outside: parsed with its schema, never trusted
 * (docs/rules/data-from-outside-is-unknown-until-checked.md).
 */
import { z } from 'zod';
import { savedRunSchema, type Input, type RunResult, type SavedRun } from './schema';

export function makeSavedRun(entry: string, input: Input, result?: RunResult): SavedRun {
  return {
    format: 'algorithm-run',
    version: 1,
    entry,
    input,
    savedAt: new Date().toISOString(),
    result
  };
}

/** Reads a saved run from JSON text: the run, or why it cannot be read. */
export function readSavedRun(
  text: string
): { ok: true; run: SavedRun } | { ok: false; error: string } {
  let raw: unknown;
  try {
    raw = JSON.parse(text);
  } catch {
    return { ok: false, error: 'The file is not JSON.' };
  }
  const parsed = savedRunSchema.safeParse(raw);
  return parsed.success
    ? { ok: true, run: parsed.data }
    : { ok: false, error: `The file is not a saved run: ${z.prettifyError(parsed.error)}` };
}

/** A CLI command and a Go snippet for a run, when the wasm build cannot say. */
export function fallbackSnippet(
  entry: { id: string; import: string; package: string; func: string },
  input: Input
): { cli: string; go: string } {
  const args: string[] = ['algorithm', 'run', entry.id];
  if (input.values?.length) args.push(...input.values.map(String));
  if (input.text) args.push(JSON.stringify(input.text));
  if (input.other) args.push(JSON.stringify(input.other));
  if (input.target !== undefined) args.push('--target', String(input.target));
  if (input.key) args.push('--key', JSON.stringify(input.key));
  if (input.shift !== undefined) args.push('--shift', String(input.shift));
  if (input.ops?.length) {
    args.push(
      '--ops',
      JSON.stringify(input.ops.map((o) => `${o.name}${o.value ? ` ${o.value}` : ''}`).join(', '))
    );
  }
  const pkg = entry.package || 'algorithm';
  const go = entry.import
    ? `import "${entry.import}"\n\n// ${entry.func ? `${pkg}.${entry.func}` : entry.id}: see the package documentation\n// for its exact signature.`
    : `// ${entry.id}`;
  return { cli: args.join(' '), go };
}
