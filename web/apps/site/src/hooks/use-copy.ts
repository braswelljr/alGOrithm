'use client';

import { useCallback, useState } from 'react';
import { copyText } from '@/lib/share';

/** Copies text and says whether it worked, for two seconds. */
export function useCopy(): [state: 'idle' | 'done' | 'failed', copy: (text: string) => void] {
  const [state, setState] = useState<'idle' | 'done' | 'failed'>('idle');
  const copy = useCallback((text: string) => {
    void copyText(text).then((ok) => {
      setState(ok ? 'done' : 'failed');
      setTimeout(() => setState('idle'), 2000);
    });
  }, []);
  return [state, copy];
}
