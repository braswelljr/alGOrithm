import * as React from 'react';
import { matchSorter } from 'match-sorter';
import { HiPlusCircle } from 'react-icons/hi';
import { HiMagnifyingGlass } from 'react-icons/hi2';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Checkbox } from '@/components/ui/checkbox';
import { InputGroup, InputGroupAddon, InputGroupInput } from '@/components/ui/input-group';
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover';
import { Separator } from '@/components/ui/separator';
import { cn } from '@/lib/utils';

export type FacetFilterOption = {
  label: string;
  value: string;
  /**
   * Rows carrying this value, as counted by the server. Read when the caller
   * passes no `counts` map, so an option list built by `facetChoices` - which
   * already carries the count on each option - needs nothing further.
   */
  count?: number;
  icon?: React.ComponentType<{ className?: string }>;
  /**
   * Colour for the icon, where the value carries one by convention: green for
   * accredited, amber for pending, red for refused. Without it every icon
   * renders the same grey, which is the colour of "no meaning attached" and
   * throws away the thing the icon was chosen for.
   */
  iconClassName?: string;
  className?: string;
};

export type FacetFilterProps = {
  title?: string;
  /**
   * Readonly so a caller can declare its options `as const` and keep the
   * literal value types, which is how most of these lists are written.
   */
  options: readonly FacetFilterOption[];
  /** Selected values (controlled). */
  value: string[];
  onValueChange: (next: string[]) => void;
  /** Optional per-value counts shown on the right of each row. */
  counts?: Record<string, number>;
  /** Show a type-to-filter box over the options. On by default, so every facet
   *  can be searched directly; pass `false` for a list too short to need it. */
  searchable?: boolean;
  /** Option count above which the search box appears when `searchable` is unset.
   *  Zero, so it always does. */
  searchThreshold?: number;
  /**
   * Trigger size, passed straight to `Button`, because that is what the trigger
   * is. It used to carry its own heights taken from the `Input` scale, which
   * runs h-7/8.5/10/12/14 where `Button` runs h-6/7/8/9/10: the trigger asked
   * for one size and rendered at another, so a facet never lined up with the
   * buttons beside it. Defaults to `sm` (h-7), the height a table header is
   * built at, so a filter row needs no height overrides at all.
   */
  size?: FacetFilterSize;
  /**
   * One value at a time, for a form's choice (a dataset's shape) rather than a
   * filter: picking an option replaces the choice, it cannot be emptied, and
   * the popover closes on a pick.
   */
  single?: boolean;
  className?: string;
};

/** The button scale, since the trigger is a button. */
export type FacetFilterSize = 'xs' | 'sm' | 'default' | 'lg' | 'xl';

/**
 * Turn the server's facet buckets into the shape `counts` takes.
 *
 * `meta.facets` arrives as `[{ value, count }]` per facet, counted over the
 * same filters the list ran under, so the numbers say how many rows each option
 * would leave rather than how many exist overall. Kept beside the component so
 * every screen converts them the same way.
 */
export function countsFor(
  buckets?: { value: string; count: number }[]
): Record<string, number> | undefined {
  if (!buckets) return undefined;
  return Object.fromEntries(buckets.map((bucket) => [bucket.value, bucket.count]));
}

/**
 * Controlled, standalone faceted filter dropdown.
 *
 * Not bound to a TanStack column: selection is driven purely by
 * `value`/`onValueChange`, which is what makes it correct on a server-driven
 * list. The column-bound filter reads its options out of the table, which only
 * knows the rows currently loaded; this one takes options and counts computed
 * across the whole result set, so on a paginated list the numbers describe
 * everything that matches rather than just the page on screen.
 *
 * The only client-side work here is ranking the option list as you type, over a
 * list already in hand. Results are never reordered here - that would shuffle
 * the current page and say nothing about the rest.
 */
/**
 * How a filter that is doing something reads.
 *
 * Solid `bg-primary`, the same active state the sidebar takes, rather than the
 * tint or the bare badge it had: a bar of seven facets is scanned for which
 * ones are on, and "has a badge on the end" is something a reader has to look
 * for where a filled control is something they see. The house rule is that an
 * active state is a solid colour, and this is one.
 *
 * `text-stone-950` because primary is a mid Go blue: white on it measures
 * about 2.8:1 and fails AA, where near-black measures about 7.4:1. The badges
 * inside go translucent-dark for the same reason - `secondary` is a light grey
 * chip and disappears on gold.
 */
export const FACET_ACTIVE =
  'bg-primary! border-primary! text-stone-950! hover:bg-primary! hover:text-stone-950!';

/** The count chip on an active trigger, which sits on gold rather than on the surface. */
export const FACET_ACTIVE_BADGE = 'bg-stone-950/10! text-stone-950! border-0';

