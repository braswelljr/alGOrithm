'use client';

import { memo, useEffect, useId, useState } from 'react';
import { Keycap, type KeycapTone } from '@braswelljr/algorithm-ui/components/keycap';
import { usePlatform, type Platform } from '@braswelljr/algorithm-ui/hooks/use-platform';
import { type Keystroke } from '@braswelljr/algorithm-ui/hooks/use-typewriter';
import { cn } from '@braswelljr/algorithm-ui/lib/utils';
import {
  ArrowBigUp,
  ArrowBigUpDash,
  ArrowRightToLine,
  CornerDownLeft,
  Delete,
  FastForward,
  Globe,
  LayoutGrid,
  Mic,
  MicOff,
  Monitor,
  Moon,
  Play,
  Rewind,
  Search,
  Sparkles,
  Sun,
  SunDim,
  Volume1,
  Volume2,
  VolumeX,
  type LucideIcon
} from 'lucide-react';

type KeyDef = {
  id: string;
  /** What is printed on the cap. */
  legend: string;
  /** The shifted character, printed above the legend on a symbol key. */
  upper?: string;
  /**
   * A modifier's symbol, printed in the corner above its name. The name is then
   * set small enough for the narrowest such key, as it is on the real board.
   */
  glyph?: string;
  /**
   * A symbol drawn in place of a name (tab, shift, return), or above a small
   * name (`label`) on a function key, or beside one on fn.
   */
  icon?: LucideIcon;
  /** The small name that goes with `icon`: `F1`, `fn`. */
  label?: string;
  /** The status light caps lock carries in its corner. */
  light?: boolean;
  /** Which side a wide key's legend sits on; a right-hand key is `end`. */
  align?: 'start' | 'end';
  /** Width in key units. */
  w?: number;
  tone?: KeycapTone;
  /** The characters the key types, plain and shifted. */
  chars?: [string, string];
  /** The `KeyboardEvent.code` values that press it. */
  codes: string[];
};

/** Half-height keys sharing one position, top to bottom. One on its own sits at the bottom. */
type Stack = { stack: KeyDef[]; w?: number };

type Cell = KeyDef | Stack;

const isStack = (cell: Cell): cell is Stack => 'stack' in cell;

const CODE_OF: Record<string, string> = {
  '`': 'Backquote',
  '-': 'Minus',
  '=': 'Equal',
  '[': 'BracketLeft',
  ']': 'BracketRight',
  '\\': 'Backslash',
  ';': 'Semicolon',
  "'": 'Quote',
  ',': 'Comma',
  '.': 'Period',
  '/': 'Slash'
};

function codeOf(c: string) {
  if (/[a-z]/.test(c)) return `Key${c.toUpperCase()}`;
  if (/[0-9]/.test(c)) return `Digit${c}`;
  return CODE_OF[c];
}

/** A symbol key: both characters printed, shifted above plain. */
const sym = (plain: string, shifted: string): KeyDef => ({
  id: plain,
  legend: plain,
  upper: shifted,
  chars: [plain, shifted],
  codes: [codeOf(plain)]
});
const letters = (row: string): KeyDef[] =>
  [...row].map((c) => ({
    id: c,
    legend: c.toUpperCase(),
    chars: [c, c.toUpperCase()],
    codes: [codeOf(c)]
  }));
const mod = (
  id: string,
  legend: string,
  w: number,
  codes: string[],
  more: Partial<KeyDef> = {}
): KeyDef => ({ id, legend, w, tone: 'mod', codes, ...more });
const digits = (): KeyDef[] => [...'1234567890'].map((c, i) => sym(c, '!@#$%^&*()'[i]));
/**
 * The Surface's function keys: F1 to F8 carry what they do over their small
 * `F` name (sound, the microphone, media, brightness, a second display), F9
 * to F12 their names alone.
 */
