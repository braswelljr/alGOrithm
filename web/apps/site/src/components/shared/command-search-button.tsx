'use client';

import { KeyCombo } from '@braswelljr/algorithm-ui/algorithm/key-combo';
import { Button } from '@braswelljr/algorithm-ui/components/button';
import { SearchIcon } from 'lucide-react';
import { keysOf } from '@/config/shortcuts';
import { useShortcuts } from '@/context/use-shortcuts';

/** Opens the command palette: search every page and entry (Cmd/Ctrl+K). */
export function CommandSearchButton() {
  const { openPalette } = useShortcuts();
  return (
    <Button
      variant="ghost"
      size="sm"
      onClick={openPalette}
      data-shortcut="palette.open"
      aria-keyshortcuts="Meta+K Control+K"
      className="gap-2 text-muted max-sm:size-8 max-sm:px-0"
    >
      <SearchIcon aria-hidden />
      <span className="max-sm:sr-only">Search</span>
      <KeyCombo
        keys={keysOf('palette.open')}
        className="max-md:hidden"
      />
    </Button>
  );
}
