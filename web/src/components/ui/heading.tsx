import * as React from 'react';
import { cva, type VariantProps } from 'class-variance-authority';
import { cn } from '@/lib/utils';

/**
 * The app's heading scale.
 *
 * Four sizes, no more. Before this, a page title was `text-xl` here and
 * `text-lg` there, and a section heading was whichever of `text-sm uppercase`,
 * `text-base` or `font-semibold` the last author reached for, so two screens
 * built a week apart did not look like the same product.
 *
 * `level` and `size` are separate on purpose. The level is the document
 * outline, which assistive technology and the semantic-HTML rule care about;
 * the size is how loud it looks. A section that needs to read smaller than its
 * parent keeps its `h2` and takes a smaller `size`, rather than being demoted to
 * an `h3` to borrow the type.
 */
const headingVariants = cva('text-fg', {
  variants: {
    size: {
      /** The one title of a page or a workspace section. */
      page: 'text-xl font-semibold',
      /** A major block within a page: a panel, a card, a region. */
      section: 'text-lg font-semibold',
      /** A block within a block. */
      subsection: 'text-base font-semibold',
      /** A label over a group. Uppercase, so it reads as a rule rather than a
       *  heading competing with the content under it. */
      overline: 'text-xsm text-muted font-semibold tracking-wide uppercase'
    }
  },
  defaultVariants: { size: 'section' }
});

/** The size a level implies when none is given, so the common case is one prop. */
const SIZE_FOR_LEVEL = {
  1: 'page',
  2: 'section',
  3: 'subsection',
  4: 'overline',
  5: 'overline',
  6: 'overline'
} as const;

export type HeadingProps = React.ComponentProps<'h2'> &
  VariantProps<typeof headingVariants> & {
    /** Where this sits in the document outline. Drives the tag, not the type. */
    level?: 1 | 2 | 3 | 4 | 5 | 6;
  };

export function Heading({ level = 2, size, className, ...props }: HeadingProps) {
  const Tag = `h${level}` as const;

  return (
    <Tag
      data-slot="heading"
      className={cn(headingVariants({ size: size ?? SIZE_FOR_LEVEL[level] }), className)}
      {...props}
    />
  );
}

export { headingVariants };