const SURFACE_FKEY_ICONS = [VolumeX, Volume1, Volume2, MicOff, Play, SunDim, Sun, Monitor];
const surfaceFkeys = (): KeyDef[] =>
  Array.from({ length: 12 }, (_, i) => {
    const name = `F${i + 1}`;
    return i < SURFACE_FKEY_ICONS.length
      ? mod(`f${i + 1}`, '', 1, [name], { icon: SURFACE_FKEY_ICONS[i], label: name })
      : mod(`f${i + 1}`, name, 1, [name]);
  });

/** The MacBook's function keys: each an icon for what it does over its small `F` name. */
const MAC_FKEY_ICONS = [
  SunDim,
  Sun,
  LayoutGrid,
  Search,
  Mic,
  Moon,
  Rewind,
  Play,
  FastForward,
  VolumeX,
  Volume1,
  Volume2
];
const macFkeys = (): KeyDef[] =>
  MAC_FKEY_ICONS.map((icon, i) =>
    mod(`f${i + 1}`, '', 1, [`F${i + 1}`], { icon, label: `F${i + 1}` })
  );
const arrow = (id: string, legend: string): KeyDef =>
  mod(id, legend, 1, [`Arrow${id[0].toUpperCase()}${id.slice(1)}`]);

const END = { align: 'end' } as const;

type Layout = {
  /** How many key units wide every row is. */
  units: number;
  rows: Cell[][];
};

/**
 * The M5 MacBook Pro keyboard, US layout: 14.5 units wide; a full-height
 * function row of icons over small F names, ending in Touch ID; symbols where
 * other boards print names (tab, caps lock with its light, shift, delete,
 * return); the globe on fn; each modifier's symbol over its name; and the
 * arrow keys as an inverted T of half-height keys.
 */
const MACBOOK: Layout = {
  units: 14.5,
  rows: [
    [mod('esc', 'esc', 1.5, ['Escape']), ...macFkeys(), mod('touchid', '', 1, [])],
    [
      sym('`', '~'),
      ...digits(),
      sym('-', '_'),
      sym('=', '+'),
      mod('backspace', '', 1.5, ['Backspace'], { icon: Delete, ...END })
    ],
    [
      mod('tab', '', 1.5, ['Tab'], { icon: ArrowRightToLine }),
      ...letters('qwertyuiop'),
      sym('[', '{'),
      sym(']', '}'),
      sym('\\', '|')
    ],
    [
      mod('caps', '', 1.75, ['CapsLock'], { icon: ArrowBigUpDash, light: true }),
      ...letters('asdfghjkl'),
      sym(';', ':'),
      sym("'", '"'),
      mod('enter', '', 1.75, ['Enter', 'NumpadEnter'], {
        icon: CornerDownLeft,
        tone: 'accent',
        ...END
      })
    ],
    [
      mod('lshift', '', 2.25, ['ShiftLeft'], { icon: ArrowBigUp }),
      ...letters('zxcvbnm'),
      sym(',', '<'),
      sym('.', '>'),
      sym('/', '?'),
      mod('rshift', '', 2.25, ['ShiftRight'], { icon: ArrowBigUp, ...END })
    ],
    [
      mod('fn', '', 1, ['Fn'], { icon: Globe, label: 'fn', align: 'start' }),
      mod('ctrl', 'control', 1, ['ControlLeft'], { glyph: '⌃' }),
      mod('opt', 'option', 1, ['AltLeft'], { glyph: '⌥' }),
      mod('cmd', 'command', 1.25, ['MetaLeft'], { glyph: '⌘' }),
      mod('space', '', 5, ['Space'], { tone: 'alpha' }),
      mod('rcmd', 'command', 1.25, ['MetaRight'], { glyph: '⌘' }),
      mod('ropt', 'option', 1, ['AltRight'], { glyph: '⌥' }),
      { stack: [arrow('left', '◀')] },
      { stack: [arrow('up', '▲'), arrow('down', '▼')] },
      { stack: [arrow('right', '▶')] }
    ]
  ]
};

