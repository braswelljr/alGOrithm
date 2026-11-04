import Link from 'next/link';
import { Badge } from '@/components/shared/badge';
import { Card } from '@/components/ui/card';
import { entryHref, isStructure } from '@/lib/entries';
import type { Entry } from '@/visualizer/core';
import { EntryBadges } from './entry-badges';

/** One entry: its name, what it does and what it costs, linking to its page. */
export function EntryCard({ entry: e }: { entry: Entry }) {
  return (
    <Card
      variant="interactive"
      size="md"
      render={<Link href={entryHref(e)} />}
      className="flex h-full flex-col gap-2 px-(--card-spacing)"
    >
      <span className="flex items-center justify-between gap-2">
        <span className="font-medium">{e.name}</span>
        {isStructure(e) ? <Badge>playground</Badge> : e.traced ? <Badge>animated</Badge> : null}
      </span>
      <span className="text-muted text-sm">{e.summary}</span>
      <EntryBadges
        entry={e}
        max={3}
      />
      <span className="text-muted mt-auto font-mono text-xs">
        time {e.time} · space {e.space}
      </span>
    </Card>
  );
}
