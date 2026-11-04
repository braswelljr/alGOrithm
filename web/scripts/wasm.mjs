#!/usr/bin/env node
/**
 * Builds the site's WebAssembly before `next build`: public/algorithm.wasm
 * (the library, which the build itself runs to render every entry's page),
 * public/terminal.wasm (the editor's terminal) and public/wasm_exec.js (from
 * the same Go that built them; the two must match). It is what `make web/wasm`
 * does, for hosts that build with `pnpm build` alone, such as Vercel.
 *
 * Go comes from go.mod's `go` line: the `go` on PATH when it is at least that
 * version, else that release downloaded from go.dev (checked against its
 * published SHA-256) into node_modules/.cache, which the host keeps between
 * builds, as it keeps Go's module and build caches.
 *
 * When the three files are already there (after `make web/wasm`, in CI) it
 * does nothing, except on Vercel (VERCEL=1) or with --force, where it always
 * builds from the checkout.
 *
 *   node scripts/wasm.mjs [--force]
 */
import { execFileSync } from 'node:child_process';
import { createHash } from 'node:crypto';
import { copyFileSync, existsSync, mkdirSync, readFileSync, rmSync, writeFileSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const web = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const root = path.resolve(web, '..');
const publicDir = path.join(web, 'public');
const cache = path.join(web, 'node_modules', '.cache', 'algorithm-go');
const outputs = ['algorithm.wasm', 'terminal.wasm', 'wasm_exec.js'].map((f) =>
  path.join(publicDir, f)
);

const force = process.argv.includes('--force') || process.env.VERCEL === '1';
if (!force && outputs.every((f) => existsSync(f))) {
  console.log('wasm: already built (make web/wasm); skipped');
  process.exit(0);
}

/** The Go release go.mod asks for, as [major, minor]. */
function wanted() {
  const mod = readFileSync(path.join(root, 'go.mod'), 'utf8');
  const m = /^go\s+(\d+)\.(\d+)/m.exec(mod);
  if (!m) throw new Error('go.mod has no go line');
  return [Number(m[1]), Number(m[2])];
}

/** The version of a go binary, as [major, minor], or null when it does not run. */
function versionOf(goBin) {
  try {
    const out = execFileSync(goBin, ['env', 'GOVERSION'], {
      encoding: 'utf8',
      env: { ...process.env, GOTOOLCHAIN: 'local' }
    });
    const m = /go(\d+)\.(\d+)/.exec(out);
    return m ? [Number(m[1]), Number(m[2])] : null;
  } catch {
    return null;
  }
}

const atLeast = (v, w) => v !== null && (v[0] > w[0] || (v[0] === w[0] && v[1] >= w[1]));

const GOOS = { linux: 'linux', darwin: 'darwin', win32: 'windows', freebsd: 'freebsd' };
const GOARCH = { x64: 'amd64', arm64: 'arm64', ia32: '386' };

/** Downloads the newest stable Go of the wanted minor release, once, and returns its go binary. */
async function download([major, minor]) {
  const os = GOOS[process.platform];
  const arch = GOARCH[process.arch];
  if (!os || !arch) throw new Error(`no Go download for ${process.platform}/${process.arch}`);
  const res = await fetch('https://go.dev/dl/?mode=json');
  if (!res.ok) throw new Error(`go.dev/dl: ${res.status}`);
  const releases = await res.json();
  const prefix = `go${major}.${minor}`;
  const release = releases.find(
    (r) => r.stable && (r.version === prefix || r.version.startsWith(`${prefix}.`))
  );
  if (!release) throw new Error(`go.dev lists no stable ${prefix}`);
  const file = release.files.find(
    (f) => f.os === os && f.arch === arch && f.kind === 'archive' && f.filename.endsWith('.tar.gz')
  );
  if (!file) throw new Error(`go.dev has no ${release.version} archive for ${os}/${arch}`);

  const dir = path.join(cache, release.version);
  const goBin = path.join(dir, 'go', 'bin', 'go');
  if (existsSync(goBin)) return goBin;

  console.log(`wasm: downloading ${file.filename}`);
  const archive = await fetch(`https://go.dev/dl/${file.filename}`);
  if (!archive.ok) throw new Error(`go.dev/dl/${file.filename}: ${archive.status}`);
  const bytes = Buffer.from(await archive.arrayBuffer());
  const sum = createHash('sha256').update(bytes).digest('hex');
  if (sum !== file.sha256) throw new Error(`${file.filename}: SHA-256 ${sum}, want ${file.sha256}`);

  rmSync(dir, { recursive: true, force: true });
  mkdirSync(dir, { recursive: true });
  const tarball = path.join(dir, file.filename);
  writeFileSync(tarball, bytes);
  execFileSync('tar', ['-xzf', tarball, '-C', dir], { stdio: 'inherit' });
  rmSync(tarball);
  return goBin;
}

/** The commit the site is built from, for `algorithm version` in the wasm. */
function version() {
  const sha = process.env.VERCEL_GIT_COMMIT_SHA;
  if (sha) return sha.slice(0, 7);
  try {
    return execFileSync('git', ['describe', '--tags', '--always'], {
      cwd: root,
      encoding: 'utf8'
    }).trim();
  } catch {
    return 'dev';
  }
}

const want = wanted();
const pathGo = process.platform === 'win32' ? 'go.exe' : 'go';
const go = atLeast(versionOf(pathGo), want) ? pathGo : await download(want);

const env = {
  ...process.env,
  GOOS: 'js',
  GOARCH: 'wasm',
  GOTOOLCHAIN: 'local',
  GOFLAGS: '-trimpath',
  GOCACHE: process.env.GOCACHE ?? path.join(cache, 'build'),
  GOMODCACHE: process.env.GOMODCACHE ?? path.join(cache, 'mod')
};
const build = (cwd, out, pkg, ldflags) => {
  console.log(`wasm: ${path.relative(web, out)}`);
  execFileSync(go, ['build', `-ldflags=${ldflags}`, '-o', out, pkg], {
    cwd,
    env,
    stdio: 'inherit'
  });
};

mkdirSync(publicDir, { recursive: true });
build(
  root,
  outputs[0],
  './cmd/algorithm-wasm',
  `-s -w -X github.com/braswelljr/alGOrithm/internal/version.Version=${version()}`
);
build(path.join(root, 'web', 'runtime'), outputs[1], './cmd/terminal', '-s -w');
const goroot = execFileSync(go, ['env', 'GOROOT'], { encoding: 'utf8', env }).trim();
copyFileSync(path.join(goroot, 'lib', 'wasm', 'wasm_exec.js'), outputs[2]);
console.log(
  `wasm: built with ${execFileSync(go, ['env', 'GOVERSION'], { encoding: 'utf8', env }).trim()}`
);
