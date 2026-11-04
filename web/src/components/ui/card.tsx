'use client';

import * as React from 'react';
import { mergeProps } from '@base-ui/react/merge-props';
import { useRender } from '@base-ui/react/use-render';
import { cva, type VariantProps } from 'class-variance-authority';
import { cn } from '@/lib/utils';

/**
 * A frame that fuses a stack of `Card`s into one panel: the frame draws the
 * outline, and the cards inside it become rows rather than separate objects.
 *
 * `Card` outlines with `ring-1`, which is a box-shadow and has no layout box to
 * pull in, so the children simply drop their own ring and radius, and
 * `overflow-hidden` on the frame clips the corners. No `clip-path` inset or
 * negative margin is needed.
 */
function CardFrame({
  className,
  render,
  ...props
}: useRender.ComponentProps<'div'>): React.ReactElement {
  const defaultProps = {
    className: cn(
      'bg-card text-fg ring-border relative flex flex-col overflow-hidden rounded-xl text-sm ring-1',
      // Children stop being their own object: no ring, no radius, no shadow.
      '*:data-[slot=card]:rounded-none *:data-[slot=card]:bg-transparent *:data-[slot=card]:shadow-none *:data-[slot=card]:ring-0',
      // A hairline between stacked rows, but not above the first one.
      '*:not-first:data-[slot=card]:border-border *:not-first:data-[slot=card]:border-t',
      className
    ),
    'data-slot': 'card-frame'
  };

  return useRender({
    defaultTagName: 'div',
    props: mergeProps<'div'>(defaultProps, props),
    render
  });
}

/** Chrome at the frame level, above the stacked cards. */
function CardFrameHeader({
  className,
  render,
  ...props
}: useRender.ComponentProps<'div'>): React.ReactElement {
  const defaultProps = {
    className: cn(
      'relative grid auto-rows-min grid-rows-[auto_auto] flex-col items-start gap-x-4 px-6 py-4 has-data-[slot=card-frame-action]:grid-cols-[1fr_auto]',
      className
    ),
    'data-slot': 'card-frame-header'
  };

  return useRender({
    defaultTagName: 'div',
    props: mergeProps<'div'>(defaultProps, props),
    render
  });
}

function CardFrameTitle({
  className,
  render,
  ...props
}: useRender.ComponentProps<'div'>): React.ReactElement {
  const defaultProps = {
    className: cn('font-heading text-fg text-sm font-semibold', className),
    'data-slot': 'card-frame-title'
  };

  return useRender({
    defaultTagName: 'div',
    props: mergeProps<'div'>(defaultProps, props),
    render
  });
}

function CardFrameDescription({
  className,
  render,
  ...props
}: useRender.ComponentProps<'div'>): React.ReactElement {
  const defaultProps = {
    className: cn('text-fg text-sm', className),
    'data-slot': 'card-frame-description'
  };

  return useRender({
    defaultTagName: 'div',
    props: mergeProps<'div'>(defaultProps, props),
    render
  });
}

function CardFrameAction({
  className,
  render,
  ...props
}: useRender.ComponentProps<'div'>): React.ReactElement {
  const defaultProps = {
    className: cn(
      'col-start-2 row-span-2 row-start-1 inline-flex self-center justify-self-end',
      className
    ),
    'data-slot': 'card-frame-action'
  };

  return useRender({
    defaultTagName: 'div',
    props: mergeProps<'div'>(defaultProps, props),
    render
  });
}

function CardFrameFooter({
  className,
  render,
  ...props
}: useRender.ComponentProps<'div'>): React.ReactElement {
  const defaultProps = {
    className: cn('px-6 py-4', className),
    'data-slot': 'card-frame-footer'
  };

  return useRender({
    defaultTagName: 'div',
    props: mergeProps<'div'>(defaultProps, props),
    render
  });
}

/**
 * The card's looks. A card is always this component (docs/rules/cards-use-the-card-component.md):
 * add a variant here rather than a bordered, rounded div elsewhere.
 *
 * - default: a raised surface with a hairline ring.
 * - outline: a ring and no fill, for grouping on a card or a tinted page.
 * - ghost: no ring, a subtle fill, for quiet panels.
 * - interactive: a card that is one big link or button; it lifts on hover
 *   (transform only), shows the focus ring of whatever it contains, and can be
 *   pressed.
 * - selected: the chosen card of a set, in the accent, said in text as well.
 * - accent: a tinted callout card (the TL;DR card).
 *
 * Sizes set the padding (--card-spacing), so nested corners follow
 * outer radius = inner radius + padding.
 */
