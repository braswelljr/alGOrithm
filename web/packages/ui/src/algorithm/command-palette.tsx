'use client';

import * as React from 'react';
import { CornerDownLeftIcon } from 'lucide-react';
import { KeyCombo } from '@braswelljr/algorithm-ui/algorithm/key-combo';
import {
  Command,
  CommandCollection,
  CommandDialog,
  CommandDialogPopup,
  CommandEmpty,
  CommandFooter,
  CommandGroup,
  CommandGroupLabel,
  CommandInput,
  CommandItem,
  CommandList,
  CommandPanel
} from '@braswelljr/algorithm-ui/components/command';

export type { PaletteGroup, PaletteItem } from '@braswelljr/algorithm-ui/algorithm/command-palette-filter';
import { filterPalette, type PaletteGroup, type PaletteItem } from '@braswelljr/algorithm-ui/algorithm/command-palette-filter';
export { filterPalette, fuzzyScore } from '@braswelljr/algorithm-ui/algorithm/command-palette-filter';


/**
 * The command palette (⌘K / Ctrl+K): search every entry, path, category, tag,
 * page and action, grouped, fuzzy-matched, with recent items first when the
 * query is empty. Built on the Base UI Command; arrows move, Enter runs,
 * Escape closes.
 *
 *   <CommandPalette open={open} onOpenChange={setOpen} groups={groups} />
 */
export function CommandPalette({
  open,
  onOpenChange,
  groups,
  placeholder = 'Search entries, paths, pages and actions'
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  groups: PaletteGroup[];
  placeholder?: string;
}) {
  const [query, setQuery] = React.useState('');
  const shown = React.useMemo(() => filterPalette(groups, query), [groups, query]);
  const run = (item: PaletteItem) => {
    onOpenChange(false);
    setQuery('');
    item.onSelect();
  };
  return (
    <CommandDialog
      open={open}
      onOpenChange={onOpenChange}
    >
      <CommandDialogPopup aria-label="Command palette">
        <Command
          items={shown}
          mode="none"
          value={query}
          onValueChange={setQuery}
          itemToStringValue={(item: unknown) => (item as PaletteItem).label}
        >
          <CommandInput placeholder={placeholder} />
          <CommandPanel>
            <CommandEmpty>Nothing matches. Try fewer letters.</CommandEmpty>
            <CommandList>
              {(group: PaletteGroup) => (
                <CommandGroup
                  key={group.label}
                  items={group.items}
                >
                  <CommandGroupLabel>{group.label}</CommandGroupLabel>
                  <CommandCollection>
                    {(item: PaletteItem) => (
                      <CommandItem
                        key={item.id}
                        value={item}
                        onClick={() => run(item)}
                        className="flex items-center gap-3"
                      >
                        {item.icon ? (
                          <span
                            aria-hidden
                            className="text-muted"
                          >
                            {item.icon}
                          </span>
                        ) : null}
                        <span className="flex min-w-0 flex-1 flex-col">
                          <span className="truncate">{item.label}</span>
                          {item.hint ? <span className="truncate text-xs text-muted">{item.hint}</span> : null}
                        </span>
                        {item.keys ? <KeyCombo keys={item.keys} /> : null}
                      </CommandItem>
                    )}
                  </CommandCollection>
                </CommandGroup>
              )}
            </CommandList>
          </CommandPanel>
          <CommandFooter className="text-xs text-muted">
            <span className="inline-flex items-center gap-1.5">
              <CornerDownLeftIcon
                aria-hidden
                className="size-3.5"
              />
              to open
            </span>
            <span className="inline-flex items-center gap-1.5">
              <KeyCombo keys="shift+slash" /> all shortcuts
            </span>
          </CommandFooter>
        </Command>
      </CommandDialogPopup>
    </CommandDialog>
  );
}
