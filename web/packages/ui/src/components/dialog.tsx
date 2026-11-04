'use client';

import type React from 'react';
import { Dialog as DialogPrimitive } from '@base-ui/react/dialog';
import { mergeProps } from '@base-ui/react/merge-props';
import { useRender } from '@base-ui/react/use-render';
import { Button } from '@braswelljr/algorithm-ui/components/button';
import { cn } from '@braswelljr/algorithm-ui/lib/utils';
import { XIcon } from 'lucide-react';

export const DialogCreateHandle: typeof DialogPrimitive.createHandle = DialogPrimitive.createHandle;

export const Dialog: typeof DialogPrimitive.Root = DialogPrimitive.Root;

export const DialogPortal: typeof DialogPrimitive.Portal = DialogPrimitive.Portal;

export function DialogTrigger(props: DialogPrimitive.Trigger.Props): React.ReactElement {
  return (
    <DialogPrimitive.Trigger
      data-slot="dialog-trigger"
      {...props}
    />
  );
}

export function DialogClose(props: DialogPrimitive.Close.Props): React.ReactElement {
  return (
    <DialogPrimitive.Close
      data-slot="dialog-close"
      {...props}
    />
  );
}

export function DialogBackdrop({
  className,
  ...props
}: DialogPrimitive.Backdrop.Props): React.ReactElement {
  return (
    <DialogPrimitive.Backdrop
      className={cn(
        'fixed inset-0 z-50 bg-fg/32 backdrop-blur-sm transition-all duration-200 data-ending-style:opacity-0 data-starting-style:opacity-0',
        className
      )}
      data-slot="dialog-backdrop"
      {...props}
    />
  );
}

export function DialogViewport({
  className,
  ...props
}: DialogPrimitive.Viewport.Props): React.ReactElement {
  return (
    <DialogPrimitive.Viewport
      className={cn(
        'fixed inset-0 z-50 grid grid-rows-[1fr_auto_3fr] justify-items-center p-4',
        className
      )}
      data-slot="dialog-viewport"
      {...props}
    />
  );
}

export function DialogPopup({
  className,
  children,
  showCloseButton = true,
  bottomStickOnMobile = true,
  closeProps,
  portalProps,
  ...props
}: DialogPrimitive.Popup.Props & {
  showCloseButton?: boolean;
  bottomStickOnMobile?: boolean;
  closeProps?: DialogPrimitive.Close.Props;
  portalProps?: DialogPrimitive.Portal.Props;
}): React.ReactElement {
  return (
    <DialogPortal {...portalProps}>
      <DialogBackdrop />
      <DialogViewport
        className={cn(bottomStickOnMobile && 'max-sm:grid-rows-[1fr_auto] max-sm:p-0 max-sm:pt-12')}
      >
        <DialogPrimitive.Popup
          className={cn(
            'group/dialog-popup relative row-start-2 flex max-h-full min-h-0 w-full max-w-lg min-w-0 origin-center flex-col rounded-2xl border border-border bg-card text-fg opacity-[calc(1-var(--nested-dialogs))] shadow-lg/5 transition-[scale,opacity,translate] duration-200 ease-(--ease-out-strong) will-change-transform outline-none not-dark:bg-clip-padding',
            // It arrives from just smaller than itself and fades in, and leaves
            // the same way, so it does not pop into being or vanish.
            'data-ending-style:scale-98 data-ending-style:opacity-0 data-starting-style:scale-98 data-starting-style:opacity-0',
            'before:pointer-events-none before:absolute before:inset-0 before:rounded-[calc(var(--radius-2xl)-1px)] before:shadow-[0_1px_--theme(--color-zinc-950/4%)] dark:before:shadow-[0_-1px_--theme(--color-white/6%)]',
            bottomStickOnMobile &&
              'max-sm:max-w-none max-sm:origin-bottom max-sm:rounded-none max-sm:border-x-0 max-sm:border-t max-sm:border-b-0 max-sm:before:hidden max-sm:before:rounded-none max-sm:data-ending-style:translate-y-4 max-sm:data-ending-style:scale-100 max-sm:data-starting-style:translate-y-4 max-sm:data-starting-style:scale-100',
            className
          )}
          data-slot="dialog-popup"
          data-close-button={showCloseButton ? '' : undefined}
          {...props}
        >
          {children}
          {showCloseButton && (
            <DialogPrimitive.Close
              aria-label="Close"
              className="inset-e-2 absolute top-2"
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
            </DialogPrimitive.Close>
          )}
        </DialogPrimitive.Popup>
      </DialogViewport>
    </DialogPortal>
  );
}

