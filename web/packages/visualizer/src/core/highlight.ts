/**
 * A small Go tokenizer for the code view: comments, strings, numbers,
 * keywords, predeclared types and function names. It is not a parser; it only
 * has to color the source the wasm build returns, without a dependency.
 */

export type TokenKind = 'plain' | 'keyword' | 'string' | 'comment' | 'number' | 'type' | 'func';

export interface Token {
  kind: TokenKind;
  text: string;
}

const keywords = new Set(
  'break case chan const continue default defer else fallthrough for func go goto if import interface map package range return select struct switch type var'.split(
    ' '
  )
);
const types = new Set(
  'any bool byte comparable complex64 complex128 error float32 float64 int int8 int16 int32 int64 rune string uint uint8 uint16 uint32 uint64 uintptr true false nil iota'.split(
    ' '
  )
);

const rules: [TokenKind, RegExp][] = [
  ['comment', /\/\/[^\n]*/y],
  ['comment', /\/\*[\s\S]*?(\*\/|$)/y],
  ['string', /"(?:[^"\\\n]|\\.)*"?/y],
  ['string', /`[^`]*`?/y],
  ['string', /'(?:[^'\\\n]|\\.)*'?/y],
  ['number', /\b(?:0[xX][\da-fA-F_]+|\d[\d_]*(?:\.\d+)?(?:[eE][+-]?\d+)?)\b/y],
  ['plain', /[A-Za-z_]\w*/y],
  ['plain', /\s+/y],
  ['plain', /[^\sA-Za-z_\d"`'/]+|\//y]
];

/** Splits Go source into tokens, then into lines of tokens. */
export function highlightGo(code: string): Token[][] {
  const tokens: Token[] = [];
  let i = 0;
  while (i < code.length) {
    let matched = false;
    for (const [kind, re] of rules) {
      re.lastIndex = i;
      const m = re.exec(code);
      if (!m || m[0] === '') continue;
      let k = kind;
      if (kind === 'plain' && /^[A-Za-z_]/.test(m[0])) {
        if (keywords.has(m[0])) k = 'keyword';
        else if (types.has(m[0])) k = 'type';
        else if (code[i + m[0].length] === '(' || code[i + m[0].length] === '[') k = 'func';
      }
      tokens.push({ kind: k, text: m[0] });
      i += m[0].length;
      matched = true;
      break;
    }
    if (!matched) {
      tokens.push({ kind: 'plain', text: code[i] });
      i++;
    }
  }
  // Split tokens that span lines (block comments, raw strings, whitespace).
  const lines: Token[][] = [[]];
  for (const t of tokens) {
    const parts = t.text.split('\n');
    parts.forEach((p, n) => {
      if (n > 0) lines.push([]);
      if (p) lines[lines.length - 1].push({ kind: t.kind, text: p });
    });
  }
  return lines;
}
