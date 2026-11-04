import Link from 'next/link';
import type { Entry } from '@braswelljr/algorithm-visualizer/core';
import { Badge } from '@/components/ui/badge';
import { words } from '@/lib/filters';

/**
 * An entry's kind, level and tags as text badges. With `links`, each tag
 * opens its place on the tags page (do not use inside another link).
 */
export function EntryBadges({
  entry,
  max,
  links = false
}: {
  entry: Pick<Entry, 'kind' | 'level' | 'tags'>;
  max?: number;
  links?: boolean;
}) {
  const tags = max ? entry.tags.slice(0, max) : entry.tags;
  return (
    <span className="flex flex-wrap gap-1">
      <Badge title="kind">{words(entry.kind)}</Badge>
      {entry.level ? <Badge title="level">{entry.level}</Badge> : null}
      {tags.map((t) =>
        links ? (
          <Link
            key={t}
            href={`/docs/tags/#tag-${t}`}
            className="rounded-full focus-visible:outline-2 focus-visible:outline-accent"
          >
            <Badge title="tag">#{t}</Badge>
          </Link>
        ) : (
          <Badge
            key={t}
            title="tag"
          >
            #{t}
          </Badge>
        )
      )}
      {max && entry.tags.length > max ? (
        <Badge title="more tags">+{entry.tags.length - max}</Badge>
      ) : null}
    </span>
  );
}
