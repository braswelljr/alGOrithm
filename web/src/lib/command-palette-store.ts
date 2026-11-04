/**
 * Whether the command palette is open.
 *
 * A store rather than state inside the palette, because two unrelated things
 * open it: the keyboard shortcut and the search control in the navbar, which
 * sit in different subtrees. Subscribers select what they use, so the navbar
 * button does not re-render on anything but the open flag. As in
 * gsl-special-courses.
 */

import { Store, useStore } from '@tanstack/react-store';

type CommandPaletteState = { open: boolean };

export const commandPaletteStore = new Store<CommandPaletteState>({ open: false });

export function openCommandPalette() {
  commandPaletteStore.setState(() => ({ open: true }));
}

export function closeCommandPalette() {
  commandPaletteStore.setState(() => ({ open: false }));
}

export function toggleCommandPalette() {
  commandPaletteStore.setState((state) => ({ open: !state.open }));
}

export function setCommandPaletteOpen(open: boolean) {
  commandPaletteStore.setState(() => ({ open }));
}

/** Whether the palette is currently open. */
export function useCommandPaletteOpen(): boolean {
  return useStore(commandPaletteStore, (state) => state.open);
}
