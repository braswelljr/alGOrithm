import { ComponentProps } from 'react';
import { cva, VariantProps } from 'class-variance-authority';
import { cn } from '@/lib/utils';

const labelVariants = cva(
  'text-sm leading-none font-medium peer-disabled:cursor-not-allowed peer-disabled:opacity-70'
);

/**
 * The asterisk a required field's label carries.
 *
 * Declared once, here, rather than written at each call site as a red `<span>`.
 * Two reasons it is worth a component: the colour is a decision (stone-500
 * would fail contrast in light mode, and the dark step mirrors red-600 as
 * red-400), and a hand-written one drifts: the app already held it in three
 * shades across two scales.
 *
 * The star is hidden from screen readers and "(required)" is read in its
 * place. Most controls here are validated by the form rather than carrying
 * `required` themselves, so the label is the only place the obligation is
 * stated, and "star" is not a word anybody should have to interpret.
 */
export function RequiredMark({ className }: { className?: string }) {
  return (
    <span className={cn('text-red-600 dark:text-red-400', className)}>
      <span aria-hidden="true">*</span>
      <span className="sr-only"> (required)</span>
    </span>
  );
}

export function Label({
  className,
  required,
  children,
  ...props
}: ComponentProps<'label'> &
  VariantProps<typeof labelVariants> & {
    /**
     * Whether the control this labels must be answered.
     *
     * Marks the label so an unanswered obligation is found by scanning rather
     * than by submitting and reading the errors. Say it here rather than
     * appending a `<span>` to the text: a label is what the reader sees, and an
     * asterisk inside the wording is part of the wording.
     */
    required?: boolean;
  }) {
  return (
    <label
      data-slot="label"
      className={cn(labelVariants(), className)}
      {...props}
    >
      {children}
      {required && <RequiredMark className="ml-0.5" />}
    </label>
  );
}
