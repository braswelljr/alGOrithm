'use client';

import * as React from 'react';
import { Select as SelectPrimitive } from '@base-ui/react/select';
import { cn } from '@braswelljr/algorithm-ui/lib/utils';
import { CheckIcon, ChevronDownIcon, ChevronUpIcon } from 'lucide-react';

/**
 * The select, showing the chosen option's label rather than its value.
 *
 * base-ui's `Select.Value` prints the raw value unless the root is handed the
 * items, so without them a reader who picked "Bank transfer" saw
 * `bank_transfer`, and one who picked a person saw their id. Almost no caller
 * passed `items`, and asking each of them to was how it came to be missing.
 *
 * So the root reads them off the `SelectItem`s written beneath it: the value
 * and the children each one already declares. They are elements in the tree
 * whether or not the popup is open, which is what lets the label show on a
 * value chosen before this render. A caller that renders its items through a
 * component of its own, where the walk cannot see them, passes `items` itself,
 * and an explicit `items` always wins.
 */
function Select<Value, Multiple extends boolean | undefined = false>({
  items,
  children,
  ...props
}: SelectPrimitive.Root.Props<Value, Multiple>) {
  const derived = React.useMemo(() => {
    if (items) return items;
    const found: Record<string, React.ReactNode> = {};
    collectItemLabels(children, found);
    return Object.keys(found).length > 0 ? found : undefined;
  }, [items, children]);
  return (
    <SelectPrimitive.Root
      items={derived}
      {...props}
    >
      {children}
    </SelectPrimitive.Root>
  );
}

/** Every `SelectItem` under a node, by value, with what it renders as its label. */
function collectItemLabels(node: React.ReactNode, found: Record<string, React.ReactNode>) {
  React.Children.forEach(node, (child) => {
    if (!React.isValidElement(child)) return;
    const props = child.props as { value?: unknown; children?: React.ReactNode };
    if (child.type === SelectItem && props.value !== undefined && props.value !== null) {
      found[String(props.value)] = props.children;
      return;
    }
    if (props.children !== undefined) collectItemLabels(props.children, found);
  });
}

function SelectGroup({ className, ...props }: SelectPrimitive.Group.Props) {
  return (
    <SelectPrimitive.Group
      data-slot="select-group"
      className={cn('scroll-my-1 p-1', className)}
      {...props}
    />
  );
}

function SelectValue({ className, ...props }: SelectPrimitive.Value.Props) {
  return (
    <SelectPrimitive.Value
      data-slot="select-value"
      className={cn(
        // `min-w-0`, or a long selected value - a name beside an email address,
        // an institution's legal name - widens the trigger past its column
        // instead of being cut by it.
        'flex min-w-0 flex-1 text-left',
        className
      )}
      {...props}
    />
  );
}

function SelectTrigger({
  className,
  size = 'md',
  // No default: with the attribute absent none of the `data-[radius=…]` rules
  // below match and the base pill stands, which is what every trigger wants.
  // The prop was declared in the type and then neither destructured nor
  // rendered, so all five of those rules matched nothing and `radius="sm"`
  // was accepted by the compiler and did nothing at all.
  radius,
  children,
  ...props
}: SelectPrimitive.Trigger.Props & {
  size?: 'xs' | 'sm' | 'default' | 'md' | 'lg';
  radius?: 'none' | 'sm' | 'default' | 'lg' | 'full';
}) {
  return (
    <SelectPrimitive.Trigger
      data-slot="select-trigger"
      data-size={size}
      data-radius={radius}
      className={cn(
        'flex w-fit items-center justify-between gap-1.5 rounded-lg border py-2 pr-2 pl-2.5 text-sm whitespace-nowrap shadow-xs shadow-black/5 transition-[color,background-color,border-color,box-shadow] outline-none select-none',
        // Zinc Palette for Trigger
        'border-border bg-card text-fg hover:bg-bg',
        // States
        'focus-visible:border-control focus-visible:ring-3 focus-visible:ring-fg/10',
        'disabled:cursor-not-allowed disabled:opacity-50 aria-invalid:border-red-600 aria-invalid:ring-red-600/20 dark:aria-invalid:border-red-400 dark:aria-invalid:ring-red-400/25',
        'data-placeholder:text-muted',
        // Sizing
        'data-[size=xs]:h-7 data-[size=xs]:px-2 data-[size=xs]:text-sm',
        'data-[size=sm]:h-8 data-[size=sm]:px-2.5 data-[size=sm]:text-sm data-[size=sm]:leading-(--text-sm--line-height)',
        'data-[size=default]:h-9 data-[size=default]:px-3 data-[size=default]:text-sm',
        'data-[size=md]:h-10 data-[size=md]:px-3 data-[size=md]:text-sm',
        'data-[size=lg]:h-11 data-[size=lg]:px-4 data-[size=lg]:text-base',
        // Radius
        'data-[radius=none]:rounded-none',
        'data-[radius=sm]:rounded-sm',
        'data-[radius=default]:rounded-md',
        'data-[radius=lg]:rounded-lg',
        'data-[radius=full]:rounded-full',
        // Content alignment
        '*:data-[slot=select-value]:line-clamp-1 *:data-[slot=select-value]:flex *:data-[slot=select-value]:items-center *:data-[slot=select-value]:gap-1.5',
        "[&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-4",

        className
      )}
      {...props}
    >
      {children}
      <SelectPrimitive.Icon
        render={<ChevronDownIcon className="pointer-events-none size-4 text-muted" />}
      />
    </SelectPrimitive.Trigger>
  );
}

