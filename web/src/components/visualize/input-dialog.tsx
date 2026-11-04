'use client';

/**
 * The input of a run, set in a dialog so its form never sits among the views,
 * the way gsl-special-courses' form dialogs are: a title and a line, the form,
 * then Cancel and the action. A cog beside a line saying what the run works on
 * opens it.
 *
 * The form is react-hook-form: it starts from the run's current input each
 * time the dialog opens, holds every change as a draft, and only Apply hands
 * the values back to be written to the link. Cancel, Escape or the backdrop
 * leave the run as it was.
 */
import { useEffect, useState, type ReactNode } from 'react';
import { useForm, type DefaultValues, type FieldValues, type UseFormReturn } from 'react-hook-form';
import { LuSettings } from 'react-icons/lu';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogBody,
  DialogClose,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogPopup,
  DialogTitle
} from '@/components/ui/dialog';
import { cn } from '@/lib/utils';
import type { Entry, Input } from '@/visualizer/core';

/** A short line for what an input holds: "24 values · target 9 · quantum 2". */
export function describeInput(input: Input | null | undefined): string {
  if (!input) return 'No data yet';
  const parts: string[] = [];
  if (input.text) {
    const lines = input.text.split('\n').filter((l) => l.trim());
    const first = lines[0]?.trim() ?? '';
    parts.push(
      lines.length > 1
        ? `${lines.length} lines`
        : `“${first.length > 40 ? `${first.slice(0, 40)}...` : first}”`
    );
  }
  if (input.other) parts.push(`“${input.other}”`);
  if (input.values?.length) parts.push(`${input.values.length} values`);
  if (input.target !== undefined) parts.push(`target ${input.target}`);
  if (input.key) parts.push(input.key);
  if (input.shift !== undefined) parts.push(`shift ${input.shift}`);
  for (const [name, value] of Object.entries(input.options ?? {})) parts.push(`${name} ${value}`);
  return parts.length ? parts.join(' · ') : 'The example';
}

export function InputDialog<T extends FieldValues>({
  entry,
  summary,
  wide = false,
  values,
  onApply,
  icon,
  lead = 'Running on',
  action,
  title,
  description,
  tour = 'configure',
  children
}: {
  entry: Entry;
  /** The bar's `data-tour` name: `configure` unless a page has two. */
  tour?: string;
  /** The button's icon; a cog unless said otherwise. */
  icon?: ReactNode;
  /** The words before the summary. */
  lead?: string;
  /** The button's accessible name. */
  action?: string;
  /** The dialog's title and line. */
  title?: string;
  description?: string;
  /** What the run works on now, in a few words. */
  summary: ReactNode;
  /** A wider dialog, for a form with a picture in it (a grid). */
  wide?: boolean;
  /** The run's current input: the form starts from it on every opening. */
  values: T;
  /** Apply was pressed: write these to the link. */
  onApply: (values: T) => void;
  /** The form's fields, given the form. */
  children: (form: UseFormReturn<T>) => ReactNode;
}) {
  const [open, setOpen] = useState(false);
  const form = useForm<T>({ defaultValues: values as DefaultValues<T> });
  const { reset } = form;
  // A fresh draft each time the dialog opens, from what the run uses now.
  useEffect(() => {
    if (open) reset(values);
    // Only on opening: the run's input does not change while the dialog is up.
  }, [open, reset]);

  return (
    <>
      <div
        data-tour={tour}
        className="bg-subtle flex w-full items-center justify-between gap-3 rounded-xl py-2 ps-4 pe-2"
      >
        <p className="min-w-0 truncate text-sm">
          <span className="text-muted">{lead} </span>
          <span className="font-medium">{summary}</span>
        </p>
        <Button
          variant="ghost"
          size="icon-sm"
          className="bg-bg hover:bg-bg/70 shrink-0"
          onClick={() => setOpen(true)}
          aria-label={action ?? `Configure ${entry.name}: its data and settings`}
          title={action ?? 'Configure the data and settings'}
        >
          {icon ?? <LuSettings aria-hidden />}
        </Button>
      </div>
      <Dialog
        open={open}
        onOpenChange={setOpen}
      >
        <DialogPopup className={cn('sm:max-w-3xl', wide && 'sm:max-w-5xl')}>
          <DialogHeader>
            <DialogTitle>{title ?? `Configure ${entry.name}`}</DialogTitle>
            <DialogDescription>
              {description ??
                'The data and the settings the run starts from. Nothing changes until you apply.'}
            </DialogDescription>
          </DialogHeader>
          {/* Not a <form>: the data tabs inside carry forms of their own, and
              forms cannot nest. Apply submits through react-hook-form instead. */}
          <DialogBody>{open ? children(form) : null}</DialogBody>
          <DialogFooter>
            <DialogClose
              render={
                <Button
                  variant="ghost"
                  type="button"
                />
              }
            >
              Cancel
            </DialogClose>
            <Button
              variant="primary"
              type="button"
              onClick={() =>
                void form.handleSubmit((v) => {
                  onApply(v);
                  setOpen(false);
                })()
              }
            >
              Apply
            </Button>
          </DialogFooter>
        </DialogPopup>
      </Dialog>
    </>
  );
}
