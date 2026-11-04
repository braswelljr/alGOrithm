'use client';

import {
  createContext,
  useCallback,
  useContext,
  useMemo,
  type ComponentProps,
  type ReactNode
} from 'react';
import { X } from 'lucide-react';
import { Badge, type BadgeProps } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  ResizablePanel,
  ResizablePanelGroup,
  ResizableHandle as RHandle
} from '@/components/ui/resizable';
import { cn } from '@/lib/utils';

// ---------------------------------------------------------------------------
// Context
// ---------------------------------------------------------------------------

type SplitContextValue = {
  open: boolean;
  setOpen: (open: boolean) => void;
  defaultLeftSize: number;
  minSize: number;
};

const SplitContext = createContext<SplitContextValue | null>(null);

function useSplitContext() {
  const ctx = useContext(SplitContext);
  if (!ctx) throw new Error('ResizableSplitView compound must be inside <ResizableSplitView.Root>');
  return ctx;
}

/**
 * Access the split-view context from any descendant component.
 *
 * @returns `{ open, setOpen, defaultLeftSize, minSize }`
 */
export function useResizableSplitView() {
  return useSplitContext();
}

// ---------------------------------------------------------------------------
// Root
// ---------------------------------------------------------------------------

type SplitRootProps = {
  children: ReactNode;
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
  defaultLeftSize?: number;
  minSize?: number;
  className?: string;
  /** The group's id, and its layout kept and restored (`useDefaultLayout`),
   *  so a split is remembered. */
  id?: string;
  defaultLayout?: ComponentProps<typeof ResizablePanelGroup>['defaultLayout'];
  onLayoutChanged?: ComponentProps<typeof ResizablePanelGroup>['onLayoutChanged'];
};

/**
 * Context provider for the resizable split-view. Renders a two-panel
 * layout when `open` is true; collapses to a single panel otherwise.
 *
 * @example
 * ```tsx
 * <ResizableSplitView.Root
 *   open={viewingDoc != null}
 *   onOpenChange={(o) => !o && setViewingDoc(null)}
 *   defaultLeftSize={60}
 * >
 *   <ResizableSplitView.Left>
 *     <ScoringPanel />
 *   </ResizableSplitView.Left>
 *   <ResizableSplitView.Handle />
 *   <ResizableSplitView.Right>
 *     <DocumentViewer.Root url={url} filename={name}>
 *       <DocumentViewer.Content />
 *     </DocumentViewer.Root>
 *   </ResizableSplitView.Right>
 * </ResizableSplitView.Root>
 * ```
 */
function SplitRoot({
  children,
  open = false,
  onOpenChange,
  defaultLeftSize = 50,
  minSize = 20,
  className,
  id,
  defaultLayout,
  onLayoutChanged
}: SplitRootProps) {
  const setOpen = useCallback((v: boolean) => onOpenChange?.(v), [onOpenChange]);

  const ctx = useMemo<SplitContextValue>(
    () => ({ open, setOpen, defaultLeftSize, minSize }),
    [open, setOpen, defaultLeftSize, minSize]
  );

  if (!open) {
    return (
      <SplitContext.Provider value={ctx}>
        <div className={cn('flex h-full w-full min-w-0 flex-col overflow-hidden', className)}>
          {children}
        </div>
      </SplitContext.Provider>
    );
  }

  return (
    <SplitContext.Provider value={ctx}>
      <ResizablePanelGroup
        orientation="horizontal"
        id={id}
        defaultLayout={defaultLayout}
        onLayoutChanged={onLayoutChanged}
        className={cn('h-full w-full min-w-0', className)}
      >
        {children}
      </ResizablePanelGroup>
    </SplitContext.Provider>
  );
}

// ---------------------------------------------------------------------------
// Left
// ---------------------------------------------------------------------------

type SplitLeftProps = {
  children: ReactNode;
  className?: string;
} & PanelExtras;

/** What a side may add: an id for a kept layout, a handle to collapse or
 *  expand it, and whether it may be collapsed. */
type PanelExtras = Pick<
  ComponentProps<typeof ResizablePanel>,
  'id' | 'panelRef' | 'collapsible' | 'collapsedSize'
>;

/**
 * Left panel. Takes `defaultSize`% when the split is open; renders as a
 * full-width `flex-1` container when closed.
 *
 * @example
 * ```tsx
 * <ResizableSplitView.Left>
 *   <EvidencePanel />
 * </ResizableSplitView.Left>
 * ```
 */
function SplitLeft({ children, className, ...extras }: SplitLeftProps) {
  const { open, defaultLeftSize, minSize } = useSplitContext();

  if (!open) {
    return <div className={cn('min-w-0 flex-1 overflow-auto', className)}>{children}</div>;
  }

  return (
    <ResizablePanel
      defaultSize={defaultLeftSize}
      minSize={minSize}
      {...extras}
      className={cn('flex min-w-0 flex-col', className)}
    >
      {children}
    </ResizablePanel>
  );
}

// ---------------------------------------------------------------------------
// Handle
// ---------------------------------------------------------------------------

type SplitHandleProps = {
  className?: string;
};

/**
 * Draggable divider between panels. Hidden when the split is closed.
 *
 * @example
 * ```tsx
 * <ResizableSplitView.Handle className="hover:bg-primary-50" />
 * ```
 */
function SplitHandle({ className }: SplitHandleProps) {
  const { open } = useSplitContext();
  if (!open) return null;

  return (
    <RHandle
      withHandle
      className={cn('hover:bg-subtle', className)}
    />
  );
}

// ---------------------------------------------------------------------------
// Right
// ---------------------------------------------------------------------------

type SplitRightProps = {
  children: ReactNode;
  className?: string;
} & PanelExtras;

