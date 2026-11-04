/**
 * Bricks' rules (a falling-blocks game), apart from drawing and timing so they
 * can be tested: a well of cells, the seven pieces, rotation with a small wall
 * kick, line clears and the score. Pieces come from a shuffled bag of seven, so
 * no piece is missing for long.
 */

export type PieceKind = 'I' | 'O' | 'T' | 'S' | 'Z' | 'J' | 'L';

/** Each piece's cells in its four turns, as [row, col] in a 4 by 4 box. */
const SHAPES: Record<PieceKind, [number, number][]> = {
  I: [
    [1, 0],
    [1, 1],
    [1, 2],
    [1, 3]
  ],
  O: [
    [0, 1],
    [0, 2],
    [1, 1],
    [1, 2]
  ],
  T: [
    [0, 1],
    [1, 0],
    [1, 1],
    [1, 2]
  ],
  S: [
    [0, 1],
    [0, 2],
    [1, 0],
    [1, 1]
  ],
  Z: [
    [0, 0],
    [0, 1],
    [1, 1],
    [1, 2]
  ],
  J: [
    [0, 0],
    [1, 0],
    [1, 1],
    [1, 2]
  ],
  L: [
    [0, 2],
    [1, 0],
    [1, 1],
    [1, 2]
  ]
};

export const KINDS = Object.keys(SHAPES) as PieceKind[];

export type Piece = { kind: PieceKind; turn: number; row: number; col: number };

export type BricksState = {
  rows: number;
  cols: number;
  /** The settled cells, row by row: a kind, or null for empty. */
  well: (PieceKind | null)[];
  piece: Piece;
  next: PieceKind;
  bag: PieceKind[];
  lines: number;
  score: number;
  over: boolean;
};

/** A piece's cells, turned: a quarter turn clockwise each time, about its box. */
export function cells(p: Piece): [number, number][] {
  const size = p.kind === 'I' ? 4 : p.kind === 'O' ? 4 : 3;
  return SHAPES[p.kind].map(([r, c]) => {
    let [y, x] = [r, c];
    if (p.kind !== 'O') for (let t = 0; t < ((p.turn % 4) + 4) % 4; t++) [y, x] = [x, size - 1 - y];
    return [p.row + y, p.col + x];
  });
}

/** Whether the piece fits: inside the well (above the top is allowed) and on no settled cell. */
export function fits(s: Pick<BricksState, 'rows' | 'cols' | 'well'>, p: Piece): boolean {
  return cells(p).every(
    ([r, c]) => c >= 0 && c < s.cols && r < s.rows && (r < 0 || s.well[r * s.cols + c] === null)
  );
}

/** The next kind from the bag, refilling it shuffled when it runs out. */
function draw(bag: PieceKind[], rand: () => number): [PieceKind, PieceKind[]] {
  let b = bag;
  if (b.length < 1) {
    b = [...KINDS];
    for (let i = b.length - 1; i > 0; i--) {
      const j = Math.floor(rand() * (i + 1));
      [b[i], b[j]] = [b[j], b[i]];
    }
  }
  return [b[0], b.slice(1)];
}

const spawn = (kind: PieceKind, cols: number): Piece => ({
  kind,
  turn: 0,
  row: -1,
  col: Math.floor(cols / 2) - 2
});

export function newBricks(rows: number, cols: number, rand: () => number): BricksState {
  const [first, bag1] = draw([], rand);
  const [next, bag] = draw(bag1, rand);
  return {
    rows,
    cols,
    well: new Array(rows * cols).fill(null),
    piece: spawn(first, cols),
    next,
    bag,
    lines: 0,
    score: 0,
    over: false
  };
}

/** Points for lines cleared at once, times the level (one more for every ten lines). */
export function linePoints(cleared: number, lines: number): number {
  return [0, 100, 300, 500, 800][cleared] * (1 + Math.floor(lines / 10));
}

/** Moves the piece by a row and a column if it fits. */
export function shift(s: BricksState, dr: number, dc: number): BricksState {
  const p = { ...s.piece, row: s.piece.row + dr, col: s.piece.col + dc };
  return s.over || !fits(s, p) ? s : { ...s, piece: p };
}

/** Turns the piece a quarter clockwise, nudging it off a wall when it would not fit. */
export function rotate(s: BricksState): BricksState {
  if (s.over) return s;
  for (const kick of [0, -1, 1, -2, 2]) {
    const p = { ...s.piece, turn: s.piece.turn + 1, col: s.piece.col + kick };
    if (fits(s, p)) return { ...s, piece: p };
  }
  return s;
}

/** Settles the piece, clears full rows, scores them, and brings on the next piece. */
export function lock(s: BricksState, rand: () => number): BricksState {
  const well = [...s.well];
  let above = false;
  for (const [r, c] of cells(s.piece)) {
    if (r < 0) above = true;
    else well[r * s.cols + c] = s.piece.kind;
  }
  const kept: (PieceKind | null)[][] = [];
  let cleared = 0;
  for (let r = 0; r < s.rows; r++) {
    const row = well.slice(r * s.cols, (r + 1) * s.cols);
    if (row.every((x) => x !== null)) cleared++;
    else kept.push(row);
  }
  const fresh: (PieceKind | null)[] = new Array(cleared * s.cols).fill(null);
  const nextWell = [...fresh, ...kept.flat()];
  const [after, bag] = draw(s.bag, rand);
  const piece = spawn(s.next, s.cols);
  const next: BricksState = {
    ...s,
    well: nextWell,
    piece,
    next: after,
    bag,
    lines: s.lines + cleared,
    score: s.score + linePoints(cleared, s.lines)
  };
  return { ...next, over: above || !fits(next, piece) };
}

/** One fall: down a row, or settle where it is. */
export function fall(s: BricksState, rand: () => number): BricksState {
  if (s.over) return s;
  const moved = shift(s, 1, 0);
  return moved === s ? lock(s, rand) : moved;
}

/** Straight to the bottom, two points a row, and settle. */
export function drop(s: BricksState, rand: () => number): BricksState {
  if (s.over) return s;
  let t = s;
  let rows = 0;
  for (;;) {
    const moved = shift(t, 1, 0);
    if (moved === t) break;
    t = moved;
    rows++;
  }
  return lock({ ...t, score: t.score + rows * 2 }, rand);
}

/** Where the piece would land: its ghost. */
export function ghost(s: BricksState): Piece {
  let p = s.piece;
  for (;;) {
    const q = { ...p, row: p.row + 1 };
    if (!fits(s, q)) return p;
    p = q;
  }
}

/** Milliseconds a row at a number of lines cleared: faster each level. */
export function fallMs(lines: number): number {
  return Math.max(90, 800 - Math.floor(lines / 10) * 70);
}
