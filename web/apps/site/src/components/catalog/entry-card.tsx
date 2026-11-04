import Link from 'next/link';
import { Card } from '@braswelljr/algorithm-ui/components/card';
import type { Entry } from '@braswelljr/algorithm-visualizer/core';
import { Badge } from '@/components/ui/badge';
import { entryHref, isStructure } from '@/lib/entries';
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
      <span className="text-sm text-muted">{e.summary}</span>
      <EntryBadges
        entry={e}
        max={3}
      />
      <span className="mt-auto font-mono text-xs text-muted">
        time {e.time} · space {e.space}
      </span>
    </Card>
  );
}
