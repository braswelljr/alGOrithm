import { CodeSample } from '@/components/ui/code-block';

/**
 * Go source with line numbers, highlighting and a copy button: the site's code
 * block, under the name the visualizer has always exported. The code is the
 * real source of the entry, as the wasm build returns it.
 */
export function CodeView({
  code,
  file,
  highlight = [],
  label
}: {
  code: string;
  file?: string;
  /** 1-based line numbers to mark. */
  highlight?: number[];
  label?: string;
}) {
  return (
    <CodeSample
      title={label ?? file ?? 'Go source'}
      language="go"
      code={code}
      highlight={highlight}
      lineNumbers
    />
  );
}