/** A filter trigger at rest: the input's hairline border, so every control in
 *  a filter row, the date filter included, reads as one set. */
export const FACET_TRIGGER =
  'flex items-center border border-stone-300 text-stone-800 dark:border-stone-700 dark:text-stone-200';

export function FacetFilter({
  title,
  options,
  value,
  onValueChange,
  counts,
  searchable,
  searchThreshold = 0,
  size = 'sm',
  single = false,
  className
}: FacetFilterProps) {
  const selected = new Set(value);
  const [term, setTerm] = React.useState('');
  const [open, setOpen] = React.useState(false);

  const showSearch = searchable ?? options.length > searchThreshold;
  const visible = React.useMemo(
    () => (term.trim() ? matchSorter(options, term, { keys: ['label', 'value'] }) : options),
    [options, term]
  );

  function toggle(v: string) {
    if (single) {
      onValueChange([v]);
      setOpen(false);
      return;
    }
    const next = new Set(selected);
    if (next.has(v)) next.delete(v);
    else next.add(v);
    onValueChange(Array.from(next));
  }

  return (
    <Popover
      open={open}
      onOpenChange={setOpen}
    >
      <PopoverTrigger
        render={
          <Button
            variant="ghost"
            size={size}
            data-type-fixed=""
            className={cn(
              // The input's border, not a heavier one of its own.
              //
              // This was `border-stone-600!` over `dark:border-stone-400!`,
              // which mirrors correctly but is three steps darker than anything
              // it stands beside: an `InputGroup` is 300/700 and a
              // `SelectTrigger` is 200/800, so a filter row drew one control in
              // outline and the rest in hairline. The `!` is gone with it -
              // forced, no row could correct it from the outside, and a caller
              // passing a border class was silently ignored.
              FACET_TRIGGER,
              selected.size > 0 && FACET_ACTIVE,
              className
            )}
          >
            <HiPlusCircle className="mr-1 size-4" />
            <span className="line-clamp-1 max-w-52 text-left">{title}</span>
            {selected.size > 0 ? (
              <>
                <Separator
                  orientation="vertical"
                  className="mx-2 h-7 bg-stone-950/20"
                />
                <div className="hidden gap-1 lg:flex">
                  {selected.size > 1 ? (
                    <Badge
                      variant="secondary"
                      case="normal"
                      className={cn('rounded-full px-1.5 text-sm font-normal', FACET_ACTIVE_BADGE)}
                    >
                      {selected.size} selected
                    </Badge>
                  ) : (
                    options
                      .filter((o) => selected.has(o.value))
                      .map((o) => (
                        <Badge
                          key={o.value}
                          variant="secondary"
                          case="normal"
                          className={cn(
                            'rounded-full px-1.5 text-sm font-normal',
                            FACET_ACTIVE_BADGE
                          )}
                        >
                          {o.label}
                        </Badge>
                      ))
                  )}
                </div>
                <Badge
                  variant="secondary"
                  case="normal"
                  className={cn(
                    'rounded-full px-1.5 text-sm font-normal lg:hidden',
                    FACET_ACTIVE_BADGE
                  )}
                >
                  {selected.size}
                </Badge>
              </>
            ) : null}
          </Button>
        }
      />
      <PopoverContent
        align="start"
        className="w-56 p-0!"
      >
        {/* The facet's name over its options: Roboto Serif for the heading,
            Matimo for the rows, as the platform sets titles and controls. */}
        {title ? (
          <h3 className="font-heading px-3 pt-2.5 pb-1 text-sm font-semibold text-stone-950 dark:text-stone-50">
            {title}
          </h3>
        ) : null}
        {showSearch ? (
          <FacetSearch
            title={title}
            term={term}
            onTermChange={setTerm}
          />
        ) : null}
        <div className="max-h-64 overflow-y-auto">
          {visible.length < 1 ? (
            <p className="px-2 py-3 text-center text-sm text-stone-600 dark:text-stone-400">
              No matches.
            </p>
          ) : null}
          {visible.map((o) => (
            <FacetRow
              key={o.value}
              option={o}
              checked={selected.has(o.value)}
              count={counts?.[o.value] ?? o.count}
              onToggle={() => toggle(o.value)}
            />
          ))}
        </div>
        {selected.size > 0 && !single ? (
          <>
            <Separator className="my-1" />
            <Button
              type="button"
              onClick={() => onValueChange([])}
              size="sm"
              className="w-full"
            >
              Clear filters
            </Button>
          </>
        ) : null}
      </PopoverContent>
    </Popover>
  );
}

