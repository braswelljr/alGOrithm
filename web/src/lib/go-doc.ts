/**
 * Go's doc comment syntax, read for the docs pages (src/components/docs/go-doc.tsx).
 */

export type GoDocBlock =
  | { kind: 'text'; text: string }
  | { kind: 'heading'; text: string }
  | { kind: 'list'; items: string[] }
  | { kind: 'code'; code: string };

const indented = (line: string) => /^[ \t]/.test(line);
const bullet = /^\s*(?:[-*+•]|\d+[.)])\s+/;

/** The lines with their shared indentation taken off. */
function dedent(lines: string[]) {
  const width = Math.min(...lines.filter((l) => l.trim()).map((l) => /^[ \t]*/.exec(l)![0].length));
  return lines.map((l) => l.slice(width)).join('\n');
}

/**
 * Go's doc comment syntax read into blocks: a paragraph is lines up to a blank
 * line, "# Title" is a heading, indented lines are code, and indented lines that
 * all start with "-" or "1." are a list.
 */
export function parseGoDoc(text: string): GoDocBlock[] {
  const blocks: GoDocBlock[] = [];
  let run: string[] = [];
  let runIndented = false;
  const flush = () => {
    if (run.length < 1) return;
    if (!runIndented) {
      const joined = run.map((l) => l.trim()).join(' ');
      if (run.length === 1 && /^#\s+\S/.test(run[0]))
        blocks.push({ kind: 'heading', text: run[0].replace(/^#\s+/, '') });
      else blocks.push({ kind: 'text', text: joined });
    } else {
      const lines = run.filter((l, i) => l.trim() || (i > 0 && i < run.length - 1));
      // A list: it opens with a bullet, and every other line is a bullet or
      // a continuation indented further than the bullets.
      const depth = (l: string) => /^[ \t]*/.exec(l)![0].length;
      const isList =
        bullet.test(lines[0]) &&
        lines.every((l) => !l.trim() || bullet.test(l) || depth(l) > depth(lines[0]));
      if (isList) blocks.push({ kind: 'list', items: listItems(lines) });
      else blocks.push({ kind: 'code', code: dedent(lines).replace(/\n+$/, '') });
    }
    run = [];
  };
  for (const line of text.replace(/\r\n/g, '\n').split('\n')) {
    if (!line.trim()) {
      // A blank line ends a paragraph, but code may hold blank lines.
      if (runIndented) run.push('');
      else flush();
      continue;
    }
    const isIndented = indented(line);
    if (run.length > 0 && isIndented !== runIndented) {
      // Drop trailing blanks kept inside code before switching.
      while (run.length && !run[run.length - 1].trim()) run.pop();
      flush();
    }
    runIndented = isIndented;
    run.push(line);
  }
  while (run.length && !run[run.length - 1].trim()) run.pop();
  flush();
  return blocks;
}

function listItems(lines: string[]) {
  const items: string[] = [];
  for (const l of lines) {
    if (!l.trim()) continue;
    if (bullet.test(l)) items.push(l.replace(bullet, '').trim());
    else if (items.length) items[items.length - 1] += ` ${l.trim()}`;
  }
  return items;
}
