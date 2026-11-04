'use client';

import * as React from 'react';
import { FlameIcon, RotateCcwIcon, SkipForwardIcon, TrophyIcon } from 'lucide-react';
import { KeyCombo, usePlatformKeys } from '@braswelljr/algorithm-ui/algorithm/key-combo';
import { NumberTicker } from '@braswelljr/algorithm-ui/algorithm/number-ticker';
import { Button } from '@braswelljr/algorithm-ui/components/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@braswelljr/algorithm-ui/components/card';
import { Keyboard } from '@braswelljr/algorithm-ui/components/keyboard';
import { Progress } from '@braswelljr/algorithm-ui/components/progress';
import { Stat, StatLabel, StatValue } from '@braswelljr/algorithm-ui/components/stat';
import { matchesChord, parseKeys, scopeNames, speakKeys, type Shortcut } from '@braswelljr/algorithm-ui/lib/shortcuts';
import { shortcuts } from '@/config/shortcuts';

const progressKey = 'algorithm:shortcut-trainer';

/** Shortcuts a keyboard can press without leaving the page (not Space alone in a list). */
const drills = shortcuts.filter((s) => !s.keys.includes(','));

type Progress = { learned: string[]; best: number };

function readProgress(): Progress {
  try {
    const v: unknown = JSON.parse(localStorage.getItem(progressKey) ?? '{}');
    const o = (v && typeof v === 'object' ? v : {}) as Record<string, unknown>;
    return {
      learned: Array.isArray(o.learned) ? o.learned.filter((x): x is string => typeof x === 'string') : [],
      best: typeof o.best === 'number' ? o.best : 0
    };
  } catch {
    return { learned: [], best: 0 };
  }
}

function pick(exclude?: string): Shortcut {
  const pool = drills.filter((d) => d.id !== exclude);
  return pool[Math.floor(Math.random() * pool.length)];
}

/**
 * The shortcut trainer: it names an action from the shortcuts registry, the
 * student presses its keys (the on-screen keyboard lights each key as it goes
 * down), and a right answer adds to the streak. While it is listening, the
 * site's own shortcuts are held back, so g then l trains instead of navigating.
 * Progress (the shortcuts learned, the best streak) stays in this browser.
 */
export function ShortcutTrainer() {
  const platform = usePlatformKeys();
  const [target, setTarget] = React.useState<Shortcut>(() => drills[0]);
  const [step, setStep] = React.useState(0);
  const [streak, setStreak] = React.useState(0);
  const [progress, setProgress] = React.useState<Progress>({ learned: [], best: 0 });
  const [status, setStatus] = React.useState<'waiting' | 'right' | 'wrong'>('waiting');
  const [active, setActive] = React.useState(false);
  const area = React.useRef<HTMLDivElement>(null);

  React.useEffect(() => {
    setProgress(readProgress());
    setTarget(pick());
  }, []);

  const save = React.useCallback((p: Progress) => {
    setProgress(p);
    try {
      localStorage.setItem(progressKey, JSON.stringify(p));
    } catch {
      // Blocked storage: progress lasts this visit.
    }
  }, []);

  const next = React.useCallback(() => {
    setTarget((t) => pick(t.id));
    setStep(0);
    setStatus('waiting');
  }, []);

  React.useEffect(() => {
    if (!active) return;
    const combo = parseKeys(target.keys)[0];
    const onKey = (e: KeyboardEvent) => {
      if (['Shift', 'Control', 'Alt', 'Meta'].includes(e.key)) return;
      if (e.key === 'Escape') {
        setActive(false);
        return;
      }
      // Hold the site's shortcuts back while training.
      e.stopPropagation();
      e.preventDefault();
      if (status !== 'waiting') return;
      if (matchesChord(e, combo[step], platform)) {
        if (step + 1 < combo.length) {
          setStep(step + 1);
          return;
        }
        const s = streak + 1;
        setStreak(s);
        setStatus('right');
        save({
          learned: progress.learned.includes(target.id) ? progress.learned : [...progress.learned, target.id],
          best: Math.max(progress.best, s)
        });
        window.setTimeout(next, 700);
      } else {
        setStreak(0);
        setStep(0);
        setStatus('wrong');
        window.setTimeout(() => setStatus('waiting'), 900);
      }
    };
    window.addEventListener('keydown', onKey, { capture: true });
    return () => window.removeEventListener('keydown', onKey, { capture: true });
  }, [active, target, step, streak, status, platform, progress, save, next]);

  const learned = progress.learned.length;
  return (
    <div className="flex flex-col gap-6">
      <div className="grid gap-3 sm:grid-cols-3">
        <Stat>
          <StatLabel>Streak</StatLabel>
          <StatValue className="flex items-center gap-2">
            <FlameIcon aria-hidden className="size-6 text-accent" />
            <NumberTicker value={streak} label="in a row" />
          </StatValue>
        </Stat>
        <Stat>
          <StatLabel>Best streak</StatLabel>
          <StatValue className="flex items-center gap-2">
            <TrophyIcon aria-hidden className="size-6 text-warning" />
            <NumberTicker value={progress.best} />
          </StatValue>
        </Stat>
        <Stat>
          <StatLabel>Learned</StatLabel>
          <StatValue>
            <NumberTicker value={learned} /> <span className="text-base text-muted">of {drills.length}</span>
          </StatValue>
          <Progress value={(learned / drills.length) * 100} aria-label={`${learned} of ${drills.length} shortcuts learned`} className="col-span-2" />
        </Stat>
      </div>

      <Card size="lg" variant={status === 'right' ? 'selected' : 'default'}>
        <CardHeader>
          <CardDescription>
            {scopeNames[target.scope]} · {target.category}
          </CardDescription>
          <CardTitle className="text-2xl">{target.description}</CardTitle>
        </CardHeader>
        <CardContent className="flex flex-col gap-4">
          <div
            ref={area}
            aria-live="polite"
            className="min-h-7 text-base"
          >
            {!active ? (
              'Press Start, then press the keys for this action.'
            ) : status === 'right' ? (
              <span className="font-medium text-success">Right: {speakKeys(target.keys)}.</span>
            ) : status === 'wrong' ? (
              <span className="font-medium text-danger">Not that one. Try again.</span>
            ) : step > 0 ? (
              'Good, now the next key.'
            ) : (
              'Listening. Press the keys. (Escape stops.)'
            )}
          </div>
          <div className="flex flex-wrap items-center gap-2">
            {active ? (
              <Button variant="ghost" onClick={() => setActive(false)}>
                Stop
              </Button>
            ) : (
              <Button variant="primary" onClick={() => setActive(true)}>
                Start
              </Button>
            )}
            <Button variant="ghost" onClick={next}>
              <SkipForwardIcon aria-hidden />
              Skip
            </Button>
            <Button
              variant="ghost"
              onClick={() => {
                setStreak(0);
                save({ learned: [], best: 0 });
              }}
            >
              <RotateCcwIcon aria-hidden />
              Reset progress
            </Button>
            <details className="ms-auto text-sm text-muted">
              <summary className="cursor-pointer">Show the answer</summary>
              <KeyCombo keys={target.keys} className="mt-2" />
            </details>
          </div>
        </CardContent>
      </Card>

      <Keyboard live={active} hands={false} className="mx-auto w-full max-w-3xl" />
    </div>
  );
}
