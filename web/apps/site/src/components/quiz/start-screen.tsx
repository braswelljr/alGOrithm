'use client';

import { Card } from '@braswelljr/algorithm-ui/components/card';
import { RadioGroup, RadioGroupItem } from '@braswelljr/algorithm-ui/components/radio-group';
import { ToggleGroup, ToggleGroupItem } from '@braswelljr/algorithm-ui/components/toggle-group';
import type { Entry, GameMode } from '@braswelljr/algorithm-visualizer/core';
import { Flame, Timer, Trophy, Zap } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { levels, modes, type Bests, type GameConfig, type QuizLevel } from '@/lib/quiz-seed';
import { cn } from '@/utils/utils';

const modeIcon = { classic: Trophy, time: Timer, streak: Flame } as const;

/** A row of radio buttons drawn as chips: one choice of several. */
function Chips<T extends string>({
  legend,
  value,
  options,
  onChange
}: {
  legend: string;
  value: T;
  options: { id: T; label: string }[];
  onChange(v: T): void;
}) {
  return (
    <fieldset className="flex flex-col gap-2">
      <legend className="mb-2 text-sm font-medium">{legend}</legend>
      <ToggleGroup
        variant="chip"
        aria-label={legend}
        value={[value]}
        onValueChange={(v) => {
          const next = options.find((o) => o.id === v[0]);
          if (next) onChange(next.id);
        }}
        className="flex-wrap gap-2"
      >
        {options.map((o) => (
          <ToggleGroupItem
            key={o.id}
            value={o.id}
            className="h-8 px-3.5 text-sm"
          >
            {o.label}
          </ToggleGroupItem>
        ))}
      </ToggleGroup>
    </fieldset>
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
        <p className="flex flex-wrap items-center justify-between gap-2 rounded-xl border border-border bg-subtle px-3.5 py-2.5 text-sm">
          <Zap
            aria-hidden="true"
            className="size-4"
          />
          Someone scored {beat.toLocaleString()} on this game. Same seed, same questions: beat it.
        </p>
      ) : null}

      {entry ? (
        <div className="flex flex-wrap items-center justify-between gap-2 rounded-xl border border-border bg-subtle px-3.5 py-2.5 text-sm">
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
          <Chips
            legend="Topic"
            value={config.domain}
            options={[
              { id: 'mixed', label: 'Mixed' },
              ...domains.map((d) => ({ id: d, label: d }))
            ]}
            onChange={(domain) => onChange({ domain })}
          />
          <Chips<QuizLevel>
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

      <fieldset>
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
                    className="size-5 text-accent"
                  />
                  <RadioGroupItem value={m.id} />
                </span>
                <span className="font-medium">{m.name}</span>
                <span className="text-sm text-muted">{m.blurb}</span>
                <span className="mt-auto text-xs text-muted">
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
        <p
          role="alert"
          className="text-sm text-danger"
        >
          {error}
        </p>
      ) : null}

      <div className="flex flex-wrap items-center gap-3">
        <Button
          variant="primary"
          className="h-11 px-6 text-base"
          onClick={onStart}
        >
          Start
        </Button>
        <span className="text-sm text-muted">
          Keys 1 to 4 answer. Every question comes from a real run of the code.
        </span>
      </div>
    </div>
  );
}