/**
 * The latest Microsoft Surface Laptop keyboard (the Copilot+ PC generation),
 * US layout: 15 units wide, a function row from Esc to Del and the power key,
 * capitalized names, the Windows key, the Copilot key where the menu key used
 * to be, and half-height up and down arrows between full-height left and
 * right.
 */
const SURFACE: Layout = {
  units: 15,
  rows: [
    [
      mod('esc', 'Esc', 1, ['Escape']),
      ...surfaceFkeys(),
      mod('del', 'Del', 1, ['Delete']),
      mod('power', '⏻', 1, [])
    ],
    [
      sym('`', '~'),
      ...digits(),
      sym('-', '_'),
      sym('=', '+'),
      mod('backspace', 'Backspace', 2, ['Backspace'], END)
    ],
    [
      mod('tab', 'Tab', 1.5, ['Tab']),
      ...letters('qwertyuiop'),
      sym('[', '{'),
      sym(']', '}'),
      { ...sym('\\', '|'), w: 1.5 }
    ],
    [
      mod('caps', 'Caps', 1.75, ['CapsLock']),
      ...letters('asdfghjkl'),
      sym(';', ':'),
      sym("'", '"'),
      mod('enter', 'Enter', 2.25, ['Enter', 'NumpadEnter'], { tone: 'accent', ...END })
    ],
    [
      mod('lshift', 'Shift', 2.25, ['ShiftLeft']),
      ...letters('zxcvbnm'),
      sym(',', '<'),
      sym('.', '>'),
      sym('/', '?'),
      mod('rshift', 'Shift', 2.75, ['ShiftRight'], END)
    ],
    [
      mod('ctrl', 'Ctrl', 1.25, ['ControlLeft']),
      mod('fn', 'Fn', 1, ['Fn']),
      mod('win', '⊞', 1, ['MetaLeft', 'MetaRight']),
      mod('alt', 'Alt', 1, ['AltLeft']),
      mod('space', '', 5.75, ['Space'], { tone: 'alpha' }),
      mod('ralt', 'Alt', 1, ['AltRight']),
      // The Copilot key sends Shift+Windows+F23; a board without one sends the menu key.
      mod('copilot', '', 1, ['F23', 'ContextMenu', 'ControlRight'], {
        icon: Sparkles
      }),
      arrow('left', '←'),
      { stack: [arrow('up', '↑'), arrow('down', '↓')] },
      arrow('right', '→')
    ]
  ]
};

/** Keys on the left half: a shifted one of these is shifted with the right hand. */
const LEFT = new Set([...'`12345qwertasdfgzxcvb']);

/**
 * The geometry every board shares, in percent of the keyboard's width (the
 * `cqw` the keys are sized in): each row is `ROW` tall with `GAP` between keys
 * and rows. The hands are drawn in the same units, so they land on the keys
 * they press.
 */
const GAP = 0.6;
const ROW = 6;
const HALF = (ROW - GAP) / 2;

type Placed = { cx: number; cy: number; width: number };

type FingerName = 'L5' | 'L4' | 'L3' | 'L2' | 'L1' | 'R1' | 'R2' | 'R3' | 'R4' | 'R5';

/** Touch typing: the keys each finger is responsible for, on either board. */
const REACH: Record<FingerName, string[]> = {
  L5: [...'`1qaz', 'esc', 'f1', 'tab', 'caps', 'lshift', 'fn', 'ctrl'],
  L4: [...'2wsx', 'f2', 'opt', 'win'],
  L3: [...'3edc', 'f3'],
  L2: [...'45rtfgvb', 'f4', 'f5'],
  L1: ['cmd', 'alt'],
  R1: ['space', 'rcmd', 'ralt'],
  R2: [...'67yuhjnm', 'f6', 'f7'],
  R3: [...'8ik,', 'f8'],
  R4: [...'9ol.', 'f9', 'f10', 'ropt', 'copilot'],
  R5: [
    ..."0-=p[]\\;'/",
    ...['f11', 'f12', 'del', 'enter', 'rshift', 'backspace'],
    ...['left', 'up', 'down', 'right']
  ]
};

