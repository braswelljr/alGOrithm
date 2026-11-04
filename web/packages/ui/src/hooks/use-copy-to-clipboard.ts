'use client';

import { useCallback, useEffect, useRef, useState } from 'react';

/** Idle, or the outcome of the last attempt. Never both. */
type CopyState = { status: 'idle' | 'copied' } | { status: 'failed'; error: Error };

const IDLE: CopyState = { status: 'idle' };

export type UseCopyToClipboardOptions = {
  /**
   * How long `isCopied` stays true, in milliseconds. `0` leaves it true until
   * the next copy.
   */
  timeout?: number;
  /** Runs after a successful copy. */
  onCopy?: () => void;
};

export type UseCopyToClipboardResult = {
  /** Copies `value` and resolves to whether it worked. */
  copyToClipboard: (value: string) => Promise<boolean>;
  /** True for a moment after a successful copy. Resets itself. */
  isCopied: boolean;
  /** The last failure, for the caller to show beside its control. */
  error: Error | null;
};

/**
 * Copies text to the clipboard, and tells the truth about whether it worked.
 *
 * This is the only place the site touches `navigator.clipboard`. The Clipboard
 * API is absent outside a secure context (plain http on a LAN address, say) and
 * a browser may refuse a write, so a copy can fail; a caller that announced
 * success unconditionally would be lying to the reader. Copied and failed are
 * one piece of state because they cannot both be true.
 *
 * There is no toast here, unlike the hook this one is shaped after: the site
 * has no toast system, and every copy control already changes to say "Copied"
 * or shows `error` in place.
 */
export function useCopyToClipboard({
  timeout = 2000,
  onCopy
}: UseCopyToClipboardOptions = {}): UseCopyToClipboardResult {
  const [state, setState] = useState<CopyState>(IDLE);
  const timeoutId = useRef<ReturnType<typeof setTimeout> | null>(null);

  // A control that unmounts mid-acknowledgement must not set state afterwards.
  useEffect(
    () => () => {
      if (timeoutId.current) clearTimeout(timeoutId.current);
    },
    []
  );

  const copyToClipboard = useCallback(
    async (value: string): Promise<boolean> => {
      if (!value) return false;

      const failed = (cause: unknown): boolean => {
        const error = cause instanceof Error ? cause : new Error('Copy failed');
        if (timeoutId.current) clearTimeout(timeoutId.current);
        setState({ status: 'failed', error });
        return false;
      };

      // Over plain http the property is simply not there, so reading
      // `.writeText` off it would throw.
      if (typeof window === 'undefined' || !navigator?.clipboard?.writeText) {
        return failed(new Error('The clipboard is unavailable in this context'));
      }

      try {
        await navigator.clipboard.writeText(value);
      } catch (cause) {
        return failed(cause);
      }

      if (timeoutId.current) clearTimeout(timeoutId.current);
      setState({ status: 'copied' });
      if (timeout !== 0) {
        timeoutId.current = setTimeout(() => {
          setState(IDLE);
          timeoutId.current = null;
        }, timeout);
      }

      onCopy?.();
      return true;
    },
    [timeout, onCopy]
  );

  return {
    copyToClipboard,
    isCopied: state.status === 'copied',
    error: state.status === 'failed' ? state.error : null
  };
}
