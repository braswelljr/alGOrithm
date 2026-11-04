'use client';

import * as React from 'react';
import { useIsomorphicLayoutEffect } from '@braswelljr/algorithm-ui/hooks/use-isomorphic-layout-effect';
import { cn } from '@braswelljr/algorithm-ui/lib/utils';
import { cva, VariantProps } from 'class-variance-authority';

const inputVariants = cva(
  `flex w-full border border-border bg-card text-fg shadow-xs shadow-black/5 transition-[color,box-shadow] file:h-full file:border-0 file:border-e file:border-solid file:border-border file:bg-transparent file:p-0 file:font-medium file:text-fg file:not-italic placeholder:text-muted/80 focus-visible:border-control focus-visible:ring-3 focus-visible:ring-fg/10 focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-60 aria-invalid:border-red-600 aria-invalid:ring-red-600/20 dark:aria-invalid:border-red-400 dark:aria-invalid:ring-red-400/25 [&::-webkit-calendar-picker-indicator]:ms-auto [&::-webkit-calendar-picker-indicator]:cursor-pointer [[readonly]]:bg-subtle [[type=file]]:py-0`,
  {
    variants: {
      variant: {
        lg: 'h-11 px-4 text-base file:me-4 file:pe-4',
        // 16 on a phone, so the browser does not zoom the page into the field.
        md: 'h-10 px-3 text-base sm:text-sm',
        default: 'h-9 px-3 text-base sm:text-sm',
        sm: 'h-8 px-2.5 text-sm leading-(--text-sm--line-height) file:me-3 file:pe-3',
        xs: 'h-7 px-2 text-sm file:me-2.5 file:pe-2.5'
      },
      radius: {
        none: 'rounded-none',
        sm: 'rounded-sm',
        default: 'rounded-md',
        lg: 'rounded-lg',
        full: 'rounded-full'
      }
    },
    defaultVariants: {
      // The system's control height is 40 (`md`). `default` stays 36 for the
      // places that ask for it by name.
      variant: 'md',
      radius: 'default'
    }
  }
);

/**
 * What each character of a mask accepts.
 *
 * `9` a digit, `a` a letter, `*` either. Anything else in the mask is a literal
 * and is typed in for the reader: the dashes in a Ghana Post address, the slash
 * in a date.
 */
const MASK_TOKENS: Record<string, RegExp> = {
  '9': /\d/,
  a: /[A-Za-z]/,
  '*': /[A-Za-z0-9]/
};

/**
 * How a mask treats the case of what was typed.
 *
 * Left alone by default, because case is the field's business rather than the
 * format's: a mask over a name must not shout it. Set where the value has a
 * conventional case of its own - a Ghana Post address and a GTEC file number are
 * both written in capitals, and a reader who types one in lower case has not
 * made a mistake worth correcting them over.
 */
export type MaskCase = 'upper' | 'lower';

/**
 * The literal run a mask opens with, punctuation and all: `CLET-` of
 * `CLET-****-****-****`, and nothing at all for `aa-999-9999`.
 */
function maskHead(mask: string): string {
  return mask.match(/^[^9a*]+/)?.[0] ?? '';
}

/**
 * `raw` reduced to the characters the mask's slots hold.
 *
 * Punctuation goes, and so does a leading literal prefix the value already
 * carries, so `CLET-Z4K8`, `cletz4k8` and `Z4K8` all describe the same body.
 * Everything the caret arithmetic and the deletion rules count is counted here,
 * which is the only way the two can agree.
 */
export function maskBody(raw: string, mask: string): string {
  const clean = raw.replace(/[^A-Za-z0-9]/g, '');
  const prefix = maskHead(mask).replace(/[^A-Za-z0-9]/g, '');
  if (prefix && clean.toUpperCase().startsWith(prefix.toUpperCase()))
    return clean.slice(prefix.length);
  return clean;
}

/** `body` laid into `mask`'s slots, with the literals written between them. */
function fillMask(body: string, mask: string): string {
  let index = 0;
  let out = '';
  // Literals wait for the slot after them to be filled before they are written.
  // A dash held up by nothing strands the caret past punctuation the reader has
  // not earned yet: typing one character into an empty box produced "GA-", and
  // a last keystroke the mask had no slot for produced "GA-123-".
  let held = '';

  for (const maskChar of mask) {
    const token = MASK_TOKENS[maskChar];
    if (!token) {
      held += maskChar;
      continue;
    }
    // Skip anything the slot will not take: a letter offered to a digit slot is
    // dropped rather than stopping the whole value.
    while (index < body.length && !token.test(body[index])) index += 1;
    if (index >= body.length) break;
    out += held + body[index];
    held = '';
    index += 1;
  }
  return out;
}

/**
 * `raw` rewritten to fit `mask`.
 *
 * Everything that is not a letter or a digit is dropped first, so a reader may
 * paste `GA 123 4567` or `ga-123-4567` and get the same answer, and the literals
 * come from the mask rather than from what was typed.
 *
 * Stops as soon as the input runs out, so a half-typed value stays half-typed
 * rather than being padded with the rest of the mask's punctuation.
 */