/**
 * The type-to-filter box, shared by both dropdowns.
 *
 * An `InputGroup` rather than an `Input` with an icon absolutely positioned
 * over it and `pl-7` to clear it. That arrangement is the house rule's worked
 * example of what not to do: the padding is a guess at the icon's width, so the
 * two drift the moment either changes, and the icon needs `pointer-events-none`
 * to stop it swallowing clicks meant for the field. The group lays them out as
 * one control and none of that is needed. The date filter's preset search is
 * built the same way, so the two read as the same control.
 */
function FacetSearch({
  title,
  term,
  onTermChange
}: {
  title?: string;
  term: string;
  onTermChange: (next: string) => void;
}) {
  return (
    <div className="border-b border-stone-200 p-1 dark:border-stone-800">
      <InputGroup className="h-8 border-0 shadow-none">
        <InputGroupAddon>
          <HiMagnifyingGlass
            className="size-3.5 text-stone-600 dark:text-stone-400"
            aria-hidden="true"
          />
        </InputGroupAddon>
        <InputGroupInput
          value={term}
          onChange={(event) => onTermChange(event.target.value)}
          placeholder={`Filter ${title?.toLowerCase() ?? 'options'}...`}
          aria-label={`Filter ${title?.toLowerCase() ?? 'options'}`}
          className="text-xs"
        />
      </InputGroup>
    </div>
  );
}

/**
 * One option row.
 *
 * The icon's default colour is a fallback, not a base to override: merging the
 * two left `dark:text-stone-400` in place, a different variant from the
 * `text-*` it was meant to lose to, so a green icon rendered green in light
 * mode and grey in dark.
 */
function FacetRow({
  option,
  checked,
  count,
  onToggle,
  indent = false,
  heading = false
}: {
  option: FacetFilterOption;
  checked: boolean;
  count?: number;
  onToggle: () => void;
  indent?: boolean;
  /** A group's parent row, set in the heading face. */
  heading?: boolean;
}) {
  return (
    <button
      type="button"
      onClick={onToggle}
      className={cn(
        'text-xsm flex h-7 w-full items-center gap-2 rounded-md px-1 hover:bg-stone-100 dark:hover:bg-stone-900',
        indent && 'pl-6'
      )}
    >
      <Checkbox
        checked={checked}
        className="pointer-events-none"
      />
      {option.icon ? (
        <option.icon
          className={cn('size-4', option.iconClassName ?? 'text-stone-600 dark:text-stone-400')}
        />
      ) : null}
      <span
        className={cn(
          'flex-1 text-left',
          heading && 'font-heading font-semibold',
          option.className
        )}
      >
        {option.label}
      </span>
      {count != null ? (
        <span className="text-sm text-stone-600 dark:text-stone-400">{count}</span>
      ) : null}
    </button>
  );
}

/**
 * A parent facet value with the narrower options that sit beneath it.
 *
 * The parent is a filter term in its own right - it is not only a heading - so
 * it carries a `value` and is sent on its own facet key.
 */
export type FacetFilterGroup = {
  value: string;
  label: string;
  count?: number;
  icon?: React.ComponentType<{ className?: string }>;
  iconClassName?: string;
  /** The narrower terms under this parent, sent on the *options* facet key. */
  options: readonly FacetFilterOption[];
};

/**
 * Turning one parent off, as its effect on both selections.
 *
 * Pure and exported because this is the only real logic in the control and the
 * interesting case is easy to get wrong: a child offered by another *selected*
 * parent has to survive, or taking one kind out of the list would silently
 * narrow a kind the reader never touched. When no parent is left selected every
 * kind is in scope again, so the children stand on their own and are all kept.
 */
export function dropGroupFromSelection({
  groups,
  groupValue,
  optionValue,
  drop
}: {
  groups: readonly FacetFilterGroup[];
  groupValue: string[];
  optionValue: string[];
  drop: string;
}): { groupValue: string[]; optionValue: string[] } {
  const remaining = groupValue.filter((v) => v !== drop);
  if (remaining.length < 1) return { groupValue: remaining, optionValue };

  const stillOffered = new Set(
    groups.filter((g) => remaining.includes(g.value)).flatMap((g) => g.options.map((o) => o.value))
  );
  return { groupValue: remaining, optionValue: optionValue.filter((v) => stillOffered.has(v)) };
}

export type GroupedFacetFilterProps = {
  title?: string;
  groups: readonly FacetFilterGroup[];
  /** Selected parent values, e.g. the accreditation kinds. */
  groupValue: string[];
  onGroupValueChange: (next: string[]) => void;
  /** Selected child values, e.g. the programmes. */
  optionValue: string[];
  onOptionValueChange: (next: string[]) => void;
  /** Counts for parents and children alike, keyed by value. */
  counts?: Record<string, number>;
  size?: FacetFilterSize;
  className?: string;
};

