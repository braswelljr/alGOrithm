'use client';

/**
 * The editor: a code editor in the browser, laid out as one. The explorer on
 * the left (the visitor's home and the repository's packages), the open files
 * as tabs over the editor, and below it a panel of tabs: the problems in the
 * open files, and as many terminals as the visitor opens (each gnux's shell of
 * its own, with `go run`, `bash`, `make`, compression and every gnux tool). A
 * title bar with Run and the settings, and a status bar. It takes the whole
 * screen, apart from the site, and on a wide screen never scrolls as a page:
 * each part scrolls inside itself. On a phone the code and the panel stack
 * down a page that scrolls, and the files open in a drawer.
 *
 * Save (Ctrl/Cmd+S) formats the file first when Format on save is on (gofmt
 * for Go, tidy whitespace for every file). Files are checked as they are
 * edited; Run (Ctrl/Cmd+Enter) saves and runs the file in the terminal (a Go
 * program with `go run`, on the Go Playground, which the button says; a
 * script with `bash`; a Makefile with `make`), and with Stop on errors on, a
 * file with errors shows them instead. Saved files are kept in this browser.
 *
 * ui-ok: primaries: Run, and Try again on the screen shown when the terminal fails.
 */
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import Link from 'next/link';
import {
  LuArrowLeft,
  LuCircle,
  LuFolderTree,
  LuPlay,
  LuPlus,
  LuRotateCcw,
  LuSave,
  LuSettings,
  LuTrash2,
  LuWandSparkles,
  LuX
} from 'react-icons/lu';
import { toast } from 'sonner';
import { ThemeToggle } from '@/components/animated/theme-toggle';
import { LogoColor } from '@/components/brand/logo';
import {
  AlertDialog,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle
} from '@/components/ui/alert-dialog';
import { Button } from '@/components/ui/button';
import {
  ContextMenu,
  ContextMenuContent,
  ContextMenuItem,
  ContextMenuSeparator,
  ContextMenuTrigger
} from '@/components/ui/context-menu';
import {
  Drawer,
  DrawerHeader,
  DrawerPanel,
  DrawerPopup,
  DrawerTitle
} from '@/components/ui/drawer';
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuSeparator,
  DropdownMenuShortcut,
  DropdownMenuTrigger
} from '@/components/ui/dropdown-menu';
import { ResizableHandle, ResizablePanel, ResizablePanelGroup } from '@/components/ui/resizable';
import { Skeleton } from '@/components/ui/skeleton';
import {
  forgetSaved,
  LANGUAGE_NAMES,
  languageOf,
  readText,
  removePath,
  renamePath,
  runLine,
  writeText
} from '@/editor/files';
import { takeHandoff } from '@/editor/handoff';
import { terminalSession } from '@/editor/session';
import { useEditorSettings, type EditorSettings } from '@/editor/settings';
import type { Diagnostic, TerminalSession } from '@/editor/terminal-runtime';
import { useMediaQuery } from '@/hooks/use-media-query';
import { cn } from '@/lib/utils';
import {
  CodeEditor,
  forgetEditorState,
  moveEditorState,
  replaceEditorText,
  revealInEditor
} from './code-editor';
import { Explorer } from './explorer';
import { FileIcon } from './file-icon';
import { NoFile } from './no-file';
import { countProblems, ProblemsPanel, SeverityIcon } from './problems-panel';
import { TerminalPanel, type TerminalHandle } from './terminal-panel';

type Tab = { path: string; text: string; saved: string };
type Term = { id: number; initial?: string };
/** The bottom panel's tab: the problems, or a terminal by id. */
type Bottom = 'problems' | number;

function name(path: string) {
  return path.slice(path.lastIndexOf('/') + 1);
}

