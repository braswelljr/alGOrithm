'use client';

import * as React from 'react';
import { cva, type VariantProps } from 'class-variance-authority';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea, type TextareaProps } from '@/components/ui/textarea';
import { cn } from '@/lib/utils';

const inputGroupVariants = cva(
  [
    'group/input-group border-border bg-card relative flex w-full items-center overflow-hidden border shadow-xs transition-[color,box-shadow] outline-none',
    'has-[>textarea]:h-auto',

    'has-[>[data-align=inline-start]]:[&>input]:pl-2',
    'has-[>[data-align=inline-end]]:[&>input]:pr-2',
    'has-[>[data-align=block-start]]:h-auto has-[>[data-align=block-start]]:flex-col has-[>[data-align=block-start]]:[&>input]:pb-3',
    'has-[>[data-align=block-end]]:h-auto has-[>[data-align=block-end]]:flex-col has-[>[data-align=block-end]]:[&>input]:pt-3',

    'hover:border-border',

    // The field's focus, the same as `Input`'s: a darker edge and a soft halo.
    'has-[[data-slot=input-group-control]:focus-visible]:border-control',
    'has-[[data-slot=input-group-control]:focus-visible]:ring-3',
    'has-[[data-slot=input-group-control]:focus-visible]:ring-fg/10',
    'has-[[data-slot=input-group-control]:focus-visible]:outline-none',
    '',
    '',

    'has-[[data-slot][aria-invalid=true]]:border-red-600',
    'has-[[data-slot][aria-invalid=true]]:ring-red-600/20',
    'dark:has-[[data-slot][aria-invalid=true]]:border-red-400',
    'dark:has-[[data-slot][aria-invalid=true]]:ring-red-400/25'
  ],
  {
    variants: {
      size: {
        // The shared control scale; see `Button`.
        lg: 'h-11',
        md: 'h-10',
        default: 'h-9',
        sm: 'h-8',
        xs: 'h-7'
      },
      radius: {
        none: 'rounded-none',
        sm: 'rounded-sm',
        default: 'rounded-md',
        lg: 'rounded-lg',
        // A pill, unless the box is tall.
        //
        // `has-[>textarea]` and the two block-aligned addons are already the
        // three things that make this group grow past a single line - the base
        // sets `h-auto` on each - and a pill curve on a box several lines tall
        // cuts into the first and last line of text. The chat composer is the
        // case that showed it: a textarea with a toolbar row beneath, drawn as
        // a lozenge.
        //
        // Written into `full` rather than the base so it only ever corrects
        // the pill: a caller who asks for `none` or `sm` still gets it, tall
        // box or not. The `:has()` selector outranks a plain class, which is
        // what makes this win without an `!`.
        full:
          'rounded-full has-[>textarea]:rounded-2xl ' +
          'has-[>[data-align=block-end]]:rounded-2xl has-[>[data-align=block-start]]:rounded-2xl'
      }
    },
    defaultVariants: {
      // 40, the system's control height; see `Input`.
      size: 'md',
      radius: 'default'
    }
  }
);

function InputGroup({
  className,
  size,
  radius,
  ...props
}: React.ComponentProps<'div'> & VariantProps<typeof inputGroupVariants>) {
  return (
    <div
      data-slot="input-group"
      role="group"
      className={cn(inputGroupVariants({ size, radius }), className)}
      {...props}
    />
  );
}

const inputGroupAddonVariants = cva(
  "text-muted flex h-auto cursor-text items-center justify-center gap-2 py-1.5 text-sm font-medium select-none group-data-[disabled=true]/input-group:opacity-50 [&>kbd]:rounded-[calc(var(--radius)-5px)] [&>svg:not([class*='size-'])]:size-4",
  {
    variants: {
      align: {
        'inline-start': 'order-first pl-3 has-[>button]:ml-[-0.45rem] has-[>kbd]:ml-[-0.35rem]',
        'inline-end': 'order-last pr-3 has-[>button]:mr-[-0.4rem] has-[>kbd]:mr-[-0.35rem]',
        'block-start':
          'order-first w-full justify-start px-3 pt-3 group-has-[>input]/input-group:pt-2.5 [.border-b]:pb-3',
        'block-end':
          'order-last w-full justify-start px-3 pb-3 group-has-[>input]/input-group:pb-2.5 [.border-t]:pt-3'
      }
    },
    defaultVariants: {
      align: 'inline-start'
    }
  }
);

function InputGroupAddon({
  className,
  align = 'inline-start',
  ...props
}: React.ComponentProps<'div'> & VariantProps<typeof inputGroupAddonVariants>) {
  return (
    <div
      role="group"
      data-slot="input-group-addon"
      data-align={align}
      className={cn(inputGroupAddonVariants({ align }), className)}
      onClick={(e) => {
        if ((e.target as HTMLElement).closest('button')) {
          return;
        }
        e.currentTarget.parentElement?.querySelector('input')?.focus();
      }}
      {...props}
    />
  );
}

const inputGroupButtonVariants = cva('flex items-center gap-2 text-sm shadow-none', {
  variants: {
    size: {
      xs: "h-6 gap-1 rounded-[calc(var(--radius)-5px)] px-2 has-[>svg]:px-2 [&>svg:not([class*='size-'])]:size-3.5",
      sm: 'h-8 gap-1.5 rounded-md px-2.5 has-[>svg]:px-2.5',
      'icon-xs': 'size-6 shrink-0 rounded-[calc(var(--radius)-5px)] p-0 has-[>svg]:p-0',
      'icon-sm': 'size-8 shrink-0 p-0 has-[>svg]:p-0'
    }
  },
  defaultVariants: {
    size: 'xs'
  }
});

function InputGroupButton({
  className,
  type = 'button',
  variant = 'ghost',
  size = 'xs',
  ...props
}: Omit<React.ComponentProps<typeof Button>, 'size'> &
  VariantProps<typeof inputGroupButtonVariants>) {
  return (
    <Button
      type={type}
      data-size={size}
      variant={variant}
      className={cn(inputGroupButtonVariants({ size }), className)}
      {...props}
    />
  );
}

function InputGroupText({ className, ...props }: React.ComponentProps<'span'>) {
  return (
    <span
      className={cn(
        "text-muted flex items-center gap-2 text-sm [&_svg]:pointer-events-none [&_svg:not([class*='size-'])]:size-4",
        className
      )}
      {...props}
    />
  );
}

function InputGroupInput({ className, ...props }: React.ComponentProps<typeof Input>) {
  return (
    <Input
      data-slot="input-group-control"
      className={cn(
        'flex-1 rounded-none border-0 bg-transparent! shadow-none',
        'focus-visible:border-transparent focus-visible:ring-0',
        className
      )}
      {...props}
    />
  );
}

function InputGroupTextarea({ className, ...props }: TextareaProps) {
  return (
    <Textarea
      data-slot="input-group-control"
      className={cn(
        'flex-1 resize-none rounded-none border-0 bg-transparent py-3 shadow-none focus-visible:ring-0 dark:bg-transparent',
        className
      )}
      {...props}
    />
  );
}

export {
  InputGroup,
  InputGroupAddon,
  InputGroupButton,
  InputGroupInput,
  InputGroupText,
  InputGroupTextarea
};