const FINGER_OF = new Map<string, FingerName>();
for (const [finger, ids] of Object.entries(REACH)) {
  for (const id of ids) FINGER_OF.set(id, finger as FingerName);
}

type Finger = { name: FingerName; tip: [number, number]; base: [number, number]; width: number };

type Hand = {
  fingers: Finger[];
  palm: { x: number; y: number; width: number; height: number };
};

/** A layout with everything worked out from it: where each key is, and how the hands rest on it. */
type Board = {
  layout: Layout;
  /** The width of one key unit, gap included. */
  unit: number;
  height: number;
  placed: Map<string, Placed>;
  /** The key, and whether shift is held, that types each character. */
  byChar: Map<string, { id: string; shift: boolean }>;
  byCode: Map<string, string>;
  hands: Record<'L' | 'R', Hand>;
};

function build(layout: Layout): Board {
  const unit = (100 + GAP) / layout.units;
  const height = layout.rows.length * ROW + (layout.rows.length - 1) * GAP;

  const placed = new Map<string, Placed>();
  const keys: KeyDef[] = [];
  layout.rows.forEach((row, r) => {
    let units = 0;
    const top = r * (ROW + GAP);
    for (const cell of row) {
      const w = cell.w ?? 1;
      const width = w * unit - GAP;
      const cx = units * unit + width / 2;
      if (isStack(cell)) {
        // Half-height keys fill their position from the bottom up.
        cell.stack.forEach((k, i) => {
          const fromBottom = cell.stack.length - 1 - i;
          placed.set(k.id, { cx, cy: top + ROW - HALF / 2 - fromBottom * (HALF + GAP), width });
          keys.push(k);
        });
      } else {
        placed.set(cell.id, { cx, cy: top + ROW / 2, width });
        keys.push(cell);
      }
      units += w;
    }
  });

  const byChar = new Map<string, { id: string; shift: boolean }>();
  const byCode = new Map<string, string>();
  for (const k of keys) {
    for (const code of k.codes) byCode.set(code, k.id);
    if (!k.chars) continue;
    byChar.set(k.chars[0], { id: k.id, shift: false });
    byChar.set(k.chars[1], { id: k.id, shift: true });
  }
  byChar.set(' ', { id: 'space', shift: false });
  byChar.set('\n', { id: 'enter', shift: false });

  const at = (id: string) => placed.get(id) as Placed;

  /** One hand resting on the home row: four fingers, a thumb on the space bar, a palm. */
  const hand = (side: 'L' | 'R'): Hand => {
    const home = side === 'L' ? ['a', 's', 'd', 'f'] : [';', 'l', 'k', 'j'];
    const widths = [3, 3.5, 3.65, 3.65];
    const tips = home.map((id) => [at(id).cx, at(id).cy] as [number, number]);
    const center = (tips[0][0] + tips[3][0]) / 2;
    const baseY = height + 2.2;
    const fingers: Finger[] = home.map((_, i) => ({
      name: `${side}${5 - i}` as FingerName,
      tip: tips[i],
      base: [tips[i][0] + (center - tips[i][0]) * 0.3, baseY],
      width: widths[i]
    }));
    const space = at('space');
    const toward = side === 'L' ? 1 : -1;
    const thumbTip: [number, number] = [space.cx + space.width * toward * -0.22, space.cy];
    fingers.push({
      name: `${side}1` as FingerName,
      tip: thumbTip,
      base: [thumbTip[0] - toward * 7, baseY + 3.4],
      width: 4
    });
    const left = Math.min(...fingers.map((f) => f.base[0])) - 2.8;
    const right = Math.max(...fingers.map((f) => f.base[0])) + 2.8;
    return { fingers, palm: { x: left, y: baseY - 1.8, width: right - left, height: 11 } };
  };

  return { layout, unit, height, placed, byChar, byCode, hands: { L: hand('L'), R: hand('R') } };
}

