'use client';

import { useEffect, useRef, type ReactNode } from 'react';
import { Caret } from '@braswelljr/algorithm-ui/components/caret';
import { TerminalWindow } from '@braswelljr/algorithm-ui/components/terminal-window';
import { useInView } from '@braswelljr/algorithm-ui/hooks/use-in-view';
import { useReducedMotion } from '@braswelljr/algorithm-ui/hooks/use-reduced-motion';
import {
  useTypewriter,
  type Keystroke,
  type TypingScene
} from '@braswelljr/algorithm-ui/hooks/use-typewriter';
import { cn } from '@braswelljr/algorithm-ui/lib/utils';

type TypingTerminalProps = {
  scenes: TypingScene[];
  /** Says what the animation shows, for readers who cannot see it. */
  label: string;
  /** The prompt before each command. */
  prompt?: ReactNode;
  /** The title bar's text. */
  title?: ReactNode;
  /** Characters per millisecond, roughly: lower is faster. */
  speed?: number;
  loop?: boolean;
  glow?: boolean;
  /** Called on every key down and key up, to drive a keyboard drawn beside it. */
  onStroke?: (stroke: Keystroke | null) => void;
  /**
   * Someone else has the keyboard: the animation holds where it is, and this
   * line, what they are typing, sits at a fresh prompt below it.
   */
  takenOver?: string;
  /**
   * What their own session has printed. While they have the keyboard it is
   * shown in place of the animation, above the line they are typing.
   */
  scrollback?: ReactNode;
  className?: string;
  /** Classes for the window frame, such as dropping its border to fill a monitor's screen. */
  windowClassName?: string;
  /** Sets the screen's height; lines fill it from the top, and older ones scroll off the top. */
  screenClassName?: string;
};

/**
 * A terminal that types real commands and prints what they print. It is a
 * picture of a session, not a session: the window is hidden from assistive
 * technology and the same commands and output are given as text beside it.
 * A caller that lets someone take the keyboard may pass their session's output
 * as `scrollback`; the accessible terminal is the playground's.
 * It only runs while on screen, and under reduced motion it shows the finished
 * session and stays still.
 */
export function TypingTerminal({
  scenes,
  label,
  prompt = '$',
  title,
  speed,
  loop = true,
  glow,
  onStroke,
  takenOver,
  scrollback,
  className,
  windowClassName,
  screenClassName
}: TypingTerminalProps) {
  const ref = useRef<HTMLElement>(null);
  const inView = useInView(ref, '0px 0px -15% 0px');
  const reduced = useReducedMotion();
  const { scenes: shown, stroke } = useTypewriter(scenes, {
    active: inView && takenOver === undefined,
    reduced,
    speed,
    loop
  });
  const typing = takenOver === undefined && shown.some((scene) => scene.typing);

  useEffect(() => onStroke?.(stroke), [stroke, onStroke]);

  return (
    <figure
      ref={ref}
      className={cn('min-w-0 text-left', className)}
    >
      <figcaption className="sr-only">{label}</figcaption>
      <TerminalWindow
        aria-hidden="true"
        title={title}
        glow={glow}
        className={windowClassName}
        data-paused={!inView}
      >
        {/* Lines start at the top, as on a sheet of paper. Once they outgrow
            the screen the newest stay in view and the oldest leave by the top:
            a reversed column overflows upward, and the bottom margin takes the
            spare room only while there is some. */}
        <div
          className={cn(
            'flex h-40 flex-col-reverse overflow-hidden p-3 font-mono text-xs leading-relaxed sm:h-56 sm:p-4 sm:text-sm',
            screenClassName
          )}
        >
          <div className="mb-auto">
            {takenOver !== undefined && scrollback}
            {(takenOver === undefined || scrollback === undefined) &&
              shown.map((scene, i) => (
                <div key={i}>
                  <p className="break-words whitespace-pre-wrap">
                    <span className="text-violet-700 select-none dark:text-violet-300">
                      {prompt}{' '}
                    </span>
                    {scene.command}
                    {scene.typing && takenOver === undefined && <Caret blinking={false} />}
                  </p>
                  {scene.output !== undefined && (
                    <p className="break-words whitespace-pre-wrap text-fd-muted-foreground">
                      {scene.output}
                    </p>
                  )}
                </div>
              ))}
            {!typing && (
              <p className="break-words whitespace-pre-wrap">
                <span className="text-violet-700 select-none dark:text-violet-300">{prompt} </span>
                {takenOver}
                <Caret blinking={!takenOver} />
              </p>
            )}
          </div>
        </div>
      </TerminalWindow>
      <ol className="sr-only">
        {scenes.map((scene) => (
          <li key={scene.command}>
            <code>{scene.command}</code>
            {scene.output ? (
              <>
                {' '}
                prints <samp>{scene.output}</samp>
              </>
            ) : (
              ' prints nothing'
            )}
          </li>
        ))}
      </ol>
    </figure>
  );
}
