import type * as React from 'react';
import {
  FieldDescription,
  FieldLabel,
  Field as UIField
} from '@braswelljr/algorithm-ui/components/field';
import { Input as UIInput } from '@braswelljr/algorithm-ui/components/input';
import { Label as UILabel } from '@braswelljr/algorithm-ui/components/label';
import { Textarea as UITextarea } from '@braswelljr/algorithm-ui/components/textarea';
import { cn } from '@/utils/utils';

/** A text field: the library's Input. */
export function Input(props: React.ComponentProps<typeof UIInput>) {
  return <UIInput {...props} />;
}

/** A multi-line field for data, in mono: the library's Textarea. */
export function Textarea({ className, ...props }: React.ComponentProps<typeof UITextarea>) {
  return (
    <UITextarea
      className={cn('min-h-24 font-mono', className)}
      {...props}
    />
  );
}

/** A label: the library's Label. */
export function Label(props: React.ComponentProps<typeof UILabel>) {
  return <UILabel {...props} />;
}

/** A label above its control, with an optional hint below: the library's Field. */
export function Field({
  label,
  htmlFor,
  hint,
  children,
  className
}: {
  label: string;
  htmlFor: string;
  hint?: React.ReactNode;
  children: React.ReactNode;
  className?: string;
}) {
  return (
    <UIField className={className}>
      <FieldLabel htmlFor={htmlFor}>{label}</FieldLabel>
      {children}
      {hint ? <FieldDescription>{hint}</FieldDescription> : null}
    </UIField>
  );
}
