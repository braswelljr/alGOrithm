import { CodeSample } from '@/components/ui/code-block';
import { Heading } from '@/components/ui/heading';
import { parseGoDoc } from '@/lib/go-doc';
import { cn } from '@/lib/utils';

/**
 * A Go doc comment, drawn at the full width: paragraphs re-flowed (the source
 * wraps at 80 columns), headings, lists, and every indented example in the
 * code block.
 */
export function GoDoc({
  text,
  importPath,
  className
}: {
  text: string;
  /** The package's import path, shown as the first code block. */
  importPath?: string;
  className?: string;
}) {
  const blocks = parseGoDoc(text);
  return (
    <div className={cn('flex w-full min-w-0 flex-col gap-3', className)}>
      {importPath ? (
        <CodeSample
          title="Import"
          language="go"
          code={`import "${importPath}"`}
          size="sm"
        />
      ) : null}
      {blocks.map((b, i) => {
        switch (b.kind) {
          case 'heading':
            return (
              <Heading
                key={i}
                level={4}
                className="mt-2 text-base font-semibold"
              >
                {b.text}
              </Heading>
            );
          case 'list':
            return (
              <ul
                key={i}
                className="flex list-disc flex-col gap-1 ps-5"
              >
                {b.items.map((item, j) => (
                  <li key={j}>{item}</li>
                ))}
              </ul>
            );
          case 'code':
            return (
              <CodeSample
                key={i}
                title="Example"
                language="go"
                code={b.code}
                size="sm"
              />
            );
          default:
            return <p key={i}>{b.text}</p>;
        }
      })}
    </div>
  );
}