function quote(arg: string) {
  return /[\s'"$`\\]/.test(arg) ? `'${arg.replace(/'/g, "'\\''")}'` : arg;
}

/** The problems a file has, as the settings choose to see them. */
function check(session: TerminalSession, path: string, text: string, s: EditorSettings) {
  return session.lint(path, text).filter((d) => s.styleHints || d.source !== 'style');
}

export function EditorApp() {
  const [session, setSession] = useState<TerminalSession | null>(null);
  const [failed, setFailed] = useState<string | null>(null);
  const [attempt, setAttempt] = useState(0);
  const [tabs, setTabs] = useState<Tab[]>([]);
  const [active, setActive] = useState<string>();
  const [version, setVersion] = useState(0);
  const [terms, setTerms] = useState<Term[]>([{ id: 0 }]);
  const [bottom, setBottom] = useState<Bottom>(0);
  const [cwds, setCwds] = useState<Record<number, string>>({});
  const [problems, setProblems] = useState<Record<string, Diagnostic[]>>({});
  const [closing, setClosing] = useState<string[] | null>(null);
  const [files, setFiles] = useState(false);
  const nextTerm = useRef(1);
  const terminals = useRef(new Map<number, TerminalHandle>());
  const narrow = useMediaQuery('(max-width: 47.99rem)');
  const settings = useEditorSettings();

  const prefs = useMemo(
    () => ({
      wordWrap: settings.wordWrap,
      fontSize: settings.fontSize,
      whitespace: settings.styleHints
    }),
    [settings.wordWrap, settings.fontSize, settings.styleHints]
  );

  useEffect(() => {
    let live = true;
    terminalSession().then(
      (s) => live && setSession(s),
      (err: unknown) => live && setFailed(err instanceof Error ? err.message : String(err))
    );
    return () => {
      live = false;
    };
  }, [attempt]);

  const open = useCallback(
    (path: string) => {
      if (!session) return;
      setTabs((ts) => {
        if (ts.some((t) => t.path === path)) return ts;
        let text: string;
        try {
          text = readText(session.fs, path);
        } catch {
          return ts;
        }
        return [...ts, { path, text, saved: text }];
      });
      setActive(path);
    },
    [session]
  );

  /** The terminal a command goes to: the one shown, else the first, else a new one. */
  const terminalFor = useCallback((): number => {
    if (typeof bottom === 'number' && terms.some((t) => t.id === bottom)) return bottom;
    if (terms.length) {
      setBottom(terms[0].id);
      return terms[0].id;
    }
    const id = nextTerm.current++;
    setTerms([{ id }]);
    setBottom(id);
    return id;
  }, [bottom, terms]);

  const newTerminal = useCallback((dir?: string) => {
    const id = nextTerm.current++;
    setTerms((ts) => [...ts, { id, initial: dir ? `cd ${quote(dir)}` : undefined }]);
    setBottom(id);
  }, []);

  const closeTerminal = (id: number) => {
    session?.close(id);
    terminals.current.delete(id);
    setTerms((ts) => {
      const next = ts.filter((t) => t.id !== id);
      if (bottom === id) setBottom(next[next.length - 1]?.id ?? 'problems');
      return next;
    });
  };

  // The first files: what another page handed over, else the README and main.go.
  useEffect(() => {
    if (!session) return;
    session.onOpen(open);
    const h = takeHandoff();
    if (h) {
      const program = `${session.home}/${h.name}/main.go`;
      writeText(session.fs, program, h.main);
      if (h.file) open(`${session.library}/${h.file}`);
      open(program);
      const shown = terminals.current.get(0);
      if (shown) void shown.run(`cd ${quote(h.name)}`);
      else void session.run(`cd ${quote(h.name)}`, 0);
      setVersion((v) => v + 1);
    } else {
      open(`${session.home}/README.md`);
      open(`${session.home}/main.go`);
    }
    return () => session.onOpen(undefined);
  }, [session, open]);

  // Check the open files a moment after they stop changing.
  useEffect(() => {
    if (!session) return;
    const timer = setTimeout(() => {
      const next: Record<string, Diagnostic[]> = {};
      if (settings.lint)
        for (const t of tabs) next[t.path] = check(session, t.path, t.text, settings);
      setProblems(next);
    }, 300);
    return () => clearTimeout(timer);
  }, [session, tabs, settings]);

  const tab = tabs.find((t) => t.path === active);

  /** The file's text formatted, or the text as it is (and why) when it cannot be. */
  const formatted = useCallback(
    (path: string, text: string) => {
      if (!session) return text;
      const r = session.format(path, text);
      if (r.error) {
        toast.warning(`${name(path)} was not formatted`, {
          description: r.error.split('\n')[0]
        });
        return text;
      }
      if (r.text !== text) replaceEditorText(path, r.text);
      return r.text;
    },
    [session]
  );

  const save = useCallback(
    (path = active) => {
      if (!session || !path) return;
      const t = tabs.find((x) => x.path === path);
      if (!t) return;
      const text = settings.formatOnSave ? formatted(path, t.text) : t.text;
      writeText(session.fs, path, text);
      setTabs((ts) => ts.map((x) => (x.path === path ? { ...x, text, saved: text } : x)));
      setVersion((v) => v + 1);
    },
    [session, active, tabs, settings.formatOnSave, formatted]
  );

  const saveAll = () => {
    for (const t of tabs) if (t.text !== t.saved) save(t.path);
  };

  const formatActive = useCallback(() => {
    if (!tab) return;
    const text = formatted(tab.path, tab.text);
    if (text === tab.text) toast.success(`${name(tab.path)} is already formatted`);
  }, [tab, formatted]);

  const run = useCallback(async () => {
    if (!session || !tab) return;
    save(tab.path);
    if (settings.stopOnErrors) {
      const errors = session
        .lint(tab.path, readText(session.fs, tab.path))
        .filter((d) => d.severity === 'error');
      if (errors.length) {
        setBottom('problems');
        toast.error(
          `${name(tab.path)} has ${errors.length} error${errors.length === 1 ? '' : 's'}`,
          { description: 'Fix them to run it, or turn off Stop on errors in the settings.' }
        );
        return;
      }
    }
    const id = terminalFor();
    const line = runLine(tab.path, session.cwd(id));
    if (!line) return;
    // The terminal may have just been made: wait for it to mount.
    await new Promise(requestAnimationFrame);
    const handle = terminals.current.get(id);
    handle?.focus();
    await (handle ? handle.run(line) : session.run(line, id));
    setVersion((v) => v + 1);
  }, [session, tab, save, settings.stopOnErrors, terminalFor]);

  const discard = (paths: string[]) => {
    for (const p of paths) forgetEditorState(p);
    setTabs((ts) => {
      const next = ts.filter((t) => !paths.includes(t.path));
      if (active && paths.includes(active)) {
        // The neighbor of the closed tab, as a code editor picks it.
        const at = ts.findIndex((t) => t.path === active);
        const after = ts.slice(at + 1).find((t) => !paths.includes(t.path));
        setActive(after?.path ?? next[next.length - 1]?.path);
      }
      return next;
    });
  };

  /** Closes tabs, asking first about any with unsaved changes. */
  const close = (paths: string[]) => {
    const dirty = tabs.filter((t) => paths.includes(t.path) && t.text !== t.saved);
    if (dirty.length) setClosing(paths);
    else discard(paths);
  };

  if (failed) {
    return (
      <div className="bg-bg text-fg grid h-full place-items-center p-6 text-center">
        <div className="flex max-w-md flex-col items-center gap-4">
          <p className="font-heading text-xl font-semibold">The terminal could not start</p>
          <p className="text-muted text-sm">{failed}</p>
          <Button
            variant="primary"
            onClick={() => {
              setFailed(null);
              setAttempt((n) => n + 1);
            }}
          >
            <LuRotateCcw aria-hidden />
            Try again
          </Button>
        </div>
      </div>
    );
  }

  const shownTerminal = typeof bottom === 'number' ? bottom : (terms[0]?.id ?? 0);
  const cwd = session ? (cwds[shownTerminal] ?? session.cwd(shownTerminal)) : '';
  const runnable = tab && session ? runLine(tab.path, cwd) !== undefined : false;
  const language = tab ? languageOf(tab.path) : undefined;
  const isGo = language === 'go';
  const count = countProblems(problems);
  const activeProblems = tab ? problems[tab.path] : undefined;
  const dirtyClosing = closing
    ? tabs.filter((t) => closing.includes(t.path) && t.text !== t.saved)
    : [];

  // The three parts, laid out side by side on a wide screen and stacked on a phone.
  const explorerPane = session ? (
    <Explorer
      fs={session.fs}
      roots={[
        { path: session.home, label: 'Home' },
        { path: session.library, label: 'alGOrithm/pkg' }
      ]}
      version={version}
      active={active}
      onOpen={(p) => {
        open(p);
        setFiles(false);
      }}
      onCreate={(path, folder) => {
        try {
          if (folder) session.fs.mkdirpSync(path);
          else {
            writeText(session.fs, path, '');
            open(path);
          }
        } catch (err) {
          toast.error(`Could not make ${name(path)}`, {
            description: err instanceof Error ? err.message : String(err)
          });
        }
        setVersion((v) => v + 1);
      }}
      onRename={(from, to) => {
        try {
          renamePath(session.fs, from, to);
        } catch (err) {
          toast.error(`Could not rename ${name(from)}`, {
            description: err instanceof Error ? err.message : String(err)
          });
          return;
        }
        const moved = (p: string) =>
          p === from || p.startsWith(`${from}/`) ? to + p.slice(from.length) : p;
        for (const t of tabs) if (moved(t.path) !== t.path) moveEditorState(t.path, moved(t.path));
        setTabs((ts) => ts.map((t) => ({ ...t, path: moved(t.path) })));
        if (active) setActive(moved(active));
        setVersion((v) => v + 1);
      }}
      onDelete={(path) => {
        removePath(session.fs, path);
        discard(tabs.map((t) => t.path).filter((p) => p === path || p.startsWith(`${path}/`)));
        setVersion((v) => v + 1);
      }}
      onDuplicate={(path) => {
        const base = name(path);
        const dot = base.lastIndexOf('.');
        const dir = path.slice(0, path.length - base.length);
        const stem = dir + (dot > 0 ? base.slice(0, dot) : base);
        const ext = dot > 0 ? base.slice(dot) : '';
        let copy = `${stem} copy${ext}`;
        for (let n = 2; session.fs.existsSync(copy); n++) copy = `${stem} copy ${n}${ext}`;
        writeText(session.fs, copy, readText(session.fs, path));
        open(copy);
        setVersion((v) => v + 1);
      }}
      onTerminal={(dir) => newTerminal(dir)}
      onRefresh={() => setVersion((v) => v + 1)}
    />
  ) : null;

  const codePane = session ? (
    <div
      data-tour="editor-code"
      className="flex h-full min-h-0 flex-col"
    >
      <div
        role="tablist"
        aria-label="Open files"
        className="border-border bg-subtle/40 scroll-quiet flex shrink-0 overflow-x-auto border-b"
      >
        {tabs.map((t, i) => {
          const on = t.path === active;
          const dirty = t.text !== t.saved;
          const errs = problems[t.path]?.filter((d) => d.severity === 'error').length ?? 0;
          return (
            <ContextMenu key={t.path}>
              <ContextMenuTrigger
                className={cn(
                  'border-border flex shrink-0 items-center border-e text-[13px]',
                  on ? 'bg-bg text-fg' : 'text-muted hover:text-fg'
                )}
              >
                {/* ui-ok: an editor tab, as a code editor's tab strip has */}
                <button
                  type="button"
                  role="tab"
                  aria-selected={on}
                  onClick={() => setActive(t.path)}
                  onAuxClick={(e) => {
                    if (e.button === 1) close([t.path]);
                  }}
                  title={t.path}
                  className={cn(
                    'flex items-center gap-1.5 py-2 ps-3 pe-1 outline-none focus-visible:underline',
                    on && 'shadow-[inset_0_2px_0_var(--color-primary)]',
                    errs > 0 && 'text-danger'
                  )}
                >
                  <FileIcon path={t.path} />
                  {name(t.path)}
                  {errs > 0 ? <span className="sr-only">, {errs} errors</span> : null}
                </button>
                <Button
                  variant="plain"
                  size="icon-xs"
                  aria-label={
                    dirty ? `${name(t.path)} is not saved; close it` : `Close ${name(t.path)}`
                  }
                  onClick={() => close([t.path])}
                  className="group/close me-1"
                >
                  {dirty ? (
                    <>
                      <LuCircle
                        aria-hidden
                        className="size-2.5 fill-current group-hover/close:hidden"
                      />
                      <LuX
                        aria-hidden
                        className="hidden group-hover/close:block"
                      />
                    </>
                  ) : (
                    <LuX aria-hidden />
                  )}
                </Button>
              </ContextMenuTrigger>
              <ContextMenuContent className="min-w-52">
                <ContextMenuItem onClick={() => close([t.path])}>Close</ContextMenuItem>
                <ContextMenuItem
                  disabled={tabs.length < 2}
                  onClick={() => close(tabs.filter((x) => x.path !== t.path).map((x) => x.path))}
                >
                  Close others
                </ContextMenuItem>
                <ContextMenuItem
                  disabled={i === tabs.length - 1}
                  onClick={() => close(tabs.slice(i + 1).map((x) => x.path))}
                >
                  Close to the right
                </ContextMenuItem>
                <ContextMenuItem
                  onClick={() => discard(tabs.filter((x) => x.text === x.saved).map((x) => x.path))}
                >
                  Close saved
                </ContextMenuItem>
                <ContextMenuItem onClick={() => close(tabs.map((x) => x.path))}>
                  Close all
                </ContextMenuItem>
                <ContextMenuSeparator />
                <ContextMenuItem
                  disabled={!dirty}
                  onClick={() => save(t.path)}
                >
                  Save
                </ContextMenuItem>
                <ContextMenuItem
                  onClick={() => {
                    void navigator.clipboard.writeText(t.path);
                    toast.success('Copied', { description: t.path });
                  }}
                >
                  Copy path
                </ContextMenuItem>
                <ContextMenuItem
                  onClick={() => newTerminal(t.path.slice(0, t.path.lastIndexOf('/')))}
                >
                  Open its folder in a terminal
                </ContextMenuItem>
              </ContextMenuContent>
            </ContextMenu>
          );
        })}
      </div>
      <div className="min-h-0 flex-1">
        {tab ? (
          <CodeEditor
            path={tab.path}
            doc={tab.text}
            language={languageOf(tab.path)}
            diagnostics={activeProblems}
            prefs={prefs}
            onChange={(text) =>
              setTabs((ts) => ts.map((x) => (x.path === tab.path ? { ...x, text } : x)))
            }
            onSave={() => save(tab.path)}
            onRun={() => void run()}
            onFormat={formatActive}
          />
        ) : (
          <NoFile
            fs={session.fs}
            home={session.home}
            library={session.library}
            onOpen={open}
            onNew={() => {
              let path = `${session.home}/untitled.go`;
              for (let n = 2; session.fs.existsSync(path); n++)
                path = `${session.home}/untitled-${n}.go`;
              writeText(session.fs, path, 'package main\n\nfunc main() {\n}\n');
              open(path);
              setVersion((v) => v + 1);
            }}
          />
        )}
      </div>
    </div>
  ) : null;

  const bottomPane = session ? (
    <section
      aria-label="Problems and terminals"
      data-tour="editor-terminal"
      className="border-border flex h-full min-h-0 flex-col border-t"
    >
      <div className="flex shrink-0 items-center gap-1 px-1 py-0.5">
        <div
          role="tablist"
          aria-label="Panels"
          className="scroll-quiet flex min-w-0 flex-1 items-center gap-0.5 overflow-x-auto"
        >
          {/* ui-ok: the panel's tabs, as a code editor's bottom panel has */}
          <button
            type="button"
            role="tab"
            aria-selected={bottom === 'problems'}
            onClick={() => setBottom('problems')}
            className={cn(
              'flex shrink-0 items-center gap-1.5 rounded-sm px-2 py-1 text-[11px] font-semibold tracking-widest uppercase outline-none focus-visible:underline',
              bottom === 'problems' ? 'bg-fg/8 text-fg' : 'text-muted hover:text-fg'
            )}
          >
            Problems
            {count.errors + count.warnings > 0 ? (
              <span className="bg-subtle text-fg rounded-full px-1.5 tracking-normal">
                {count.errors + count.warnings}
              </span>
            ) : null}
          </button>
          {terms.map((t, i) => (
            <span
              key={t.id}
              className={cn(
                'group/term flex shrink-0 items-center rounded-sm',
                bottom === t.id ? 'bg-fg/8 text-fg' : 'text-muted hover:text-fg'
              )}
            >
              {/* ui-ok: a terminal's tab in the panel's tab strip */}
              <button
                type="button"
                role="tab"
                aria-selected={bottom === t.id}
                onClick={() => {
                  setBottom(t.id);
                  requestAnimationFrame(() => terminals.current.get(t.id)?.focus());
                }}
                className="py-1 ps-2 pe-1 text-[11px] font-semibold tracking-widest uppercase outline-none focus-visible:underline"
              >
                Terminal {i + 1}
              </button>
              <Button
                variant="plain"
                size="icon-xs"
                aria-label={`Close terminal ${i + 1}`}
                onClick={() => closeTerminal(t.id)}
                className="me-0.5 opacity-60 group-hover/term:opacity-100"
              >
                <LuX aria-hidden />
              </Button>
            </span>
          ))}
        </div>
        <Button
          variant="plain"
          size="icon-xs"
          aria-label="New terminal"
          title="New terminal"
          onClick={() => newTerminal()}
        >
          <LuPlus aria-hidden />
        </Button>
        <Button
          variant="plain"
          size="icon-xs"
          aria-label="Clear the terminal"
          title="Clear the terminal (Ctrl+L)"
          disabled={typeof bottom !== 'number'}
          onClick={() => {
            if (typeof bottom === 'number') terminals.current.get(bottom)?.clear();
          }}
        >
          <LuTrash2 aria-hidden />
        </Button>
      </div>
      <div className="relative min-h-0 flex-1">
        <div className={cn('absolute inset-0', bottom !== 'problems' && 'hidden')}>
          <ProblemsPanel
            problems={problems}
            home={session.home}
            onGo={(path, line, col) => {
              revealInEditor(path, line, col);
              open(path);
            }}
          />
        </div>
        {terms.map((t) => (
          <div
            key={t.id}
            className={cn('absolute inset-0', bottom !== t.id && 'hidden')}
          >
            <TerminalPanel
              ref={(h) => {
                if (h) terminals.current.set(t.id, h);
                else terminals.current.delete(t.id);
              }}
              session={session}
              terminal={t.id}
              initial={t.initial}
              onCwd={(dir) => setCwds((c) => ({ ...c, [t.id]: dir }))}
            />
          </div>
        ))}
        {terms.length === 0 && bottom !== 'problems' ? (
          <div className="grid h-full place-items-center">
            <Button
              variant="ghost"
              size="sm"
              onClick={() => newTerminal()}
            >
              <LuPlus aria-hidden />
              New terminal
            </Button>
          </div>
        ) : null}
      </div>
    </section>
  ) : null;

  return (
    <div className="bg-bg text-fg flex min-h-dvh flex-col md:h-full md:min-h-0 md:overflow-hidden">
      <header className="border-border bg-bg sticky top-0 z-20 flex h-11 shrink-0 items-center gap-2 border-b px-2">
        {narrow && session ? (
          <Button
            variant="ghost"
            size="icon-sm"
            aria-label="Files"
            title="Files"
            onClick={() => setFiles(true)}
          >
            <LuFolderTree aria-hidden />
          </Button>
        ) : null}
        <Link
          href="/"
          aria-label="alGOrithm: back to the site"
          className="focus-visible:ring-accent flex items-center gap-2 rounded-md px-1 outline-none focus-visible:ring-2"
        >
          <LogoColor className="size-6" />
          <span className="hidden text-sm font-semibold sm:inline">Editor</span>
        </Link>
        <span className="text-muted min-w-0 flex-1 truncate font-mono text-xs">
          {tab ? tab.path.replace(session?.home ?? '', '~') : ''}
        </span>
        {isGo ? (
          <span className="text-muted hidden text-xs lg:inline">Go runs on the Go Playground</span>
        ) : null}
        <Button
          variant="primary"
          size="sm"
          onClick={() => void run()}
          disabled={!runnable}
          title={
            isGo
              ? 'Run (Ctrl+Enter): sent to the Go Playground, play.golang.org'
              : 'Run (Ctrl+Enter)'
          }
          data-tour="editor-run"
        >
          <LuPlay aria-hidden />
          Run
        </Button>
        <Button
          variant="ghost"
          size="icon-sm"
          onClick={() => save()}
          disabled={!tab || tab.text === tab.saved}
          aria-label="Save (Ctrl+S)"
          title="Save (Ctrl+S)"
        >
          <LuSave aria-hidden />
        </Button>
        <DropdownMenu>
          <DropdownMenuTrigger
            render={
              <Button
                variant="ghost"
                size="icon-sm"
                aria-label="Editor settings"
                title="Editor settings"
                data-tour="editor-settings"
              />
            }
          >
            <LuSettings aria-hidden />
          </DropdownMenuTrigger>
          <DropdownMenuContent
            align="end"
            className="min-w-64"
          >
            <DropdownMenuGroup>
              <DropdownMenuLabel>On save and run</DropdownMenuLabel>
              <DropdownMenuCheckboxItem
                checked={settings.formatOnSave}
                onCheckedChange={(v) => settings.set({ formatOnSave: v })}
              >
                Format on save
              </DropdownMenuCheckboxItem>
              <DropdownMenuCheckboxItem
                checked={settings.stopOnErrors}
                onCheckedChange={(v) => settings.set({ stopOnErrors: v })}
              >
                Stop on errors when running
              </DropdownMenuCheckboxItem>
            </DropdownMenuGroup>
            <DropdownMenuSeparator />
            <DropdownMenuGroup>
              <DropdownMenuLabel>Checks</DropdownMenuLabel>
              <DropdownMenuCheckboxItem
                checked={settings.lint}
                onCheckedChange={(v) => settings.set({ lint: v })}
              >
                Flag problems as I type
              </DropdownMenuCheckboxItem>
              <DropdownMenuCheckboxItem
                checked={settings.styleHints}
                onCheckedChange={(v) => settings.set({ styleHints: v })}
              >
                Style hints (trailing spaces, final newline)
              </DropdownMenuCheckboxItem>
            </DropdownMenuGroup>
            <DropdownMenuSeparator />
            <DropdownMenuGroup>
              <DropdownMenuLabel>View</DropdownMenuLabel>
              <DropdownMenuCheckboxItem
                checked={settings.wordWrap}
                onCheckedChange={(v) => settings.set({ wordWrap: v })}
              >
                Word wrap
              </DropdownMenuCheckboxItem>
              <DropdownMenuRadioGroup
                value={String(settings.fontSize)}
                onValueChange={(v) =>
                  settings.set({ fontSize: Number(v) as EditorSettings['fontSize'] })
                }
              >
                {[12, 13, 14, 16].map((n) => (
                  <DropdownMenuRadioItem
                    key={n}
                    value={String(n)}
                  >
                    {n}px text
                  </DropdownMenuRadioItem>
                ))}
              </DropdownMenuRadioGroup>
            </DropdownMenuGroup>
            <DropdownMenuSeparator />
            <DropdownMenuItem
              disabled={!tab}
              onClick={formatActive}
            >
              <LuWandSparkles aria-hidden />
              Format this file
              <DropdownMenuShortcut>Shift+Alt+F</DropdownMenuShortcut>
            </DropdownMenuItem>
            <DropdownMenuItem
              disabled={!tabs.some((t) => t.text !== t.saved)}
              onClick={saveAll}
            >
              <LuSave aria-hidden />
              Save all
            </DropdownMenuItem>
            <DropdownMenuItem
              onClick={() => {
                forgetSaved();
                toast.success('Forgotten', {
                  description: 'The next visit starts from the starting files.'
                });
              }}
            >
              <LuTrash2 aria-hidden />
              Forget my files
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => settings.reset()}>
              <LuRotateCcw aria-hidden />
              Reset the settings
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
        <ThemeToggle variant="diagonal" />
        <Button
          variant="ghost"
          size="sm"
          render={<Link href="/" />}
          className="hidden md:inline-flex"
        >
          <LuArrowLeft aria-hidden />
          The site
        </Button>
      </header>

      {!session ? (
        <div className="flex flex-1 flex-col gap-3 p-4">
          <Skeleton className="h-full w-full rounded-xl" />
          <p className="text-muted text-center text-sm">Starting the terminal...</p>
        </div>
      ) : narrow ? (
        // A phone: the code, then the panel, down a page that scrolls; the
        // files in a drawer.
        <div className="flex flex-col">
          <div className="flex h-[62dvh] min-h-80 flex-col">{codePane}</div>
          <div className="h-[48dvh] min-h-72">{bottomPane}</div>
          <Drawer
            open={files}
            onOpenChange={setFiles}
          >
            <DrawerPopup
              showBar
              showCloseButton
            >
              <DrawerHeader>
                <DrawerTitle>Files</DrawerTitle>
              </DrawerHeader>
              <DrawerPanel className="h-[60dvh] px-0">{explorerPane}</DrawerPanel>
            </DrawerPopup>
          </Drawer>
        </div>
      ) : (
        <ResizablePanelGroup
          orientation="horizontal"
          className="min-h-0 flex-1"
        >
          <ResizablePanel
            defaultSize="22%"
            minSize="12%"
            className="bg-subtle/40"
          >
            {explorerPane}
          </ResizablePanel>
          <ResizableHandle />
          <ResizablePanel minSize="30%">
            <ResizablePanelGroup orientation="vertical">
              <ResizablePanel
                defaultSize="64%"
                minSize="15%"
              >
                {codePane}
              </ResizablePanel>
              <ResizableHandle />
              <ResizablePanel
                defaultSize="36%"
                minSize="12%"
              >
                {bottomPane}
              </ResizablePanel>
            </ResizablePanelGroup>
          </ResizablePanel>
        </ResizablePanelGroup>
      )}

      <footer className="border-border text-muted flex h-6 shrink-0 items-center gap-3 border-t px-2 text-[11px]">
        {/* ui-ok: the status bar's problem count, as a code editor's opens its panel */}
        <button
          type="button"
          onClick={() => setBottom('problems')}
          className="hover:text-fg flex items-center gap-2 rounded-sm px-1 outline-none focus-visible:underline"
          aria-label={`${count.errors} errors, ${count.warnings} warnings: show the problems`}
        >
          <span className="flex items-center gap-1">
            <SeverityIcon severity="error" />
            {count.errors}
          </span>
          <span className="flex items-center gap-1">
            <SeverityIcon severity="warning" />
            {count.warnings}
          </span>
        </button>
        <span>
          {tab && tab.text !== tab.saved ? 'unsaved' : tab ? 'saved in this browser' : ''}
        </span>
        <span className="ms-auto hidden sm:inline">
          {settings.formatOnSave ? 'Format on save' : 'No format on save'}
        </span>
        <span>{language ? LANGUAGE_NAMES[language] : ''}</span>
        <span className="font-mono">{session ? cwd.replace(session.home, '~') : ''}</span>
      </footer>

      <AlertDialog
        open={closing !== null}
        onOpenChange={(o) => {
          if (!o) setClosing(null);
        }}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>
              {dirtyClosing.length === 1
                ? `Save the changes to ${name(dirtyClosing[0].path)}?`
                : `Save the changes to ${dirtyClosing.length} files?`}
            </AlertDialogTitle>
            <AlertDialogDescription>
              Your changes are lost if you close without saving.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <Button
              variant="ghost"
              onClick={() => {
                if (closing) discard(closing);
                setClosing(null);
              }}
            >
              Don&apos;t save
            </Button>
            <Button
              variant="primary"
              onClick={() => {
                for (const t of dirtyClosing) save(t.path);
                if (closing) discard(closing);
                setClosing(null);
              }}
            >
              Save
            </Button>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
