import assert from 'node:assert/strict';
import { test } from 'node:test';
import { rankOf } from '../src/lib/arena/rank.ts';
import {
  gridText,
  newGame,
  pilotMove,
  rng,
  step,
  type SnakeState
} from '../src/lib/arena/snake.ts';
import { fewestSwaps } from '../src/lib/arena/swaps.ts';

test('a snake moves, eats and grows, and dies at a wall or itself', () => {
  const rand = rng(1);
  let s = newGame(8, 6, true, rand);
  const head = s.body[0];
  s = step(s, null, rand);
  assert.equal(s.body[0], head + 1);
  assert.equal(s.body.length, 3);
  // A turn straight back is ignored.
  s = step(s, 'left', rand);
  assert.equal(s.body[0], head + 2);
  // Food just ahead is eaten, and the snake grows by one.
  s = { ...s, food: s.body[0] + 1 };
  s = step(s, null, rand);
  assert.equal(s.eaten, 1);
  assert.equal(s.body.length, 4);
  assert.ok(!s.body.includes(s.food));
  // Straight on into the wall.
  while (s.alive) s = step({ ...s, food: -1 }, 'right', rand);
  assert.equal(s.alive, false);
});

test('without walls the snake wraps round the edge', () => {
  const rand = rng(2);
  let s: SnakeState = { ...newGame(5, 3, false, rand), body: [9, 8, 7], dir: 'right', food: 0 };
  s = step(s, null, rand);
  assert.equal(s.body[0], 5);
});

test('the board reads as a grid, and the pilot takes the path’s first step', () => {
  const s = { ...newGame(5, 3, true, rng(3)), body: [6, 5], dir: 'right' as const, food: 8 };
  assert.equal(gridText(s), '.....\n#S.E.\n.....');
  assert.equal(pilotMove(s, [6, 7, 8]), 'right');
  // No path: any move that does not end the game now.
  assert.ok(pilotMove(s, []) !== null);
});

test('the fewest swaps is the length less the cycles', () => {
  assert.equal(fewestSwaps([1, 2, 3]), 0);
  assert.equal(fewestSwaps([2, 1, 3]), 1);
  assert.equal(fewestSwaps([3, 1, 2]), 2);
  assert.equal(fewestSwaps([4, 3, 2, 1]), 2);
});

test('ranks climb with points', () => {
  assert.equal(rankOf(0).name, 'Novice');
  assert.equal(rankOf(100).name, 'Apprentice');
  assert.equal(rankOf(99).toNext, 1);
  assert.equal(rankOf(1_000_000).next, null);
});

test('a car waits its thinking time, then pays each cell’s cost', async () => {
  const { arrivals, positionAt, places, racePoints } = await import('../src/lib/arena/race.ts');
  const cost = [1, 1, 9, 1];
  const t = arrivals([0, 1, 2, 3], cost, 100, 50);
  assert.deepEqual(t, [50, 150, 1050, 1150]);
  assert.deepEqual(positionAt([0, 1, 2, 3], t, 600), { from: 1, to: 2, along: 0.5, done: false });
  assert.equal(positionAt([0, 1, 2, 3], t, 2000)?.done, true);
  assert.deepEqual(places([300, 100, Infinity, 200]), [3, 1, 4, 2]);
  assert.equal(racePoints(1, 4), 460);
});

test('bricks fall, turn, clear lines and end at the top', async () => {
  const b = await import('../src/lib/arena/bricks.ts');
  const rand = rng(7);
  let s = b.newBricks(6, 4, rand);
  // A full bottom row but one cell clears when an I fills it.
  s = { ...s, well: [...new Array(20).fill(null), 'O', 'O', 'O', null], piece: { kind: 'I', turn: 1, row: 0, col: 1 } };
  assert.ok(b.fits(s, s.piece));
  s = b.drop(s, rand);
  assert.equal(s.lines, 1);
  assert.ok(s.score >= 100);
  // A turn against the wall is kicked off it, not refused.
  const t = { ...s, piece: { kind: 'T' as const, turn: 0, row: 2, col: 2 } };
  assert.notEqual(b.rotate(t).piece.turn, 0);
  // Stacking past the top ends the game.
  let full = b.newBricks(4, 4, rand);
  full = { ...full, well: ['O', 'O', null, 'O', 'O', 'O', null, 'O', 'O', 'O', null, 'O', 'O', 'O', null, 'O'] };
  for (let i = 0; i < 6 && !full.over; i++) full = b.drop(full, rand);
  assert.equal(full.over, true);
  assert.equal(b.linePoints(4, 0), 800);
});
