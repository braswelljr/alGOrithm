/**
 * Tab completion for the playground's prompt: command names in command
 * position, file names from the virtual filesystem everywhere else.
 */
import type { MemFS } from './memfs';

export interface CompletionRequest {
  line: string;
  /** Caret offset in `line`. */
  cursor: number;
  /** Command names: the gnux tools plus the shell's builtins. */
  commands: string[];
  fs: MemFS;
  home: string;
}

export interface Completion {
  line: string;
  cursor: number;
  /** Every match, when there is more than one, for the caller to list. */
  candidates: string[];
}

/** Characters that end a word, besides whitespace. */
const BREAKS = '|;&<>';

/** Characters that need a backslash to stay part of a word. */
const SPECIAL = /[\s'"\\$*?[\]|;&<>#~(){}]/g;

function escapeName(name: string): string {
  return name.replace(SPECIAL, '\\$&');
}

function unescapeWord(word: string): string {
  return word.replace(/\\(.)/g, '$1');
}

function commonPrefix(names: string[]): string {
  let prefix = names[0] ?? '';
  for (const name of names) {
    let i = 0;
    while (i < prefix.length && i < name.length && prefix[i] === name[i]) i++;
    prefix = prefix.slice(0, i);
  }
  return prefix;
}

export function complete(req: CompletionRequest): Completion {
  const { line, cursor } = req;
  const unchanged: Completion = { line, cursor, candidates: [] };

  // The word under the caret starts after the nearest unescaped break.
  let start = cursor;
  while (start > 0) {
    const ch = line[start - 1];
    const escaped = start > 1 && line[start - 2] === '\\';
    if (!escaped && (/\s/.test(ch) || BREAKS.includes(ch))) break;
    start--;
  }
  const rawWord = line.slice(start, cursor);
  const word = unescapeWord(rawWord);

  // A word is a command when nothing but a connector comes before it.
  const before = line.slice(0, start).trimEnd();
  const lastChar = before.slice(-1);
  const commandPosition = before === '' || '|;&'.includes(lastChar);
  const afterRedirect = lastChar !== '' && '<>'.includes(lastChar);
  const firstWord = before.split(/[|;&]/).pop()?.trim().split(/\s+/)[0] ?? '';

  let matches: { name: string; dir: boolean }[];
  let typed: string;
  let prefix = '';

  if (commandPosition && !afterRedirect && !word.includes('/')) {
    typed = word;
    matches = [...new Set(req.commands)]
      .filter((name) => name.startsWith(word))
      .sort()
      .map((name) => ({ name, dir: false }));
  } else {
    const slash = word.lastIndexOf('/');
    prefix = slash === -1 ? '' : word.slice(0, slash + 1);
    typed = word.slice(slash + 1);
    let dir = prefix === '' ? '.' : prefix;
    if (dir === '~/' || dir.startsWith('~/')) dir = `${req.home}/${dir.slice(2)}`;

    let names: string[];
    try {
      names = req.fs.readdirSync(dir);
    } catch {
      return unchanged;
    }
    const dirsOnly = firstWord === 'cd' || firstWord === 'rmdirx' || firstWord === 'rmdir';
    matches = names
      .filter((name) => name.startsWith(typed) && (typed.startsWith('.') || !name.startsWith('.')))
      .sort()
      .map((name) => {
        let isDir = false;
        try {
          isDir = req.fs.statSync(`${dir}/${name}`).isDirectory();
        } catch {
          // A dangling link completes as a plain name.
        }
        return { name, dir: isDir };
      })
      .filter((m) => !dirsOnly || m.dir);
  }

  if (matches.length === 0) return unchanged;

  let completed: string;
  let candidates: string[] = [];
  if (matches.length === 1) {
    // A directory invites the next path component; anything else ends the word.
    completed = escapeName(matches[0].name) + (matches[0].dir ? '/' : ' ');
  } else {
    const shared = commonPrefix(matches.map((m) => m.name));
    completed = escapeName(shared.length > typed.length ? shared : typed);
    candidates = matches.map((m) => (m.dir ? `${m.name}/` : m.name));
  }

  const head = line.slice(0, start) + escapeName(prefix).replace(/^\\~/, '~') + completed;
  return { line: head + line.slice(cursor), cursor: head.length, candidates };
}
