'use client';

/**
 * The editor's terminal: gnux's shell (web/runtime) in a scrollback with a
 * prompt, as the bottom panel of a code editor has. Output keeps its ANSI
 * colors (the theme's palette), Up and Down walk the history, Tab completes
 * commands and files, Ctrl+L clears, and the editor's Run sends a line through
 * here so it shows as if typed.
 *
 * Each panel is one terminal of the session (its own shell, variables and
 * folder); the editor shows several as tabs. The scrollback scrolls inside the
 * panel and the prompt stays at its foot: focusing it never scrolls the page.
 */
import { forwardRef, useCallback, useEffect, useImperativeHandle, useRef, useState } from 'react';
import { parseAnsi, type AnsiSpan } from '@/editor/ansi';
import { complete } from '@/editor/complete';
import type { TerminalSession } from '@/editor/terminal-runtime';
import { cn } from '@/lib/utils';

/** The ANSI palette, as classes on the site's tokens: normal 0-7, bright 8-15. */
const FG = [
  'text-stone-500',
  'text-red-500',
  'text-emerald-500',
  'text-amber-500',
  'text-sky-500',
  'text-fuchsia-500',
  'text-cyan-500',
  'text-fg',
  'text-stone-400',
  'text-red-400',
  'text-emerald-400',
  'text-amber-400',
  'text-sky-400',
  'text-fuchsia-400',
  'text-cyan-400',
  'text-fg'
];

type Line = { id: number; spans: AnsiSpan[]; kind: 'out' | 'err' | 'cmd' };

export type TerminalHandle = {
  /** Runs a line as if typed, and resolves its status. */
  run: (line: string) => Promise<number>;
  focus: () => void;
  clear: () => void;
};

/** Where the prompt says the shell is: ~ for the home folder. */
function pretty(cwd: string, home: string): string {
  if (cwd === home) return '~';
  return cwd.startsWith(`${home}/`) ? `~${cwd.slice(home.length)}` : cwd;
}

export const TerminalPanel = forwardRef<
  TerminalHandle,
  {
    session: TerminalSession;
    /** Which of the session's terminals this is. */
    terminal: number;
    /** Tells the editor where the terminal is, after each line. */
    onCwd?: (dir: string) => void;
    /** A line to run when the terminal opens (cd into a folder). */
    initial?: string;
  }
