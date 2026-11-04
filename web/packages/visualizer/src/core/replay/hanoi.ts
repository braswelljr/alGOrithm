/** The Tower of Hanoi after some moves, from the moves the Go side returns. */

export interface HanoiMove {
  from: number;
  to: number;
}

export interface HanoiState {
  pegs: [number[], number[], number[]];
  move: HanoiMove | null;
}

const PEGS = ['A', 'B', 'C'];

/** Parses moves written as "A -> B". Unreadable lines are skipped. */
export function parseMoves(lines: string[]): HanoiMove[] {
  const out: HanoiMove[] = [];
  for (const line of lines) {
    const m = /^\s*([ABC])\s*->\s*([ABC])\s*$/.exec(line);
    if (m) out.push({ from: PEGS.indexOf(m[1]), to: PEGS.indexOf(m[2]) });
  }
  return out;
}

/** Plays the first step moves of n disks, all starting on peg A. */
export function hanoiAt(n: number, moves: HanoiMove[], step: number): HanoiState {
  const pegs: [number[], number[], number[]] = [[], [], []];
  for (let d = n; d >= 1; d--) pegs[0].push(d);
  const k = Math.max(0, Math.min(step, moves.length));
  for (let i = 0; i < k; i++) {
    const disk = pegs[moves[i].from].pop();
    if (disk !== undefined) pegs[moves[i].to].push(disk);
  }
  return { pegs, move: k > 0 ? moves[k - 1] : null };
}

/** The number of disks the moves imply: 2^n - 1 moves. */
export function disksFor(moves: number): number {
  return Math.max(1, Math.round(Math.log2(moves + 1)));
}

/** The peg's letter. */
export function pegName(i: number): string {
  return PEGS[i] ?? '?';
}
