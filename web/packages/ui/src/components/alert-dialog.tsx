'use client';

import * as React from 'react';
import { AlertDialog as AlertDialogPrimitive } from '@base-ui/react/alert-dialog';
import { Button } from '@braswelljr/algorithm-ui/components/button';
import { cn } from '@braswelljr/algorithm-ui/lib/utils';

function AlertDialog({ ...props }: AlertDialogPrimitive.Root.Props) {
  return (
    <AlertDialogPrimitive.Root
      data-slot="alert-dialog"
      {...props}
    />
  );
}

function AlertDialogTrigger({ ...props }: AlertDialogPrimitive.Trigger.Props) {
  return (
    <AlertDialogPrimitive.Trigger
      data-slot="alert-dialog-trigger"
      {...props}
    />
  );
}

function AlertDialogPortal({ ...props }: AlertDialogPrimitive.Portal.Props) {
  return (
    <AlertDialogPrimitive.Portal
      data-slot="alert-dialog-portal"
      {...props}
    />
  );
}

function AlertDialogOverlay({ className, ...props }: AlertDialogPrimitive.Backdrop.Props) {
  return (
    <AlertDialogPrimitive.Backdrop
      data-slot="alert-dialog-overlay"
      className={cn(
        'fixed inset-0 isolate z-50 bg-fg/32 duration-150 data-closed:animate-out data-closed:fade-out-0 data-open:animate-in data-open:fade-in-0 supports-backdrop-filter:backdrop-blur-xs',
        className
      )}
      {...props}
    />
  );
}

function AlertDialogContent({
  className,
  size = 'default',
  ...props
}: AlertDialogPrimitive.Popup.Props & {
  size?: 'default' | 'sm';
}) {
  return (
    <AlertDialogPortal>
      <AlertDialogOverlay />
      <AlertDialogPrimitive.Popup
        data-slot="alert-dialog-content"
        data-size={size}
        className={cn(
          // The same frame as `DialogPopup`: 24 of padding, a 2xl corner, and
          // never closer than 16 to the edge of a phone's screen.
          'group/alert-dialog-content fixed top-1/2 left-1/2 z-50 grid max-h-[calc(100dvh-2rem)] w-[calc(100%-2rem)] -translate-x-1/2 -translate-y-1/2 gap-4 overflow-y-auto rounded-2xl p-6 shadow-lg/5 ring-1 duration-150 ease-(--ease-out-strong) outline-none',
          'bg-card text-fg ring-fg/10',
          'data-closed:animate-out data-closed:fade-out-0 data-closed:zoom-out-98',
          'data-open:animate-in data-open:fade-in-0 data-open:zoom-in-98',
          'data-[size=default]:max-w-md data-[size=sm]:max-w-xs',
          className
        )}
        {...props}
      />
    </AlertDialogPortal>
  );
}

function AlertDialogHeader({ className, ...props }: React.ComponentProps<'div'>) {
  return (
    <div
      data-slot="alert-dialog-header"
      className={cn(
        'grid grid-rows-[auto_1fr] place-items-center gap-2 text-center has-data-[slot=alert-dialog-media]:grid-rows-[auto_auto_1fr] has-data-[slot=alert-dialog-media]:gap-x-4 sm:group-data-[size=default]/alert-dialog-content:place-items-start sm:group-data-[size=default]/alert-dialog-content:text-left sm:group-data-[size=default]/alert-dialog-content:has-data-[slot=alert-dialog-media]:grid-rows-[auto_1fr]',
        className
      )}
      {...props}
    />
  );
}

function AlertDialogFooter({ className, ...props }: React.ComponentProps<'div'>) {
  return (
    <div
      data-slot="alert-dialog-footer"
      className={cn(
        '-mx-6 -mb-6 flex flex-col-reverse gap-2 rounded-b-[calc(var(--radius-2xl)-1px)] border-t px-6 py-4 sm:flex-row sm:justify-end',
        'border-border bg-bg/50',
        'group-data-[size=sm]/alert-dialog-content:grid group-data-[size=sm]/alert-dialog-content:grid-cols-2',
        className
      )}
      {...props}
    />
  );
}

function AlertDialogMedia({ className, ...props }: React.ComponentProps<'div'>) {
  return (
    <div
      data-slot="alert-dialog-media"
      className={cn(
        "mb-2 inline-flex size-10 items-center justify-center rounded-md sm:group-data-[size=default]/alert-dialog-content:row-span-2 *:[svg:not([class*='size-'])]:size-6",
        'bg-subtle',
        className
      )}
      {...props}
    />
  );
}

function AlertDialogTitle({
  className,
  ...props
}: React.ComponentProps<typeof AlertDialogPrimitive.Title>) {
  return (
    <AlertDialogPrimitive.Title
      data-slot="alert-dialog-title"
      className={cn(
        'font-heading text-lg leading-snug font-semibold text-balance sm:group-data-[size=default]/alert-dialog-content:group-has-data-[slot=alert-dialog-media]/alert-dialog-content:col-start-2',
        className
      )}
      {...props}
    />
  );
}

function AlertDialogDescription({
  className,
  ...props
}: React.ComponentProps<typeof AlertDialogPrimitive.Description>) {
  return (
    <AlertDialogPrimitive.Description
      data-slot="alert-dialog-description"
      className={cn(
        'text-sm text-balance text-muted md:text-pretty',
        '*:[a]:underline *:[a]:underline-offset-3 *:[a]:hover:text-fg',
        className
      )}
      {...props}
    />
  );
}

/**
 * The dialog's answer: runs its `onClick`, then closes the dialog.
 *
 * It was a plain `Button`, so pressing it ran the action and left the dialog
 * standing over the page that had already changed beneath it. Every "Remove"
 * and "Delete" in the app read as having done nothing, because the question was
 * still on screen. A close renders the same button and shuts the dialog after
 * the handler, which is what the word on the button promises.
 *
 * A dialog that must stay open while an async action runs, and close only on
 * success, uses a plain `Button` and closes itself: see `ConfirmDialog`.
 */
function AlertDialogAction({
  className,
  variant,
  size,
  ...props
}: AlertDialogPrimitive.Close.Props &
  Pick<React.ComponentProps<typeof Button>, 'variant' | 'size'>) {
  return (
    <AlertDialogPrimitive.Close
      data-slot="alert-dialog-action"
      className={cn(className)}
      render={
        <Button
          variant={variant}
          size={size}
        />
      }
      {...props}
    />
  );
}

/**
 * Backing out of a dialog, in the quiet variant.
 *
 * `ghost` rather than `outline`, and matching every other Cancel in the app.
 * A dialog asks one question and offers one answer to it; the way out is not a
 * second choice of equal weight, and an outlined button beside a filled one
 * reads as two things to decide between. This is the one the reader falls back
 * to, so it recedes.
 *
 * The variant carries the colour, the hover and the dark mode. It used to be
 * `outline` with a neutral palette written over it by hand, which is a second
 * definition of a variant that already exists - and one that does not move when
 * the variant does.
 */
function AlertDialogCancel({
  className,
  variant = 'ghost',
  size = 'default',
  ...props
}: AlertDialogPrimitive.Close.Props &
  Pick<React.ComponentProps<typeof Button>, 'variant' | 'size'>) {
  return (
    <AlertDialogPrimitive.Close
      data-slot="alert-dialog-cancel"
      className={cn(className)}
      render={
        <Button
          variant={variant}
          size={size}
        />
      }
      {...props}
    />
  );
}

export {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogMedia,
  AlertDialogOverlay,
  AlertDialogPortal,
  AlertDialogTitle,
  AlertDialogTrigger
};
