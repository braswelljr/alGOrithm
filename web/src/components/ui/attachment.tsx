'use client';

/**
 * A file, as a row (from gsl-special-courses' attachment): the file's mark,
 * its name and a line about it, and its actions at the end. `state` says where
 * it is in its life: `idle` is the empty slot waiting for one, `error` a file
 * that was refused, `done` one in hand.
 *
 *   <Attachment>
 *     <AttachmentMedia><LuFileText /></AttachmentMedia>
 *     <AttachmentContent>
 *       <AttachmentTitle>data.csv</AttachmentTitle>
 *       <AttachmentDescription>CSV · 2.1 KB</AttachmentDescription>
 *     </AttachmentContent>
 *     <AttachmentActions><AttachmentAction aria-label="Remove">...</AttachmentAction></AttachmentActions>
 *   </Attachment>
 */
import * as React from 'react';
import { cva, type VariantProps } from 'class-variance-authority';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';

const attachmentVariants = cva(
  'group/attachment text-fg relative flex w-full min-w-0 flex-wrap items-center rounded-xl transition-colors',
  {
    variants: {
      variant: {
        default: 'bg-subtle',
        muted: 'bg-bg',
        success: 'bg-success/10',
        error: 'bg-danger/10'
      },
      size: {
        sm: 'gap-2.5 p-1.5 pe-2 text-sm',
        default: 'gap-3 p-2 pe-3 text-sm',
        lg: 'gap-3 p-2.5 pe-3 text-base'
      }
    },
    defaultVariants: { variant: 'default', size: 'default' }
  }
);

function Attachment({
  className,
  state = 'done',
  variant,
  size,
  ...props
}: React.ComponentProps<'div'> &
  VariantProps<typeof attachmentVariants> & {
    state?: 'idle' | 'error' | 'done';
  }) {
  return (
    <div
      data-slot="attachment"
      data-state={state}
      data-size={size ?? 'default'}
      className={cn(
        attachmentVariants({ variant: state === 'error' ? 'error' : variant, size }),
        className
      )}
      {...props}
    />
  );
}

function AttachmentMedia({ className, ...props }: React.ComponentProps<'div'>) {
  return (
    <div
      data-slot="attachment-media"
      className={cn(
        'bg-bg text-accent flex aspect-square w-10 shrink-0 items-center justify-center rounded-lg group-data-[size=lg]/attachment:w-12 group-data-[size=sm]/attachment:w-8',
        'group-data-[state=error]/attachment:text-danger [&_svg]:pointer-events-none [&_svg:not([class*=size-])]:size-5',
        className
      )}
      {...props}
    />
  );
}

function AttachmentContent({ className, ...props }: React.ComponentProps<'div'>) {
  return (
    <div
      data-slot="attachment-content"
      className={cn('min-w-0 flex-1 leading-tight', className)}
      {...props}
    />
  );
}

function AttachmentTitle({ className, ...props }: React.ComponentProps<'span'>) {
  return (
    <span
      data-slot="attachment-title"
      className={cn('block min-w-0 truncate font-medium', className)}
      {...props}
    />
  );
}

function AttachmentDescription({ className, ...props }: React.ComponentProps<'span'>) {
  return (
    <span
      data-slot="attachment-description"
      className={cn(
        'text-muted group-data-[state=error]/attachment:text-danger mt-0.5 block min-w-0 truncate text-xs',
        className
      )}
      {...props}
    />
  );
}

function AttachmentActions({ className, ...props }: React.ComponentProps<'div'>) {
  return (
    <div
      data-slot="attachment-actions"
      className={cn('flex shrink-0 items-center gap-1', className)}
      {...props}
    />
  );
}

/** A button at the end of the row: ghost and small unless told otherwise. */
function AttachmentAction({
  variant = 'ghost',
  size = 'sm',
  ...props
}: React.ComponentProps<typeof Button>) {
  return (
    <Button
      data-slot="attachment-action"
      variant={variant}
      size={size}
      {...props}
    />
  );
}

export {
  Attachment,
  AttachmentMedia,
  AttachmentContent,
  AttachmentTitle,
  AttachmentDescription,
  AttachmentActions,
  AttachmentAction
};