const cardVariants = cva(
  'group/card text-fg smooth-corners relative flex flex-col gap-(--card-spacing) rounded-xl py-(--card-spacing) text-sm has-data-[slot=card-footer]:pb-0 has-[>img:first-child]:pt-0 *:[img:first-child]:rounded-t-xl *:[img:last-child]:rounded-b-xl',
  {
    variants: {
      variant: {
        default: 'bg-card ring-border shadow-xs ring-1',
        outline: 'ring-border bg-transparent ring-1',
        ghost: 'bg-subtle',
        interactive:
          'bg-card ring-border hover:ring-control has-focus-visible:ring-accent shadow-xs ring-1 transition-[transform,box-shadow] duration-(--duration-base) ease-(--ease-out-expo) hover:-translate-y-0.5 hover:shadow-md active:translate-y-0 has-focus-visible:ring-2 motion-reduce:hover:translate-y-0',
        selected: 'bg-accent/5 ring-accent shadow-xs ring-2',
        accent: 'bg-accent/5 ring-accent/30 ring-1'
      },
      size: {
        sm: '[--card-spacing:--spacing(3)]',
        md: '[--card-spacing:--spacing(4)]',
        default: '[--card-spacing:--spacing(5)]',
        lg: '[--card-spacing:--spacing(6)]'
      }
    },
    defaultVariants: { variant: 'default', size: 'default' }
  }
);

export type CardVariant = NonNullable<VariantProps<typeof cardVariants>['variant']>;

function Card({
  className,
  size = 'default',
  variant = 'default',
  render,
  ...props
}: useRender.ComponentProps<'div'> & VariantProps<typeof cardVariants>): React.ReactElement {
  // `render` lets a card be another element (an <article>, the <label> of a
  // choice, a link) while keeping the card's own look.
  //
  // No `overflow-hidden`: a card is a flex item in the page's column, and
  // `overflow: hidden` sets its automatic minimum height to zero, so a card on
  // a long page could shrink and clip its own content.
  const defaultProps = {
    'data-slot': 'card',
    'data-size': size,
    'data-variant': variant,
    className: cn(cardVariants({ variant, size }), className)
  };
  return useRender({
    defaultTagName: 'div',
    render,
    props: mergeProps<'div'>(defaultProps, props)
  });
}

function CardHeader({ className, ...props }: React.ComponentProps<'div'>) {
  return (
    <div
      data-slot="card-header"
      className={cn(
        'group/card-header @container/card-header grid auto-rows-min items-start gap-1 rounded-t-xl px-(--card-spacing) has-data-[slot=card-action]:grid-cols-[1fr_auto] has-data-[slot=card-description]:grid-rows-[auto_auto] [.border-b]:pb-(--card-spacing)',
        className
      )}
      {...props}
    />
  );
}

function CardTitle({ className, ...props }: React.ComponentProps<'div'>) {
  return (
    <div
      data-slot="card-title"
      className={cn(
        'font-heading text-base leading-snug font-medium group-data-[size=sm]/card:text-sm',
        className
      )}
      {...props}
    />
  );
}

function CardDescription({ className, ...props }: React.ComponentProps<'div'>) {
  return (
    <div
      data-slot="card-description"
      className={cn('text-muted text-sm', className)}
      {...props}
    />
  );
}

function CardAction({ className, ...props }: React.ComponentProps<'div'>) {
  return (
    <div
      data-slot="card-action"
      className={cn('col-start-2 row-span-2 row-start-1 self-start justify-self-end', className)}
      {...props}
    />
  );
}

function CardContent({ className, ...props }: React.ComponentProps<'div'>) {
  return (
    <div
      data-slot="card-content"
      className={cn('px-(--card-spacing)', className)}
      {...props}
    />
  );
}

function CardFooter({ className, ...props }: React.ComponentProps<'div'>) {
  return (
    <div
      data-slot="card-footer"
      className={cn(
        'border-border bg-subtle flex items-center rounded-b-xl border-t p-(--card-spacing)',
        className
      )}
      {...props}
    />
  );
}

export {
  cardVariants,
  Card,
  CardHeader,
  CardFooter,
  CardTitle,
  CardAction,
  CardDescription,
  CardContent,
  // Back-compat: `CardPanel` was the previous name for the card body slot.
  CardContent as CardPanel,
  CardFrame,
  CardFrameHeader,
  CardFrameTitle,
  CardFrameDescription,
  CardFrameAction,
  CardFrameFooter
};
