'use client';

import type React from 'react';
import { Dialog as SheetPrimitive } from '@base-ui/react/dialog';
import { mergeProps } from '@base-ui/react/merge-props';
import { useRender } from '@base-ui/react/use-render';
import { Button } from '@braswelljr/algorithm-ui/components/button';
import { ScrollArea } from '@braswelljr/algorithm-ui/components/scroll-area';
import { cn } from '@braswelljr/algorithm-ui/lib/utils';
import { XIcon } from 'lucide-react';

export const Sheet: typeof SheetPrimitive.Root = SheetPrimitive.Root;

export const SheetPortal: typeof SheetPrimitive.Portal = SheetPrimitive.Portal;

export function SheetTrigger(props: SheetPrimitive.Trigger.Props): React.ReactElement {
  return (
    <SheetPrimitive.Trigger
      data-slot="sheet-trigger"
      {...props}
    />
  );
}

export function SheetClose(props: SheetPrimitive.Close.Props): React.ReactElement {
  return (
    <SheetPrimitive.Close
      data-slot="sheet-close"
      {...props}
    />
  );
}

export function SheetBackdrop({
  className,
  ...props
}: SheetPrimitive.Backdrop.Props): React.ReactElement {
  return (
    <SheetPrimitive.Backdrop
      className={cn(
        'fixed inset-0 z-50 bg-fg/32 backdrop-blur-sm transition-all duration-200 data-ending-style:opacity-0 data-starting-style:opacity-0',
        className
      )}
      data-slot="sheet-backdrop"
      {...props}
    />
  );
}

export function SheetViewport({
  className,
  side,
  variant = 'default',
  ...props
}: SheetPrimitive.Viewport.Props & {
  side?: 'right' | 'left' | 'top' | 'bottom';
  variant?: 'default' | 'inset';
}): React.ReactElement {
  return (
    <SheetPrimitive.Viewport
      className={cn(
        'fixed inset-0 z-50 grid',
        side === 'bottom' && 'grid grid-rows-[1fr_auto] pt-12',
        side === 'top' && 'grid grid-rows-[auto_1fr] pb-12',
        side === 'left' && 'flex justify-start',
        side === 'right' && 'flex justify-end',
        variant === 'inset' && 'sm:p-4',
        className
      )}
      data-slot="sheet-viewport"
      {...props}
    />
  );
}

export function SheetPopup({
  className,
  children,
  showCloseButton = true,
  side = 'right',
  variant = 'default',
  closeProps,
  portalProps,
  ...props
}: SheetPrimitive.Popup.Props & {
  showCloseButton?: boolean;
  side?: 'right' | 'left' | 'top' | 'bottom';
  variant?: 'default' | 'inset';
  closeProps?: SheetPrimitive.Close.Props;
  portalProps?: SheetPrimitive.Portal.Props;
}): React.ReactElement {
  return (
    <SheetPortal {...portalProps}>
      <SheetBackdrop />
      <SheetViewport
        side={side}
        variant={variant}
      >
        <SheetPrimitive.Popup
          className={cn(
            // Changed background and text to Zinc
            'relative flex max-h-full min-h-0 w-full min-w-0 flex-col bg-card text-fg',
            'shadow-lg/5 transition-[opacity,translate] duration-200 ease-in-out will-change-transform outline-none not-dark:bg-clip-padding',
            // Zinc-based Shadows
            'before:pointer-events-none before:absolute before:inset-0 before:shadow-[0_1px_--theme(--color-zinc-950/4%)] dark:before:shadow-[0_-1px_--theme(--color-white/6%)]',
            'data-ending-style:opacity-0 data-starting-style:opacity-0 max-sm:before:hidden',
            // Side-specific borders (Zinc)
            side === 'bottom' &&
              'row-start-2 border-t border-border data-ending-style:translate-y-8 data-starting-style:translate-y-8',
            side === 'top' &&
              'border-b border-border data-ending-style:-translate-y-8 data-starting-style:-translate-y-8',
            side === 'left' &&
              'w-[calc(100%-(--spacing(12)))] max-w-md border-e border-border data-ending-style:-translate-x-8 data-starting-style:-translate-x-8',
            side === 'right' &&
              'col-start-2 w-[calc(100%-(--spacing(12)))] max-w-md border-s border-border data-ending-style:translate-x-8 data-starting-style:translate-x-8',
            // Inset variant handling
            variant === 'inset' &&
              'before:hidden sm:rounded-2xl sm:border sm:border-border sm:before:rounded-[calc(var(--radius-2xl)-1px)]',
            className
          )}
          data-slot="sheet-popup"
          {...props}
        >
          {children}
          {showCloseButton && (
            <SheetPrimitive.Close
              aria-label="Close"
              className="inset-e-4 absolute top-4"
              render={
                <Button
                  size="icon"
                  variant="ghost"
                  className="text-muted hover:text-fg"
                />
              }
              {...closeProps}
            >
              <XIcon className="size-4" />
            </SheetPrimitive.Close>
          )}
        </SheetPrimitive.Popup>
      </SheetViewport>
    </SheetPortal>
  );
}

