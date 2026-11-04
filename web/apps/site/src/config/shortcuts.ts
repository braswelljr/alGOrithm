import type { Shortcut } from '@braswelljr/algorithm-ui/lib/shortcuts';

/**
 * Every keyboard shortcut on the site. The help overlay (?), the key hints in
 * tooltips and the palette, the trainer on /shortcuts and the tests all read
 * this list; a key bound anywhere else is a bug. Keys use react-hotkeys-hook
 * syntax: `mod` is Command on a Mac and Ctrl elsewhere, `g>l` is g then l.
 */
export const shortcuts: readonly Shortcut[] = [
  // Everywhere
  {
    id: 'palette.open',
    keys: 'mod+k',
    description: 'Open the command palette',
    scope: 'global',
    category: 'Find'
  },
  {
    id: 'help.open',
    keys: 'shift+slash',
    description: 'Show all keyboard shortcuts',
    scope: 'global',
    category: 'Find'
  },
  {
    id: 'search.focus',
    keys: 'slash',
    description: 'Focus the search box',
    scope: 'global',
    category: 'Find'
  },
  {
    id: 'go.home',
    keys: 'g>h',
    description: 'Go to the catalog',
    scope: 'global',
    category: 'Navigation'
  },
  {
    id: 'go.learn',
    keys: 'g>l',
    description: 'Go to the learning paths',
    scope: 'global',
    category: 'Navigation'
  },
  {
    id: 'go.quiz',
    keys: 'g>q',
    description: 'Go to the quiz',
    scope: 'global',
    category: 'Navigation'
  },
  {
    id: 'go.docs',
    keys: 'g>d',
    description: 'Go to the docs',
    scope: 'global',
    category: 'Navigation'
  },
  {
    id: 'go.bigo',
    keys: 'g>b',
    description: 'Go to Big O',
    scope: 'global',
    category: 'Navigation'
  },
  {
    id: 'go.cheatsheets',
    keys: 'g>c',
    description: 'Go to the cheat sheets',
    scope: 'global',
    category: 'Navigation'
  },
  {
    id: 'theme.toggle',
    keys: 't',
    description: 'Switch between light and dark',
    scope: 'global',
    category: 'Display'
  },

  // The player
  {
    id: 'player.toggle',
    keys: 'space',
    description: 'Play or pause',
    scope: 'player',
    category: 'Playback'
  },
  {
    id: 'player.forward',
    keys: 'arrowright',
    description: 'Step forward',
    scope: 'player',
    category: 'Playback'
  },
  {
    id: 'player.back',
    keys: 'arrowleft',
    description: 'Step back',
    scope: 'player',
    category: 'Playback'
  },
  {
    id: 'player.forward10',
    keys: 'shift+arrowright',
    description: 'Jump forward 10 steps',
    scope: 'player',
    category: 'Playback'
  },
  {
    id: 'player.back10',
    keys: 'shift+arrowleft',
    description: 'Jump back 10 steps',
    scope: 'player',
    category: 'Playback'
  },
  {
    id: 'player.first',
    keys: 'home',
    description: 'Go to the first step',
    scope: 'player',
    category: 'Playback'
  },
  {
    id: 'player.last',
    keys: 'end',
    description: 'Go to the last step',
    scope: 'player',
    category: 'Playback'
  },
  {
    id: 'player.replay',
    keys: 'r',
    description: 'Replay from the start',
    scope: 'player',
    category: 'Playback'
  },
  { id: 'player.slower', keys: '[', description: 'Slower', scope: 'player', category: 'Speed' },
  { id: 'player.faster', keys: ']', description: 'Faster', scope: 'player', category: 'Speed' },
  {
    id: 'player.code',
    keys: 'c',
    description: 'Show or hide the code',
    scope: 'player',
    category: 'Panels'
  },
  {
    id: 'player.predict',
    keys: 'p',
    description: 'Turn predict mode on or off',
    scope: 'player',
    category: 'Panels'
  },
  {
    id: 'player.export',
    keys: 'e',
    description: 'Export the run as JSON',
    scope: 'player',
    category: 'Panels'
  },

  // The quiz
  {
    id: 'quiz.answer',
    keys: '1,2,3,4',
    description: 'Pick answer 1, 2, 3 or 4',
    scope: 'quiz',
    category: 'Answering'
  },
  {
    id: 'quiz.next',
    keys: 'enter',
    description: 'Next question',
    scope: 'quiz',
    category: 'Answering'
  },

  // Lists
  { id: 'list.down', keys: 'j', description: 'Next item', scope: 'list', category: 'Moving' },
  { id: 'list.up', keys: 'k', description: 'Previous item', scope: 'list', category: 'Moving' },
  {
    id: 'list.open',
    keys: 'enter',
    description: 'Open the item',
    scope: 'list',
    category: 'Moving'
  }
];

/** One shortcut by id; a typo is a type error at the call site's test. */
export function shortcut(id: string): Shortcut {
  const s = shortcuts.find((x) => x.id === id);
  if (!s) throw new Error(`no shortcut ${id}`);
  return s;
}

/** The keys of a shortcut, for react-hotkeys-hook and the hints. */
export const keysOf = (id: string) => shortcut(id).keys;