function SelectContent({
  className,
  children,
  side = 'bottom',
  sideOffset = 4,
  align = 'center',
  alignOffset = 0,
  alignItemWithTrigger = true,
  ...props
}: SelectPrimitive.Popup.Props &
  Pick<
    SelectPrimitive.Positioner.Props,
    'align' | 'alignOffset' | 'side' | 'sideOffset' | 'alignItemWithTrigger'
  >) {
  return (
    <SelectPrimitive.Portal>
      <SelectPrimitive.Positioner
        side={side}
        sideOffset={sideOffset}
        align={align}
        alignOffset={alignOffset}
        alignItemWithTrigger={alignItemWithTrigger}
        className="isolate z-50"
      >
        <SelectPrimitive.Popup
          data-slot="select-content"
          data-align-trigger={alignItemWithTrigger}
          className={cn(
            'relative isolate z-50 max-h-(--available-height) w-(--anchor-width) min-w-36 origin-(--transform-origin) overflow-x-hidden overflow-y-auto scroll-quiet rounded-lg p-1 shadow-md ring-1 duration-100',
            'bg-card text-fg ring-fg/10',
            'data-closed:animate-out data-closed:fade-out-0 data-closed:zoom-out-95 data-open:animate-in data-open:fade-in-0 data-open:zoom-in-95 data-[align-trigger=true]:animate-none data-[side=bottom]:slide-in-from-top-2 data-[side=inline-end]:slide-in-from-left-2 data-[side=inline-start]:slide-in-from-right-2 data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2',
            className
          )}
          {...props}
        >
          <SelectScrollUpButton />
          <SelectPrimitive.List>{children}</SelectPrimitive.List>
          <SelectScrollDownButton />
        </SelectPrimitive.Popup>
      </SelectPrimitive.Positioner>
    </SelectPrimitive.Portal>
  );
}

function SelectLabel({ className, ...props }: SelectPrimitive.GroupLabel.Props) {
  return (
    <SelectPrimitive.GroupLabel
      data-slot="select-label"
      className={cn('px-1.5 py-1 text-sm font-medium text-muted', className)}
      {...props}
    />
  );
}

function SelectItem({ className, children, ...props }: SelectPrimitive.Item.Props) {
  return (
    <SelectPrimitive.Item
      data-slot="select-item"
      className={cn(
        'relative flex w-full cursor-default items-center gap-1.5 rounded-md py-1 pr-8 pl-1.5 text-sm outline-hidden select-none',
        // Hover/Focus state
        'focus:bg-subtle focus:text-fg data-highlighted:bg-subtle data-highlighted:text-fg',
        "data-disabled:pointer-events-none data-disabled:opacity-50 [&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-4 *:[span]:last:flex *:[span]:last:items-center *:[span]:last:gap-2",
        className
      )}
      {...props}
    >
      <SelectPrimitive.ItemText className="flex flex-1 shrink-0 gap-2 whitespace-nowrap">
        {children}
      </SelectPrimitive.ItemText>
      <SelectPrimitive.ItemIndicator
        render={
          <span className="pointer-events-none absolute right-2 flex size-4 items-center justify-center">
            <CheckIcon className="size-3.5" />
          </span>
        }
      />
    </SelectPrimitive.Item>
  );
}

function SelectSeparator({ className, ...props }: SelectPrimitive.Separator.Props) {
  return (
    <SelectPrimitive.Separator
      data-slot="select-separator"
      className={cn('pointer-events-none -mx-1 my-1 h-px bg-subtle', className)}
      {...props}
    />
  );
}

function SelectScrollUpButton({
  className,
  ...props
}: React.ComponentProps<typeof SelectPrimitive.ScrollUpArrow>) {
  return (
    <SelectPrimitive.ScrollUpArrow
      data-slot="select-scroll-up-button"
      className={cn(
        'top-0 z-10 flex w-full cursor-default items-center justify-center bg-card/80 py-1 backdrop-blur-sm',
        "[&_svg:not([class*='size-'])]:size-4",
        className
      )}
      {...props}
    >
      <ChevronUpIcon />
    </SelectPrimitive.ScrollUpArrow>
  );
}

function SelectScrollDownButton({
  className,
  ...props
}: React.ComponentProps<typeof SelectPrimitive.ScrollDownArrow>) {
  return (
    <SelectPrimitive.ScrollDownArrow
      data-slot="select-scroll-down-button"
      className={cn(
        'bottom-0 z-10 flex w-full cursor-default items-center justify-center bg-card/80 py-1 backdrop-blur-sm',
        "[&_svg:not([class*='size-'])]:size-4",
        className
      )}
      {...props}
    >
      <ChevronDownIcon />
    </SelectPrimitive.ScrollDownArrow>
  );
}

export {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectLabel,
  SelectScrollDownButton,
  SelectScrollUpButton,
  SelectSeparator,
  SelectTrigger,
  SelectValue
};