/** The board for each platform: a MacBook Pro's for Apple devices, a Surface's for the rest. */
const BOARDS: Record<Platform, Board> = { apple: build(MACBOOK), pc: build(SURFACE) };

/**
 * The keys the visitor is holding down on their own keyboard, by key id. It
 * only listens while `active`, and lets go of everything when the window
 * loses focus, since the matching keyup will never arrive.
 */
function useHeldKeys(active: boolean, byCode: Board['byCode']) {
  const [held, setHeld] = useState<ReadonlySet<string>>(() => new Set());

  useEffect(() => {
    if (!active) return;
    const change = (code: string, down: boolean) => {
      const id = byCode.get(code);
      if (!id) return;
      setHeld((prev) => {
        if (prev.has(id) === down) return prev;
        const next = new Set(prev);
        if (down) next.add(id);
        else next.delete(id);
        return next;
      });
    };
    const clear = () => setHeld(new Set());
    const onDown = (e: KeyboardEvent) => change(e.code, true);
    const onUp = (e: KeyboardEvent) => {
      // macOS sends no keyup for a key released while Command is down, so
      // letting go of Command lets go of everything.
      if (e.key === 'Meta') clear();
      else change(e.code, false);
    };
    const onHide = () => {
      if (document.hidden) clear();
    };
    window.addEventListener('keydown', onDown);
    window.addEventListener('keyup', onUp);
    window.addEventListener('blur', clear);
    document.addEventListener('visibilitychange', onHide);
    return () => {
      window.removeEventListener('keydown', onDown);
      window.removeEventListener('keyup', onUp);
      window.removeEventListener('blur', clear);
      document.removeEventListener('visibilitychange', onHide);
      clear();
    };
  }, [active, byCode]);

  return held;
}

/**
 * Where each finger is now: the one pressing a key reaches it, and its hand
 * follows a little. A wide key is pressed near its edge, as a real finger does.
 */
function reach(board: Board, down: ReadonlySet<string>) {
  const offsets = new Map<FingerName, [number, number]>();
  const handShift: Record<'L' | 'R', [number, number]> = { L: [0, 0], R: [0, 0] };
  for (const id of down) {
    const finger = FINGER_OF.get(id);
    if (!finger) continue;
    const side = finger[0] as 'L' | 'R';
    const f = board.hands[side].fingers.find((x) => x.name === finger);
    const target = board.placed.get(id);
    if (!f || !target) continue;
    const tx =
      target.width > board.unit * 1.5
        ? f.tip[0] +
          (target.cx > f.tip[0] ? 1 : -1) *
            Math.min(Math.abs(target.cx - f.tip[0]), target.width / 2 - 2.1)
        : target.cx;
    const dx = tx - f.tip[0];
    const dy = target.cy - f.tip[1] + 0.3;
    offsets.set(finger, [dx, dy]);
    handShift[side] = [dx * 0.22, dy * 0.22];
  }
  return { offsets, handShift };
}

