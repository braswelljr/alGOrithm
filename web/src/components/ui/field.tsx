'use client';

import { useId, useMemo, useRef } from 'react';
import { cva, type VariantProps } from 'class-variance-authority';
import { Label } from '@/components/ui/label';
import { Separator } from '@/components/ui/separator';
import { useIsomorphicLayoutEffect } from '@/hooks/use-isomorphic-layout-effect';
import { cn } from '@/lib/utils';

function FieldSet({ className, ...props }: React.ComponentProps<'fieldset'>) {
  return (
    <fieldset
      data-slot="field-set"
      className={cn(
        'flex flex-col gap-6',
        'has-[>[data-slot=checkbox-group]]:gap-3 has-[>[data-slot=radio-group]]:gap-3',
        className
      )}
      {...props}
    />
  );
}

function FieldLegend({
  className,
  variant = 'legend',
  ...props
}: React.ComponentProps<'legend'> & { variant?: 'legend' | 'label' }) {
  return (
    <legend
      data-slot="field-legend"
      data-variant={variant}
      className={cn(
        'mb-3 font-medium',
        'data-[variant=legend]:text-base',
        'data-[variant=label]:text-sm',
        className
      )}
      {...props}
    />
  );
}

function FieldGroup({ className, ...props }: React.ComponentProps<'div'>) {
  return (
    <div
      data-slot="field-group"
      className={cn(
        'group/field-group @container/field-group flex w-full flex-col gap-7 data-[slot=checkbox-group]:gap-3 *:data-[slot=field-group]:gap-4',
        className
      )}
      {...props}
    />
  );
}

const fieldVariants = cva(
  'group/field flex w-full gap-3 data-[invalid=true]:text-red-600 dark:data-[invalid=true]:text-red-400',
  {
    variants: {
      orientation: {
        vertical: ['flex-col *:w-full [&>.sr-only]:w-auto'],
        horizontal: [
          'flex-row items-center',
          '*:data-[slot=field-label]:flex-auto',
          'has-[>[data-slot=field-content]]:items-start has-[>[data-slot=field-content]]:[&>[role=checkbox],[role=radio]]:mt-px'
        ],
        responsive: [
          'flex-col *:w-full @md/field-group:flex-row @md/field-group:items-center @md/field-group:*:w-auto [&>.sr-only]:w-auto',
          '@md/field-group:*:data-[slot=field-label]:flex-auto',
          '@md/field-group:has-[>[data-slot=field-content]]:items-start @md/field-group:has-[>[data-slot=field-content]]:[&>[role=checkbox],[role=radio]]:mt-px'
        ]
      }
    },
    defaultVariants: {
      orientation: 'vertical'
    }
  }
);

const FIELD_CONTROL =
  'input:not([type="hidden"]),textarea,select,[role="combobox"],[role="checkbox"],[role="switch"],[role="radiogroup"],[role="slider"]';

/**
 * Ties a field's description and error to its control: `aria-describedby`
 * names them, and `aria-invalid` is set while an error shows, unless the page
 * set its own. Done on the DOM after each render, so every form gets it
 * without wiring ids by hand. Only this field's own parts count, not a nested
 * field's.
 */
function useDescribedControl(ref: React.RefObject<HTMLDivElement | null>) {
  const base = useId();
  useIsomorphicLayoutEffect(() => {
    const root = ref.current;
    if (!root) return;
    const own = (el: Element) => el.closest('[data-slot="field"]') === root;
    const control = Array.from(root.querySelectorAll<HTMLElement>(FIELD_CONTROL)).find(own);
    if (!control) return;
    const parts = Array.from(
      root.querySelectorAll<HTMLElement>(
        '[data-slot="field-description"],[data-slot="field-error"]'
      )
    ).filter(own);
    const ids = parts.map((part, i) => {
      if (!part.id) part.id = `${base}-part-${i}`;
      return part.id;
    });
    const ours = (control.getAttribute('data-field-describedby') ?? '').split(' ').filter(Boolean);
    const theirs = (control.getAttribute('aria-describedby') ?? '')
      .split(' ')
      .filter((id) => id && !ours.includes(id));
    const next = [...theirs, ...ids.filter((id) => !theirs.includes(id))].join(' ');
    if (next) control.setAttribute('aria-describedby', next);
    else control.removeAttribute('aria-describedby');
    control.setAttribute('data-field-describedby', ids.join(' '));

    const erring = parts.some((p) => p.getAttribute('data-slot') === 'field-error');
    const setByUs = control.getAttribute('data-field-invalid') === 'true';
    if (erring && !control.hasAttribute('aria-invalid')) {
      control.setAttribute('aria-invalid', 'true');
      control.setAttribute('data-field-invalid', 'true');
    } else if (!erring && setByUs) {
      control.removeAttribute('aria-invalid');
      control.removeAttribute('data-field-invalid');
    }
  });
}

