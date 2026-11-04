import { createHash } from 'node:crypto';
import { existsSync, readFileSync } from 'node:fs';
import type { NextConfig } from 'next';

// A short hash of the wasm build, put in its URL (?v=), so the browser and
// the CDN may cache algorithm.wasm forever and a new build is a new URL: a
// page never pairs new JavaScript with an old wasm.
function wasmVersion(): string {
  const file = 'public/algorithm.wasm';
  if (!existsSync(file)) return 'dev';
  return createHash('sha256').update(readFileSync(file)).digest('hex').slice(0, 12);
}

// A static export: the site is plain files plus the wasm build, so it can be
// hosted anywhere, and nothing runs on a server that could receive a visitor's
// data (docs/rules/the-users-data-never-leaves-the-browser.md).
const config: NextConfig = {
  output: 'export',
  env: { NEXT_PUBLIC_WASM_VERSION: wasmVersion() },
  images: { unoptimized: true },
  trailingSlash: true,
  reactStrictMode: true,
  // The component library ships TypeScript source; Next compiles it.
  transpilePackages: ['@braswelljr/algorithm-ui']
};

export default config;