/** The two hands, drawn over the keys in the keyboard's own units. */
function Hands({ board, down }: { board: Board; down: ReadonlySet<string> }) {
  const { offsets, handShift } = reach(board, down);
  const HEIGHT = board.height;
  const fade = `${useId().replace(/:/g, '')}-fade`;

  return (
    <svg
      viewBox={`0 0 100 ${HEIGHT}`}
      focusable="false"
      className="pointer-events-none absolute inset-x-0 top-0 h-auto w-full overflow-visible"
    >
      <defs>
        <linearGradient
          id={`${fade}-g`}
          x1="0"
          y1="0"
          x2="0"
          y2="1"
        >
          <stop
            offset="0.6"
            stopColor="white"
          />
          <stop
            offset="0.85"
            stopColor="white"
            stopOpacity="0"
          />
        </linearGradient>
        <mask
          id={fade}
          maskUnits="userSpaceOnUse"
          x="-10"
          y="-5"
          width="120"
          height={HEIGHT + 30}
        >
          <rect
            x="-10"
            y="-5"
            width="120"
            height={HEIGHT + 30}
            fill={`url(#${fade}-g)`}
          />
        </mask>
      </defs>
      <g
        mask={`url(#${fade})`}
        opacity="0.78"
      >
        {(['L', 'R'] as const).map((side) => {
          const { fingers, palm } = board.hands[side];
          const [hx, hy] = handShift[side];
          const move = (finger: FingerName, ms: number) => {
            const [x, y] = offsets.get(finger) ?? [hx, hy];
            return {
              transform: `translate(${x}px, ${y}px)`,
              transition: `transform ${ms}ms ease-out`
            };
          };
          const palmStyle = {
            transform: `translate(${hx}px, ${hy}px)`,
            transition: 'transform 120ms ease-out'
          };
          const capsule = (f: Finger, extra: number, paint: string) => (
            <line
              key={`${f.name}-${extra}`}
              x1={f.tip[0]}
              y1={f.tip[1]}
              x2={f.base[0]}
              y2={f.base[1]}
              stroke={paint}
              strokeWidth={f.width + extra}
              strokeLinecap="round"
              style={move(f.name, 90)}
            />
          );
          return (
            <g key={side}>
              {/* Outlines first, then fills over them, so the hand reads as one shape. */}
              <rect
                {...palm}
                rx="6.7"
                fill="none"
                stroke="var(--hand-stroke)"
                strokeWidth="0.5"
                style={palmStyle}
              />
              {fingers.map((f) => capsule(f, 0.5, 'var(--hand-stroke)'))}
              <rect
                {...palm}
                rx="6.7"
                fill="var(--hand-fill)"
                style={palmStyle}
              />
              {fingers.map((f) => capsule(f, 0, 'var(--hand-fill)'))}
              {fingers.map((f) =>
                offsets.has(f.name) ? (
                  <circle
                    key={`${f.name}-tip`}
                    cx={f.tip[0]}
                    cy={f.tip[1] - f.width * 0.15}
                    r={f.width * 0.24}
                    fill="var(--cap-accent-face)"
                    style={move(f.name, 90)}
                  />
                ) : null
              )}
            </g>
          );
        })}
      </g>
    </svg>
  );
}

/** The `KeyboardEvent.key` a key without a character sends when it is clicked. */
const NAMED_KEY: Record<string, string> = {
  enter: 'Enter',
  backspace: 'Backspace',
  del: 'Delete',
  esc: 'Escape',
  tab: 'Tab',
  up: 'ArrowUp',
  down: 'ArrowDown',
  left: 'ArrowLeft',
  right: 'ArrowRight'
};

/**
 * The small names that share a cap with a symbol ("control", "F1"). They have
 * to fit a one-unit key at every width, which no step of the type scale does,
 * so they are sized against the board itself.
 */
const SMALL = 'text-[1.1cqw]';
const ICON = 'size-[1.6cqw] shrink-0';

