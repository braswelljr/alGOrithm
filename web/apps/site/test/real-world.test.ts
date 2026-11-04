/** The "Real-world code" section's helpers. */
import assert from 'node:assert/strict';
import { test } from 'node:test';
import { examplePrograms, programName, runCommand } from '../src/lib/real-world.ts';

test('examplePrograms keeps each linked program once, in order', () => {
  const uses = [
    { title: 'a', where: 'x.', example: 'examples/http-cache' },
    { title: 'b', where: 'y.', example: '' },
    { title: 'c', where: 'z.', example: 'examples/http-cache' },
    { title: 'd', where: 'w.', example: 'examples/leaderboard' }
  ];
  assert.deepEqual(examplePrograms(uses), ['examples/http-cache', 'examples/leaderboard']);
  assert.deepEqual(examplePrograms([]), []);
});

test('runCommand and programName', () => {
  assert.equal(runCommand('examples/leaderboard'), 'go run ./examples/leaderboard');
  assert.equal(runCommand('./examples/leaderboard/'), 'go run ./examples/leaderboard');
  assert.equal(programName('examples/leaderboard/'), 'leaderboard');
});
