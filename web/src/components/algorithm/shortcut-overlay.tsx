'use client';

import * as React from 'react';
import { SearchIcon } from 'lucide-react';
import { KeyCombo, usePlatformKeys } from '@/components/algorithm/key-combo';
import {
  Dialog,
  DialogDescription,
  DialogHeader,
  DialogPopup,
  DialogTitle
} from '@/components/ui/dialog';
import { Heading } from '@/components/ui/heading';
import { InputGroup, InputGroupAddon, InputGroupInput } from '@/components/ui/input-group';
import { formatKeys, scopeNames, type Shortcut, type ShortcutScope } from '@/lib/shortcuts';

/**
 * Every keyboard shortcut (the `?` key), grouped by scope, the current page's
 * scopes first, searchable by what it does or by its keys. Keys are drawn as
 * keycaps and read aloud by name.
 *
 *   <ShortcutOverlay open={open} onOpenChange={setOpen} shortcuts={shortcuts} scopes={['player']} />
 */
export function ShortcutOverlay({
  open,
  onOpenChange,
  shortcuts,
  scopes = [],
  trainerHref
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  shortcuts: readonly Shortcut[];
  /** The scopes active on this page, listed first. */
  scopes?: ShortcutScope[];
  /** A link to the shortcut trainer, if there is one. */
  trainerHref?: string;
}) {
  const [q, setQ] = React.useState('');
  const platform = usePlatformKeys();
  const query = q.toLowerCase().trim();
  const match = (s: Shortcut) =>
    !query ||
    s.description.toLowerCase().includes(query) ||
    formatKeys(s.keys, platform).toLowerCase().includes(query) ||
    s.category.toLowerCase().includes(query);
  const order: ShortcutScope[] = [
    ...scopes,
    ...(['global', 'player', 'quiz', 'list', 'palette'] as ShortcutScope[]).filter(
      (s) => !scopes.includes(s)
    )
  ];
  const groups = order
    .map((scope) => ({ scope, items: shortcuts.filter((s) => s.scope === scope && match(s)) }))
    .filter((g) => g.items.length);
  return (
    <Dialog
      open={open}
      onOpenChange={onOpenChange}
    >
      <DialogPopup className="sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>Keyboard shortcuts</DialogTitle>
          <DialogDescription>
            Learn a few and the whole site is a key away.
            {trainerHref ? (
              <>
                {' '}
                <a
                  href={trainerHref}
                  className="text-accent underline-offset-4 hover:underline"
                >
                  Practice them in the trainer
                </a>
                .
              </>
            ) : null}
          </DialogDescription>
        </DialogHeader>
        <InputGroup>
          <InputGroupAddon>
            <SearchIcon aria-hidden />
          </InputGroupAddon>
          <InputGroupInput
            aria-label="Search shortcuts"
            value={q}
            onChange={(e) => setQ(e.target.value)}
            placeholder="Search: play, theme, g l"
          />
        </InputGroup>
        <div className="scroll-quiet -mx-1 max-h-[60vh] overflow-y-auto px-1">
          {groups.length < 1 ? (
            <p className="text-muted py-6 text-center text-sm">No shortcut matches.</p>
          ) : null}
          {groups.map((g) => (
            <section
              key={g.scope}
              className="mt-4 first:mt-1"
            >
              <Heading
                level={3}
                className="text-muted mb-2 text-xs font-semibold tracking-wide uppercase"
              >
                {scopeNames[g.scope]}
                {scopes.includes(g.scope) ? (
                  <span className="text-accent ms-2 normal-case">on this page</span>
                ) : null}
              </Heading>
              <dl className="grid gap-x-6 gap-y-1 sm:grid-cols-2">
                {g.items.map((s) => (
                  <div
                    key={s.id}
                    className="hover:bg-subtle flex items-center justify-between gap-3 rounded-md px-2 py-1.5"
                  >
                    <dt className="text-fg text-sm">{s.description}</dt>
                    <dd>
                      <KeyCombo keys={s.keys} />
                    </dd>
                  </div>
                ))}
              </dl>
            </section>
          ))}
        </div>
      </DialogPopup>
    </Dialog>
  );
}