/** What is printed on a key, laid out the way its kind of key prints it. */
function Legend({ k }: { k: KeyDef }) {
  const end = k.align === 'end';
  const Icon = k.icon;

  // A symbol key: the shifted character over the plain one.
  if (k.upper) {
    return (
      <span className="flex size-full flex-col items-center justify-center gap-[0.4cqw]">
        <span className="opacity-60">{k.upper}</span>
        {k.legend}
      </span>
    );
  }
  // A function key: what it does over its F name.
  if (Icon && k.label && k.align === undefined) {
    return (
      <span className="flex size-full flex-col items-center justify-center gap-[0.5cqw]">
        <Icon
          aria-hidden="true"
          className={ICON}
        />
        <span className={SMALL}>{k.label}</span>
      </span>
    );
  }
  // A one-unit key that is only a symbol (Copilot): the symbol, centered.
  if (Icon && !k.label && k.align === undefined && (k.w ?? 1) === 1) {
    return (
      <span className="flex size-full items-center justify-center">
        <Icon
          aria-hidden="true"
          className={ICON}
        />
      </span>
    );
  }
  // tab, caps lock, shift, delete, return, fn: a symbol in the bottom corner,
  // caps lock's light in the top one, and fn's name beside its globe.
  if (Icon) {
    return (
      <span
        className={cn('flex size-full flex-col justify-between', end ? 'items-end' : 'items-start')}
      >
        <span className={cn('size-[0.6cqw] rounded-full', k.light && 'bg-current opacity-60')} />
        <span className="flex items-center gap-[0.4cqw]">
          <Icon
            aria-hidden="true"
            className={ICON}
          />
          {k.label && <span className={SMALL}>{k.label}</span>}
        </span>
      </span>
    );
  }
  // control, option, command: the symbol over the name, centered.
  if (k.glyph) {
    return (
      <span className="flex size-full flex-col items-center justify-center gap-[0.5cqw]">
        <span>{k.glyph}</span>
        <span className={SMALL}>{k.legend}</span>
      </span>
    );
  }
  // A name or a letter: a wide key's in its corner, anything else centered.
  const cornered = (k.w ?? 1) > 1 || k.align !== undefined;
  return (
    <span
      className={cn(
        'flex size-full',
        cornered ? 'items-end' : 'items-center justify-center',
        cornered && (end ? 'justify-end' : 'justify-start')
      )}
    >
      {k.legend}
    </span>
  );
}

/** One key: its cap, the face that lights when it is down, and its legend. */
function Key({
  k,
  on,
  half,
  onPress
}: {
  k: KeyDef;
  on: boolean;
  half?: boolean;
  onPress?: (k: KeyDef) => void;
}) {
  return (
    <Keycap
      tone={k.tone}
      pressed={on}
      style={half ? undefined : { gridColumn: `span ${(k.w ?? 1) * 4}` }}
      onPointerDown={
        onPress &&
        ((event) => {
          // Keep the focus where it is, and the page from selecting text.
          event.preventDefault();
          onPress(k);
        })
      }
      className={cn(
        'flex overflow-hidden rounded-[0.9cqw] font-mono [--cap-depth:0.5cqw]',
        half && 'h-[2.7cqw]',
        onPress && 'cursor-pointer'
      )}
    >
      {/* The lit face: on at once, fading out after the key comes back up. */}
      <span
        className={cn(
          'absolute inset-0 bg-lit-face transition-opacity motion-reduce:transition-none',
          on ? 'opacity-90 duration-0' : 'opacity-0 duration-700'
        )}
      />
      <span
        className={cn(
          'relative flex size-full p-[0.7cqw] text-xs leading-none transition-colors motion-reduce:transition-none',
          on ? 'text-white duration-0' : 'duration-700'
        )}
      >
        <Legend k={k} />
      </span>
    </Keycap>
  );
}

type KeyboardProps = {
  /** The key an animation is pressing now, if any. */
  stroke?: Keystroke | null;
  /** Also press the keys the visitor presses on their own keyboard. */
  live?: boolean;
  /** Draw the two hands that type. */
  hands?: boolean;
  /**
   * Makes the keys clickable: each click sends the `KeyboardEvent.key` the
   * key would (`a`, `A`, `Enter`, `Backspace`, `ArrowUp`, ...). Shift and caps
   * lock are held by the board itself: a click on shift shifts the next
   * character, and caps lock stays on until it is clicked again.
   */
  onPress?: (key: string) => void;
  className?: string;
};

