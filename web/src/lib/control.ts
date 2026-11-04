/**
 * The one size scale for every form control, so an Input, a Select trigger, a
 * Combobox, an Autocomplete, a NumberField, a NativeSelect, an InputGroup, a
 * Toggle and a Button of the same size line up to the pixel in a row:
 *
 *   xs 28px   sm 32px   default 36px   md 40px   lg 44px
 *
 * Every component reads these records instead of writing its own heights. The
 * strings are literal on purpose: Tailwind finds classes by scanning source, so
 * a class built at runtime (`h-${n}`) would never be generated.
 */

export type ControlSize = 'xs' | 'sm' | 'default' | 'md' | 'lg';

/** Height, inline padding and text size of a control's box, by size. */
export const controlSize: Record<ControlSize, string> = {
  xs: 'h-7 px-2 text-sm',
  sm: 'h-8 px-2.5 text-sm',
  default: 'h-9 px-3 text-base sm:text-sm',
  md: 'h-10 px-3 text-base sm:text-sm',
  lg: 'h-11 px-4 text-base'
};

/** The same heights without padding, for wrappers whose children pad
 *  themselves (InputGroup, NumberField's group). */
export const controlHeight: Record<ControlSize, string> = {
  xs: 'h-7',
  sm: 'h-8',
  default: 'h-9',
  md: 'h-10',
  lg: 'h-11'
};

/** A square of the control's height, for icon-only buttons and toggles. */
export const controlSquare: Record<ControlSize, string> = {
  xs: 'size-7',
  sm: 'size-8',
  default: 'size-9',
  md: 'size-10',
  lg: 'size-11'
};

/**
 * The same heights keyed by a `data-size` attribute, for components that set
 * their size on a part they don't render themselves (Select's trigger, the
 * NumberField's input). Literal, like the records above.
 */
export const controlSizeByData =
  'data-[size=xs]:h-7 data-[size=xs]:px-2 data-[size=xs]:text-sm ' +
  'data-[size=sm]:h-8 data-[size=sm]:px-2.5 data-[size=sm]:text-sm ' +
  'data-[size=default]:h-9 data-[size=default]:px-3 data-[size=default]:text-base sm:data-[size=default]:text-sm ' +
  'data-[size=md]:h-10 data-[size=md]:px-3 data-[size=md]:text-base sm:data-[size=md]:text-sm ' +
  'data-[size=lg]:h-11 data-[size=lg]:px-4 data-[size=lg]:text-base';

/** One row in any popup list (Select, Combobox, Autocomplete, Command,
 *  DropdownMenu), so every list reads the same. 32px on touch, 28px with a
 *  pointer, plus a 44px hit area on coarse pointers. */
export const controlItem =
  'min-h-8 px-2 py-1 text-base sm:min-h-7 sm:text-sm pointer-coarse:min-h-11';
