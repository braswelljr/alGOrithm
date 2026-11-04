/** Where the site serves the wasm build, and the base path it lives under. */
export const basePath = process.env.NEXT_PUBLIC_BASE_PATH ?? '';

/** The wasm build's content hash (next.config.ts), so its URL changes with it. */
export const wasmVersion = process.env.NEXT_PUBLIC_WASM_VERSION ?? 'dev';

export const wasmOptions = {
  wasmUrl: `${basePath}/algorithm.wasm?v=${wasmVersion}`,
  execUrl: `${basePath}/wasm_exec.js?v=${wasmVersion}`,
  // The release's builds (GitHub Pages) when published, else the site's own:
  // lib/wasm-source.ts.
  resolve: () => import('@/lib/wasm-source').then((m) => m.wasmUrl('algorithm'))
};

export const repoUrl = 'https://github.com/braswelljr/alGOrithm';

/**
 * Where the site is published, for canonical links, the sitemap and social
 * cards. Set NEXT_PUBLIC_SITE_URL when deploying anywhere else.
 */
export const siteUrl = (process.env.NEXT_PUBLIC_SITE_URL ?? 'https://algorithm.vercel.app').replace(
  /\/+$/,
  ''
);

export const siteName = 'alGOrithm';
export const siteTagline = 'Algorithms and data structures, step by step';
export const siteDescription =
  'Run hundreds of algorithms, data structures and operating-system simulations on your own data and watch every step. The real Go code, compiled to WebAssembly, running in your browser.';

/** The author, credited in the metadata, the footer and the structured data. */
export const author = {
  name: 'braswelljr',
  url: 'https://braswelljr.vercel.app',
  github: 'https://github.com/braswelljr'
} as const;

export const npmUrl = 'https://www.npmjs.com/package/@/visualizer';
export const licenseUrl = `${repoUrl}/blob/main/LICENSE`;

/** An absolute URL on this site for a path ("/docs/"). */
export function absoluteUrl(path: string): string {
  return `${siteUrl}${basePath}${path.startsWith('/') ? path : `/${path}`}`;
}