/**
 * A facet over two levels: a parent term, and the narrower terms under it.
 *
 * # Why two value sets rather than one
 *
 * The two levels are different questions, and the API asks them on different
 * keys - `?kind=institutional&programme=lpt`. So the component drives two
 * selections, not one list of mixed values: `groupValue` for the parents and
 * `optionValue` for the children. Flattening them into a single set is how a
 * facet ends up sending a programme where the server expects a kind.
 *
 * # What the parent toggle does
 *
 * Turning a parent *off* turns off everything under it: the parent term is
 * dropped and so are any of its children that no remaining parent offers. That
 * is the behaviour the control is for - one click to take an entire kind of
 * accreditation out of the list, rather than hunting its children individually.
 *
 * Turning a parent *on* adds only the parent. It deliberately does not tick
 * every child, because an empty child selection already means "every one of
 * them": pre-ticking them all would say the same thing in a way the reader then
 * has to undo before they can narrow it.
 *
 * # Children are shared, and that is not a bug
 *
 * A child value means the same thing under either parent - LPT is LPT - so it
 * is one term on one key, drawn under each parent it applies to and ticked in
 * both places when selected. The alternative is a composite `institutional:lpt`
 * term, which the API has no key for.
 */
export function GroupedFacetFilter({
  title,
  groups,
  groupValue,
  onGroupValueChange,
  optionValue,
  onOptionValueChange,
  counts,
  size = 'sm',
  className
}: GroupedFacetFilterProps) {
  const selectedGroups = new Set(groupValue);
  const selectedOptions = new Set(optionValue);
  const total = selectedGroups.size + selectedOptions.size;

  function toggleGroup(group: FacetFilterGroup) {
    if (selectedGroups.has(group.value)) {
      const next = dropGroupFromSelection({
        groups,
        groupValue,
        optionValue,
        drop: group.value
      });
      onGroupValueChange(next.groupValue);
      if (next.optionValue.length !== optionValue.length) onOptionValueChange(next.optionValue);
      return;
    }
    onGroupValueChange([...groupValue, group.value]);
  }

  function toggleOption(value: string) {
    const next = new Set(selectedOptions);
    if (next.has(value)) next.delete(value);
    else next.add(value);
    onOptionValueChange(Array.from(next));
  }

  return (
    <Popover>
      <PopoverTrigger
        render={
          <Button
            variant="ghost"
            size={size}
            className={cn(
              // The same border as the single-select trigger above, for the same
              // reason: a filter row is one row of controls and this is one of
              // them.
              'flex items-center border border-stone-300 text-stone-800 dark:border-stone-700 dark:text-stone-200',
              total > 0 && FACET_ACTIVE,
              className
            )}
          >
            <HiPlusCircle className="mr-1 size-4" />
            <span className="line-clamp-1 max-w-52 text-left">{title}</span>
            {total > 0 ? (
              <>
                <Separator
                  orientation="vertical"
                  className="mx-2 h-7 bg-stone-950/20"
                />
                <Badge
                  variant="secondary"
                  case="normal"
                  className={cn('rounded-full px-1.5 text-sm font-normal', FACET_ACTIVE_BADGE)}
                >
                  {total} selected
                </Badge>
              </>
            ) : null}
          </Button>
        }
      />
      <PopoverContent
        align="start"
        className="w-64 p-0!"
      >
        {/* The facet's name over its options: Roboto Serif for the heading,
            Matimo for the rows, as the platform sets titles and controls. */}
        {title ? (
          <h3 className="font-heading px-3 pt-2.5 pb-1 text-sm font-semibold text-stone-950 dark:text-stone-50">
            {title}
          </h3>
        ) : null}
        <div className="max-h-72 overflow-y-auto py-1">
          {groups.map((group) => (
            <div key={group.value}>
              <FacetRow
                heading
                option={group}
                checked={selectedGroups.has(group.value)}
                count={counts?.[group.value] ?? group.count}
                onToggle={() => toggleGroup(group)}
              />
              {group.options.map((option) => (
                <FacetRow
                  // Keyed by both, because the same child is drawn under every
                  // parent that offers it.
                  key={`${group.value}:${option.value}`}
                  option={option}
                  checked={selectedOptions.has(option.value)}
                  count={counts?.[option.value] ?? option.count}
                  onToggle={() => toggleOption(option.value)}
                  indent
                />
              ))}
            </div>
          ))}
        </div>
        {total > 0 ? (
          <>
            <Separator className="my-1" />
            <Button
              type="button"
              onClick={() => {
                onGroupValueChange([]);
                onOptionValueChange([]);
              }}
              size="sm"
              className="w-full"
            >
              Clear filters
            </Button>
          </>
        ) : null}
      </PopoverContent>
    </Popover>
  );
}
