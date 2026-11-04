/**
 * Writes src/styles/colors.css from src/lib/colors.ts, the one source of the
 * color system. Run it after changing a color: `pnpm --dir web colors`.
 * test/colors.test.ts fails when the file is stale.
 */
import { writeFile } from 'node:fs/promises';
import { fileURLToPath } from 'node:url';
import { renderColorsCss } from './render-colors.ts';

const out = fileURLToPath(new URL('../src/styles/colors.css', import.meta.url));
await writeFile(out, renderColorsCss());
console.log(`wrote ${out}`);
