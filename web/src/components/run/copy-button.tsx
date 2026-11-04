'use client';

import type { ReactNode } from 'react';
import { Button } from '@/components/shared/button';
import { useCopy } from '@/hooks/use-copy';

/** A button that copies text and says, in words, whether it worked. */
export function CopyButton({ text, children }: { text: () => string; children: ReactNode }) {
  const [state, copy] = useCopy();
  return (
    <Button
      size="sm"
      onClick={() => copy(text())}
      aria-live="polite"
    >
      {state === 'done' ? 'Copied' : state === 'failed' ? 'Copy failed' : children}
    </Button>
  );
}