export function SheetHeader({
  className,
  render,
  ...props
}: useRender.ComponentProps<'div'>): React.ReactElement {
  const defaultProps = {
    className: cn(
      'flex flex-col gap-2 p-6 in-[[data-slot=sheet-popup]:has([data-slot=sheet-panel])]:pb-3 max-sm:pb-4',
      className
    ),
    'data-slot': 'sheet-header'
  };

  return useRender({
    defaultTagName: 'div',
    props: mergeProps<'div'>(defaultProps, props),
    render
  });
}

export function SheetFooter({
  className,
  variant = 'default',
  render,
  ...props
}: useRender.ComponentProps<'div'> & {
  variant?: 'default' | 'bare';
}): React.ReactElement {
  const defaultProps = {
    className: cn(
      'flex flex-col-reverse gap-2 px-6 sm:flex-row sm:justify-end',
      // Changed bg-subtle to Zinc-50/900
      variant === 'default' && 'border-t border-border bg-bg/72 py-4',
      variant === 'bare' &&
        'pt-4 pb-6 in-[[data-slot=sheet-popup]:has([data-slot=sheet-panel])]:pt-3',
      className
    ),
    'data-slot': 'sheet-footer'
  };

  return useRender({
    defaultTagName: 'div',
    props: mergeProps<'div'>(defaultProps, props),
    render
  });
}

export function SheetTitle({
  className,
  ...props
}: SheetPrimitive.Title.Props): React.ReactElement {
  return (
    <SheetPrimitive.Title
      className={cn('font-heading text-xl leading-none font-semibold text-fg', className)}
      data-slot="sheet-title"
      {...props}
    />
  );
}

export function SheetDescription({
  className,
  ...props
}: SheetPrimitive.Description.Props): React.ReactElement {
  return (
    <SheetPrimitive.Description
      // Changed text-muted to Zinc-500/400
      className={cn('text-sm text-muted', className)}
      data-slot="sheet-description"
      {...props}
    />
  );
}

export function SheetPanel({
  className,
  scrollFade = true,
  render,
  ...props
}: useRender.ComponentProps<'div'> & {
  scrollFade?: boolean;
}): React.ReactElement {
  const defaultProps = {
    // Even padding on all four sides.
    //
    // It collapsed the top to `pt-1` whenever a `SheetHeader` was present and
    // the bottom to `pb-1` whenever an unruled `SheetFooter` was, on the
    // reasoning that those already provide the spacing. They do not: the header
    // and the footer carry their own, and what the panel lost was the gap
    // between them and the content - so every sheet with a header opened with
    // its first line jammed under the title, and the two ends of the same panel
    // were padded differently from each other and from the sides.
    className: cn('p-6', className),
    'data-slot': 'sheet-panel'
  };

  return (
    <ScrollArea scrollFade={scrollFade}>
      {useRender({
        defaultTagName: 'div',
        props: mergeProps<'div'>(defaultProps, props),
        render
      })}
    </ScrollArea>
  );
}

export { SheetPrimitive, SheetBackdrop as SheetOverlay, SheetPopup as SheetContent };