function Field({
  className,
  orientation = 'vertical',
  ref: outer,
  ...props
}: React.ComponentProps<'div'> & VariantProps<typeof fieldVariants>) {
  const ref = useRef<HTMLDivElement>(null);
  useDescribedControl(ref);
  return (
    <div
      ref={(node) => {
        ref.current = node;
        if (typeof outer === 'function') outer(node);
        else if (outer) outer.current = node;
      }}
      role="group"
      data-slot="field"
      data-orientation={orientation}
      className={cn(fieldVariants({ orientation }), className)}
      {...props}
    />
  );
}

function FieldContent({ className, ...props }: React.ComponentProps<'div'>) {
  return (
    <div
      data-slot="field-content"
      className={cn(
        // `min-w-0` so a long description or validation message wraps inside
        // the field rather than widening the form.
        'group/field-content flex min-w-0 flex-1 flex-col gap-1.5 leading-snug',
        className
      )}
      {...props}
    />
  );
}

function FieldLabel({ className, ...props }: React.ComponentProps<typeof Label>) {
  return (
    <Label
      data-slot="field-label"
      className={cn(
        'group/field-label peer/field-label flex w-fit gap-2 leading-snug group-data-[disabled=true]/field:opacity-50',
        'has-[>[data-slot=field]]:w-full has-[>[data-slot=field]]:flex-col has-[>[data-slot=field]]:rounded-md has-[>[data-slot=field]]:border *:data-[slot=field]:p-4',
        'has-data-[state=checked]:border-control has-data-[state=checked]:bg-fg/5',
        className
      )}
      {...props}
    />
  );
}

function FieldTitle({ className, ...props }: React.ComponentProps<'div'>) {
  return (
    <div
      data-slot="field-label"
      className={cn(
        'flex w-fit items-center gap-2 text-sm leading-snug font-medium group-data-[disabled=true]/field:opacity-50',
        className
      )}
      {...props}
    />
  );
}

function FieldDescription({ className, ...props }: React.ComponentProps<'p'>) {
  return (
    <p
      data-slot="field-description"
      className={cn(
        'text-muted text-sm leading-normal font-normal group-has-data-[orientation=horizontal]/field:text-balance',
        'last:mt-0 nth-last-2:-mt-1 [[data-variant=legend]+&]:-mt-1.5',
        '[&>a:hover]:text-fg dark:[&>a:hover]:text-bg [&>a]:underline [&>a]:underline-offset-4',
        className
      )}
      {...props}
    />
  );
}

function FieldSeparator({
  children,
  className,
  ...props
}: React.ComponentProps<'div'> & {
  children?: React.ReactNode;
}) {
  return (
    <div
      data-slot="field-separator"
      data-content={!!children}
      className={cn(
        'relative -my-2 h-5 text-sm group-data-[variant=outline]/field-group:-mb-2',
        className
      )}
      {...props}
    >
      <Separator className="absolute inset-0 top-1/2" />
      {children && (
        <span
          className="bg-card text-muted relative mx-auto block w-fit px-2"
          data-slot="field-separator-content"
        >
          {children}
        </span>
      )}
    </div>
  );
}

function FieldError({
  className,
  children,
  errors,
  ...props
}: React.ComponentProps<'div'> & {
  errors?: Array<{ message?: string } | undefined>;
}) {
  const content = useMemo(() => {
    if (children) {
      return children;
    }

    if (!errors) {
      return null;
    }

    if (errors?.length === 1 && errors[0]?.message) {
      return errors[0].message;
    }

    return (
      <ul className="ml-4 flex list-disc flex-col gap-1">
        {errors.map((error, index) => error?.message && <li key={index}>{error.message}</li>)}
      </ul>
    );
  }, [children, errors]);

  if (!content) {
    return null;
  }

  return (
    <div
      role="alert"
      data-slot="field-error"
      className={cn('text-sm font-normal text-red-600 dark:text-red-400', className)}
      {...props}
    >
      {content}
    </div>
  );
}

export {
  Field,
  FieldLabel,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLegend,
  FieldSeparator,
  FieldSet,
  FieldContent,
  FieldTitle
};
