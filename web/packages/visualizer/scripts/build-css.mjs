// Writes dist/styles.css: React Flow's base styles, the package's own, then the
// views' Tailwind utilities compiled from src/tailwind.css (prefixed `av:`, no
// Preflight). One file to import, and no Tailwind needed by consumers.
import { execFileSync } from 'node:child_process';
import { mkdir, readFile, rename, rm, writeFile } from 'node:fs/promises';
import { createRequire } from 'node:module';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

const here = dirname(fileURLToPath(import.meta.url));
const require = createRequire(import.meta.url);
const flow = await readFile(require.resolve('@xyflow/react/dist/base.css'), 'utf8');
const own = await readFile(join(here, '../src/styles.css'), 'utf8');
await mkdir(join(here, '../dist'), { recursive: true });
const twOut = join(here, '../dist/.tailwind.css');
// The CLI's own entry, from its package.json "bin", run with this Node.
const cliPkg = require.resolve('@tailwindcss/cli/package.json');
const { bin } = JSON.parse(await readFile(cliPkg, 'utf8'));
const cli = join(dirname(cliPkg), typeof bin === 'string' ? bin : bin.tailwindcss);
execFileSync(
  process.execPath,
  [cli, '-i', join(here, '../src/tailwind.css'), '-o', twOut, '--minify'],
  { cwd: join(here, '..'), stdio: ['ignore', 'ignore', 'inherit'] }
);
const tw = await readFile(twOut, 'utf8');
await rm(twOut);
// Written beside it and renamed over it, so a reader never sees half a file.
const tmp = join(here, '../dist/.styles.css');
await writeFile(
  tmp,
  `/* @xyflow/react base styles (MIT) */\n${flow}\n/* @braswelljr/algorithm-visualizer */\n${own}\n/* The views' Tailwind utilities (av: prefix) */\n${tw}`
);
await rename(tmp, join(here, '../dist/styles.css'));
