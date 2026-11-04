'use client';

import * as React from 'react';
import { ContextMenu as ContextMenuPrimitive } from '@base-ui/react/context-menu';
import { Menu as MenuPrimitive } from '@base-ui/react/menu';
import { ChevronRightIcon } from 'lucide-react';
import { cn } from '@/lib/utils';

/**
 * Right-click menu, built on Base UI to match the rest of this library.
 *
 * The niko-table tree arrived expecting the Radix flavour of this component,
 * which differs in two ways that matter: it takes a child via `asChild` where
 * Base UI takes an element via `render`, and it types `className` as a plain
 * string where Base UI also allows a `(state) => string` callback. Mixing the
 * two produced type errors wherever a menu prop crossed between them, so this
 * is the Base UI version.
 *
 * Base UI splits the pieces differently from Radix: `ContextMenu.Root` and
 * `.Trigger` open the menu, but everything inside the popup is the shared
 * `Menu` primitive. That is why both are imported here. The styling is copied
 * from `dropdown-menu.tsx` deliberately, a context menu and a dropdown should
 * be indistinguishable once open.
 */
function ContextMenu({ ...props }: ContextMenuPrimitive.Root.Props) {
  return (
    <ContextMenuPrimitive.Root
      data-slot="context-menu"
      {...props}
    />
  );
}

function ContextMenuTrigger({ ...props }: ContextMenuPrimitive.Trigger.Props) {
  return (
    <ContextMenuPrimitive.Trigger
      data-slot="context-menu-trigger"
      {...props}
    />
  );
}

function ContextMenuPortal({ ...props }: MenuPrimitive.Portal.Props) {
  return (
    <MenuPrimitive.Portal
      data-slot="context-menu-portal"
      {...props}
    />
  );
}

function ContextMenuContent({
  align = 'start',
  alignOffset = 0,
  side = 'bottom',
  sideOffset = 4,
  className,
  ...props
}: MenuPrimitive.Popup.Props &
  Pick<MenuPrimitive.Positioner.Props, 'align' | 'alignOffset' | 'side' | 'sideOffset'>) {
  return (
    <MenuPrimitive.Portal>
      <MenuPrimitive.Positioner
        className="isolate z-50 outline-none"
        align={align}
        alignOffset={alignOffset}
        side={side}
        sideOffset={sideOffset}
      >
        <MenuPrimitive.Popup
          data-slot="context-menu-content"
          className={cn(
            'scroll-quiet z-50 max-h-(--available-height) min-w-32 origin-(--transform-origin) overflow-x-hidden overflow-y-auto rounded-lg p-1 shadow-md ring-1 duration-100 outline-none',
            'bg-card text-fg ring-fg/10',
            'data-closed:animate-out data-closed:fade-out-0 data-closed:zoom-out-95 data-open:animate-in data-open:fade-in-0 data-open:zoom-in-95 data-[side=bottom]:slide-in-from-top-2 data-[side=inline-end]:slide-in-from-left-2 data-[side=inline-start]:slide-in-from-right-2 data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2 data-closed:overflow-hidden',
            className
          )}
          {...props}
        />
      </MenuPrimitive.Positioner>
    </MenuPrimitive.Portal>
  );
}

function ContextMenuGroup({ ...props }: MenuPrimitive.Group.Props) {
  return (
    <MenuPrimitive.Group
      data-slot="context-menu-group"
      {...props}
    />
  );
}

function ContextMenuLabel({
  className,
  inset,
  ...props
}: MenuPrimitive.GroupLabel.Props & {
  inset?: boolean;
}) {
  return (
    <MenuPrimitive.Group data-slot="context-menu-label-group">
      <MenuPrimitive.GroupLabel
        data-slot="context-menu-label"
        data-inset={inset}
        className={cn('text-muted px-1.5 py-1 text-sm font-medium data-inset:pl-7', className)}
        {...props}
      />
    </MenuPrimitive.Group>
  );
}

