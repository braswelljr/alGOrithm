'use client';

import { useId } from 'react';
import { CircleAlert, Flame, Timer, Trophy, Zap } from 'lucide-react';
import { Button } from '@/components/shared/button';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Card } from '@/components/ui/card';
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group';
import { Tabs, TabsList, TabsTab } from '@/components/ui/tabs';
import { levels, modes, type Bests, type GameConfig, type QuizLevel } from '@/lib/quiz-seed';
import type { Entry, GameMode } from '@/visualizer/core';

const modeIcon = { classic: Trophy, time: Timer, streak: Flame } as const;

/**
 * One choice of several, drawn as tabs: the choice changes what the quiz asks,
 * and the tab strip scrolls sideways when the options outrun the screen.
 */
function ChoiceTabs<T extends string>({
  legend,
  tour,
  value,
  options,
  onChange
}: {
  legend: string;
  /** The `data-tour` name. */
  tour?: string;
  value: T;
  options: { id: T; label: string }[];
  onChange(v: T): void;
}) {
  const id = useId();
  return (
    <div
      data-tour={tour}
      className="flex w-full min-w-0 flex-col gap-2"
    >
      <p
        id={id}
        className="text-sm font-medium"
      >
        {legend}
      </p>
      <Tabs
        value={value}
        onValueChange={(v) => {
          const next = options.find((o) => o.id === v);
          if (next) onChange(next.id);
        }}
      >
        <TabsList aria-labelledby={id}>
          {options.map((o) => (
            <TabsTab
              key={o.id}
              value={o.id}
            >
              {o.label}
            </TabsTab>
          ))}
        </TabsList>
      </Tabs>
    </div>
  );
}

/**
 * Choose what to be asked about, how hard, and how to play, then start. The
 * best scores in this browser sit beside each mode.
 */
export function StartScreen({
  config,
  domains,
  entry,
  bests,
  beat,
  error,
  onChange,
  onStart
}: {
  config: GameConfig;
  domains: string[];
  /** The entry of a game about one entry. */
  entry: Entry | undefined;
  bests: Bests;
  /** A score from a shared link, to beat on the same questions. */
  beat: number | null;
  error: string | null;
  onChange(next: Partial<GameConfig>): void;
  onStart(): void;
}) {
  return (
    <div className="flex flex-col gap-6">
      {beat !== null ? (
        <p className="border-border bg-subtle flex flex-wrap items-center justify-between gap-2 rounded-xl border px-3.5 py-2.5 text-sm">
          <Zap
            aria-hidden="true"
            className="size-4"
          />
          Someone scored {beat.toLocaleString()} on this game. Same seed, same questions: beat it.
        </p>
      ) : null}

      {entry ? (
        <div className="border-border bg-subtle flex flex-wrap items-center justify-between gap-2 rounded-xl border px-3.5 py-2.5 text-sm">
          <p>
            A 5-question game about <strong>{entry.name}</strong>, on its example.
          </p>
          <Button
            size="sm"
            variant="ghost"
            onClick={() => onChange({ entry: null })}
          >
            Quiz on everything instead
          </Button>
        </div>
      ) : (
        <>
          <ChoiceTabs
            tour="quiz-topic"
            legend="Topic"
            value={config.domain}
            options={[
              { id: 'mixed', label: 'Mixed' },
              ...domains.map((d) => ({ id: d, label: d }))
            ]}
            onChange={(domain) => onChange({ domain })}
          />
          <ChoiceTabs<QuizLevel>
            tour="quiz-difficulty"
            legend="Difficulty"
            value={config.level}
            options={levels.map((l) => ({
              id: l,
              label: l === 'any' ? 'Any' : l[0].toUpperCase() + l.slice(1)
            }))}
            onChange={(level) => onChange({ level })}
          />
        </>
      )}

      <fieldset data-tour="quiz-mode">
        <legend className="mb-2 text-sm font-medium">Mode</legend>
        <RadioGroup
          value={config.mode}
          onValueChange={(v) => onChange({ mode: v as GameMode })}
          aria-label="Mode"
          className="grid gap-3 sm:grid-cols-3"
        >
          {modes.map((m) => {
            const Icon = modeIcon[m.id];
            const best = bests[m.id];
            return (
              <Card
                key={m.id}
                variant={config.mode === m.id ? 'selected' : 'interactive'}
                size="md"
                className="min-h-36 cursor-pointer gap-1.5 px-(--card-spacing)"
                render={<label />}
              >
                <span className="flex items-center justify-between">
                  <Icon
                    aria-hidden="true"
                    className="text-accent size-5"
                  />
                  <RadioGroupItem value={m.id} />
                </span>
                <span className="font-medium">{m.name}</span>
                <span className="text-muted text-sm">{m.blurb}</span>
                <span className="text-muted mt-auto text-xs">
                  {best
                    ? `Best: ${best.score.toLocaleString()} points, streak ${best.streak}`
                    : 'No best yet'}
                </span>
              </Card>
            );
          })}
        </RadioGroup>
      </fieldset>

      {error ? (
        <Alert variant="destructive">
          <CircleAlert aria-hidden />
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      ) : null}

      <div
        data-tour="quiz-start"
        className="flex flex-wrap items-center gap-3"
      >
        <Button
          variant="primary"
          className="h-11 px-6 text-base"
          onClick={onStart}
        >
          Start
        </Button>
        <span className="text-muted text-sm">
          Keys 1 to 4 answer. Every question comes from a real run of the code.
        </span>
      </div>
    </div>
  );
}
