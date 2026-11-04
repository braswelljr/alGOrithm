import { useId } from 'react';
import { Combobox } from '@/components/ui/combobox';
import { SPEEDS } from '../core/machine';
import type { Player } from '../hooks/use-player';
import {
  PauseIcon,
  PlayIcon,
  ReplayIcon,
  SkipBackIcon,
  SkipForwardIcon,
  StepBackIcon,
  StepForwardIcon
} from './icons';

/**
 * Play, pause, step, replay, scrub and speed. Every control is a native one,
 * so it works from the keyboard and with a screen reader; each button is
 * enabled only when its event would change the machine's state.
 */
export function PlayerControls({ player, label = 'Step' }: { player: Player; label?: string }) {
  const id = useId();
  const { state, send } = player;
  const playing = state.status === 'playing';
  const finished = state.status === 'finished';
  return (
    <div className="av-controls">
      <div className="av-controls-row">
        <button
          type="button"
          className="av-btn av-btn-icon"
          aria-label="Go to the start"
          disabled={!player.can({ type: 'SEEK', index: 0 })}
          onClick={() => send({ type: 'SEEK', index: 0 })}
        >
          <SkipBackIcon />
        </button>
        <button
          type="button"
          className="av-btn av-btn-icon"
          aria-label="Step back"
          disabled={!player.can({ type: 'STEP_BACK' })}
          onClick={() => send({ type: 'STEP_BACK' })}
        >
          <StepBackIcon />
        </button>
        <button
          type="button"
          className="av-btn av-btn-primary"
          disabled={!player.can({ type: 'TOGGLE' })}
          onClick={() => send({ type: 'TOGGLE' })}
        >
          {playing ? <PauseIcon /> : <PlayIcon />}
          {playing ? 'Pause' : finished ? 'Play again' : 'Play'}
        </button>
        <button
          type="button"
          className="av-btn av-btn-icon"
          aria-label="Step forward"
          disabled={!player.can({ type: 'STEP_FORWARD' })}
          onClick={() => send({ type: 'STEP_FORWARD' })}
        >
          <StepForwardIcon />
        </button>
        <button
          type="button"
          className="av-btn av-btn-icon"
          aria-label="Go to the end"
          disabled={!player.can({ type: 'SEEK', index: Infinity })}
          onClick={() => send({ type: 'SEEK', index: Infinity })}
        >
          <SkipForwardIcon />
        </button>
        <button
          type="button"
          className="av-btn"
          disabled={!player.can({ type: 'REPLAY' })}
          onClick={() => send({ type: 'REPLAY' })}
        >
          <ReplayIcon />
          Replay
        </button>
        <div className="av-speed">
          <label
            htmlFor={`${id}-speed`}
            className="av-muted"
          >
            Speed
          </label>
          <Combobox
            id={`${id}-speed`}
            size="sm"
            className="w-32"
            value={String(state.speed)}
            onChange={(value) => send({ type: 'SET_SPEED', speed: Number(value) })}
            options={SPEEDS.map((s) => ({
              value: String(s),
              label: `${s} ${s === 1 ? 'step' : 'steps'}/s`
            }))}
          />
        </div>
      </div>
      <div className="av-scrub">
        <label
          htmlFor={`${id}-scrub`}
          className="av-sr-only"
        >
          {label}
        </label>
        <input
          id={`${id}-scrub`}
          type="range"
          className="av-range"
          min={0}
          max={state.total}
          value={state.index}
          disabled={state.total === 0}
          onChange={(e) => send({ type: 'SEEK', index: Number(e.target.value) })}
          aria-valuetext={`${label} ${state.index} of ${state.total}`}
        />
        <span className="av-count">
          {state.index} / {state.total}
        </span>
      </div>
    </div>
  );
}
