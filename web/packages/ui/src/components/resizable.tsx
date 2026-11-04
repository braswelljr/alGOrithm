'use client';

import * as React from 'react';
import { cn } from '@braswelljr/algorithm-ui/lib/utils';
import { GripVertical } from 'lucide-react';
import * as ResizablePrimitive from 'react-resizable-panels';

function ResizablePanelGroup({
  className,
  ...props
}: React.ComponentProps<typeof ResizablePrimitive.Group>) {
  return (
    <ResizablePrimitive.Group
      className={cn('flex h-full w-full data-[orientation=vertical]:flex-col', className)}
      {...props}
    />
  );
}

function ResizablePanel({
  className,
  ...props
}: React.ComponentProps<typeof ResizablePrimitive.Panel>) {
  return (
    <ResizablePrimitive.Panel
      className={cn('min-w-0 overflow-auto', className)}
      {...props}
    />
  );
}

function ResizableHandle({
  withHandle,
  className,
  ...props
}: React.ComponentProps<typeof ResizablePrimitive.Separator> & {
  withHandle?: boolean;
}) {
  return (
    <ResizablePrimitive.Separator
      className={cn(
        'relative flex w-2 items-center justify-center bg-subtle transition-colors after:absolute after:inset-y-0 after:left-1/2 after:w-px after:-translate-x-1/2 focus-visible:ring-1 focus-visible:ring-border focus-visible:ring-offset-1 focus-visible:outline-none data-[orientation=vertical]:h-2 data-[orientation=vertical]:w-full data-[orientation=vertical]:after:inset-x-0 data-[orientation=vertical]:after:top-1/2 data-[orientation=vertical]:after:h-px data-[orientation=vertical]:after:w-full data-[orientation=vertical]:after:-translate-y-1/2 [&[data-orientation=vertical]>div]:rotate-90',
        className
      )}
      {...props}
    >
      {withHandle && (
        <div className="z-10 flex h-6 w-1.5 items-center justify-center rounded-full border border-border bg-card shadow-sm">
          <GripVertical className="size-3 text-muted" />
        </div>
      )}
    </ResizablePrimitive.Separator>
  );
}

export { ResizablePanelGroup, ResizablePanel, ResizableHandle };
