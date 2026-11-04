'use client';

import * as React from 'react';
import { Button } from '@braswelljr/algorithm-ui/components/button';
import { useHighlightedCode } from '@braswelljr/algorithm-ui/hooks/use-highlighted-code';
import { cn } from '@braswelljr/algorithm-ui/lib/utils';
import { cva, type VariantProps } from 'class-variance-authority';
import { CheckIcon, CopyIcon } from 'lucide-react';

const codeBlockVariants = cva(
  'group/code-block relative flex w-full min-w-0 flex-col overflow-hidden rounded-lg text-sm has-[>[data-slot=code-block-content]:focus-visible]:ring-3 has-[>[data-slot=code-block-content]:focus-visible]:ring-ring/50 data-[size=sm]:text-xs',
  {
    variants: {
      variant: {
        default:
          'border border-input bg-card text-foreground has-[>[data-slot=code-block-content]:focus-visible]:border-ring',
        destructive: 'border border-destructive/30 text-destructive dark:border-destructive/40'
      }
    },
    defaultVariants: {
      variant: 'default'
    }
  }
);

function CodeBlock({
  className,
  variant = 'default',
  size = 'default',
  ...props
}: React.ComponentProps<'div'> &
  VariantProps<typeof codeBlockVariants> & { size?: 'default' | 'sm' }) {
  return (
    <div
      data-slot="code-block"
      data-variant={variant}
      data-size={size}
      className={cn(codeBlockVariants({ variant }), className)}
      {...props}
    />
  );
}

function CodeBlockHeader({ className, ...props }: React.ComponentProps<'div'>) {
  return (
    <div
      data-slot="code-block-header"
      className={cn(
        'flex min-w-0 shrink-0 items-center gap-2 px-4 pt-4 text-muted-foreground group-data-[size=sm]/code-block:px-3 group-data-[size=sm]/code-block:pt-3 group-data-[variant=destructive]/code-block:text-destructive *:has-[>[data-slot=code-block-title]]:min-w-0',
        className
      )}
      {...props}
    />
  );
}

function CodeBlockTitle({ className, ...props }: React.ComponentProps<'div'>) {
  return (
    <div
      data-slot="code-block-title"
      className={cn(
        "min-w-0 truncate font-sans text-sm font-medium text-foreground group-data-[size=sm]/code-block:text-xs group-data-[variant=destructive]/code-block:text-destructive [&_svg]:pointer-events-none [&_svg]:inline-block [&_svg]:align-middle [&_svg:not([class*='size-'])]:size-4 group-data-[size=sm]/code-block:[&_svg:not([class*='size-'])]:size-3.5 *:[svg]:me-1.5",
        className
      )}
      {...props}
    />
  );
}

function CodeBlockActions({ className, ...props }: React.ComponentProps<'div'>) {
  return (
    <div
      data-slot="code-block-actions"
      className={cn(
        '-my-1 ms-auto -me-1.5 flex shrink-0 items-center gap-1 [[data-slot=code-block]>&]:absolute [[data-slot=code-block]>&]:end-3 [[data-slot=code-block]>&]:top-3 [[data-slot=code-block]>&]:z-10 [[data-slot=code-block]>&]:m-0 [[data-slot=code-block]>&]:opacity-0 [[data-slot=code-block]>&]:transition-opacity [[data-slot=code-block]>&]:group-focus-within/code-block:opacity-100 [[data-slot=code-block]>&]:group-hover/code-block:opacity-100 group-data-[size=sm]/code-block:[[data-slot=code-block]>&]:end-2 group-data-[size=sm]/code-block:[[data-slot=code-block]>&]:top-2 [[data-slot=code-block]>&]:pointer-coarse:opacity-100 [[data-slot=code-block]>&]:*:[button]:bg-background [[data-slot=code-block]>&]:*:[button]:ring-1 [[data-slot=code-block]>&]:*:[button]:ring-foreground/10',
        className
      )}
      {...props}
    />
  );
}

function CodeBlockCopyButton({
  value,
  variant = 'ghost',
  size = 'icon-xs',
  className,
  onClick,
  ...props
}: React.ComponentProps<typeof Button> & { value: string }) {
  const [copied, setCopied] = React.useState(false);

  React.useEffect(() => {
    if (!copied) {
      return;
    }

    const timeout = setTimeout(() => setCopied(false), 2000);
    return () => clearTimeout(timeout);
  }, [copied]);

  return (
    <>
      <Button
        data-slot="code-block-copy-button"
        data-copied={copied || undefined}
        variant={variant}
        size={size}
        aria-label={copied ? 'Copied' : 'Copy'}
        className={cn(className)}
        onClick={(event) => {
          onClick?.(event);
          navigator.clipboard
            .writeText(value)
            .then(() => setCopied(true))
            .catch(() => {});
        }}
        {...props}
      >
        {copied ? <CheckIcon /> : <CopyIcon />}
      </Button>
      <span
        aria-live="polite"
        className="sr-only"
      >
        {copied ? 'Copied' : ''}
      </span>
    </>
  );
}

