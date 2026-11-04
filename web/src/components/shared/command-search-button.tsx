'use client';

import { SearchNormal1 } from 'iconsax-reactjs';
import { KeyCombo } from '@/components/algorithm/key-combo';
import { Button } from '@/components/ui/button';
import { keysOf } from '@/config/shortcuts';
import { useShortcuts } from '@/context/use-shortcuts';
import { prefetchPalette } from '@/lib/prefetch-palette';

/**
 * The visible way into the command palette, as in gsl-certificates' navbar.
 * The shortcut is the fast path and this is the discoverable one: the hint
 * printed on it is where people learn the keystroke. Solid (secondary), not
 * outlined.
 */
export function CommandSearchButton() {
  const { openPalette } = useShortcuts();
  return (
    <Button
      type="button"
      variant="secondary"
      size="sm"
      onClick={openPalette}
      // Warms the palette's code and its index on the way to a click.
      onPointerEnter={prefetchPalette}
      onFocus={prefetchPalette}
      data-shortcut="palette.open"
      aria-label="Search and commands"
      aria-keyshortcuts="Meta+K Control+K"
      className="text-muted gap-2 px-2.5"
    >
      <SearchNormal1
        aria-hidden
        className="size-4 shrink-0"
      />
      <span className="hidden text-sm font-medium sm:inline">Search</span>
      {/* Hidden on a phone, where there is no keyboard to press it with. */}
      <KeyCombo
        keys={keysOf('palette.open')}
        className="hidden sm:inline-flex"
      />
    </Button>
  );
}
