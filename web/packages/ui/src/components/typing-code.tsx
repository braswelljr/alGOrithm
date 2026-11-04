'use client';

import { useRef, useState, type ReactNode } from 'react';
import { Caret } from '@braswelljr/algorithm-ui/components/caret';
import { TerminalWindow } from '@braswelljr/algorithm-ui/components/terminal-window';
import { useInView } from '@braswelljr/algorithm-ui/hooks/use-in-view';
import { useReducedMotion } from '@braswelljr/algorithm-ui/hooks/use-reduced-motion';
import { useTypewriter, type TypingScene } from '@braswelljr/algorithm-ui/hooks/use-typewriter';
import { cn } from '@braswelljr/algorithm-ui/lib/utils';

type TypingCodeProps = {
  /** The source to type. */
  code: string;
  /** The command that runs it, and what that prints. */
  run: TypingScene;
  /** The file name in the title bar. */
  file: string;
  /** Says what the animation shows, for readers who cannot see it. */
  label: string;
  className?: string;
};

const TOKEN = /("[^"\n]*"?|\b[a-z]+(?=\.)|\b[A-Z]\w*|:=|\/\/[^\n]*)/g;

/** Just enough Go coloring to read: strings, packages, exported names, comments. */
function tint(source: string): ReactNode[] {
  return source.split(TOKEN).map((part, i) => {
    if (i % 2 === 0) return part;
    const tone = part.startsWith('//')
      ? 'text-fd-muted-foreground italic'
      : part.startsWith('"')
        ? 'text-amber-700 dark:text-amber-300'
        : part === ':='
          ? 'text-fd-muted-foreground'
          : /^[A-Z]/.test(part)
            ? 'text-violet-700 dark:text-violet-300'
            : 'text-cyan-700 dark:text-cyan-300';
    return (
      <span
        key={i}
        className={tone}
      >
        {part}
      </span>
    );
  });
}

/**
 * An editor that types a short Go program, then a terminal strip that runs it.
 * Like `TypingTerminal` it is a picture: hidden from assistive technology,
 * with the program and its output given as text beside it.
 */
export function TypingCode({ code, run, file, label, className }: TypingCodeProps) {
  const ref = useRef<HTMLElement>(null);
  const inView = useInView(ref, '0px 0px -15% 0px');
  const reduced = useReducedMotion();
  const [scenes] = useState<TypingScene[]>(() => [{ command: code }, run]);
  const { scenes: shown } = useTypewriter(scenes, { active: inView, reduced, speed: 28 });
  const [source, command] = shown;

  return (
    <figure
      ref={ref}
      className={cn('min-w-0 text-left', className)}
    >
      <figcaption className="sr-only">{label}</figcaption>
      <TerminalWindow
        aria-hidden="true"
        title={file}
        data-paused={!inView}
      >
        <pre className="h-36 overflow-hidden p-3 font-mono text-xs leading-relaxed sm:h-44 sm:p-4 sm:text-sm">
          {tint(source?.command ?? '')}
          {source?.typing && <Caret blinking={false} />}
        </pre>
        <div className="h-20 overflow-hidden border-t bg-fd-secondary/40 p-3 font-mono text-xs leading-relaxed sm:h-24 sm:px-4">
          {command && (
            <>
              <p>
                <span className="text-violet-700 select-none dark:text-violet-300">$ </span>
                {command.command}
                {command.typing && <Caret blinking={false} />}
              </p>
              {command.output !== undefined && (
                <p className="whitespace-pre-wrap text-fd-muted-foreground">{command.output}</p>
              )}
            </>
          )}
        </div>
      </TerminalWindow>
      <div className="sr-only">
        <pre>
          <code>{code}</code>
        </pre>
        <p>
          <code>{run.command}</code> prints <samp>{run.output}</samp>
        </p>
      </div>
    </figure>
  );
}