function CodeBlockContent({
  className,
  children,
  style,
  wrap = false,
  lineNumbers = false,
  startLine = 1,
  ...props
}: React.ComponentProps<'pre'> & {
  wrap?: boolean;
  lineNumbers?: boolean;
  startLine?: number;
}) {
  return (
    <pre
      data-slot="code-block-content"
      data-wrap={wrap || undefined}
      data-numbered={lineNumbers || undefined}
      dir="ltr"
      tabIndex={0}
      style={{ counterReset: `line ${startLine - 1}`, ...style }}
      className={cn(
        'min-h-0 min-w-0 overflow-auto scroll-quiet px-4 py-4 text-start font-mono leading-relaxed outline-none group-has-data-[slot=code-block-footer]/code-block:pb-2 group-has-data-[slot=code-block-header]/code-block:pt-2 group-data-[size=sm]/code-block:px-3 group-data-[size=sm]/code-block:py-3 group-data-[size=sm]/code-block:group-has-data-[slot=code-block-header]/code-block:pt-1.5 data-wrap:wrap-break-word data-wrap:whitespace-pre-wrap',
        className
      )}
      {...props}
    >
      <code
        data-slot="code-block-code"
        className="grid min-h-lh min-w-max content-start in-data-numbered:grid-cols-[auto_1fr] in-data-wrap:min-w-0"
      >
        {lineNumbers && typeof children === 'string'
          ? children
              .replace(/\n$/, '')
              .split('\n')
              .map((line, index) => <CodeBlockLine key={index}>{line}</CodeBlockLine>)
          : children}
      </code>
    </pre>
  );
}

const codeBlockLineVariants = cva(
  "relative -mx-4 block min-h-lh px-4 group-data-[size=sm]/code-block:-mx-3 group-data-[size=sm]/code-block:px-3 before:min-w-6 before:text-end before:text-muted-foreground after:absolute after:start-1 after:top-0 after:text-muted-foreground after:select-none in-data-numbered:col-span-2 in-data-numbered:grid in-data-numbered:grid-cols-subgrid in-data-numbered:gap-x-3 in-data-numbered:[counter-increment:line] in-data-numbered:before:content-[counter(line)] in-data-numbered:before:select-none in-data-numbered:supports-[content:counter(line)_/_'']:before:content-[counter(line)_/_'']",
  {
    variants: {
      variant: {
        default: '',
        added: "bg-subtle after:content-['+']",
        removed: "bg-destructive/10 after:text-destructive after:content-['−']",
        highlighted: 'bg-accent/10 shadow-[inset_2px_0_0_var(--accent)]'
      }
    },
    defaultVariants: {
      variant: 'default'
    }
  }
);

function CodeBlockLine({
  className,
  variant = 'default',
  children,
  ...props
}: React.ComponentProps<'span'> & VariantProps<typeof codeBlockLineVariants>) {
  // An empty line holds a newline so copying a selection keeps blank lines.
  const empty =
    children === undefined ||
    children === null ||
    children === '' ||
    (Array.isArray(children) && children.length === 0);

  return (
    <span
      data-slot="code-block-line"
      data-variant={variant}
      className={cn(codeBlockLineVariants({ variant }), className)}
      {...props}
    >
      <span
        data-slot="code-block-line-content"
        className="min-w-0"
      >
        {empty ? '\n' : children}
      </span>
    </span>
  );
}

function CodeBlockFooter({ className, ...props }: React.ComponentProps<'div'>) {
  return (
    <div
      data-slot="code-block-footer"
      className={cn(
        'flex min-w-0 shrink-0 items-center gap-2 px-4 pb-4 text-muted-foreground group-data-[size=sm]/code-block:px-3 group-data-[size=sm]/code-block:pb-3',
        className
      )}
      {...props}
    />
  );
}

/**
 * A code sample in one piece: a title (the file name or "Shell"), a copy
 * button, highlighting with shiki, optional line numbers and highlighted lines.
 * The composition parts above (CodeBlock, CodeBlockHeader, ...) build anything
 * else. Lines are 1-based in `highlight`.
 *
 *   <CodeSample title="main.go" language="go" code={src} highlight={[3, 4]} lineNumbers />
 */
function CodeSample({
  code,
  language = 'text',
  title,
  copyValue,
  highlight = [],
  lineNumbers = false,
  wrap = false,
  size = 'default',
  className
}: {
  code: string;
  /** A shiki language: go, bash, json, ts, tsx, ...; plain text otherwise. */
  language?: string;
  /** Shown above the code, and the block's accessible name. */
  title?: string;
  /** What the copy button copies, when it should differ from what is shown. */
  copyValue?: string;
  highlight?: number[];
  lineNumbers?: boolean;
  wrap?: boolean;
  size?: 'default' | 'sm';
  className?: string;
}) {
  const lines = useHighlightedCode(code, language);
  const marked = new Set(highlight);
  return (
    <CodeBlock
      size={size}
      className={className}
      role="figure"
      aria-label={title ?? `${language} code`}
    >
      {title ? (
        <CodeBlockHeader>
          <CodeBlockTitle>{title}</CodeBlockTitle>
          <CodeBlockActions>
            <CodeBlockCopyButton value={copyValue ?? code} />
          </CodeBlockActions>
        </CodeBlockHeader>
      ) : (
        <CodeBlockActions>
          <CodeBlockCopyButton value={copyValue ?? code} />
        </CodeBlockActions>
      )}
      <CodeBlockContent
        lineNumbers={lineNumbers}
        wrap={wrap}
      >
        {lines.map((line, i) => (
          <CodeBlockLine
            key={i}
            variant={marked.has(i + 1) ? 'highlighted' : 'default'}
            data-line={i + 1}
          >
            {line}
          </CodeBlockLine>
        ))}
      </CodeBlockContent>
    </CodeBlock>
  );
}

export {
  CodeSample,
  CodeBlock,
  CodeBlockHeader,
  CodeBlockTitle,
  CodeBlockActions,
  CodeBlockCopyButton,
  CodeBlockContent,
  CodeBlockLine,
  CodeBlockFooter
};
