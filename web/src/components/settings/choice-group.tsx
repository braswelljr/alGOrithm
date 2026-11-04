'use client';

import type { ReactNode } from 'react';
import { Label } from '@/components/ui/label';
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group';

/**
 * One preference as a set of choices, each with the sentence that says what it
 * does and, where one is given, a picture of what it looks like. A fieldset, so
 * a screen reader announces the question before the answers; the legend is read
 * rather than shown, under a section that titles it.
 */
export function ChoiceGroup<T extends string>({
  legend,
  options,
  value,
  onChange,
  preview
}: {
  legend: string;
  options: { value: T; label: string; hint: string }[];
  value: T;
  onChange: (value: T) => void;
  /** Draws each option, so it can be chosen by looking at it. */
  preview?: (value: T, selected: boolean) => ReactNode;
}) {
  return (
    <fieldset className="flex flex-col gap-2">
      {/* Named for a screen reader; the section's own title already names it
          on screen, and saying it twice reads as two questions. */}
      <legend className="sr-only">{legend}</legend>
      <RadioGroup
        value={value}
        onValueChange={(next) => {
          const chosen = options.find((option) => option.value === next);
          if (chosen) onChange(chosen.value);
        }}
        // As many as fit at 275px or more each, each growing to share the row,
        // so a card is never a fixed size. The min() lets one card take the
        // whole width on a screen narrower than that rather than overflow it.
        className="grid grid-cols-[repeat(auto-fill,minmax(min(275px,100%),1fr))] gap-3"
      >
        {options.map((option) => (
          <Label
            key={option.value}
            className="has-data-checked:border-primary flex cursor-pointer flex-col items-stretch gap-3 rounded-lg border border-stone-200 p-3 transition-colors hover:bg-stone-50 dark:border-stone-800 dark:hover:bg-stone-950"
          >
            {preview ? (
              // A picture, so it is hidden from a screen reader: the label and
              // the sentence under it already say what it shows.
              <div
                aria-hidden="true"
                className="relative aspect-16/10 w-full overflow-hidden rounded-md border border-stone-200 bg-white dark:border-stone-800 dark:bg-stone-950"
              >
                {preview(option.value, option.value === value)}
              </div>
            ) : null}
            <span className="flex items-start gap-3">
              <RadioGroupItem
                value={option.value}
                className="mt-0.5"
              />
              <span className="flex flex-col gap-0.5">
                <span className="text-sm font-medium text-stone-900 dark:text-stone-100">
                  {option.label}
                </span>
                <span className="text-sm text-stone-600 dark:text-stone-400">{option.hint}</span>
              </span>
            </span>
          </Label>
        ))}
      </RadioGroup>
    </fieldset>
  );
}
