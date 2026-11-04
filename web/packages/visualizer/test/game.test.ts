/** The quiz game's rules and every transition, legal and illegal. */
import assert from 'node:assert/strict';
import { describe, test } from 'node:test';
import {
  BASE_POINTS,
  GAME_RULES,
  gameReducer,
  gameSummary,
  initialGameState,
  multiplier,
  points,
  questionCount,
  questionSchema,
  questionsFor,
  TIME_BONUS,
  type GameEvent,
  type GameState
} from '../dist/core.js';

const run = (events: GameEvent[], from: GameState = initialGameState()) =>
  events.reduce(gameReducer, from);
const start = (mode: 'classic' | 'time' | 'streak' = 'classic', n = 12, questions?: number) =>
  run([
    {
      type: 'START',
      mode,
      answers: Array.from({ length: n }, (_, i) => i % 3),
      choices: Array.from({ length: n }, () => 4),
      questions
    }
  ]);
const right = (s: GameState, ms = 1000): GameEvent => ({
  type: 'ANSWER',
  choice: s.answers[s.index],
  ms
});
const wrong = (s: GameState): GameEvent => ({
  type: 'ANSWER',
  choice: (s.answers[s.index] + 1) % 4,
  ms: 1000
});
/** Answers right (or wrong) and moves on. */
const step = (s: GameState, ok = true) => run([ok ? right(s) : wrong(s), { type: 'NEXT' }], s);

describe('scoring', () => {
  test('the multiplier grows every three in a row, up to x4', () => {
    assert.deepEqual([0, 1, 3, 4, 6, 7, 10, 40].map(multiplier), [1, 1, 1, 2, 2, 3, 4, 4]);
  });
  test('points are base x multiplier + a time bonus for the time left', () => {
    assert.equal(points(1, 0, 20_000), BASE_POINTS + TIME_BONUS);
    assert.equal(points(1, 20_000, 20_000), BASE_POINTS);
    assert.equal(points(4, 10_000, 20_000), 2 * BASE_POINTS + TIME_BONUS / 2);
    assert.equal(points(1, 99_999, 20_000), BASE_POINTS, 'late answers get no negative bonus');
  });
});

