/**
 * Keyboard shortcuts as data. An app keeps one list of `Shortcut`s (the site's
 * is web/src/lib/shortcuts.ts) and the help overlay, the tooltips'
 * key hints, the trainer and the docs all read it, so they cannot drift apart.
 *
 * Keys use react-hotkeys-hook's syntax: `mod+k` (Command on a Mac, Ctrl
 * elsewhere), `shift+arrowright`, a sequence as `g>h` (g, then h), and
 * alternatives separated by commas (`1,2,3,4`).
 */

export type ShortcutScope = 'global' | 'player' | 'quiz' | 'list' | 'palette';

export type Shortcut = {
  /** A stable name, such as "palette.open". */
  id: string;
  /** react-hotkeys-hook keys: "mod+k", "g>l", "shift+arrowleft". */
  keys: string;
  /** What it does, as a verb phrase: "Open the command palette". */
  description: string;
  scope: ShortcutScope;
  /** A heading in the help: "Navigation", "Playback", "Answering". */
  category: string;
};

export const scopeNames: Record<ShortcutScope, string> = {
  global: 'Everywhere',
  player: 'Player',
  quiz: 'Quiz',
  list: 'Lists',
  palette: 'Command palette'
};

export type Platform = 'apple' | 'pc';

/** Whether the visitor is on an Apple device, for ⌘ against Ctrl. */
export function detectPlatform(nav?: { platform?: string; userAgent?: string }): Platform {
  const n = nav ?? (typeof navigator === 'undefined' ? undefined : navigator);
  const s = `${n?.platform ?? ''} ${n?.userAgent ?? ''}`;
  return /Mac|iPhone|iPad|iPod/i.test(s) ? 'apple' : 'pc';
}

const names: Record<string, { apple: string; pc: string }> = {
  mod: { apple: '⌘', pc: 'Ctrl' },
  meta: { apple: '⌘', pc: 'Win' },
  ctrl: { apple: '⌃', pc: 'Ctrl' },
  alt: { apple: '⌥', pc: 'Alt' },
  shift: { apple: '⇧', pc: 'Shift' },
  enter: { apple: '↩', pc: 'Enter' },
  escape: { apple: 'Esc', pc: 'Esc' },
  space: { apple: 'Space', pc: 'Space' },
  arrowleft: { apple: '←', pc: '←' },
  arrowright: { apple: '→', pc: '→' },
  arrowup: { apple: '↑', pc: '↑' },
  arrowdown: { apple: '↓', pc: '↓' },
  home: { apple: 'Home', pc: 'Home' },
  end: { apple: 'End', pc: 'End' },
  slash: { apple: '/', pc: '/' },
  backspace: { apple: '⌫', pc: 'Backspace' }
};

/** Spoken names, for screen readers and the trainer's prompt. */
const spoken: Record<string, string> = {
  mod: 'Command or Control',
  shift: 'Shift',
  alt: 'Option or Alt',
  enter: 'Enter',
  escape: 'Escape',
  space: 'Space',
  arrowleft: 'Left arrow',
  arrowright: 'Right arrow',
  arrowup: 'Up arrow',
  arrowdown: 'Down arrow',
  slash: 'Slash'
};

/** One alternative of a binding: the steps of a sequence, each a set of keys pressed together. */
export type Chord = string[];
export type Combo = Chord[];

/** Splits "g>h, mod+k" into alternatives, each a sequence of chords. */
export function parseKeys(keys: string): Combo[] {
  return keys
    .split(',')
    .map((alt) => alt.trim())
    .filter(Boolean)
    .map((alt) => alt.split('>').map((step) => step.split('+').map((k) => k.trim().toLowerCase())));
}

/** The labels to draw on keycaps for one chord: ["⌘", "K"]. */
export function keyLabels(chord: Chord, platform: Platform): string[] {
  return chord.map((k) => names[k]?.[platform] ?? (k.length === 1 ? k.toUpperCase() : k));
}

/** The whole binding as text: "⌘ K", "G then L", "1, 2, 3 or 4". */
export function formatKeys(keys: string, platform: Platform): string {
  const alts = parseKeys(keys).map((combo) =>
    combo
      .map((chord) => keyLabels(chord, platform).join(platform === 'apple' ? '' : '+'))
      .join(' then ')
  );
  return alts.length > 2
    ? `${alts.slice(0, -1).join(', ')} or ${alts[alts.length - 1]}`
    : alts.join(' or ');
}

/** The binding read aloud: "Command or Control plus K". */
export function speakKeys(keys: string): string {
  return parseKeys(keys)
    .map((combo) =>
      combo
        .map((chord) =>
          chord.map((k) => spoken[k] ?? (k.length === 1 ? k.toUpperCase() : k)).join(' plus ')
        )
        .join(', then ')
    )
    .join(', or ');
}

/** True when a key press belongs to a field the visitor is typing in. */
export function isTypingTarget(target: EventTarget | null): boolean {
  if (!target || typeof (target as { closest?: unknown }).closest !== 'function') return false;
  const el = target as Element;
  if (el.closest('[contenteditable=""], [contenteditable="true"]')) return true;
  const field = el.closest('input, textarea, select');
  if (!field) return false;
  const type = (field.getAttribute('type') ?? 'text').toLowerCase();
  // Buttons and checkboxes take no text, so shortcuts still work on them.
  return !['button', 'checkbox', 'radio', 'submit', 'reset', 'range', 'color'].includes(type);
}

/** Shortcuts whose keys collide within a scope (or with a global one). */
export function conflicts(list: readonly Shortcut[]): [Shortcut, Shortcut][] {
  const out: [Shortcut, Shortcut][] = [];
  const norm = (keys: string) =>
    parseKeys(keys).map((c) => c.map((ch) => [...ch].sort().join('+')).join('>'));
  for (let i = 0; i < list.length; i++) {
    for (let j = i + 1; j < list.length; j++) {
      const a = list[i];
      const b = list[j];
      const overlap = a.scope === b.scope || a.scope === 'global' || b.scope === 'global';
      if (!overlap) continue;
      const ka = new Set(norm(a.keys));
      if (norm(b.keys).some((k) => ka.has(k))) out.push([a, b]);
    }
  }
  return out;
}

/** Does a keyboard event press this chord? (For the trainer.) */
export function matchesChord(
  e: { key: string; metaKey: boolean; ctrlKey: boolean; altKey: boolean; shiftKey: boolean },
  chord: Chord,
  platform: Platform
): boolean {
  const want = new Set(chord);
  const mod = platform === 'apple' ? e.metaKey : e.ctrlKey;
  if (want.has('mod') !== mod) return false;
  if (
    want.has('shift') !== e.shiftKey &&
    !(e.key.length === 1 && /[?!@#$%^&*()_+{}|:"<>~]/.test(e.key))
  )
    return false;
  if (want.has('alt') !== e.altKey) return false;
  const main = chord.find((k) => !['mod', 'shift', 'alt', 'ctrl', 'meta'].includes(k));
  if (!main) return false;
  const key = e.key.toLowerCase();
  const alias: Record<string, string> = { ' ': 'space', '/': 'slash' };
  return key === main || alias[key] === main;
}