function ContextMenuItem({
  className,
  inset,
  variant = 'default',
  ...props
}: MenuPrimitive.Item.Props & {
  inset?: boolean;
  variant?: 'default' | 'destructive' | 'destructive-secondary';
}) {
  return (
    <MenuPrimitive.Item
      data-slot="context-menu-item"
      data-inset={inset}
      data-variant={variant}
      className={cn(
        "group/context-menu-item relative flex min-h-8 cursor-default items-center gap-1.5 rounded-md px-1.5 py-1 text-sm outline-hidden select-none data-disabled:pointer-events-none data-disabled:opacity-50 data-inset:pl-7 sm:min-h-7 pointer-coarse:min-h-11 [&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-4",
        'focus:bg-subtle focus:text-fg',
        'data-[variant=destructive]:bg-red-600! data-[variant=destructive]:text-white! data-[variant=destructive]:focus:bg-red-700! data-[variant=destructive]:*:[svg]:text-white!',
        'data-[variant=destructive-secondary]:bg-red-50 data-[variant=destructive-secondary]:text-red-600 data-[variant=destructive-secondary]:focus:bg-red-100! data-[variant=destructive-secondary]:focus:text-red-700! dark:data-[variant=destructive-secondary]:bg-red-950 dark:data-[variant=destructive-secondary]:focus:bg-red-900! data-[variant=destructive-secondary]:*:[svg]:text-current!',
        className
      )}
      {...props}
    />
  );
}

function ContextMenuSub({ ...props }: MenuPrimitive.SubmenuRoot.Props) {
  return (
    <MenuPrimitive.SubmenuRoot
      data-slot="context-menu-sub"
      {...props}
    />
  );
}

function ContextMenuSubTrigger({
  className,
  inset,
  children,
  ...props
}: MenuPrimitive.SubmenuTrigger.Props & {
  inset?: boolean;
}) {
  return (
    <MenuPrimitive.SubmenuTrigger
      data-slot="context-menu-sub-trigger"
      data-inset={inset}
      className={cn(
        "flex min-h-8 cursor-default items-center gap-1.5 rounded-md px-1.5 py-1 text-sm outline-hidden select-none data-inset:pl-7 sm:min-h-7 pointer-coarse:min-h-11 [&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-4",
        'focus:bg-subtle focus:text-fg data-open:bg-subtle data-popup-open:bg-subtle',
        className
      )}
      {...props}
    >
      {children}
      <ChevronRightIcon className="cn-rtl-flip text-muted ml-auto" />
    </MenuPrimitive.SubmenuTrigger>
  );
}

function ContextMenuSubContent({
  align = 'start',
  alignOffset = -3,
  side = 'right',
  sideOffset = 0,
  className,
  ...props
}: React.ComponentProps<typeof ContextMenuContent>) {
  return (
    <ContextMenuContent
      data-slot="context-menu-sub-content"
      className={cn('w-auto min-w-24', className)}
      align={align}
      alignOffset={alignOffset}
      side={side}
      sideOffset={sideOffset}
      {...props}
    />
  );
}

function ContextMenuSeparator({ className, ...props }: MenuPrimitive.Separator.Props) {
  return (
    <MenuPrimitive.Separator
      data-slot="context-menu-separator"
      className={cn('bg-fg/10 -mx-1 my-1 h-px', className)}
      {...props}
    />
  );
}

function ContextMenuShortcut({ className, ...props }: React.ComponentProps<'span'>) {
  return (
    <span
      data-slot="context-menu-shortcut"
      className={cn('text-muted ml-auto text-sm tracking-widest', className)}
      {...props}
    />
  );
}

export {
  ContextMenu,
  ContextMenuContent,
  ContextMenuGroup,
  ContextMenuItem,
  ContextMenuLabel,
  ContextMenuPortal,
  ContextMenuSeparator,
  ContextMenuShortcut,
  ContextMenuSub,
  ContextMenuSubContent,
  ContextMenuSubTrigger,
  ContextMenuTrigger
};
