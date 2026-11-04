/** Every transition of the player machine, including the illegal ones. */
import assert from 'node:assert/strict';
import { describe, test } from 'node:test';
import {
  can,
  createPlayerMachine,
  initialPlayerState,
  MAX_SPEED,
  MIN_SPEED,
  playerReducer,
  type PlayerEvent,
  type PlayerState,
  type PlayerStatus
} from '../dist/core.js';

const run = (events: PlayerEvent[], from: PlayerState = initialPlayerState()) =>
  events.reduce(playerReducer, from);
const loaded = (total = 10, index = 0) => run([{ type: 'LOAD' }, { type: 'LOADED', total, index }]);
const playing = (total = 10) => run([{ type: 'PLAY' }], loaded(total));

const all: PlayerEvent[] = [
  { type: 'LOAD' },
  { type: 'LOADED', total: 5 },
  { type: 'FAIL', error: 'x' },
  { type: 'PLAY' },
  { type: 'PAUSE' },
  { type: 'TOGGLE' },
  { type: 'STEP_FORWARD' },
  { type: 'STEP_BACK' },
  { type: 'SEEK', index: 3 },
  { type: 'SET_SPEED', speed: 25 },
  { type: 'TICK', dt: 1000 },
  { type: 'REPLAY' },
  { type: 'RESET' }
];

