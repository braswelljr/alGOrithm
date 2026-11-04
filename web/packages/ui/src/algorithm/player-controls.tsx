'use client';

import { Slider } from '@base-ui/react/slider';
import {
  ChevronFirstIcon,
  ChevronLastIcon,
  PauseIcon,
  PlayIcon,
  RotateCcwIcon,
  StepBackIcon,
  StepForwardIcon
} from 'lucide-react';
import { KeyCombo } from '@braswelljr/algorithm-ui/algorithm/key-combo';
import { Card } from '@braswelljr/algorithm-ui/components/card';
import { Button } from '@braswelljr/algorithm-ui/components/button';
import { Tooltip, TooltipContent, TooltipTrigger } from '@braswelljr/algorithm-ui/components/tooltip';
import { cn } from '@braswelljr/algorithm-ui/lib/utils';

function Control({
  label,
  keys,
  onClick,
  disabled,
  children
}: {
  label: string;
  keys?: string;
  onClick: () => void;
  disabled?: boolean;
  children: React.ReactNode;
}) {
  return (
    <Tooltip>
      <TooltipTrigger
        render={
          <Button
            variant="ghost"
            size="icon-sm"
            aria-label={label}
            aria-keyshortcuts={keys?.replaceAll('+', '+')}
            disabled={disabled}
            onClick={onClick}
          />
        }
      >
        {children}
      </TooltipTrigger>
      <TooltipContent className="flex items-center gap-2">
        {label}
        {keys ? <KeyCombo keys={keys} /> : null}
      </TooltipContent>
    </Tooltip>
  );
}

/**
 * Play, pause, step, jump, replay, a scrubber and a speed slider, each with
 * its key in the tooltip (from the shortcuts registry's player keys). The
 * scrubber is a real slider: arrow keys move a step, Page keys ten.
 *
 *   <PlayerControls playing={p} index={i} total={n} speed={s} onToggle={...} onSeek={...} ... />
 */
export function PlayerControls({
  playing,
  index,
  total,
  speed,
  speeds = [0.25, 0.5, 1, 2, 4, 8],
  onToggle,
  onStep,
  onSeek,
  onSpeed,
  onReplay,
  unit = 'step',
  className
}: {
  playing: boolean;
  index: number;
  total: number;
  speed: number;
  speeds?: number[];
  onToggle: () => void;
  onStep: (by: number) => void;
  onSeek: (index: number) => void;
  onSpeed: (speed: number) => void;
  onReplay: () => void;
  unit?: string;
  className?: string;
}) {
  const at = speeds.indexOf(speed);
  return (
    <Card
      size="sm"
      role="group"
      aria-label="Playback"
      className={cn('gap-2 px-(--card-spacing)', className)}
    >
      <Slider.Root
        value={index}
        min={0}
        max={Math.max(total, 1)}
        step={1}
        largeStep={10}
        onValueChange={(v) => onSeek(Array.isArray(v) ? v[0] : v)}
        aria-label={`${unit} ${index} of ${total}`}
        className="px-2"
      >
        <Slider.Control className="flex h-5 w-full touch-none items-center select-none">
          <Slider.Track className="relative h-1.5 w-full rounded-full bg-subtle">
            <Slider.Indicator className="rounded-full bg-accent" />
            <Slider.Thumb
              aria-label={`${unit} ${index} of ${total}`}
              className="size-4 rounded-full border-2 border-accent bg-card shadow-sm outline-none focus-visible:ring-4 focus-visible:ring-accent/30"
            />
          </Slider.Track>
        </Slider.Control>
      </Slider.Root>
      <div className="flex flex-wrap items-center gap-1">
        <Control label="First" keys="home" onClick={() => onSeek(0)} disabled={index === 0}>
          <ChevronFirstIcon />
        </Control>
        <Control label="Step back" keys="arrowleft" onClick={() => onStep(-1)} disabled={index === 0}>
          <StepBackIcon />
        </Control>
        <Control label={playing ? 'Pause' : 'Play'} keys="space" onClick={onToggle} disabled={total === 0}>
          {playing ? <PauseIcon /> : <PlayIcon />}
        </Control>
        <Control label="Step forward" keys="arrowright" onClick={() => onStep(1)} disabled={index >= total}>
          <StepForwardIcon />
        </Control>
        <Control label="Last" keys="end" onClick={() => onSeek(total)} disabled={index >= total}>
          <ChevronLastIcon />
        </Control>
        <Control label="Replay" keys="r" onClick={onReplay}>
          <RotateCcwIcon />
        </Control>
        <span className="ms-2 font-mono text-xs text-muted tabular-nums" aria-live="off">
          {index} / {total}
        </span>
        <span className="ms-auto flex items-center gap-1 text-xs text-muted">
          <Control label="Slower" keys="[" onClick={() => at > 0 && onSpeed(speeds[at - 1])} disabled={at <= 0}>
            <span aria-hidden className="font-mono">-</span>
          </Control>
          <span className="w-10 text-center font-mono tabular-nums" aria-label={`speed ${speed} times`}>
            {speed}x
          </span>
          <Control
            label="Faster"
            keys="]"
            onClick={() => at < speeds.length - 1 && onSpeed(speeds[at + 1])}
            disabled={at < 0 || at >= speeds.length - 1}
          >
            <span aria-hidden className="font-mono">+</span>
          </Control>
        </span>
      </div>
    </Card>
  );
}
