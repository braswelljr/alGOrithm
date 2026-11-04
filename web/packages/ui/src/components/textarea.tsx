import * as React from 'react';
import { cn } from '@braswelljr/algorithm-ui/lib/utils';
import { cva, VariantProps } from 'class-variance-authority';

const textareaVariants = cva(
  `flex w-full border border-border bg-card px-3 py-2 text-fg shadow-xs shadow-black/5 transition-[color,box-shadow] placeholder:text-muted/80 focus-visible:border-control focus-visible:ring-3 focus-visible:ring-fg/10 focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-60 aria-invalid:border-red-600 aria-invalid:ring-red-600/20 dark:aria-invalid:border-red-400 dark:aria-invalid:ring-red-400/25 [[readonly]]:bg-subtle`,
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
