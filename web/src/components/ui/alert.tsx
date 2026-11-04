import * as React from 'react';
import { cva, type VariantProps } from 'class-variance-authority';
import { cn } from '@/lib/utils';

type AlertPosition = 'left' | 'right' | 'top';

const alertVariants = cva(
  [
    'relative flex w-full gap-3 rounded-lg p-3',
    // A title and a description written straight inside the alert, without
    // `AlertContent`, stack under each other (they used to sit side by side
    // as two columns), with an icon before them spanning both lines.
    'has-[>[data-slot=alert-title],>[data-slot=alert-description]]:grid has-[>[data-slot=alert-title],>[data-slot=alert-description]]:gap-y-1',
    'grid-cols-[minmax(0,1fr)] has-[>svg]:grid-cols-[auto_minmax(0,1fr)]',
    '[&>svg]:row-span-2 [&>svg]:shrink-0 [&>svg:not([class*=size-])]:size-5 [&>svg~*]:col-start-2'
  ],
  {
    variants: {
      variant: {
        // Solid fills with the ink that belongs on each, AA in both modes; no outline.
        default: 'bg-subtle text-fg',
        destructive: 'bg-red-50 text-red-800 dark:bg-red-950 dark:text-red-200',
        info: 'bg-blue-50 text-blue-800 dark:bg-blue-950 dark:text-blue-200',
        warning: 'bg-yellow-50 text-yellow-800 dark:bg-yellow-950 dark:text-yellow-200',
        success: 'bg-green-50 text-green-800 dark:bg-green-950 dark:text-green-200'
      },
      mediaPosition: {
        // 2. Use 'flex-row-reverse' for right alignment instead of absolute positioning
        left: 'flex-row',
        right: 'flex-row-reverse'
      }
    },
    defaultVariants: {
      variant: 'default',
      mediaPosition: 'left'
    }
  }
);

const Alert = React.forwardRef<
  HTMLDivElement,
  React.HTMLAttributes<HTMLDivElement> & VariantProps<typeof alertVariants>
>(({ className, variant, mediaPosition, children, ...props }, ref) => {
  return (
    <div
      ref={ref}
      // Only a failure interrupts: a destructive alert is announced at once
      // (`alert`), every other one politely when it appears (`status`). A
      // caller passes `role` to say otherwise, `region` for one that is part
      // of the page rather than news.
      role={variant === 'destructive' ? 'alert' : 'status'}
      data-slot="alert"
      data-variant={variant ?? 'default'}
      className={cn(alertVariants({ variant, mediaPosition }), className)}
      {...props}
    >
      {children}
    </div>
  );
});
Alert.displayName = 'Alert';

const AlertContent = React.forwardRef<HTMLDivElement, React.HTMLAttributes<HTMLDivElement>>(
  ({ className, ...props }, ref) => (
    <div
      ref={ref}
      className={cn('flex min-w-0 flex-1 flex-col gap-1', className)}
      {...props}
    />
  )
);
AlertContent.displayName = 'AlertContent';

const AlertTitle = React.forwardRef<HTMLParagraphElement, React.HTMLAttributes<HTMLHeadingElement>>(
  ({ className, ...props }, ref) => (
    <h5
      ref={ref}
      data-slot="alert-title"
      className={cn('leading-snug font-medium', className)}
      {...props}
    />
  )
);
AlertTitle.displayName = 'AlertTitle';

const AlertDescription = React.forwardRef<
  HTMLParagraphElement,
  React.HTMLAttributes<HTMLParagraphElement>
>(({ className, ...props }, ref) => (
  <div
    ref={ref}
    data-slot="alert-description"
    className={cn('text-sm [&_p]:leading-relaxed', className)}
    {...props}
  />
));
AlertDescription.displayName = 'AlertDescription';

const alertMediaVariants = cva(
  'flex h-full w-max shrink-0 flex-col items-center justify-center [&_svg]:pointer-events-none [&_svg]:shrink-0',
  {
    variants: {
      variant: {
        default: 'bg-transparent',
        icon: 'bg-subtle text-fg flex size-10 shrink-0 items-center justify-center rounded-lg [&_svg:not([class*="size-"])]:size-6',
        destructive:
          'flex size-10 shrink-0 items-center justify-center rounded-lg bg-red-100 text-red-700 dark:bg-red-900/50 dark:text-red-300 [&_svg:not([class*="size-"])]:size-6',
        info: 'flex size-10 shrink-0 items-center justify-center rounded-lg bg-blue-100 text-blue-700 dark:bg-blue-900/50 dark:text-blue-300 [&_svg:not([class*="size-"])]:size-6',
        warning:
          'flex size-10 shrink-0 items-center justify-center rounded-lg bg-yellow-100 text-yellow-700 dark:bg-yellow-900/50 dark:text-yellow-300 [&_svg:not([class*="size-"])]:size-6',
        success:
          'flex size-10 shrink-0 items-center justify-center rounded-lg bg-green-100 text-green-700 dark:bg-green-900/50 dark:text-green-300 [&_svg:not([class*="size-"])]:size-6'
      },
      position: {
        left: '',
        right: ''
      }
    },
    defaultVariants: {
      variant: 'default',
      position: 'left'
    }
  }
);

const AlertMedia = React.forwardRef<
  HTMLDivElement,
  React.HTMLAttributes<HTMLDivElement> & VariantProps<typeof alertMediaVariants>
>(({ className, variant, ...props }, ref) => (
  <div
    ref={ref}
    data-slot="alert-media"
    data-variant={variant}
    className={cn(alertMediaVariants({ variant }), className)}
    {...props}
  />
));
AlertMedia.displayName = 'AlertMedia';

// An alert's actions always sit at its end: below the text with
// `position="bottom"`, or pushed to the far end of the row with `end`, so a
// button never sits under the first word of the message.
const alertActionsVariants = cva('flex flex-wrap items-center justify-end gap-2', {
  variants: {
    position: {
      // Below the text, inside AlertContent
      bottom: 'mt-2 self-stretch',
      // Directly inside Alert, at the far end of its row
      end: 'ms-auto shrink-0 self-center'
    }
  },
  defaultVariants: {
    position: 'bottom'
  }
});

type AlertActionsProps = React.HTMLAttributes<HTMLDivElement> &
  VariantProps<typeof alertActionsVariants>;

const AlertActions = React.forwardRef<HTMLDivElement, AlertActionsProps>(
  ({ className, position, ...props }, ref) => (
    <div
      ref={ref}
      data-slot="alert-actions"
      className={cn(alertActionsVariants({ position }), className)}
      {...props}
    />
  )
);
AlertActions.displayName = 'AlertActions';

export { Alert, AlertTitle, AlertDescription, AlertMedia, AlertContent, AlertActions };

export type { AlertPosition };
