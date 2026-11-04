import { cn } from '@/lib/utils';

function Skeleton({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return (
    <div
      // Only a shape: whatever is loading is announced by the page around it.
      aria-hidden="true"
      data-slot="skeleton"
      className={cn('bg-subtle animate-pulse rounded-md', className)}
      {...props}
    />
  );
}

export { Skeleton };