/**
 * A laptop keyboard built from keycaps, with two hands typing on it: a MacBook
 * Pro's for a visitor on an Apple device and a Microsoft Surface's for anyone
 * else, each with its own keys, widths and legends. A key goes down and lights
 * when an animation types it (`stroke`) or, with `live`, when the visitor
 * presses it, and the finger touch typing assigns to that key reaches for it;
 * a shifted character holds the shift key under the other hand. The hands are
 * translucent so every legend stays readable. Both boards are the same height,
 * so the page does not move when the first client render picks the visitor's.
 * It never takes focus or stops a keystroke, and it is decorative, so it is
 * hidden from assistive technology: clicking its keys (`onPress`) is a
 * convenience for a pointer, beside the real keyboard that does the same.
 */
export const Keyboard = memo(function Keyboard({
  stroke,
  live = false,
  hands = true,
  onPress,
  className
}: KeyboardProps) {
  const board = BOARDS[usePlatform()];
  const held = useHeldKeys(live, board.byCode);
  // What clicks hold down: a shift for the next character, caps lock until
  // clicked again, and the key under the pointer while the button is down.
  const [shift, setShift] = useState(false);
  const [caps, setCaps] = useState(false);
  const [clicked, setClicked] = useState<string | null>(null);

  useEffect(() => {
    if (clicked === null) return;
    const release = () => setClicked(null);
    window.addEventListener('pointerup', release);
    window.addEventListener('pointercancel', release);
    return () => {
      window.removeEventListener('pointerup', release);
      window.removeEventListener('pointercancel', release);
    };
  }, [clicked]);

  const press =
    onPress &&
    ((k: KeyDef) => {
      setClicked(k.id);
      if (k.id === 'lshift' || k.id === 'rshift') {
        setShift((s) => !s);
        return;
      }
      if (k.id === 'caps') {
        setCaps((c) => !c);
        return;
      }
      let key = NAMED_KEY[k.id];
      if (k.id === 'space') key = ' ';
      if (k.chars) {
        const letter = /[a-z]/.test(k.chars[0]);
        key = k.chars[shift !== (caps && letter) ? 1 : 0];
        setShift(false);
      }
      if (key !== undefined) onPress(key);
    });

  const down = new Set(held);
  if (clicked !== null) down.add(clicked);
  if (shift) down.add('lshift').add('rshift');
  if (caps) down.add('caps');
  const typed = stroke ? board.byChar.get(stroke.key) : undefined;
  if (typed) {
    down.add(typed.id);
    if (typed.shift) down.add(LEFT.has(typed.id) ? 'rshift' : 'lshift');
  }

  return (
    <div
      aria-hidden="true"
      className={cn('rounded-2xl plate p-2 select-none sm:p-3', className)}
    >
      <div className="@container">
        <div className="relative flex flex-col gap-[0.6cqw]">
          {board.layout.rows.map((row, r) => (
            <div
              key={r}
              className="grid h-[6cqw] gap-[0.6cqw]"
              style={{ gridTemplateColumns: `repeat(${board.layout.units * 4}, minmax(0, 1fr))` }}
            >
              {row.map((cell) =>
                isStack(cell) ? (
                  <div
                    key={cell.stack[0].id}
                    style={{ gridColumn: `span ${(cell.w ?? 1) * 4}` }}
                    className="flex flex-col justify-end gap-[0.6cqw]"
                  >
                    {cell.stack.map((k) => (
                      <Key
                        key={k.id}
                        k={k}
                        on={down.has(k.id)}
                        half
                        onPress={press}
                      />
                    ))}
                  </div>
                ) : (
                  <Key
                    key={cell.id}
                    k={cell}
                    on={down.has(cell.id)}
                    onPress={press}
                  />
                )
              )}
            </div>
          ))}
          {hands && (
            <Hands
              board={board}
              down={down}
            />
          )}
        </div>
      </div>
    </div>
  );
});
