'use client';

import * as React from 'react';
import { AnimatePresence, motion } from 'motion/react';
import { XIcon } from 'lucide-react';
import { Card } from '@braswelljr/algorithm-ui/components/card';
import { useReducedMotion } from '@braswelljr/algorithm-ui/hooks/use-reduced-motion';
import { softSpring } from '@braswelljr/algorithm-ui/lib/motion';
import { cn } from '@braswelljr/algorithm-ui/lib/utils';

/**
 * A card that grows into a dialog: the trigger and the panel share a layoutId,
 * so opening morphs one into the other (FLIP, transform only) instead of
 * popping. It is a modal dialog: focus moves into it, Escape and the backdrop
 * close it, and focus returns to the trigger. Under reduced motion it fades.
 *
 *   <MorphingDialog title="Quick Sort" trigger={<EntrySummary />}>details</MorphingDialog>
 */
export function MorphingDialog({
  title,
  trigger,
  children,
  className
}: {
  title: string;
  trigger: React.ReactNode;
  children: React.ReactNode;
  className?: string;
}) {
  const [open, setOpen] = React.useState(false);
  const reduce = useReducedMotion();
  const id = React.useId();
  const opener = React.useRef<HTMLButtonElement>(null);
  const panel = React.useRef<HTMLDivElement>(null);
  React.useEffect(() => {
    if (!open) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') setOpen(false);
    };
    window.addEventListener('keydown', onKey);
    panel.current?.focus();
    return () => window.removeEventListener('keydown', onKey);
  }, [open]);
  const close = () => {
    setOpen(false);
    opener.current?.focus();
  };
  return (
    <>
      <Card
        variant="interactive"
        size="md"
        className={cn('w-full px-(--card-spacing) text-left', className)}
        render={
          <motion.button
            ref={opener}
            type="button"
            layoutId={reduce ? undefined : `${id}-card`}
            aria-haspopup="dialog"
            aria-expanded={open}
            onClick={() => setOpen(true)}
          />
        }
      >
        {trigger}
      </Card>
      <AnimatePresence>
        {open ? (
          <div className="fixed inset-0 z-50 grid place-items-center p-4">
            <motion.div
              aria-hidden
              className="absolute inset-0 bg-black/40 backdrop-blur-sm"
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              onClick={close}
            />
            <Card
              size="lg"
              className="max-h-[85vh] w-full max-w-xl overflow-y-auto rounded-2xl px-(--card-spacing) shadow-xl outline-none"
              render={
                <motion.div
                  ref={panel}
                  role="dialog"
                  aria-modal="true"
                  aria-labelledby={`${id}-title`}
                  tabIndex={-1}
                  layoutId={reduce ? undefined : `${id}-card`}
                  initial={reduce ? { opacity: 0 } : undefined}
                  animate={reduce ? { opacity: 1 } : undefined}
                  exit={reduce ? { opacity: 0 } : undefined}
                  transition={softSpring}
                />
              }
            >
              <div className="flex items-start justify-between gap-4">
                <h2
                  id={`${id}-title`}
                  className="text-lg font-semibold text-fg"
                >
                  {title}
                </h2>
                <button
                  type="button"
                  onClick={close}
                  aria-label="Close"
                  className="rounded-md p-1 text-muted hover:bg-subtle hover:text-fg"
                >
                  <XIcon className="size-5" />
                </button>
              </div>
              <div>{children}</div>
            </Card>
          </div>
        ) : null}
      </AnimatePresence>
    </>
  );
}
