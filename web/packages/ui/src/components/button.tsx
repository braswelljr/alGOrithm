'use client';

import type * as React from 'react';
import { mergeProps } from '@base-ui/react/merge-props';
import { useRender } from '@base-ui/react/use-render';
import { Spinner } from '@braswelljr/algorithm-ui/components/spinner';
import { cn } from '@braswelljr/algorithm-ui/lib/utils';
import { cva, type VariantProps } from 'class-variance-authority';

export const buttonVariants = cva(
  'relative inline-flex shrink-0 items-center justify-center gap-2 text-sm font-medium whitespace-nowrap transition-[color,background-color,border-color,box-shadow,scale] duration-150 ease-out focus-visible:ring-2 focus-visible:ring-accent focus-visible:ring-offset-2 focus-visible:ring-offset-bg focus-visible:outline-none not-disabled:not-aria-disabled:active:scale-[0.98] disabled:pointer-events-none disabled:opacity-50 disabled:saturate-50 has-[input:focus-visible]:ring-2 has-[input:focus-visible]:ring-offset-2 [&_svg]:pointer-events-none [&_svg]:size-4 [&_svg]:shrink-0 [&:has(+input:focus-visible)]:ring-2 [&:has(+input:focus-visible)]:ring-offset-2',
  {
    variants: {
      variant: {
        primary:
          'bg-accent text-accent-fg shadow-xs hover:bg-primary-700 dark:hover:bg-primary-400',
        destructive: 'bg-danger text-white hover:bg-danger/90 dark:text-bg',
        // A destructive action among others, such as Suspend in a bulk bar or
        // Revoke on a row: shaped like its neighbours, red text on a faint red,
        // a step deeper on hover. The main destructive action is `destructive`.
        'destructive-secondary': 'bg-danger/10 text-danger hover:bg-danger/15',
        success: 'bg-success text-white hover:bg-success/90 dark:text-bg',
        info: 'bg-primary-600 text-accent-fg hover:bg-primary-700',
        warning: 'bg-warning text-white hover:bg-warning/90 dark:text-bg',
        outline: 'border border-control bg-transparent text-fg hover:bg-subtle',
        secondary: 'border border-border bg-card text-fg shadow-xs hover:bg-subtle',
        ghost: 'bg-transparent text-fg hover:bg-fg/8 active:bg-fg/12 aria-pressed:bg-fg/10 data-pressed:bg-fg/10',
        link: 'text-accent underline-offset-4 hover:underline',
        'link-hover-in':
          'relative after:absolute after:bottom-2 after:h-px after:w-2/3 after:origin-bottom-left after:scale-x-100 after:bg-accent after:transition-transform after:duration-300 after:ease-in-out hover:after:origin-bottom-right hover:after:scale-x-0',
        'link-hover-out':
          'relative after:absolute after:bottom-2 after:h-px after:w-2/3 after:origin-bottom-right after:scale-x-0 after:bg-accent after:transition-transform after:duration-300 after:ease-in-out hover:after:origin-bottom-left hover:after:scale-x-100'
      },
      size: {
        // One height scale with the fields, by name: xs 28, sm 32, default 36,
        // md 40, lg 44, xl 48 - the same as `Input`, `InputGroup`,
        // `SelectTrigger` and `NumberField`. `md` is the default everywhere,
        // and the plain `icon` is 40 too, so a button or an icon button set
        // beside a field of the same size lines up with it.
        default:
          'h-9 gap-1.5 px-2.5 has-data-[icon=inline-end]:pr-2 has-data-[icon=inline-start]:pl-2',
        // 40, the system's control height and the size a button takes unless
        // it asks for another, so a button beside a default input lines up.
        md: 'h-10 gap-1.5 px-3 has-data-[icon=inline-end]:pr-2.5 has-data-[icon=inline-start]:pl-2.5',
        xs: 'h-7 gap-1 rounded-md px-2 text-sm in-data-[slot=button-group]:rounded-lg has-data-[icon=inline-end]:pr-1.5 has-data-[icon=inline-start]:pl-1.5 [&_svg:not([class*="size-"])]:size-3',
        sm: 'h-8 gap-1 rounded-md px-2.5 text-xsm in-data-[slot=button-group]:rounded-lg has-data-[icon=inline-end]:pr-1.5 has-data-[icon=inline-start]:pl-1.5 [&_svg:not([class*="size-"])]:size-3.5',
        lg: 'h-11 gap-1.5 px-3.5 has-data-[icon=inline-end]:pr-3 has-data-[icon=inline-start]:pl-3',
        xl: 'h-12 gap-2 px-4 has-data-[icon=inline-end]:pr-3.5 has-data-[icon=inline-start]:pl-3.5',
        // Icon buttons never shrink in a tight row, so they stay square.
        icon: 'size-10 shrink-0',
        'icon-xs':
          'size-7 shrink-0 rounded-md in-data-[slot=button-group]:rounded-lg [&_svg:not([class*="size-"])]:size-3',
        'icon-sm': 'size-8 shrink-0 rounded-md in-data-[slot=button-group]:rounded-lg',
        'icon-lg': 'size-11 shrink-0'
      },
      /**
       * The corner.
       *
       * `default` platform-wide: the pill is the register's language, not the
       * whole design system's, so the surfaces that want one ask for it -
       * `rounded="full"` here, or the descendant rule the data-table toolbar
       * uses to dress every control in the bar at once.
       */
      rounded: {
        none: 'rounded-none',
        default: 'rounded',
        sm: 'rounded-sm',
        md: 'rounded-md',
        lg: 'rounded-xl',
        xl: 'rounded-2xl',
        full: 'rounded-full'
      }
    },
    defaultVariants: {
      variant: 'primary',
      size: 'md',
      rounded: 'default'
    }
  }
);
export type ButtonProps = useRender.ComponentProps<'button'> & {
  variant?: VariantProps<typeof buttonVariants>['variant'];
  size?: VariantProps<typeof buttonVariants>['size'];
  /**
   * The corner.
   *
   * The variant has always been there and nothing could reach it: `rounded`
   * was missing from these props, so `<Button rounded="full">` was a type
   * error and the only way to change a corner was a `className` fighting the
   * base. This is what the register surfaces use to ask for a pill.
   */
  rounded?: VariantProps<typeof buttonVariants>['rounded'];
  loading?: boolean;
};
export function Button({
  className,
  variant,
  size,
  rounded,
  render,
  children,
  loading = false,
  disabled: disabledProp,
  ...props
}: ButtonProps): React.ReactElement {
  const isDisabled: boolean = Boolean(loading || disabledProp);
  const typeValue: React.ButtonHTMLAttributes<HTMLButtonElement>['type'] = render
    ? undefined
    : 'button';
  const defaultProps = {
    children: (
      <>
        {/* While it works, the label keeps its room and its name (screen
            readers still hear "Save") but is not drawn, so the spinner sits
            alone in the middle rather than over the words. */}
        {loading ? (
          <span
            className="inline-flex items-center justify-center gap-[inherit] opacity-0"
            data-slot="button-label"
          >
            {children}
          </span>
        ) : (
          children
        )}
        {loading && (
          <Spinner
            className="pointer-events-none absolute"
            data-slot="button-loading-indicator"
          />
        )}
      </>
    ),
    className: cn(buttonVariants({ className, rounded, size, variant })),
    'aria-busy': loading || undefined,
    'aria-disabled': loading || undefined,
    /**
     * Marks a button that is nothing but a glyph, so a surface can square it
     * off without having to guess from the markup.
     *
     * The toolbar tried to infer it with `:has(> svg:only-child)` and got it
     * wrong, because `:only-child` counts elements and ignores text: a button
     * reading `<svg/>Reset` has the svg as its only element child, matched,
     * and was squashed into a square with its label clipped. The size is the
     * thing that actually says "this is an icon", so the size says it.
     */
    'data-icon-only': size?.startsWith('icon') ? '' : undefined,
    'data-loading': loading ? '' : undefined,
    'data-slot': 'button',
    disabled: isDisabled,
    type: typeValue
  };
  return useRender({
    defaultTagName: 'button',
    props: mergeProps<'button'>(defaultProps, props),
    render
  });
}
