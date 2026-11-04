'use client';

import { useEffect, useState } from 'react';

/** One command typed at a prompt, and what it printed. */
export type TypingScene = {
  command: string;
  /** Printed after Enter. Omit for a command that prints nothing. */
  output?: string;
};

/** A key going down: the character typed (`\n` for Enter), and a counter so a repeat still registers. */
export type Keystroke = { key: string; n: number };

/** A scene as it should be drawn right now. */
export type VisibleScene = TypingScene & {
  /** True while its command is still being typed: the caret sits on this line. */
  typing: boolean;
};

type Phase = 'typing' | 'output' | 'end';

type Machine = { scene: number; typed: number; phase: Phase; strokes: number };

export type TypewriterOptions = {
  /** Whether time moves. Off screen, the session holds where it is. */
  active: boolean;
  /** Show the finished session at once and never move. */
  reduced: boolean;
  /** Milliseconds per character, on average. */
  speed?: number;
  /** Start again from a clear screen after the last scene. */
  loop?: boolean;
};

export type Typewriter = {
  scenes: VisibleScene[];
  /** The key that is down, or null between keystrokes. */
  stroke: Keystroke | null;
  /** True once every scene is shown and the session is not looping. */
  finished: boolean;
};

const START: Machine = { scene: 0, typed: 0, phase: 'typing', strokes: 0 };
const KEY_DOWN_MS = 85;

/** A steady, human-looking rhythm: the same jitter on every run, so terminals side by side stay in step. */
function delayFor(char: string, index: number, speed: number) {
  const jitter = 0.6 + ((index * 7919) % 11) / 12;
  return Math.round(speed * jitter * (char === ' ' ? 1.4 : 1));
}

/**
 * Drives a terminal that types its commands one character at a time and then
 * prints their output, scene after scene. It is a small state machine: each
 * state schedules exactly one timer, so pausing is simply not scheduling.
 */
export function useTypewriter(
  all: TypingScene[],
  { active, reduced, speed = 70, loop = true }: TypewriterOptions
): Typewriter {
  const [machine, setMachine] = useState<Machine>(START);
  const [stroke, setStroke] = useState<Keystroke | null>(null);

  useEffect(() => {
    if (!active || reduced || all.length === 0) return;
    const { scene, typed, phase } = machine;
    const current = all[scene];
    let wait: number;
    let next: (m: Machine) => Machine;
    let key: string | null = null;

    if (phase === 'typing' && typed < current.command.length) {
      const char = current.command[typed];
      wait = typed === 0 ? 650 : delayFor(char, typed, speed);
      key = char;
      next = (m) => ({ ...m, typed: m.typed + 1, strokes: m.strokes + 1 });
    } else if (phase === 'typing') {
      wait = 420;
      key = '\n';
      next = (m) => ({ ...m, phase: 'output', strokes: m.strokes + 1 });
    } else if (phase === 'output') {
      const lines = current.output ? current.output.split('\n').length : 0;
      wait = 1300 + lines * 110;
      next = (m) =>
        m.scene + 1 < all.length
          ? { ...m, scene: m.scene + 1, typed: 0, phase: 'typing' }
          : { ...m, phase: 'end' };
    } else {
      if (!loop) return;
      wait = 2200;
      next = (m) => ({ ...START, strokes: m.strokes });
    }

    const timer = window.setTimeout(() => {
      setMachine(next);
      if (key !== null) setStroke({ key, n: machine.strokes + 1 });
    }, wait);
    return () => window.clearTimeout(timer);
  }, [machine, active, reduced, all, speed, loop]);

  // Every key comes back up shortly after it went down.
  useEffect(() => {
    if (stroke === null) return;
    const timer = window.setTimeout(() => setStroke(null), KEY_DOWN_MS);
    return () => window.clearTimeout(timer);
  }, [stroke]);

  if (reduced) {
    return {
      scenes: all.map((scene) => ({ ...scene, typing: false })),
      stroke: null,
      finished: true
    };
  }

  const { scene, typed, phase } = machine;
  const shown = phase === 'end' ? all.length : scene + 1;
  const scenes = all.slice(0, shown).map((s, i) => {
    if (i < scene || phase === 'end') return { ...s, typing: false };
    return {
      command: s.command.slice(0, typed),
      output: phase === 'output' ? s.output : undefined,
      typing: phase === 'typing'
    };
  });

  return { scenes, stroke, finished: phase === 'end' && !loop };
}