export function DialogHeader({
  className,
  render,
  ...props
}: useRender.ComponentProps<'div'>): React.ReactElement {
  const defaultProps = {
    className: cn(
      'flex shrink-0 flex-col gap-2 p-6 max-sm:pb-4',
      // Room for the close button in the corner, so a long title wraps before
      // it rather than running under it.
      'group-data-close-button/dialog-popup:pe-14',
      // The rule under the title, on the component rather than copied into every
      // call site. It is what separates the header from the body in every dialog
      // in the app; a dialog that had to opt in is a dialog that eventually
      // forgets to. Pass a `border-b-0` className to drop it where a dialog is
      // genuinely a single block of prose.
      'border-b border-border',
      className
    ),
    'data-slot': 'dialog-header'
  };

  return useRender({
    defaultTagName: 'div',
    props: mergeProps<'div'>(defaultProps, props),
    render
  });
}

export function DialogFooter({
  className,
  variant = 'default',
  render,
  ...props
}: useRender.ComponentProps<'div'> & {
  variant?: 'default' | 'bare';
}): React.ReactElement {
  const defaultProps = {
    className: cn(
      'flex shrink-0 flex-col-reverse gap-2 px-6 sm:flex-row sm:justify-end sm:rounded-b-[calc(var(--radius-2xl)-1px)]',
      variant === 'default' && 'border-t border-border bg-bg/72 py-4',
      variant === 'bare' && 'pt-4 pb-6',
      // Clear of the home indicator on a phone, where the popup sits on the
      // bottom edge.
      'max-sm:pb-[max(--spacing(4),env(safe-area-inset-bottom))]',
      className
    ),
    'data-slot': 'dialog-footer'
  };

  return useRender({
    defaultTagName: 'div',
    props: mergeProps<'div'>(defaultProps, props),
    render
  });
}

export function DialogTitle({
  className,
  ...props
}: DialogPrimitive.Title.Props): React.ReactElement {
  return (
    <DialogPrimitive.Title
      className={cn('font-heading text-lg leading-snug font-semibold text-balance', className)}
      data-slot="dialog-title"
      {...props}
    />
  );
}

export function DialogDescription({
  className,
  ...props
}: DialogPrimitive.Description.Props): React.ReactElement {
  return (
    <DialogPrimitive.Description
      className={cn('text-sm text-muted', className)}
      data-slot="dialog-description"
      {...props}
    />
  );
}

/**
 * The body of a dialog: everything between the header and the footer.
 *
 * # The shape, and why it is a component
 *
 * `flex min-h-0 flex-1 flex-col gap-4 overflow-y-auto p-6` - it takes the space
 * the header and footer leave, scrolls inside itself rather than growing the
 * popup past the viewport, and spaces its children evenly so a form does not
 * have to margin each field.
 *
 * Those are the classes the assign-task dialog settled on and the shape every
 * dialog in the app now uses. It is a component rather than a string to copy
 * because a string gets copied wrong: before this, three dialogs out of forty
 * nine agreed on their body, and the rest each invented padding and spacing.
 *
 * There is one body component on purpose. `DialogPanel` used to be a second -
 * the same region wrapped in a `ScrollArea` - which meant two ways to write the
 * same part of a dialog and no rule for choosing. Three call sites used it,
 * forty six did not, and none of them agreed on padding.
 *
 * # Forms
 *
 * A form dialog renders the body as the form, and keeps the footer outside it:
 *
 * ```tsx
 * <DialogBody render={<form id="edit-x" onSubmit={submit} />}>…</DialogBody>
 * <DialogFooter>
 *   <Button type="submit" form="edit-x">Save</Button>
 * </DialogFooter>
 * ```
 *
 * A `<form>` wrapped around the body and the footer breaks the flex chain from
 * the popup down: the form grows to fit its fields, the body never scrolls, and
 * the popup clips the footer off the bottom of a long form.
 */
export function DialogBody({
  className,
  render,
  ...props
}: useRender.ComponentProps<'div'>): React.ReactElement {
  const defaultProps = {
    className: cn('flex min-h-0 flex-1 flex-col gap-4 overflow-y-auto scroll-quiet p-6', className),
    'data-slot': 'dialog-body'
  };

  return useRender({
    defaultTagName: 'div',
    props: mergeProps<'div'>(defaultProps, props),
    render
  });
}

export { DialogPrimitive, DialogBackdrop as DialogOverlay, DialogPopup as DialogContent };
