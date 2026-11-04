'use client';

import { AlertTriangleIcon, ChevronDownIcon, ClockIcon, LightbulbIcon, TargetIcon, ZapIcon } from 'lucide-react';
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger
} from '@braswelljr/algorithm-ui/components/collapsible';
import { Card } from '@braswelljr/algorithm-ui/components/card';
import { cn } from '@braswelljr/algorithm-ui/lib/utils';

export type EntryNotes = { tldr: string; keyIdea: string; whenToUse: string; pitfall: string };

/**
 * The shortcut to an entry: the TL;DR always shown, and the key idea, the
 * cost and why, when to use it and the pitfall behind a disclosure, open by
 * default. Each part has an icon and a heading, so none depends on color.
 *
 *   <TLDRCard notes={entry.notes} time={entry.time} why={entry.why} />
 */
export function TLDRCard({
  notes,
  time,
  space,
  why,
  defaultOpen = true,
  className
}: {
  notes: EntryNotes;
  time?: string;
  space?: string;
  why?: string;
  defaultOpen?: boolean;
  className?: string;
}) {
  const parts = [
    { icon: LightbulbIcon, title: 'Key idea', text: notes.keyIdea },
    {
      icon: ClockIcon,
      title: 'Cost',
      text: [time && `${time}${space ? ` time, ${space} space` : ''}.`, why].filter(Boolean).join(' ')
    },
    { icon: TargetIcon, title: 'When to use it', text: notes.whenToUse },
    { icon: AlertTriangleIcon, title: 'Pitfall', text: notes.pitfall }
  ].filter((p) => p.text);
  return (
    <Collapsible
      defaultOpen={defaultOpen}
      render={
        <Card
          variant="accent"
          size="md"
          className={cn('gap-0 px-(--card-spacing)', className)}
        />
      }
    >
      <div className="flex items-start gap-3">
        <ZapIcon
          aria-hidden
          className="mt-0.5 size-5 shrink-0 text-accent"
        />
        <div className="min-w-0 flex-1">
          <p className="text-xs font-semibold tracking-wide text-accent uppercase">TL;DR</p>
          <p className="mt-0.5 text-base text-pretty text-fg">{notes.tldr}</p>
        </div>
        <CollapsibleTrigger className="group inline-flex shrink-0 items-center gap-1 rounded-md px-2 py-1 text-sm text-muted hover:bg-subtle hover:text-fg">
          <span className="group-data-panel-open:hidden">More</span>
          <span className="hidden group-data-panel-open:inline">Less</span>
          <ChevronDownIcon
            aria-hidden
            className="size-4 transition-transform group-data-panel-open:rotate-180"
          />
        </CollapsibleTrigger>
      </div>
      <CollapsibleContent>
        <dl className="mt-4 grid gap-3 sm:grid-cols-2">
          {parts.map((p) => (
            <Card
              key={p.title}
              size="sm"
              className="gap-1 px-(--card-spacing)"
            >
              <dt className="flex items-center gap-1.5 text-sm font-semibold text-fg">
                <p.icon
                  aria-hidden
                  className="size-4 text-muted"
                />
                {p.title}
              </dt>
              <dd className="text-sm text-pretty text-muted">{p.text}</dd>
            </Card>
          ))}
        </dl>
      </CollapsibleContent>
    </Collapsible>
  );
}
