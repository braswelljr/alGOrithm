import { ArrowRightIcon } from 'lucide-react';
import { Card } from '@braswelljr/algorithm-ui/components/card';
import { cn } from '@braswelljr/algorithm-ui/lib/utils';

/**
 * A category of the catalog: its name, summary and how many entries it has,
 * linking to the category.
 *
 *   <CategoryCard name="Sorting" summary="..." count={23} href="/#sorting" />
 */
export function CategoryCard({
  name,
  summary,
  count,
  href,
  className
}: {
  name: string;
  summary: string;
  count: number;
  href: string;
  className?: string;
}) {
  return (
    <Card
      variant="interactive"
      size="md"
      className={cn('gap-2 px-(--card-spacing)', className)}
      render={<a href={href} />}
    >
      <span className="flex items-center justify-between gap-2">
        <span className="font-semibold text-fg">{name}</span>
        <span className="text-xs text-muted tabular-nums">
          {count} {count === 1 ? 'entry' : 'entries'}
        </span>
      </span>
      <span className="text-sm text-pretty text-muted">{summary}</span>
      <span className="mt-auto inline-flex items-center gap-1 text-sm font-medium text-accent">
        Explore
        <ArrowRightIcon
          aria-hidden
          className="size-4 transition-transform group-hover:translate-x-0.5"
        />
      </span>
    </Card>
  );
}
