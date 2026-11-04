'use client';

import type * as React from 'react';
import { Dialog as CommandDialogPrimitive } from '@base-ui/react/dialog';
import {
  Autocomplete,
  AutocompleteCollection,
  AutocompleteEmpty,
  AutocompleteGroup,
  AutocompleteGroupLabel,
  AutocompleteInput,
  AutocompleteItem,
  AutocompleteList,
  AutocompleteSeparator
} from '@braswelljr/algorithm-ui/components/autocomplete';
import { cn } from '@braswelljr/algorithm-ui/lib/utils';
import { SearchIcon } from 'lucide-react';

export const CommandDialog: typeof CommandDialogPrimitive.Root = CommandDialogPrimitive.Root;

export const CommandDialogPortal: typeof CommandDialogPrimitive.Portal =
  CommandDialogPrimitive.Portal;

export const CommandCreateHandle: typeof CommandDialogPrimitive.createHandle =
  CommandDialogPrimitive.createHandle;

export function CommandDialogTrigger(
  props: CommandDialogPrimitive.Trigger.Props
): React.ReactElement {
  return (
    <CommandDialogPrimitive.Trigger
      data-slot="command-dialog-trigger"
      {...props}
    />
  );
}

export function CommandDialogBackdrop({
  className,
  ...props
}: CommandDialogPrimitive.Backdrop.Props): React.ReactElement {
  return (
    <CommandDialogPrimitive.Backdrop
      className={cn(
        'fixed inset-0 z-50 bg-black/50 backdrop-blur-sm transition-all duration-200 data-ending-style:opacity-0 data-starting-style:opacity-0',
        className
      )}
      data-slot="command-dialog-backdrop"
      {...props}
    />
  );
}

export function CommandDialogViewport({
  className,
  ...props
}: CommandDialogPrimitive.Viewport.Props): React.ReactElement {
  return (
    <CommandDialogPrimitive.Viewport
      className={cn(
        'fixed inset-0 z-50 flex flex-col items-center px-4 py-[max(--spacing(4),4vh)] sm:py-[10vh]',
        className
      )}
      data-slot="command-dialog-viewport"
      {...props}
    />
  );
}

export function CommandDialogPopup({
  className,
  viewportClassName,
  children,
  portalProps,
  ...props
}: CommandDialogPrimitive.Popup.Props & {
  portalProps?: CommandDialogPrimitive.Portal.Props;
  /** Classes for the area the popup is placed in, such as no padding on a phone. */
  viewportClassName?: string;
}): React.ReactElement {
  return (
    <CommandDialogPortal {...portalProps}>
      <CommandDialogBackdrop />
      <CommandDialogViewport className={viewportClassName}>
        <CommandDialogPrimitive.Popup
          className={cn(
            'relative row-start-2 flex max-h-105 min-h-0 w-full max-w-xl min-w-0 translate-y-[calc(-1.25rem*var(--nested-dialogs))] flex-col rounded-2xl border border-border bg-card text-fg shadow-lg transition-[translate] duration-200 ease-in-out will-change-transform outline-none before:pointer-events-none before:absolute before:inset-0 before:rounded-[calc(var(--radius-2xl)-1px)] before:bg-bg/50 before:shadow-[0_1px_2px_rgba(0,0,0,0.05)] data-ending-style:translate-y-0 data-nested:data-ending-style:translate-y-8 data-nested-dialog-open:origin-top data-starting-style:translate-y-0 data-nested:data-starting-style:translate-y-8 **:data-[slot=scroll-area-viewport]:data-has-overflow-y:pe-1 dark:before:shadow-[0_-1px_1px_rgba(255,255,255,0.05)]',
            className
          )}
          data-slot="command-dialog-popup"
          {...props}
        >
          <div className="flex min-h-0 flex-1 scale-[calc(1-0.1*var(--nested-dialogs))] flex-col opacity-[calc(1-0.1*var(--nested-dialogs))] transition-[scale,opacity] duration-200 ease-in-out data-ending-style:scale-100 data-ending-style:opacity-100 data-starting-style:scale-98 data-starting-style:opacity-0">
            {children}
          </div>
        </CommandDialogPrimitive.Popup>
      </CommandDialogViewport>
    </CommandDialogPortal>
  );
}

export function Command({
  autoHighlight = 'always',
  keepHighlight = true,
  ...props
}: React.ComponentProps<typeof Autocomplete>): React.ReactElement {
  return (
    <Autocomplete
      autoHighlight={autoHighlight}
      inline
      keepHighlight={keepHighlight}
      open
      {...props}
    />
  );
}