>(function TerminalPanel({ session, terminal, onCwd, initial }, ref) {
  const [lines, setLines] = useState<Line[]>([]);
  const [input, setInput] = useState('');
  const [busy, setBusy] = useState(false);
  const [cwd, setCwd] = useState(() => session.cwd(terminal));
  const [hint, setHint] = useState('');
  const history = useRef<string[]>([]);
  const back = useRef(-1);
  const next = useRef(0);
  const partial = useRef(['', '']);
  const field = useRef<HTMLInputElement>(null);
  const scroller = useRef<HTMLDivElement>(null);
  const decoder = useRef([new TextDecoder(), new TextDecoder()]);

  const push = useCallback((text: string, kind: Line['kind']) => {
    const add = text.split('\n').map((t) => ({ id: next.current++, spans: parseAnsi(t), kind }));
    setLines((ls) => [...ls, ...add].slice(-2000));
  }, []);

  // The shell's output arrives in chunks; whole lines are shown as they come.
  useEffect(
    () =>
      session.listen(terminal, (fd, chunk) => {
        const i = fd - 1;
        partial.current[i] += decoder.current[i].decode(chunk, { stream: true });
        const parts = partial.current[i].split('\n');
        partial.current[i] = parts.pop() ?? '';
        if (parts.length) push(parts.join('\n'), fd === 1 ? 'out' : 'err');
      }),
    [session, terminal, push]
  );

  // Keep the prompt in view inside the panel (never the page).
  useEffect(() => {
    const el = scroller.current;
    if (el) el.scrollTop = el.scrollHeight;
  }, [lines, busy, hint]);

  const flush = useCallback(() => {
    partial.current.forEach((rest, i) => {
      if (rest) push(rest, i === 0 ? 'out' : 'err');
    });
    partial.current = ['', ''];
  }, [push]);

  const run = useCallback(
    async (line: string) => {
      const where = pretty(session.cwd(terminal), session.home);
      push(`\u001b[32mguest\u001b[0m:\u001b[34m${where}\u001b[0m$ ${line}`, 'cmd');
      if (line.trim())
        history.current = [...history.current.filter((h) => h !== line), line].slice(-200);
      back.current = -1;
      if (line.trim() === 'clear') {
        setLines([]);
        return 0;
      }
      // The prompt is disabled while the line runs; it gets the focus back
      // after if it had it, without moving the page.
      const refocus = document.activeElement === field.current;
      setBusy(true);
      try {
        return await session.run(line, terminal);
      } finally {
        flush();
        setBusy(false);
        if (refocus) requestAnimationFrame(() => field.current?.focus({ preventScroll: true }));
        const dir = session.cwd(terminal);
        setCwd(dir);
        onCwd?.(dir);
      }
    },
    [session, terminal, push, flush, onCwd]
  );

  const started = useRef(false);
  useEffect(() => {
    if (started.current || !initial) return;
    started.current = true;
    void run(initial);
  }, [initial, run]);

  useImperativeHandle(
    ref,
    () => ({
      run,
      focus: () => field.current?.focus({ preventScroll: true }),
      clear: () => setLines([])
    }),
    [run]
  );

  const onKey = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') {
      e.preventDefault();
      if (busy) return;
      const line = input;
      setInput('');
      setHint('');
      void run(line);
    } else if (e.key === 'ArrowUp' || e.key === 'ArrowDown') {
      e.preventDefault();
      const h = history.current;
      if (!h.length) return;
      back.current =
        e.key === 'ArrowUp'
          ? Math.min(h.length - 1, back.current + 1)
          : Math.max(-1, back.current - 1);
      setInput(back.current < 0 ? '' : h[h.length - 1 - back.current]);
    } else if (e.key === 'Tab') {
      e.preventDefault();
      const c = complete({
        line: input,
        cursor: e.currentTarget.selectionStart ?? input.length,
        commands: session.commands,
        fs: session.fs,
        home: session.home
      });
      setInput(c.line);
      setHint(c.candidates.length > 1 ? c.candidates.join('   ') : '');
      requestAnimationFrame(() => field.current?.setSelectionRange(c.cursor, c.cursor));
    } else if (e.key === 'l' && e.ctrlKey) {
      e.preventDefault();
      setLines([]);
    }
  };

  return (
    <div
      className="flex h-full min-h-0 flex-col font-mono text-[13px] leading-5"
      onClick={() => {
        if (!window.getSelection()?.toString()) field.current?.focus({ preventScroll: true });
      }}
    >
      <div
        ref={scroller}
        role="log"
        aria-live="polite"
        aria-label="Terminal output"
        className="min-h-0 flex-1 overflow-auto overscroll-contain px-3 py-2 break-all whitespace-pre-wrap"
      >
        {lines.map((l) => (
          <div
            key={l.id}
            className={cn(l.kind === 'err' && 'text-danger', l.kind === 'cmd' && 'mt-1')}
          >
            {l.spans.length
              ? l.spans.map((s, i) => (
                  <span
                    key={i}
                    className={cn(
                      s.fg !== undefined && FG[s.fg],
                      s.bold && 'font-bold',
                      s.dim && 'opacity-70',
                      s.underline && 'underline'
                    )}
                    style={s.color ? { color: s.color } : undefined}
                  >
                    {s.text}
                  </span>
                ))
              : '\u00a0'}
          </div>
        ))}
        {hint ? <div className="text-muted">{hint}</div> : null}
        <label className="flex items-center gap-2">
          <span
            aria-hidden
            className="shrink-0"
          >
            <span className="text-emerald-500">guest</span>:
            <span className="text-sky-500">{pretty(cwd, session.home)}</span>$
          </span>
          <span className="sr-only">Command</span>
          {/* ui-ok: the terminal's prompt is a bare line, as a terminal's is */}
          <input
            ref={field}
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={onKey}
            disabled={busy}
            spellCheck={false}
            autoCapitalize="off"
            autoComplete="off"
            className="text-fg caret-primary min-w-0 flex-1 bg-transparent outline-none disabled:opacity-60"
          />
        </label>
        {busy ? <div className="text-muted animate-pulse">running...</div> : null}
      </div>
    </div>
  );
});