export function applyMask(raw: string, mask: string, casing?: MaskCase): string {
  const cased =
    casing === 'upper' ? raw.toUpperCase() : casing === 'lower' ? raw.toLowerCase() : raw;
  const clean = cased.replace(/[^A-Za-z0-9]/g, '');
  const prefix = maskHead(mask).replace(/[^A-Za-z0-9]/g, '');

  // While the reader is still typing a mask's own prefix the value is returned
  // as typed rather than fed to the slots, so `C`, `CL`, `CLE`, `CLET` read as
  // the letters they typed and not as `CLET-C`.
  if (prefix && prefix.toUpperCase().startsWith(clean.toUpperCase())) return clean;

  return fillMask(maskBody(cased, mask), mask);
}

/**
 * Where the caret belongs once `count` body characters sit behind it.
 *
 * Counting bodies rather than positions is the whole point: the typed string and
 * the masked one do not line up (that is what the mask is for), so a caret
 * measured against one and applied to the other lands anywhere. Measuring it in
 * characters the reader actually typed survives every literal the field writes
 * for them.
 */
function caretAfter(masked: string, count: number, mask: string): number {
  // Nothing typed yet, so the caret belongs at the first slot rather than in
  // front of a prefix the reader cannot edit.
  if (count < 1) return Math.min(maskHead(mask).length, masked.length);

  let seen = 0;
  for (let i = 0; i < masked.length; i += 1) {
    const maskChar = mask[i];
    if (maskChar && MASK_TOKENS[maskChar]) {
      seen += 1;
      if (seen === count) return i + 1;
    }
  }
  return masked.length;
}

/** One keystroke, as it reaches the field. */
type MaskEdit = {
  /** The raw value the browser left in the box. */
  typed: string;
  /** Where the browser left the caret in `typed`. */
  caret: number;
  /** The masked value the field showed before the keystroke. */
  before: string;
  mask: string;
  casing?: MaskCase;
  /**
   * The `inputType` of the event. It answers the one question the strings
   * cannot: backspace and delete both shorten a value by one, and they remove
   * opposite characters.
   */
  how?: string;
};

/**
 * What a keystroke leaves in a masked field, and where the caret goes.
 *
 * Pure, and separate from the component, because this is where the field either
 * behaves or does not: every one of its rules is a rule about an edit rather
 * than about a value, and none of them can be checked by masking a string.
 */
export function maskEdit({ typed, caret, before, mask, casing, how = '' }: MaskEdit): {
  value: string;
  caret: number;
} {
  const removing = how.startsWith('delete') || (!how && typed.length < before.length);
  const forwards = how === 'deleteContentForward';

  let next = applyMask(typed, mask, casing);
  // The caret, counted in characters the reader typed rather than in positions
  // of a string the mask is about to rewrite.
  let count = maskBody(typed.slice(0, caret), mask).length;

  const head = maskHead(mask);
  const wrote = head.length > 0 && before.startsWith(head);

  // A keystroke that nudged the value by one character, rather than one that
  // replaced a selection or pasted over it. Only the first kind can land inside
  // the prefix by accident, and only the first kind is refused below: selecting
  // the field and typing over it has to keep working.
  const nudged = removing
    ? typed.length === before.length - 1
    : how === 'insertText' && typed.length === before.length + 1;

  // Editing inside a mask's own prefix.
  //
  // Backspacing over the `-` of `CLET-Z4K8` leaves `CLET` reading as body rather
  // than as the prefix, and the next mask turns `CLEZ4K8` into `CLET-CLEZ-4K8`.
  // The prefix is the field's, not the reader's, so an edit inside it is refused
  // and the caret is parked at the first slot.
  const inHead = wrote && nudged && (removing ? caret < head.length : caret <= head.length);

  if (inHead) {
    return { value: before, caret: head.length };
  }

  // Deleting a literal.
  //
  // Removing the dash in "GA-123" leaves "GA123", which masks straight back to
  // "GA-123": the reader presses backspace and nothing happens, twice, before
  // they give up. So when a deletion produced no change, take the value
  // character the key was aimed at instead, which is what they meant.
  if (removing && next === before) {
    const body = maskBody(typed, mask);
    const at = maskBody(before.slice(0, caret), mask).length;
    const cut = forwards ? at : at - 1;
    if (cut > -1 && cut < body.length) {
      next = applyMask(body.slice(0, cut) + body.slice(cut + 1), mask, casing);
      count = cut;
    }
  }

  // Deleting the last character the reader typed takes the prefix with it. A
  // field left showing `CLET` alone reads as a value, and the next backspace
  // would start eating letters the reader never typed.
  if (removing && wrote && maskBody(next, mask).length < 1) {
    return { value: '', caret: 0 };
  }

  return { value: next, caret: caretAfter(next, count, mask) };
}

