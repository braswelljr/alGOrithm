import { execFileSync } from 'node:child_process';
import { readFile, writeFile } from 'node:fs/promises';
import { defineConfig } from 'tsup';

// Two entry points sharing one chunk, so `.` and `./core` are the same module
// instance (one wasm loader, one catalog cache). `core` has no React: it works
// in Node, in a Next.js server component at build time, and under any
// framework. Only `.` (the React components) is marked "use client" for
// Next.js, added after the build since tsup's banner would mark every file,
// shared chunks included, and so `core` with them.
export default defineConfig((options) => ({
  entry: { index: 'src/index.ts', core: 'src/core/index.ts', react: 'src/react.ts' },
  format: ['esm'],
  dts: true,
  sourcemap: true,
  // The site's dev server reads dist/ while this rebuilds, so `tsup --watch`
  // (pnpm dev) never cleans, and a one-off build keeps styles.css: Turbopack
  // caches a failed `@import` of it until main.css itself changes.
  clean: options.watch ? false : ['!styles.css'],
  splitting: true,
  target: 'es2022',
  external: ['react', 'react-dom', 'react/jsx-runtime', '@xyflow/react', 'zod'],
  async onSuccess() {
    for (const file of ['dist/index.js', 'dist/react.js']) {
      const code = await readFile(file, 'utf8');
      if (!code.startsWith('"use client"')) await writeFile(file, `"use client";\n${code}`);
    }
    // The stylesheet is rebuilt with the code, so `tsup --watch` (pnpm dev)
    // picks up new Tailwind classes in the views too.
    execFileSync(process.execPath, ['scripts/build-css.mjs'], { stdio: 'inherit' });
  }
}));
