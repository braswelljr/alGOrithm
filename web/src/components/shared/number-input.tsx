'use client';

/**
 * A number typed or stepped: the ui NumberField with its minus and plus
 * buttons, in the shape the site's forms use. Every number input on the site
 * is this, so they all step, clamp and read the same way.
 *
 *   <NumberInput id={id} value={size} onValueChange={setSize} min={1} max={500} />
 *
 * The value is `null` while the field is empty, so "nothing typed" is told
 * apart from 0. A field that should act only once the reader is done (a page
 * number) passes `defaultValue` and `onValueCommitted`, which runs on blur,
 * Enter and the buttons.
 */
import {
  NumberField,
  NumberFieldDecrement,
  NumberFieldGroup,
  NumberFieldIncrement,
  NumberFieldInput
} from '@/components/ui/number-field';
import { cn } from '@/lib/utils';

export function NumberInput({
  id,
  value,
  defaultValue,
  onValueChange,
  onValueCommitted,
  min,
  max,
  step = 1,
  size = 'md',
  disabled,
  required,
  className,
  groupClassName,
  'aria-label': ariaLabel,
  'aria-invalid': ariaInvalid
}: {
  id?: string;
  value?: number | null;
  defaultValue?: number;
  onValueChange?: (value: number | null) => void;
  onValueCommitted?: (value: number | null) => void;
  min?: number;
  max?: number;
  step?: number;
  size?: 'xs' | 'sm' | 'md' | 'lg';
  disabled?: boolean;
  required?: boolean;
  className?: string;
  groupClassName?: string;
  'aria-label'?: string;
  'aria-invalid'?: boolean;
}) {
  return (
    <NumberField
      id={id}
      value={value}
      defaultValue={defaultValue}
      onValueChange={onValueChange ? (v) => onValueChange(v) : undefined}
      onValueCommitted={onValueCommitted ? (v) => onValueCommitted(v) : undefined}
      min={min}
      max={max}
      step={step}
      size={size}
      disabled={disabled}
      required={required}
      className={cn('w-full', className)}
    >
      <NumberFieldGroup className={groupClassName}>
        <NumberFieldDecrement aria-label="Less" />
        <NumberFieldInput
          aria-label={ariaLabel}
          aria-invalid={ariaInvalid || undefined}
        />
        <NumberFieldIncrement aria-label="More" />
      </NumberFieldGroup>
    </NumberField>
  );
}
