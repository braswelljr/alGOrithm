'use client';

import { Label } from '@/components/ui/label';
import {
  NumberField,
  NumberFieldDecrement,
  NumberFieldGroup,
  NumberFieldIncrement,
  NumberFieldInput,
  NumberFieldScrubArea
} from '@/components/ui/number-field';
import { Switch } from '@/components/ui/switch';
import type { EntryOption } from '@/visualizer/core';

/** An option that is only on or off (min 0, max 1). */
const isToggle = (o: EntryOption) => o.min === 0 && o.max === 1;

/**
 * The "try it" knobs of a simulation: a switch for each on/off option and a
 * number field (drag its label to scrub) for the rest, each within the
 * entry's own bounds. Every change re-runs the simulation in the browser.
 */
export function Knobs({
  options,
  values,
  onChange
}: {
  options: EntryOption[];
  values: Record<string, number>;
  onChange: (name: string, value: number) => void;
}) {
  if (options.length < 1) return null;
  return (
    <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
      {options.map((o) => {
        const value = values[o.name] ?? o.default;
        if (isToggle(o)) {
          const id = `knob-${o.name}`;
          return (
            <div
              key={o.name}
              className="flex items-center justify-between gap-3"
            >
              <Label htmlFor={id}>{o.label}</Label>
              <Switch
                id={id}
                checked={value !== 0}
                onCheckedChange={(on) => onChange(o.name, on ? 1 : 0)}
              />
            </div>
          );
        }
        return (
          <NumberField
            key={o.name}
            value={value}
            min={o.min}
            max={o.max}
            size="sm"
            onValueChange={(v) => {
              if (v !== null && Number.isFinite(v)) onChange(o.name, Math.round(v));
            }}
          >
            <NumberFieldScrubArea label={o.label} />
            <NumberFieldGroup>
              <NumberFieldDecrement aria-label={`Less ${o.label}`} />
              <NumberFieldInput />
              <NumberFieldIncrement aria-label={`More ${o.label}`} />
            </NumberFieldGroup>
          </NumberField>
        );
      })}
    </div>
  );
}
