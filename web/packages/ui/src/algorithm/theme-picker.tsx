'use client';

import * as React from 'react';
import { CheckIcon, MonitorIcon, MoonIcon, PaletteIcon, SunIcon } from 'lucide-react';
import { useTheme } from 'next-themes';
import { Button } from '@braswelljr/algorithm-ui/components/button';
import { Popover, PopoverPopup, PopoverTrigger } from '@braswelljr/algorithm-ui/components/popover';
import { accents, css, neutrals, vizPalettes } from '@braswelljr/algorithm-ui/lib/colors';
import { applyPrefs, defaultPrefs, prefsKey, readPrefs, type ThemePrefs } from '@braswelljr/algorithm-ui/lib/theme-prefs';
import { cn } from '@braswelljr/algorithm-ui/lib/utils';

/** The reader's color choices, saved in localStorage and set on <html>. */
export function useThemePrefs(): [ThemePrefs, (p: Partial<ThemePrefs>) => void] {
  const [prefs, setPrefs] = React.useState<ThemePrefs>(defaultPrefs);
  React.useEffect(() => {
    let raw: string | null = null;
    try {
      raw = localStorage.getItem(prefsKey);
    } catch {
      raw = null;
    }
    setPrefs(readPrefs(raw));
  }, []);
  const update = React.useCallback((p: Partial<ThemePrefs>) => {
    setPrefs((old) => {
      const next = { ...old, ...p };
      applyPrefs(next);
      try {
        localStorage.setItem(prefsKey, JSON.stringify(next));
      } catch {
        // Private windows and blocked storage: the choice lasts this visit.
      }
      return next;
    });
  }, []);
  return [prefs, update];
}

function Swatches<T extends { id: string; name: string }>({
  label,
  items,
  value,
  onChange,
  color
}: {
  label: string;
  items: readonly T[];
  value: string;
  onChange: (id: string) => void;
  color: (item: T) => string;
}) {
  return (
    <fieldset>
      <legend className="mb-2 text-xs font-semibold tracking-wide text-muted uppercase">{label}</legend>
      <div className="flex flex-wrap gap-2">
        {items.map((item) => {
          const on = item.id === value;
          return (
            <label
              key={item.id}
              className={cn(
                'relative flex cursor-pointer items-center gap-2 rounded-lg border px-2.5 py-1.5 text-sm has-focus-visible:ring-2 has-focus-visible:ring-accent',
                on ? 'border-accent bg-accent/10 text-fg' : 'border-border text-muted hover:border-control'
              )}
            >
              <input
                type="radio"
                name={label}
                value={item.id}
                checked={on}
                onChange={() => onChange(item.id)}
                className="sr-only"
              />
              <span
                aria-hidden
                className="size-4 rounded-full ring-1 ring-black/10"
                style={{ background: color(item) }}
              />
              {item.name}
              {on ? (
                <CheckIcon
                  aria-hidden
                  className="size-3.5 text-accent"
                />
              ) : null}
            </label>
          );
        })}
      </div>
    </fieldset>
  );
}

/**
 * Light, dark or system, a neutral, an accent and a visualization palette
 * (default, colorblind-safe, high contrast), in a popover from the header.
 * Every choice is a radio with its name beside its swatch.
 *
 *   <ThemePicker />
 */
export function ThemePicker({ className }: { className?: string }) {
  const { theme, setTheme } = useTheme();
  const [prefs, update] = useThemePrefs();
  const modes = [
    { id: 'light', name: 'Light', icon: SunIcon },
    { id: 'dark', name: 'Dark', icon: MoonIcon },
    { id: 'system', name: 'System', icon: MonitorIcon }
  ];
  return (
    <Popover>
      <PopoverTrigger
        render={
          <Button
            variant="ghost"
            size="icon-sm"
            aria-label="Theme and colors"
            className={className}
          />
        }
      >
        <PaletteIcon />
      </PopoverTrigger>
      <PopoverPopup className="w-80">
        <div className="flex flex-col gap-4">
          <fieldset>
            <legend className="mb-2 text-xs font-semibold tracking-wide text-muted uppercase">Mode</legend>
            <div className="grid grid-cols-3 gap-1 rounded-lg bg-subtle p-1">
              {modes.map((m) => (
                <label
                  key={m.id}
                  className={cn(
                    'flex cursor-pointer items-center justify-center gap-1.5 rounded-md py-1.5 text-sm has-focus-visible:ring-2 has-focus-visible:ring-accent',
                    theme === m.id ? 'bg-card text-fg shadow-sm' : 'text-muted hover:text-fg'
                  )}
                >
                  <input
                    type="radio"
                    name="mode"
                    value={m.id}
                    checked={theme === m.id}
                    onChange={() => setTheme(m.id)}
                    className="sr-only"
                  />
                  <m.icon
                    aria-hidden
                    className="size-4"
                  />
                  {m.name}
                </label>
              ))}
            </div>
          </fieldset>
          <Swatches
            label="Accent"
            items={accents}
            value={prefs.accent}
            onChange={(accent) => update({ accent })}
            color={(a) => css(a.light)}
          />
          <Swatches
            label="Neutral"
            items={neutrals}
            value={prefs.neutral}
            onChange={(neutral) => update({ neutral })}
            color={(n) => `oklch(0.6 ${n.c * 4} ${n.h})`}
          />
          <Swatches
            label="Visualization colors"
            items={vizPalettes}
            value={prefs.viz}
            onChange={(viz) => update({ viz })}
            color={(p) =>
              `conic-gradient(${css(p.colors.light.compare)} 0 33%, ${css(p.colors.light.swap)} 0 66%, ${css(p.colors.light.sorted)} 0)`
            }
          />
          <a
            href="/colors/"
            className="text-sm text-accent underline-offset-4 hover:underline"
          >
            Explore the colors
          </a>
        </div>
      </PopoverPopup>
    </Popover>
  );
}
