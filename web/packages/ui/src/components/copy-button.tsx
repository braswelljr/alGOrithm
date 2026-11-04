'use client';

import { Button, type ButtonProps } from '@braswelljr/algorithm-ui/components/button';
import { useCopyToClipboard } from '@braswelljr/algorithm-ui/hooks/use-copy-to-clipboard';
import { cn } from '@braswelljr/algorithm-ui/lib/utils';
import { Check, Copy } from 'lucide-react';

type CopyButtonProps = Omit<ButtonProps, 'children' | 'onClick' | 'value'> & {
  /** The text to copy. */
  value: string;
  /** What is being copied, for the accessible name: "Copy install command". */
  label: string;
};

/**
 * An icon button that copies `value` and says whether it worked: a check mark
 * and "Copied" on success, the reason beside it on failure.
 */
export function CopyButton({ value, label, className, ...props }: CopyButtonProps) {
  const { copyToClipboard, isCopied, error } = useCopyToClipboard();

  return (
    <span className={cn('inline-flex items-center gap-2', className)}>
      {/* Announced when it changes; a sighted reader sees the icon swap. */}
      <span
        role="status"
        className={cn('text-xs text-fd-muted-foreground', !error && 'sr-only')}
      >
        {error ? 'Could not copy. Select the text instead.' : isCopied ? 'Copied' : ''}
      </span>
      <Button
        variant="ghost"
        size="icon-sm"
        aria-label={isCopied ? 'Copied' : label}
        title={label}
        onClick={() => void copyToClipboard(value)}
        {...props}
      >
        {isCopied ? <Check aria-hidden="true" /> : <Copy aria-hidden="true" />}
      </Button>
    </span>
  );
}
