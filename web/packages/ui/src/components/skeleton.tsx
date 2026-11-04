import { cn } from '@braswelljr/algorithm-ui/lib/utils';

function Skeleton({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return (
    <div
      // Only a shape: whatever is loading is announced by the page around it.
      aria-hidden="true"
      data-slot="skeleton"
      className={cn('animate-pulse rounded-md bg-subtle', className)}
      {...props}
    />
  );
}

export { Skeleton };
