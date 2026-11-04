/**
 * The editor's files: listing a folder, reading and writing text, the
 * language of a file, and the visitor's work kept in this browser.
 *
 * What the visitor saves (a new program, an edited algorithm) and what they
 * delete is kept in localStorage and laid over the starting files on the next
 * visit. It is a convenience, best effort: storage that is full, blocked or in
 * a private window leaves the editor working, unremembered.
 */
import type { MemFS } from './memfs';

/** One entry of a folder. */
export type Entry = { name: string; path: string; dir: boolean };

/** A folder's entries, folders first, then by name. */
export function listDir(fs: MemFS, path: string): Entry[] {
  let names: string[];
  try {
    names = fs.readdirSync(path);
  } catch {
    return [];
  }
  return names
    .map((name) => {
      const p = `${path.replace(/\/$/, '')}/${name}`;
      let dir = false;
      try {
        dir = fs.statSync(p).isDirectory();
      } catch {
        // A dangling link lists as a file.
      }
      return { name, path: p, dir };
    })
    .sort((a, b) => (a.dir === b.dir ? a.name.localeCompare(b.name) : a.dir ? -1 : 1));
}

const decoder = new TextDecoder();

export function readText(fs: MemFS, path: string): string {
  return decoder.decode(fs.readFileSync(path));
}

/** Writes a file, making its folder, and keeps it for the next visit. */
export function writeText(fs: MemFS, path: string, text: string): void {
  fs.mkdirpSync(path.slice(0, path.lastIndexOf('/')) || '/');
  fs.writeFileSync(path, text, 0o644);
  remember((s) => {
    s.files[path] = text;
    s.deleted = s.deleted.filter((d) => d !== path);
  });
}

/** Removes a file or a folder (and what is in it), and keeps that too. */
export function removePath(fs: MemFS, path: string): void {
  const walk = (p: string) => {
    if (fs.statSync(p).isDirectory()) {
      for (const name of fs.readdirSync(p)) walk(`${p}/${name}`);
      fs.rmdirSync(p);
    } else fs.unlinkSync(p);
  };
  walk(path);
  remember((s) => {
    for (const k of Object.keys(s.files))
      if (k === path || k.startsWith(`${path}/`)) delete s.files[k];
    if (!s.deleted.includes(path)) s.deleted.push(path);
  });
}

/** Renames a file or folder, keeping the move. */
export function renamePath(fs: MemFS, from: string, to: string): void {
  fs.mkdirpSync(to.slice(0, to.lastIndexOf('/')) || '/');
  fs.renameSync(from, to);
  remember((s) => {
    for (const [k, v] of Object.entries(s.files)) {
      if (k === from || k.startsWith(`${from}/`)) {
        delete s.files[k];
        s.files[to + k.slice(from.length)] = v;
      }
    }
    if (!s.deleted.includes(from)) s.deleted.push(from);
    // A file moved somewhere new is saved there, even if it was never edited.
    try {
      if (!fs.statSync(to).isDirectory()) s.files[to] = readText(fs, to);
    } catch {
      // Gone already: nothing to keep.
    }
  });
}

/** The language a file is edited and run as. */
export type Language =
  | 'go'
  | 'shell'
  | 'make'
  | 'markdown'
  | 'json'
  | 'yaml'
  | 'toml'
  | 'dockerfile'
  | 'javascript'
  | 'typescript'
  | 'css'
  | 'html'
  | 'xml'
  | 'python'
  | 'rust'
  | 'sql'
  | 'diff'
  | 'properties'
  | 'protobuf'
  | 'lua'
  | 'ruby'
  | 'text';

/** Whole file names that say their language, as code editors know them. */
const NAMES: Record<string, Language> = {
  makefile: 'make',
  gnumakefile: 'make',
  dockerfile: 'dockerfile',
  containerfile: 'dockerfile',
  '.bashrc': 'shell',
  '.profile': 'shell',
  '.zshrc': 'shell',
  '.env': 'properties',
  '.editorconfig': 'properties',
  '.gitignore': 'properties',
  'go.mod': 'text',
  'go.sum': 'text'
};

const EXTENSIONS: Record<string, Language> = {
  go: 'go',
  sh: 'shell',
  bash: 'shell',
  zsh: 'shell',
  gnux: 'shell',
  mk: 'make',
  md: 'markdown',
  markdown: 'markdown',
  json: 'json',
  jsonc: 'json',
  yaml: 'yaml',
  yml: 'yaml',
  toml: 'toml',
  js: 'javascript',
  mjs: 'javascript',
  cjs: 'javascript',
  jsx: 'javascript',
  ts: 'typescript',
  tsx: 'typescript',
  mts: 'typescript',
  css: 'css',
  scss: 'css',
  html: 'html',
  htm: 'html',
  xml: 'xml',
  svg: 'xml',
  py: 'python',
  rs: 'rust',
  sql: 'sql',
  diff: 'diff',
  patch: 'diff',
  ini: 'properties',
  cfg: 'properties',
  conf: 'properties',
  properties: 'properties',
  env: 'properties',
  proto: 'protobuf',
  lua: 'lua',
  rb: 'ruby'
};