describe('player machine', () => {
  test('starts idle, paused at 0, never playing', () => {
    const s = createPlayerMachine().initial;
    assert.equal(s.status, 'idle');
    assert.equal(s.index, 0);
    assert.equal(s.total, 0);
  });

  test('idle: only LOAD, SET_SPEED and nothing else change it', () => {
    const s = initialPlayerState();
    for (const e of all) {
      const legal = e.type === 'LOAD' || e.type === 'SET_SPEED';
      assert.equal(can(s, e), legal, `${e.type} from idle`);
    }
  });

  test('loading -> ready, or -> error', () => {
    const l = run([{ type: 'LOAD' }]);
    assert.equal(l.status, 'loading');
    assert.equal(playerReducer(l, { type: 'LOADED', total: 7 }).status, 'ready');
    const err = playerReducer(l, { type: 'FAIL', error: 'nope' });
    assert.equal(err.status, 'error');
    assert.equal(err.error, 'nope');
    // Nothing plays or steps while loading.
    for (const e of all.filter((e) => !['LOADED', 'FAIL', 'RESET', 'SET_SPEED'].includes(e.type))) {
      assert.equal(playerReducer(l, e), l, `${e.type} while loading`);
    }
  });

  test('LOADED clamps the starting index and rests where it lands', () => {
    assert.deepEqual([loaded(10, 4).status, loaded(10, 4).index], ['paused', 4]);
    assert.deepEqual([loaded(10, 99).status, loaded(10, 99).index], ['finished', 10]);
    assert.deepEqual([loaded(10, -3).status, loaded(10, -3).index], ['ready', 0]);
    assert.equal(loaded(0).status, 'ready');
  });

  test('FAIL and LOADED outside loading are no-ops', () => {
    const s = loaded();
    assert.equal(playerReducer(s, { type: 'FAIL', error: 'x' }), s);
    assert.equal(playerReducer(s, { type: 'LOADED', total: 3 }), s);
  });

  test('error: only RESET, LOAD and SET_SPEED do anything', () => {
    const e = run([{ type: 'LOAD' }, { type: 'FAIL', error: 'x' }]);
    for (const ev of all) {
      const legal = ev.type === 'RESET' || ev.type === 'LOAD' || ev.type === 'SET_SPEED';
      assert.equal(can(e, ev), legal, `${ev.type} from error`);
    }
  });

  test('PLAY, PAUSE and TOGGLE', () => {
    const p = playing();
    assert.equal(p.status, 'playing');
    assert.equal(playerReducer(p, { type: 'PLAY' }), p, 'PLAY while playing');
    const q = playerReducer(p, { type: 'PAUSE' });
    assert.equal(q.status, 'paused');
    assert.equal(playerReducer(q, { type: 'PAUSE' }), q, 'PAUSE while paused');
    assert.equal(playerReducer(q, { type: 'TOGGLE' }).status, 'playing');
    assert.equal(playerReducer(p, { type: 'TOGGLE' }).status, 'paused');
  });

  test('a run with no steps cannot play or replay', () => {
    const z = loaded(0);
    for (const e of [{ type: 'PLAY' }, { type: 'TOGGLE' }, { type: 'REPLAY' }] as PlayerEvent[]) {
      assert.equal(playerReducer(z, e), z, e.type);
    }
  });

  test('TICK advances by speed and dt, carrying fractions, only while playing', () => {
    let s = run([{ type: 'SET_SPEED', speed: 10 }], playing(100));
    s = playerReducer(s, { type: 'TICK', dt: 150 }); // 1.5 steps
    assert.equal(s.index, 1);
    s = playerReducer(s, { type: 'TICK', dt: 50 }); // +0.5 carried = 2
    assert.equal(s.index, 2);
    assert.equal(s.status, 'playing');
    const paused = playerReducer(s, { type: 'PAUSE' });
    assert.equal(playerReducer(paused, { type: 'TICK', dt: 1000 }), paused);
    assert.equal(playerReducer(s, { type: 'TICK', dt: 0 }), s, 'zero dt');
    assert.equal(playerReducer(s, { type: 'TICK', dt: -5 }), s, 'negative dt');
  });

  test('TICK reaching the end finishes', () => {
    const s = playerReducer(playing(3), { type: 'TICK', dt: 10_000 });
    assert.equal(s.status, 'finished');
    assert.equal(s.index, 3);
  });

  test('PLAY from finished starts over', () => {
    const f = loaded(5, 5);
    assert.equal(f.status, 'finished');
    const p = playerReducer(f, { type: 'PLAY' });
    assert.deepEqual([p.status, p.index], ['playing', 0]);
  });

  test('stepping pauses, clamps, and finishes at the end', () => {
    const p = playing(5);
    const f = playerReducer(p, { type: 'STEP_FORWARD' });
    assert.deepEqual([f.status, f.index], ['paused', 1]);
    const end = playerReducer(f, { type: 'STEP_FORWARD', by: 99 });
    assert.deepEqual([end.status, end.index], ['finished', 5]);
    assert.equal(playerReducer(end, { type: 'STEP_FORWARD' }), end, 'past the end');
    const back = playerReducer(end, { type: 'STEP_BACK', by: 2 });
    assert.deepEqual([back.status, back.index], ['paused', 3]);
    const start = loaded(5);
    assert.equal(playerReducer(start, { type: 'STEP_BACK' }), start, 'before the start');
    assert.equal(playerReducer(start, { type: 'STEP_FORWARD' }).status, 'paused');
  });

  test('SEEK keeps playing unless it lands on the end', () => {
    const p = playing(10);
    const s = playerReducer(p, { type: 'SEEK', index: 4 });
    assert.deepEqual([s.status, s.index], ['playing', 4]);
    const e = playerReducer(p, { type: 'SEEK', index: Infinity });
    assert.deepEqual([e.status, e.index], ['finished', 10]);
    const r = loaded(10);
    const q = playerReducer(r, { type: 'SEEK', index: 7.9 });
    assert.deepEqual([q.status, q.index], ['paused', 7]);
    assert.equal(playerReducer(q, { type: 'SEEK', index: 7 }), q, 'seek to where it is');
    assert.equal(playerReducer(q, { type: 'SEEK', index: Number.NaN }).index, 0);
  });

  test('REPLAY restarts from 0 and plays', () => {
    for (const from of [loaded(5), loaded(5, 3), loaded(5, 5), playing(5)]) {
      const s = playerReducer(from, { type: 'REPLAY' });
      assert.deepEqual([s.status, s.index], ['playing', 0]);
    }
  });

  test('SET_SPEED clamps and ignores no-change', () => {
    const s = loaded();
    assert.equal(playerReducer(s, { type: 'SET_SPEED', speed: 0 }).speed, MIN_SPEED);
    assert.equal(playerReducer(s, { type: 'SET_SPEED', speed: 1e9 }).speed, MAX_SPEED);
    assert.equal(playerReducer(s, { type: 'SET_SPEED', speed: s.speed }), s);
  });

  test('RESET goes idle from every state and keeps the speed', () => {
    const states: PlayerState[] = [
      loaded(),
      playing(),
      loaded(5, 5),
      loaded(5, 2),
      run([{ type: 'LOAD' }]),
      run([{ type: 'LOAD' }, { type: 'FAIL', error: 'x' }])
    ];
    for (const s of states) {
      const r = playerReducer({ ...s, speed: 25 }, { type: 'RESET' });
      assert.equal(r.status, 'idle');
      assert.equal(r.index, 0);
      assert.equal(r.speed, 25);
    }
    const idle = initialPlayerState();
    assert.equal(playerReducer(idle, { type: 'RESET' }), idle);
  });

  test('every status is reachable', () => {
    const seen = new Set<PlayerStatus>([
      initialPlayerState().status,
      run([{ type: 'LOAD' }]).status,
      loaded().status,
      playing().status,
      loaded(5, 2).status,
      loaded(5, 5).status,
      run([{ type: 'LOAD' }, { type: 'FAIL', error: 'x' }]).status
    ]);
    assert.equal(seen.size, 7);
  });
});