export function CommandInput({
  className,
  containerClassName,
  leading,
  trailing,
  placeholder = undefined,
  ...props
}: React.ComponentProps<typeof AutocompleteInput> & {
  /** Classes for the row the input sits in, for a full-screen search bar. */
  containerClassName?: string;
  /** Rendered before the input in the same row, such as a scope chip. */
  leading?: React.ReactNode;
  /** Rendered after the input in the same row, such as a close control. */
  trailing?: React.ReactNode;
}): React.ReactElement {
  return (
    <div className={cn('px-2.5 py-1.5', containerClassName)}>
      {leading}
      <AutocompleteInput
        autoFocus
        className={cn(
          'border-transparent! bg-transparent! shadow-none before:hidden has-focus-visible:ring-0',
          className
        )}
        placeholder={placeholder}
        size="lg"
        startAddon={<SearchIcon className="text-muted" />}
        {...props}
      />
      {trailing}
    </div>
  );
}

export function CommandList({
  className,
  ...props
}: React.ComponentProps<typeof AutocompleteList>): React.ReactElement {
  return (
    <AutocompleteList
      className={cn('not-empty:scroll-py-2 not-empty:p-2', className)}
      data-slot="command-list"
      {...props}
    />
  );
}

export function CommandEmpty({
  className,
  ...props
}: React.ComponentProps<typeof AutocompleteEmpty>): React.ReactElement {
  return (
    <AutocompleteEmpty
      className={cn('text-muted not-empty:py-6', className)}
      data-slot="command-empty"
      {...props}
    />
  );
}

export function CommandPanel({
  className,
  ...props
}: React.ComponentProps<'div'>): React.ReactElement {
  return (
    <div
      className={cn(
        'relative -mx-px min-h-0 rounded-t-xl border border-b-0 border-border bg-card bg-clip-padding shadow-sm [clip-path:inset(0_1px)] not-has-[+[data-slot=command-footer]]:-mb-px not-has-[+[data-slot=command-footer]]:rounded-b-2xl not-has-[+[data-slot=command-footer]]:[clip-path:inset(0_1px_1px_1px_round_0_0_calc(var(--radius-2xl)-1px)_calc(var(--radius-2xl)-1px))] before:pointer-events-none before:absolute before:inset-0 before:rounded-t-[calc(var(--radius-xl)-1px)] **:data-[slot=scroll-area-scrollbar]:mt-2',
        className
      )}
      {...props}
    />
  );
}

export function CommandGroup({
  className,
  ...props
}: React.ComponentProps<typeof AutocompleteGroup>): React.ReactElement {
  return (
    <AutocompleteGroup
      className={cn('text-fg', className)}
      data-slot="command-group"
      {...props}
    />
  );
}

export function CommandGroupLabel({
  className,
  ...props
}: React.ComponentProps<typeof AutocompleteGroupLabel>): React.ReactElement {
  return (
    <AutocompleteGroupLabel
      className={cn('text-muted', className)}
      data-slot="command-group-label"
      {...props}
    />
  );
}

export function CommandCollection({
  ...props
}: React.ComponentProps<typeof AutocompleteCollection>): React.ReactElement {
  return (
    <AutocompleteCollection
      data-slot="command-collection"
      {...props}
    />
  );
}

export function CommandItem({
  className,
  ...props
}: React.ComponentProps<typeof AutocompleteItem>): React.ReactElement {
  return (
    <AutocompleteItem
      className={cn('py-1.5', className)}
      data-slot="command-item"
      {...props}
    />
  );
}

export function CommandSeparator({
  className,
  ...props
}: React.ComponentProps<typeof AutocompleteSeparator>): React.ReactElement {
  return (
    <AutocompleteSeparator
      className={cn('my-2', className)}
      data-slot="command-separator"
      {...props}
    />
  );
}

export function CommandShortcut({
  className,
  ...props
}: React.ComponentProps<'kbd'>): React.ReactElement {
  return (
    <kbd
      className={cn('ml-auto font-sans text-sm font-medium tracking-widest text-muted', className)}
      data-slot="command-shortcut"
      {...props}
    />
  );
}

export function CommandFooter({
  className,
  ...props
}: React.ComponentProps<'div'>): React.ReactElement {
  return (
    <div
      className={cn(
        'flex items-center justify-between gap-2 rounded-b-[calc(var(--radius-2xl)-1px)] border-t border-border bg-card px-5 py-3 text-sm',
        className
      )}
      data-slot="command-footer"
      {...props}
    />
  );
}

export { CommandDialogPrimitive };
