'use client';

/**
 * The editor itself: CodeMirror 6, dressed like the site's code block (the
 * same surface, mono face and line numbers), editable. Go, JSON and Markdown
 * have full parsers; shell, Makefiles, YAML, TOML, Dockerfiles, JavaScript,
 * TypeScript, CSS, HTML, Python, Rust, SQL and the rest have CodeMirror's
 * stream modes. Tokens get classes styled in styles/editor.css, in the site's
 * palette for light and dark. Each open file keeps its own state (its undo
 * history, its selection, scroll and flags), so switching tabs loses nothing.
 *
 * Problems come from the terminal's checks (web/runtime/tooling) as a prop,
 * and show as underlines, gutter flags and a list (Ctrl/Cmd+Shift+M).
 *
 * Keys: Ctrl/Cmd+S saves, Ctrl/Cmd+Enter runs, Shift+Alt+F formats, F8 goes
 * to the next problem, Tab indents (with a tab in Go and Makefiles).
 */
import { useEffect, useRef } from 'react';
import { defaultKeymap, history, historyKeymap, indentWithTab } from '@codemirror/commands';
import { go } from '@codemirror/lang-go';
import { json } from '@codemirror/lang-json';
import { markdown } from '@codemirror/lang-markdown';
import {
  bracketMatching,
  foldGutter,
  indentOnInput,
  indentUnit,
  StreamLanguage,
  syntaxHighlighting,
  type StreamParser
} from '@codemirror/language';
import { css } from '@codemirror/legacy-modes/mode/css';
import { diff } from '@codemirror/legacy-modes/mode/diff';
import { dockerFile } from '@codemirror/legacy-modes/mode/dockerfile';
import { javascript, typescript } from '@codemirror/legacy-modes/mode/javascript';
import { lua } from '@codemirror/legacy-modes/mode/lua';
import { properties } from '@codemirror/legacy-modes/mode/properties';
import { protobuf } from '@codemirror/legacy-modes/mode/protobuf';
import { python } from '@codemirror/legacy-modes/mode/python';
import { ruby } from '@codemirror/legacy-modes/mode/ruby';
import { rust } from '@codemirror/legacy-modes/mode/rust';
import { shell } from '@codemirror/legacy-modes/mode/shell';
import { standardSQL } from '@codemirror/legacy-modes/mode/sql';
import { toml } from '@codemirror/legacy-modes/mode/toml';
import { html, xml } from '@codemirror/legacy-modes/mode/xml';
import { yaml } from '@codemirror/legacy-modes/mode/yaml';
import { lintGutter, lintKeymap, setDiagnostics, type Diagnostic as Flag } from '@codemirror/lint';
import { Compartment, EditorState, type Extension, type Text } from '@codemirror/state';
import {
  crosshairCursor,
  drawSelection,
  EditorView,
  highlightActiveLine,
  highlightActiveLineGutter,
  highlightSpecialChars,
  highlightTrailingWhitespace,
  keymap,
  lineNumbers,
  rectangularSelection
} from '@codemirror/view';
import { classHighlighter } from '@lezer/highlight';
import type { Language } from '@/editor/files';
import type { Diagnostic } from '@/editor/terminal-runtime';