describe('game machine', () => {
  test('starts at the intro with the mode rules', () => {
    const s = initialGameState('classic');
    assert.equal(s.status, 'intro');
    assert.equal(s.lives, 3);
    assert.deepEqual(initialGameState('streak').rules, GAME_RULES.streak);
  });

  test('START asks the first question; with no questions it does nothing', () => {
    const s = start();
    assert.deepEqual([s.status, s.index, s.score], ['question', 0, 0]);
    const empty = initialGameState();
    assert.equal(
      gameReducer(empty, { type: 'START', mode: 'classic', answers: [], choices: [] }),
      empty
    );
  });

  test('classic asks 10 questions, and a smaller count is honored', () => {
    assert.equal(questionCount(start('classic', 30)), 10);
    assert.equal(questionCount(start('classic', 30, 5)), 5);
    assert.equal(questionCount(start('time', 30)), 30);
  });

  test('a right ANSWER scores and grows the streak', () => {
    const s0 = start();
    const s = gameReducer(s0, right(s0, 0));
    assert.deepEqual(
      [s.status, s.streak, s.bestStreak, s.score, s.lives],
      ['answered', 1, 1, 150, 3]
    );
    assert.deepEqual(s.results, [{ pick: 0, right: true, points: 150, ms: 0 }]);
  });

  test('a wrong ANSWER costs a life and the streak, and scores nothing', () => {
    let s = step(start());
    s = gameReducer(s, wrong(s));
    assert.deepEqual([s.streak, s.bestStreak, s.lives, s.results[1].points], [0, 1, 2, 0]);
  });

  test('ANSWER with a choice out of range, or outside a question, does nothing', () => {
    const s = start();
    assert.equal(gameReducer(s, { type: 'ANSWER', choice: 4, ms: 0 }), s);
    assert.equal(gameReducer(s, { type: 'ANSWER', choice: -1, ms: 0 }), s);
    const answered = gameReducer(s, right(s));
    assert.equal(gameReducer(answered, right(answered)), answered);
    const intro = initialGameState();
    assert.equal(gameReducer(intro, { type: 'ANSWER', choice: 0, ms: 0 }), intro);
  });

  test('TIMEOUT on a question is a miss; without a question timer it does nothing', () => {
    const s = gameReducer(start(), { type: 'TIMEOUT', scope: 'question' });
    assert.deepEqual(
      [s.status, s.lives, s.results[0].pick, s.results[0].ms],
      ['answered', 2, null, 20_000]
    );
    const timed = start('time');
    assert.equal(gameReducer(timed, { type: 'TIMEOUT', scope: 'question' }), timed);
  });

  test('TIMEOUT of the game ends time attack only', () => {
    const t = gameReducer(start('time'), { type: 'TIMEOUT', scope: 'game' });
    assert.deepEqual([t.status, t.ended], ['finished', 'time']);
    const c = start('classic');
    assert.equal(gameReducer(c, { type: 'TIMEOUT', scope: 'game' }), c);
  });

  test('REVEAL goes from answered to revealing, nowhere else', () => {
    const s = start();
    assert.equal(gameReducer(s, { type: 'REVEAL' }), s);
    const r = run([right(s), { type: 'REVEAL' }], s);
    assert.equal(r.status, 'revealing');
    assert.equal(gameReducer(r, { type: 'REVEAL' }), r);
  });

  test('NEXT asks the next question, from answered or revealing', () => {
    const s = start();
    assert.equal(gameReducer(s, { type: 'NEXT' }), s);
    const a = run([right(s), { type: 'NEXT' }], s);
    const b = run([right(s), { type: 'REVEAL' }, { type: 'NEXT' }], s);
    assert.deepEqual([a.status, a.index, b.status, b.index], ['question', 1, 'question', 1]);
  });

  test('classic ends after its questions, or with no lives left', () => {
    let s = start('classic', 12);
    for (let i = 0; i < 10; i++) s = step(s);
    assert.deepEqual([s.status, s.ended, s.results.length], ['finished', 'done', 10]);
    let l = start('classic');
    for (let i = 0; i < 3; i++) l = step(l, false);
    assert.deepEqual([l.status, l.ended, l.lives], ['finished', 'lives', 0]);
  });

  test('streak ends at the first miss', () => {
    let s = start('streak', 20);
    s = step(step(s));
    s = step(s, false);
    assert.deepEqual([s.status, s.ended, s.bestStreak], ['finished', 'lives', 2]);
  });

  test('a streak of five is a milestone; NEXT clears it', () => {
    let s = start('time', 20);
    for (let i = 0; i < 4; i++) s = step(s);
    s = gameReducer(s, right(s));
    assert.equal(s.milestone, true);
    assert.equal(gameReducer(s, { type: 'NEXT' }).milestone, false);
  });

  test('RESTART goes back to the intro of the same mode', () => {
    const s = gameReducer(step(start('streak')), { type: 'RESTART' });
    assert.deepEqual([s.status, s.mode, s.score, s.results], ['intro', 'streak', 0, []]);
    const intro = initialGameState();
    assert.equal(gameReducer(intro, { type: 'RESTART' }), intro);
  });

  test('the summary counts accuracy, streak and time', () => {
    let s = start();
    s = step(s);
    s = step(s, false);
    const sum = gameSummary(s);
    assert.deepEqual(
      [sum.asked, sum.right, sum.accuracy, sum.bestStreak, sum.elapsedMs],
      [2, 1, 0.5, 1, 2000]
    );
  });
});

describe('questions', () => {
  const q = (unit: string, step: number) =>
    questionSchema.parse({
      kind: 'x',
      prompt: 'p',
      step,
      unit,
      choices: ['a', 'b', 'c'],
      answer: 0,
      explanation: 'e'
    });
  test('questionsFor keeps the questions of a frame unit within its steps', () => {
    assert.deepEqual(
      questionsFor([q('step', 1), q('time', 2), q('step', 99)], 'Step', 10).map((x) => x.step),
      [1]
    );
  });
  test('the schema rejects an answer that is not a choice', () => {
    const r = questionSchema.safeParse({
      kind: 'x',
      prompt: 'p',
      step: 0,
      unit: 'step',
      choices: ['a', 'b'],
      answer: 2,
      explanation: 'e'
    });
    assert.equal(r.success, false);
  });
});