/**
 * Right panel. Takes `100 - defaultLeftSize`% of the available width. Hidden
 * entirely when the split is closed. Includes a left border and white
 * background by default.
 *
 * @example
 * ```tsx
 * <ResizableSplitView.Right className="bg-bg">
 *   <ResizableSplitView.Toolbar>...</ResizableSplitView.Toolbar>
 *   <ResizableSplitView.Content>...</ResizableSplitView.Content>
 * </ResizableSplitView.Right>
 * ```
 */
function SplitRight({ children, className, ...extras }: SplitRightProps) {
  const { open, defaultLeftSize, minSize } = useSplitContext();
  if (!open) return null;

  return (
    <ResizablePanel
      defaultSize={100 - defaultLeftSize}
      minSize={minSize}
      {...extras}
      className={cn('border-border bg-card flex min-w-0 flex-col border-l', className)}
    >
      {children}
    </ResizablePanel>
  );
}

// ---------------------------------------------------------------------------
// Close
// ---------------------------------------------------------------------------

type SplitCloseProps = ComponentProps<typeof Button> & {
  className?: string;
};

/**
 * Button that closes the viewer panel. Renders an `X` icon by default.
 * Accepts all `Button` props.
 *
 * @example
 * ```tsx
 * <ResizableSplitView.Close />
 * <ResizableSplitView.Close variant="outline" size="sm">Done</ResizableSplitView.Close>
 * ```
 */
function SplitClose({ className, children, ...props }: SplitCloseProps) {
  const { setOpen } = useSplitContext();

  return (
    <Button
      variant="ghost"
      size="icon-xs"
      onClick={() => setOpen(false)}
      aria-label="Close viewer"
      className={cn('shrink-0', className)}
      {...props}
    >
      {children ?? <X className="size-4" />}
    </Button>
  );
}

// ---------------------------------------------------------------------------
// Toolbar
// ---------------------------------------------------------------------------

type SplitToolbarProps = {
  children: ReactNode;
  className?: string;
};

/**
 * Header bar inside the right panel. Place filename, badges, and action
 * buttons here.
 *
 * @example
 * ```tsx
 * <ResizableSplitView.Toolbar>
 *   <span className="truncate text-sm font-medium">report.pdf</span>
 *   <ResizableSplitView.Badge variant="outline" size="sm">PDF</ResizableSplitView.Badge>
 *   <div className="flex-1" />
 *   <ResizableSplitView.Close />
 * </ResizableSplitView.Toolbar>
 * ```
 */
function SplitToolbar({ children, className }: SplitToolbarProps) {
  return (
    <div
      className={cn(
        'border-border flex shrink-0 items-center justify-between gap-2 border-b px-3 py-2',
        className
      )}
    >
      {children}
    </div>
  );
}

// ---------------------------------------------------------------------------
// Content
// ---------------------------------------------------------------------------

type SplitContentProps = {
  children: ReactNode;
  className?: string;
};

/**
 * Scrollable body region inside the right panel. Drop a `DocumentViewer.Root`
 * here.
 *
 * @example
 * ```tsx
 * <ResizableSplitView.Content>
 *   <DocumentViewer.Root url={url} filename={name}>
 *     <DocumentViewer.Content />
 *     <DocumentViewer.Footer />
 *   </DocumentViewer.Root>
 * </ResizableSplitView.Content>
 * ```
 */
function SplitContent({ children, className }: SplitContentProps) {
  return <div className={cn('min-h-0 flex-1 overflow-auto', className)}>{children}</div>;
}

// ---------------------------------------------------------------------------
// Badge, inherits from `Badge` and adds split-view presets
// ---------------------------------------------------------------------------

/**
 * Status badge for the viewer toolbar. Inherits all `Badge` props
 * (`variant`, `size`, `className`) plus a `state` shorthand for common
 * document states.
 *
 * @example
 * ```tsx
 * // Via state shortcuts
 * <ResizableSplitView.Badge state="processed" />
 * <ResizableSplitView.Badge state="pending" />
 *
 * // Full badge API
 * <ResizableSplitView.Badge variant="outline" size="xs">PDF</ResizableSplitView.Badge>
 * ```
 */
function SplitBadge({
  state,
  variant,
  size = 'sm',
  className,
  children,
  ...props
}: BadgeProps & {
  state?: 'processed' | 'pending' | 'error' | 'reviewed' | 'submitted';
}) {
  const stateVariant: Record<string, { variant: BadgeProps['variant']; label: string }> = {
    processed: { variant: 'info', label: 'Processed' },
    pending: { variant: 'warning', label: 'Pending' },
    error: { variant: 'destructive', label: 'Error' },
    reviewed: { variant: 'secondary', label: 'Reviewed' },
    submitted: { variant: 'default', label: 'Submitted' }
  };

  if (state) {
    const preset = stateVariant[state];
    return (
      <Badge
        variant={preset.variant}
        size={size}
        className={cn('shrink-0', className)}
        {...props}
      >
        {children ?? preset.label}
      </Badge>
    );
  }

  return (
    <Badge
      variant={variant}
      size={size}
      className={cn('shrink-0', className)}
      {...props}
    >
      {children}
    </Badge>
  );
}

// ---------------------------------------------------------------------------
// Compound export
// ---------------------------------------------------------------------------

export const ResizableSplitView = {
  Root: SplitRoot,
  Left: SplitLeft,
  Handle: SplitHandle,
  Right: SplitRight,
  Close: SplitClose,
  Toolbar: SplitToolbar,
  Content: SplitContent,
  Badge: SplitBadge
};

export {
  SplitRoot,
  SplitLeft,
  SplitHandle,
  SplitRight,
  SplitClose,
  SplitToolbar,
  SplitContent,
  SplitBadge
};
