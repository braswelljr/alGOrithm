/**
 * A small reader for the ANSI escape sequences a command-line tool emits.
 * Colors and bold (SGR) are kept as styled spans; every other escape sequence
 * (cursor movement, titles, ...) is dropped, since the playground's scrollback
 * is not a real terminal screen.
 */

export interface AnsiSpan {
  text: string;
  /** Foreground color index: 0-7 normal, 8-15 bright. */
  fg?: number;
  /** A foreground given exactly (24-bit, or the 256-color cube), as #rrggbb. */
  color?: string;
  bold?: boolean;
  dim?: boolean;
  underline?: boolean;
}

// CSI sequences (ESC [ ... final byte), OSC sequences (ESC ] ... BEL or ESC \),
// and two-character escapes.
// eslint-disable-next-line no-control-regex -- matching the ESC and BEL control characters is the point
const ESCAPE = /\x1b\[([0-9;:?]*)([ -/]*[@-~])|\x1b\][^\x07\x1b]*(?:\x07|\x1b\\)|\x1b[@-Z\\^_]/g;

/** Removes every escape sequence, leaving the text a terminal would show. */
export function stripAnsi(input: string): string {
  return input.replace(ESCAPE, '');
}

/** Splits text into runs of uniform style. */
export function parseAnsi(input: string): AnsiSpan[] {
  const spans: AnsiSpan[] = [];
  let style: Omit<AnsiSpan, 'text'> = {};
  let last = 0;

  const push = (text: string) => {
    if (text !== '') spans.push({ text, ...style });
  };

  for (const match of input.matchAll(ESCAPE)) {
    push(input.slice(last, match.index));
    last = match.index + match[0].length;
    // Only "select graphic rendition" (final byte m) changes the style.
    if (match[2] === 'm') style = applySgr(style, match[1]);
  }
  push(input.slice(last));
  return spans;
}

function applySgr(style: Omit<AnsiSpan, 'text'>, params: string): Omit<AnsiSpan, 'text'> {
  const next = { ...style };
  const codes = params === '' ? [0] : params.split(';').map((p) => Number(p) || 0);

  for (let i = 0; i < codes.length; i++) {
    const code = codes[i];
    if (code === 0) {
      delete next.fg;
      delete next.color;
      delete next.bold;
      delete next.dim;
      delete next.underline;
    } else if (code === 1) next.bold = true;
    else if (code === 2) next.dim = true;
    else if (code === 4) next.underline = true;
    else if (code === 22) {
      delete next.bold;
      delete next.dim;
    } else if (code === 24) delete next.underline;
    else if (code >= 30 && code <= 37) setIndex(next, code - 30);
    else if (code >= 90 && code <= 97) setIndex(next, code - 90 + 8);
    else if (code === 39) {
      delete next.fg;
      delete next.color;
    } else if (code === 38 || code === 48) {
      // Extended colors carry arguments (5;n or 2;r;g;b), which must not be
      // read as codes of their own. Only the foreground is drawn: the
      // terminal keeps its own background, as the tools expect.
      const kind = codes[i + 1];
      if (code === 38 && kind === 2) {
        next.color = hex(codes[i + 2], codes[i + 3], codes[i + 4]);
        delete next.fg;
      } else if (code === 38 && kind === 5) {
        const n = codes[i + 2] ?? 0;
        if (n < 16) setIndex(next, n);
        else {
          next.color = cube(n);
          delete next.fg;
        }
      }
      i += kind === 5 ? 2 : kind === 2 ? 4 : 0;
    }
  }
  return next;
}

function setIndex(style: Omit<AnsiSpan, 'text'>, n: number) {
  style.fg = n;
  delete style.color;
}

function hex(r = 0, g = 0, b = 0): string {
  const part = (v: number) => Math.max(0, Math.min(255, v)).toString(16).padStart(2, '0');
  return `#${part(r)}${part(g)}${part(b)}`;
}

/** A color of the 256-color palette past the first 16: the 6x6x6 cube, then 24 greys. */
function cube(n: number): string {
  if (n >= 232) {
    const v = 8 + (Math.min(n, 255) - 232) * 10;
    return hex(v, v, v);
  }
  const i = n - 16;
  const level = (v: number) => (v === 0 ? 0 : 55 + v * 40);
  return hex(level(Math.floor(i / 36)), level(Math.floor(i / 6) % 6), level(i % 6));
}
