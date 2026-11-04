import { type ComponentProps } from 'react';
import { cn } from '@braswelljr/algorithm-ui/lib/utils';

const tones = {
  /** A letter key: the default. */
  alpha: '',
  /** A modifier: shift, tab. Quieter, for labels and secondary surfaces. */
  mod: 'keycap-mod',
  /** Return: the one key on a board, or action on a screen, that matters most. */
  accent: 'keycap-accent'
};

export type KeycapTone = keyof typeof tones;

type KeycapProps = ComponentProps<'span'> & {
  tone?: KeycapTone;
  /** Held down by something other than the pointer: a keystroke, an animation. */
  pressed?: boolean;
};

/**
 * A keycap: a face over a skirt, which sinks when pressed. It draws a key and
 * nothing else; an interactive key is a `Button`, which wears the same cap.
 */
export function Keycap({ tone = 'alpha', pressed, className, ...props }: KeycapProps) {
  return (
    <span
      data-pressed={pressed || undefined}
      className={cn('keycap', tones[tone], className)}
      {...props}
    />
  );
}

/**
 * A key named in running text: `rm`, `Enter`, `Ctrl+C`. It is a real `kbd`,
 * so a screen reader hears it as keyboard input.
 */
export function Kbd({ className, ...props }: ComponentProps<'kbd'>) {
  return (
    <kbd
      className={cn(
        'keycap inline-flex items-center rounded-md px-1.5 py-px font-mono text-xs keycap-mod [--cap-depth:2px]',
        className
      )}
      {...props}
    />
  );
}