/** A Makefile: targets, variables and their references, directives, recipes. */
const makefile: StreamParser<{ recipe: boolean }> = {
  name: 'makefile',
  startState: () => ({ recipe: false }),
  token(stream, state) {
    if (stream.sol()) state.recipe = stream.peek() === '\t';
    if (stream.match(/^\$\$/)) return null;
    if (stream.match(/^\$[({][^)}]*[)}]/) || stream.match(/^\$[@<^?*%+|]/)) return 'variable-2';
    if (state.recipe) {
      if (stream.sol() && stream.match(/^\t[@\-+]+/)) return 'meta';
      if (stream.match(/^"(?:[^"\\]|\\.)*"?/) || stream.match(/^'[^']*'?/)) return 'string';
      if (stream.match(/^#.*/)) return 'comment';
      stream.next();
      return null;
    }
    if (stream.match(/^#.*/)) return 'comment';
    if (
      stream.sol() &&
      stream.match(
        /^(?:include|-include|sinclude|ifeq|ifneq|ifdef|ifndef|else|endif|define|endef|export|unexport|override|vpath)\b/
      )
    )
      return 'keyword';
    if (stream.sol() && stream.match(/^[A-Za-z_][\w.]*(?=\s*(?:::?=|\?=|\+=|!=|=))/))
      return 'variable';
    if (stream.sol() && stream.match(/^\.[A-Z_]+(?=\s*:)/)) return 'keyword';
    if (stream.sol() && stream.match(/^[^\s:=#][^:=#]*(?=:(?!=))/)) return 'def';
    if (stream.match(/^(?:::?=|\?=|\+=|!=|=|::?|;)/)) return 'operator';
    if (stream.match(/^"(?:[^"\\]|\\.)*"?/)) return 'string';
    stream.next();
    return null;
  }
};

function languageExtension(lang: Language): Extension {
  const stream = (p: StreamParser<unknown>) => StreamLanguage.define(p);
  switch (lang) {
    case 'go':
      return go();
    case 'markdown':
      return markdown();
    case 'json':
      return json();
    case 'shell':
      return stream(shell);
    case 'make':
      return StreamLanguage.define(makefile);
    case 'yaml':
      return stream(yaml);
    case 'toml':
      return stream(toml);
    case 'dockerfile':
      return stream(dockerFile);
    case 'javascript':
      return stream(javascript);
    case 'typescript':
      return stream(typescript);
    case 'css':
      return stream(css);
    case 'html':
      return stream(html);
    case 'xml':
      return stream(xml);
    case 'python':
      return stream(python);
    case 'rust':
      return stream(rust);
    case 'sql':
      return stream(standardSQL);
    case 'diff':
      return stream(diff);
    case 'properties':
      return stream(properties);
    case 'protobuf':
      return stream(protobuf);
    case 'lua':
      return stream(lua);
    case 'ruby':
      return stream(ruby);
    default:
      return [];
  }
}

/** Tabs where the language needs them (gofmt, make), spaces elsewhere. */
function indentFor(lang: Language): string {
  if (lang === 'go' || lang === 'make') return '\t';
  return lang === 'python' || lang === 'rust' ? '    ' : '  ';
}

/** The editor's look, on the site's tokens (light and dark follow the page). */
const theme = EditorView.theme({
  '&': { height: '100%', backgroundColor: 'transparent', color: 'var(--fg)' },
  '.cm-scroller': { fontFamily: 'var(--font-mono)', lineHeight: '1.6' },
  '.cm-content': { caretColor: 'var(--color-primary)', padding: '10px 0' },
  '.cm-cursor': { borderLeftColor: 'var(--color-primary)', borderLeftWidth: '2px' },
  '.cm-gutters': {
    backgroundColor: 'transparent',
    color: 'color-mix(in oklab, var(--muted) 70%, transparent)',
    border: 'none'
  },
  '.cm-activeLine': { backgroundColor: 'color-mix(in oklab, var(--fg) 5%, transparent)' },
  '.cm-activeLineGutter': { backgroundColor: 'transparent', color: 'var(--fg)' },
  '&.cm-focused .cm-selectionBackground, .cm-selectionBackground, ::selection': {
    backgroundColor: 'color-mix(in oklab, var(--color-primary) 25%, transparent) !important'
  },
  '.cm-matchingBracket': {
    backgroundColor: 'color-mix(in oklab, var(--color-primary) 20%, transparent)',
    outline: 'none'
  },
  '.cm-tooltip': {
    backgroundColor: 'var(--card)',
    color: 'var(--fg)',
    border: '1px solid color-mix(in oklab, var(--fg) 12%, transparent)',
    borderRadius: '8px',
    overflow: 'hidden'
  },
  '.cm-diagnostic': { fontFamily: 'var(--font-mono)', fontSize: '12px', padding: '4px 8px' },
  '.cm-panels': { backgroundColor: 'var(--card)', color: 'var(--fg)' },
  '&.cm-focused': { outline: 'none' }
});

/** What the settings change, reconfigured without losing the file's state. */
export type EditorPrefs = { wordWrap: boolean; fontSize: number; whitespace: boolean };

const prefsCompartment = new Compartment();

const DEFAULT_PREFS: EditorPrefs = { wordWrap: false, fontSize: 13, whitespace: false };

// p may be missing for an editor state made before a hot reload added it.
function prefsExtension(p: EditorPrefs = DEFAULT_PREFS): Extension {
  return [
    p.wordWrap ? EditorView.lineWrapping : [],
    p.whitespace ? highlightTrailingWhitespace() : [],
    EditorView.theme({ '&': { fontSize: `${p.fontSize}px` } })
  ];
}

type Handlers = {
  onChange: (text: string) => void;
  onSave: () => void;
  onRun: () => void;
  onFormat: () => void;
};

function extensions(lang: Language, h: () => Handlers, prefs: EditorPrefs): Extension[] {
  const command = (run: () => void) => () => {
    run();
    return true;
  };
  return [
    lineNumbers(),
    lintGutter(),
    highlightActiveLineGutter(),
    foldGutter(),
    highlightSpecialChars(),
    history(),
    drawSelection(),
    indentOnInput(),
    bracketMatching(),
    rectangularSelection(),
    crosshairCursor(),
    highlightActiveLine(),
    indentUnit.of(indentFor(lang)),
    EditorState.tabSize.of(4),
    syntaxHighlighting(classHighlighter),
    languageExtension(lang),
    theme,
    prefsCompartment.of(prefsExtension(prefs)),
    keymap.of([
      { key: 'Mod-s', preventDefault: true, run: command(() => h().onSave()) },
      { key: 'Mod-Enter', preventDefault: true, run: command(() => h().onRun()) },
      { key: 'Shift-Alt-f', preventDefault: true, run: command(() => h().onFormat()) },
      indentWithTab,
      ...lintKeymap,
      ...defaultKeymap,
      ...historyKeymap
    ]),
    EditorView.updateListener.of((u) => {
      if (u.docChanged) h().onChange(u.state.doc.toString());
    })
  ];
}

/** Each open file's editor state, kept while its tab is open (one editor a page). */
const states = new Map<string, EditorState>();
/** The editor on the page, and the file it shows. */
let live: { view: EditorView; path: string } | null = null;

/** Drops a closed file's state, so reopening it reads the file afresh. */
export function forgetEditorState(path: string): void {
  states.delete(path);
}

/** Moves a file's state to its new path, after a rename. */
export function moveEditorState(from: string, to: string): void {
  const s = states.get(from);
  if (s) {
    states.delete(from);
    states.set(to, s);
  }
  if (live?.path === from) live.path = to;
}

/**
 * Replaces a file's text (after formatting it), changing only the part that
 * differs, so the cursor, the undo history and the scroll stay where they were.
 */
export function replaceEditorText(path: string, text: string): void {
  const apply = (doc: Text) => {
    const old = doc.toString();
    if (old === text) return null;
    let start = 0;
    while (start < old.length && start < text.length && old[start] === text[start]) start++;
    let end = 0;
    while (
      end < old.length - start &&
      end < text.length - start &&
      old[old.length - 1 - end] === text[text.length - 1 - end]
    )
      end++;
    return { from: start, to: old.length - end, insert: text.slice(start, text.length - end) };
  };
  if (live?.path === path) {
    const change = apply(live.view.state.doc);
    if (change) live.view.dispatch({ changes: change });
    return;
  }
  const s = states.get(path);
  const change = s && apply(s.doc);
  if (s && change) states.set(path, s.update({ changes: change }).state);
}

/** Where a file's cursor goes when it is next shown, from the problems list. */
const gotos = new Map<string, { line: number; col: number }>();

function goTo(v: EditorView, line: number, col: number) {
  const doc = v.state.doc;
  const l = doc.line(Math.min(Math.max(1, line), doc.lines));
  const pos = l.from + Math.min(Math.max(0, col - 1), l.length);
  v.dispatch({
    selection: { anchor: pos },
    effects: EditorView.scrollIntoView(pos, { y: 'center' })
  });
  v.focus();
}

/** Puts the cursor at a line and column of a file, now or when it is shown. */
export function revealInEditor(path: string, line: number, col: number): void {
  if (live?.path === path) goTo(live.view, line, col);
  else gotos.set(path, { line, col });
}

/** The terminal's problems as CodeMirror's, on the document as it is now. */
function toFlags(doc: Text, diagnostics: Diagnostic[]): Flag[] {
  const at = (line: number, col: number) => {
    const l = doc.line(Math.min(Math.max(1, line), doc.lines));
    return l.from + Math.min(Math.max(0, col - 1), l.length);
  };
  return diagnostics.map((d) => {
    const from = at(d.line, d.col);
    let to: number;
    if (d.endLine) to = Math.max(from, at(d.endLine, d.endCol ?? 1));
    else {
      // The word the problem points at, or one character.
      const lineEnd = doc.lineAt(from).to;
      const rest = doc.sliceString(from, lineEnd);
      const word = /^[\w.]+/.exec(rest)?.[0].length ?? (rest.length ? 1 : 0);
      to = from + word;
    }
    return {
      from,
      to,
      severity: d.severity,
      message: d.message,
      source: d.source
    };
  });
}

export function CodeEditor({
  path,
  doc,
  language,
  diagnostics,
  prefs,
  onChange,
  onSave,
  onRun,
  onFormat
}: {
  /** The open file; each keeps its own state. */
  path: string;
  /** The file's text when it is first opened. */
  doc: string;
  language: Language;
  /** The file's problems, from the latest check. */
  diagnostics?: Diagnostic[];
  prefs: EditorPrefs;
} & Handlers) {
  const host = useRef<HTMLDivElement>(null);
  const view = useRef<EditorView | null>(null);
  // The handlers change every render; the editor's keymap calls the latest.
  const handlers = useRef<Handlers>({ onChange, onSave, onRun, onFormat });
  useEffect(() => {
    handlers.current = { onChange, onSave, onRun, onFormat };
  });
  const prefsRef = useRef(prefs);
  useEffect(() => {
    prefsRef.current = prefs;
  });

  useEffect(() => {
    if (!host.current) return;
    const v = new EditorView({ parent: host.current });
    view.current = v;
    return () => {
      if (live?.view === v) live = null;
      v.destroy();
      view.current = null;
    };
  }, []);

  // Show the file's own state (made on first opening), and keep it when
  // another file is shown.
  useEffect(() => {
    const v = view.current;
    if (!v) return;
    v.setState(
      states.get(path) ??
        EditorState.create({
          doc,
          extensions: extensions(language, () => handlers.current, prefsRef.current)
        })
    );
    v.dispatch({ effects: prefsCompartment.reconfigure(prefsExtension(prefsRef.current)) });
    live = { view: v, path };
    const g = gotos.get(path);
    gotos.delete(path);
    if (g) goTo(v, g.line, g.col);
    else v.focus();
    return () => {
      if (live?.view === v && live.path === path) states.set(path, v.state);
    };
    // Only a new path opens a file; its text after that lives in its state.
  }, [path]);

  useEffect(() => {
    view.current?.dispatch({ effects: prefsCompartment.reconfigure(prefsExtension(prefs)) });
  }, [prefs]);

  useEffect(() => {
    const v = view.current;
    if (!v || live?.path !== path) return;
    v.dispatch(setDiagnostics(v.state, toFlags(v.state.doc, diagnostics ?? [])));
  }, [diagnostics, path]);

  return (
    <div
      ref={host}
      className="editor-surface h-full min-h-0 overflow-hidden"
    />
  );
}