export function languageOf(path: string): Language {
  const name = path.slice(path.lastIndexOf('/') + 1).toLowerCase();
  if (name in NAMES) return NAMES[name];
  if (name.startsWith('dockerfile.') || name.endsWith('.dockerfile')) return 'dockerfile';
  const dot = name.lastIndexOf('.');
  return dot > 0 || (dot === 0 && name.length > 1)
    ? (EXTENSIONS[name.slice(dot + 1)] ?? 'text')
    : 'text';
}

/** The language's name, as the status bar shows it. */
export const LANGUAGE_NAMES: Record<Language, string> = {
  go: 'Go',
  shell: 'Shell',
  make: 'Makefile',
  markdown: 'Markdown',
  json: 'JSON',
  yaml: 'YAML',
  toml: 'TOML',
  dockerfile: 'Dockerfile',
  javascript: 'JavaScript',
  typescript: 'TypeScript',
  css: 'CSS',
  html: 'HTML',
  xml: 'XML',
  python: 'Python',
  rust: 'Rust',
  sql: 'SQL',
  diff: 'Diff',
  properties: 'Properties',
  protobuf: 'Protocol Buffers',
  lua: 'Lua',
  ruby: 'Ruby',
  text: 'Plain text'
};

function quote(arg: string): string {
  return /[\s'"$`\\]/.test(arg) ? `'${arg.replace(/'/g, "'\\''")}'` : arg;
}

/** The terminal line that runs a file, if the file is something to run. */
export function runLine(path: string, cwd: string): string | undefined {
  const rel = path.startsWith(`${cwd}/`) ? path.slice(cwd.length + 1) : path;
  switch (languageOf(path)) {
    case 'go':
      return path.endsWith('_test.go') ? undefined : `go run ${quote(rel)}`;
    case 'shell':
      return `bash ${quote(rel)}`;
    case 'make': {
      const slash = rel.lastIndexOf('/');
      const file = rel.slice(slash + 1);
      const dir = slash < 0 ? '' : rel.slice(0, slash) || '/';
      const named = ['Makefile', 'makefile', 'GNUmakefile'].includes(file)
        ? ''
        : ` -f ${quote(file)}`;
      return `make${dir ? ` -C ${quote(dir)}` : ''}${named}`;
    }
    default:
      return undefined;
  }
}

const KEY = 'algorithm:editor-files';
/** The most text kept in this browser, so the page's storage is not filled. */
const BUDGET = 4 * 1024 * 1024;

type Saved = { files: Record<string, string>; deleted: string[] };

function read(): Saved {
  try {
    const raw: unknown = JSON.parse(localStorage.getItem(KEY) ?? 'null');
    if (raw && typeof raw === 'object') {
      const r = raw as Partial<Saved>;
      const files: Record<string, string> = {};
      for (const [k, v] of Object.entries(r.files ?? {})) if (typeof v === 'string') files[k] = v;
      const deleted = Array.isArray(r.deleted)
        ? r.deleted.filter((d) => typeof d === 'string')
        : [];
      return { files, deleted };
    }
  } catch {
    // Unreadable or blocked storage: start from nothing.
  }
  return { files: {}, deleted: [] };
}

function remember(change: (s: Saved) => void): void {
  const s = read();
  change(s);
  try {
    const json = JSON.stringify(s);
    if (json.length <= BUDGET) localStorage.setItem(KEY, json);
  } catch {
    // Full or blocked: the change holds for this visit only.
  }
}

/** Lays the visitor's saved files and deletions over a freshly seeded filesystem. */
export function restoreSaved(fs: MemFS): void {
  const s = read();
  for (const path of s.deleted) {
    try {
      const walk = (p: string) => {
        if (fs.statSync(p).isDirectory()) {
          for (const n of fs.readdirSync(p)) walk(`${p}/${n}`);
          fs.rmdirSync(p);
        } else fs.unlinkSync(p);
      };
      walk(path);
    } catch {
      // Already gone.
    }
  }
  for (const [path, text] of Object.entries(s.files)) {
    try {
      fs.mkdirpSync(path.slice(0, path.lastIndexOf('/')) || '/');
      fs.writeFileSync(path, text, 0o644);
    } catch {
      // A path the filesystem refuses (outside home): skipped.
    }
  }
}

/** Forgets everything kept, so the next visit starts from the starting files. */
export function forgetSaved(): void {
  try {
    localStorage.removeItem(KEY);
  } catch {
    // Nothing kept.
  }
}
