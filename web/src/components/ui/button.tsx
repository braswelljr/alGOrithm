'use client';

import type * as React from 'react';
import { mergeProps } from '@base-ui/react/merge-props';
import { useRender } from '@base-ui/react/use-render';
import { cva, type VariantProps } from 'class-variance-authority';
import { Spinner } from '@/components/ui/spinner';
import { cn } from '@/lib/utils';

export const buttonVariants = cva(
  'focus-visible:ring-accent focus-visible:ring-offset-bg relative inline-flex shrink-0 items-center justify-center gap-2 text-sm font-medium whitespace-nowrap transition-[color,background-color,border-color,box-shadow,transform,filter] duration-100 ease-out focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:outline-none disabled:pointer-events-none disabled:opacity-50 disabled:saturate-50 has-[input:focus-visible]:ring-2 has-[input:focus-visible]:ring-offset-2 [&_svg]:pointer-events-none [&_svg]:size-4 [&_svg]:shrink-0 [&:has(+input:focus-visible)]:ring-2 [&:has(+input:focus-visible)]:ring-offset-2',
  {
    variants: {
      variant: {
        // Every button is a keycap (after gnux): a lit top over a skirt it
        // sinks into when pressed. Primary is the accent cap, as Return is.
        primary: 'keycap keycap-primary hover:brightness-105',
        destructive: 'keycap keycap-danger hover:brightness-105',
        // A destructive action among others, such as Suspend in a bulk bar or
        // Revoke on a row: shaped like its neighbours, red text on a faint red,
        // a step deeper on hover. The main destructive action is `destructive`.
        'destructive-secondary':
          'keycap keycap-mod text-danger! hover:brightness-[1.03] dark:hover:brightness-110',
        success: 'keycap keycap-success hover:brightness-105',
        info: 'keycap keycap-primary hover:brightness-105',
        warning: 'keycap keycap-warning hover:brightness-105',
        outline: 'keycap hover:brightness-[1.03] dark:hover:brightness-110',
        // Solid fills, never outlines (as on braswelljr.vercel.app and gnux):
        // secondary is the brand's chocolate tint, ghost the neutral surface.
        // Every pair is AA (test/ui/buttons.test.ts).
        secondary: 'keycap keycap-secondary hover:brightness-[1.03] dark:hover:brightness-110',
        // The modifier cap: the quiet key, held down while pressed or on.
        ghost:
          'keycap keycap-mod hover:brightness-[1.03] aria-pressed:brightness-95 data-pressed:brightness-95 dark:hover:brightness-110',
        // Not a keycap: a menu's resting item, whose chosen one is the cap.
        plain: 'text-fg hover:bg-fg/10 active:bg-fg/14 bg-transparent',
        link: 'text-accent underline-offset-4 hover:underline',
        'link-hover-in':
          'after:bg-primary relative after:absolute after:bottom-2 after:h-px after:w-2/3 after:origin-bottom-left after:scale-x-100 after:transition-transform after:duration-300 after:ease-in-out hover:after:origin-bottom-right hover:after:scale-x-0',
        'link-hover-out':
          'after:bg-primary relative after:absolute after:bottom-2 after:h-px after:w-2/3 after:origin-bottom-right after:scale-x-0 after:transition-transform after:duration-300 after:ease-in-out hover:after:origin-bottom-left hover:after:scale-x-100'
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
        sm: 'text-xsm h-8 gap-1 rounded-md px-2.5 in-data-[slot=button-group]:rounded-lg has-data-[icon=inline-end]:pr-1.5 has-data-[icon=inline-start]:pl-1.5 [&_svg:not([class*="size-"])]:size-3.5',
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
