/**
 * Snake's rules, apart from drawing and timing so they can be tested: a board
 * of cells numbered row by row, a snake as its cells from the head, and one
 * step at a time. The autopilot reads the board as a grid a pathfinder from
 * the catalog can run on (lib/grid-text's format: # wall, . open, S, E).
 */

export type Dir = 'up' | 'down' | 'left' | 'right';

export type SnakeState = {
  cols: number;
  rows: number;
  /** The snake's cells, the head first. */
  body: number[];
  dir: Dir;
  food: number;
  /** Whether the edges kill (true) or wrap round (false). */
  walls: boolean;
  alive: boolean;
  /** Food eaten. */
  eaten: number;
};

const OPPOSITE: Record<Dir, Dir> = { up: 'down', down: 'up', left: 'right', right: 'left' };

/** A seeded random number generator (mulberry32), so a game can be replayed. */
export function rng(seed: number) {
  let a = seed >>> 0;
  return () => {
    a = (a + 0x6d2b79f5) >>> 0;
    let t = a;
    t = Math.imul(t ^ (t >>> 15), t | 1);
    t ^= t + Math.imul(t ^ (t >>> 7), t | 61);
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
  };
}

/** A free cell for the food; -1 when the snake fills the board. */
export function placeFood(cols: number, rows: number, body: number[], rand: () => number): number {
  const taken = new Set(body);
  const free: number[] = [];
  for (let i = 0; i < cols * rows; i++) if (!taken.has(i)) free.push(i);
  return free.length ? free[Math.floor(rand() * free.length)] : -1;
}

/** A new game: a short snake in the middle heading right. */
export function newGame(
  cols: number,
  rows: number,
  walls: boolean,
  rand: () => number
): SnakeState {
  const mid = Math.floor(rows / 2) * cols + Math.floor(cols / 3);
  const body = [mid, mid - 1, mid - 2];
  return {
    cols,
    rows,
    body,
    dir: 'right',
    food: placeFood(cols, rows, body, rand),
    walls,
    alive: true,
    eaten: 0
  };
}

/** The cell one step from `cell` in `dir`, or -1 off a walled board. */
export function neighbor(
  s: Pick<SnakeState, 'cols' | 'rows' | 'walls'>,
  cell: number,
  dir: Dir
): number {
  let r = Math.floor(cell / s.cols);
  let c = cell % s.cols;
  if (dir === 'up') r--;
  else if (dir === 'down') r++;
  else if (dir === 'left') c--;
  else c++;
  if (r < 0 || r >= s.rows || c < 0 || c >= s.cols) {
    if (s.walls) return -1;
    r = (r + s.rows) % s.rows;
    c = (c + s.cols) % s.cols;
  }
  return r * s.cols + c;
}

/** The direction from one cell to the next one beside it, if it is beside it. */
export function dirTo(
  s: Pick<SnakeState, 'cols' | 'rows' | 'walls'>,
  from: number,
  to: number
): Dir | null {
  for (const d of ['up', 'down', 'left', 'right'] as Dir[])
    if (neighbor(s, from, d) === to) return d;
  return null;
}

/**
 * One step. A turn straight back is ignored. Eating grows the snake and moves
 * the food; running into a wall or itself ends the game. The tail's cell is
 * free on the step it leaves, so following the tail is safe.
 */
export function step(s: SnakeState, turn: Dir | null, rand: () => number): SnakeState {
  if (!s.alive) return s;
  const dir = turn && turn !== OPPOSITE[s.dir] ? turn : s.dir;
  const head = neighbor(s, s.body[0], dir);
  const eats = head === s.food;
  const body = eats ? s.body : s.body.slice(0, -1);
  if (head < 0 || body.includes(head)) return { ...s, dir, alive: false };
  const next = [head, ...body];
  return {
    ...s,
    dir,
    body: next,
    eaten: s.eaten + (eats ? 1 : 0),
    food: eats ? placeFood(s.cols, s.rows, next, rand) : s.food
  };
}

/**
 * The board as a grid for a pathfinder: the head is S, the food E, the body
 * walls (the tail end too, to be safe). Wrapping edges cannot be said in a
 * grid, so the autopilot plans as if the edges were walls.
 */
export function gridText(s: SnakeState): string {
  const body = new Set(s.body.slice(1));
  const rows: string[] = [];
  for (let r = 0; r < s.rows; r++) {
    let line = '';
    for (let c = 0; c < s.cols; c++) {
      const i = r * s.cols + c;
      line += i === s.body[0] ? 'S' : i === s.food ? 'E' : body.has(i) ? '#' : '.';
    }
    rows.push(line);
  }
  return rows.join('\n');
}

/**
 * The autopilot's move: the first step of the pathfinder's path from the head
 * to the food, or, with no path, any move that does not end the game now.
 */
export function pilotMove(s: SnakeState, path: number[]): Dir | null {
  const at = path.indexOf(s.body[0]);
  const nextCell = at >= 0 ? path[at + 1] : path[0] !== s.body[0] ? path[0] : path[1];
  if (nextCell !== undefined) {
    const d = dirTo(s, s.body[0], nextCell);
    if (d && d !== OPPOSITE[s.dir]) return d;
  }
  for (const d of [s.dir, 'up', 'right', 'down', 'left'] as Dir[]) {
    if (d === OPPOSITE[s.dir]) continue;
    const cell = neighbor(s, s.body[0], d);
    if (cell >= 0 && !s.body.slice(0, -1).includes(cell)) return d;
  }
  return null;
}
