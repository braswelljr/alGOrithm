import * as React from 'react';
import { cva, VariantProps } from 'class-variance-authority';
import { cn } from '@/lib/utils';

const textareaVariants = cva(
  `border-border bg-card text-fg placeholder:text-muted/80 focus-visible:border-control focus-visible:ring-fg/10 [[readonly]]:bg-subtle flex w-full border px-3 py-2 shadow-xs shadow-black/5 transition-[color,box-shadow] focus-visible:ring-3 focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-60 aria-invalid:border-red-600 aria-invalid:ring-red-600/20 dark:aria-invalid:border-red-400 dark:aria-invalid:ring-red-400/25`,
  {
    variants: {
      size: {
        sm: 'min-h-16 text-sm',
        // 16 on a phone, so the browser does not zoom the page into the field.
        default: 'min-h-24 text-base sm:text-sm',
        lg: 'min-h-36 text-base'
      },
      radius: {
        none: 'rounded-none',
        sm: 'rounded-sm',
        default: 'rounded-md',
        lg: 'rounded-lg',
        full: 'rounded-lg'
      }
    },
    defaultVariants: {
      size: 'default',
      radius: 'default'
    }
  }
);

type TextareaProps = React.ComponentProps<'textarea'> & VariantProps<typeof textareaVariants>;

function Textarea({ className, size, radius, ...props }: TextareaProps) {
  return (
    <textarea
      data-slot="textarea"
      className={cn(textareaVariants({ size, radius }), className)}
      {...props}
    />
  );
}

export { Textarea, type TextareaProps };