type InputProps = React.ComponentProps<'input'> &
  VariantProps<typeof inputVariants> & {
    /**
     * A format the value is held to as it is typed.
     *
     * `9` a digit, `a` a letter, `*` either; anything else is punctuation the
     * field writes for the reader. A Ghana Post digital address is `aa-999-9999`.
     *
     * Absent, the input is exactly what it was, so no existing field changes
     * behaviour by this being here.
     *
     * # Why it lives on `Input` rather than in a `MaskedInput`
     *
     * A separate component is a second input to keep in step with this one: its
     * variants, its focus ring, its invalid state, its file styling. They drift,
     * and the drift shows up as one field in a form looking unlike the field
     * beside it.
     *
     * # It stays transparent to react-hook-form
     *
     * The masked value is written back onto the event before `onChange` runs, so
     * `{...field}` from a `Controller` stores the formatted string with no
     * adapter. A component that reported through its own `onValueChange` would
     * need every call site to know it is special.
     */
    mask?: string;
    /** Forces the case of a masked value. Ignored without a `mask`. */
    maskCase?: MaskCase;
  };

/** Hands the element to a caller's ref, a callback or an object. Kept outside
 *  the component so writing to the caller's ref is not a render-time mutation. */
function assignRef<T>(ref: React.Ref<T> | undefined, value: T | null) {
  if (typeof ref === 'function') ref(value);
  else if (ref) (ref as React.RefObject<T | null>).current = value;
}

function Input({ className, type, variant, radius, mask, maskCase, ...props }: InputProps) {
  const inner = React.useRef<HTMLInputElement | null>(null);
  // The value the caret was measured against, and where it goes once that value
  // is on screen. Held together so a render caused by something else cannot drag
  // the caret back to where an earlier keystroke left it.
  const pending = React.useRef<{ value: string; caret: number } | null>(null);
  const { ref: forwardedRef, onChange, value, defaultValue, ...rest } = props;

  // The masked value is what the field shows, so a serial arriving unpunctuated
  // from a URL reads the same as one typed in. Masking is idempotent, so a value
  // that is already formatted passes through untouched.
  const shown =
    mask && value !== undefined && value !== null
      ? applyMask(String(value), mask, maskCase)
      : value;
  const shownDefault =
    mask && defaultValue !== undefined && defaultValue !== null
      ? applyMask(String(defaultValue), mask, maskCase)
      : defaultValue;

  const previous = React.useRef(String(shownDefault ?? ''));

  const setRef = React.useCallback(
    (element: HTMLInputElement | null) => {
      inner.current = element;
      assignRef(forwardedRef, element);
    },
    [forwardedRef]
  );

  // Restoring the caret is the whole reason this is not four lines.
  //
  // Rewriting the value on every keystroke sends the caret to the end, so
  // correcting the third character of a GTEC number moved the cursor past the
  // last one and the next keystroke landed in the wrong place. Layout effect
  // rather than an effect, so it happens before the browser paints and the caret
  // is never seen in the wrong spot.
  useIsomorphicLayoutEffect(() => {
    const element = inner.current;
    const target = pending.current;
    if (!mask || !element || !target) return;
    // Only once the field is showing the value the caret was measured against.
    if (element.value !== target.value) return;
    // A type that has no selection to set: `email` and `number` throw rather
    // than ignore it.
    if (element.selectionStart === null) return;
    pending.current = null;
    element.setSelectionRange(target.caret, target.caret);
  });

  const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    if (!mask) {
      onChange?.(event);
      return;
    }

    const element = event.target;
    const { value: next, caret } = maskEdit({
      typed: element.value,
      caret: element.selectionStart ?? element.value.length,
      // What the field showed before this keystroke. `element.defaultValue` is
      // not it: on an uncontrolled field that is the value it started with, so
      // every edit after the first was measured against the wrong string.
      before: value !== undefined && value !== null ? String(shown ?? '') : previous.current,
      mask,
      casing: maskCase,
      how: (event.nativeEvent as Partial<InputEvent>).inputType ?? ''
    });

    // Written back onto the event, so a `Controller` spreading `{...field}`
    // stores the masked string without knowing the mask exists.
    element.value = next;
    previous.current = next;
    // Placed now as well as in the layout effect: when the masked value matches
    // what the field already held there is no render to restore it in.
    if (element.selectionStart !== null) element.setSelectionRange(caret, caret);
    pending.current = { value: next, caret };
    // And forgotten at the end of this keystroke, so a later render cannot
    // reapply a caret the reader has since moved.
    queueMicrotask(() => {
      pending.current = null;
    });

    onChange?.(event);
  };

  return (
    <input
      ref={setRef}
      type={type}
      data-slot="input"
      value={shown}
      defaultValue={shownDefault}
      onChange={handleChange}
      className={cn(inputVariants({ variant, radius }), className)}
      {...rest}
    />
  );
}

export { Input, type InputProps };
