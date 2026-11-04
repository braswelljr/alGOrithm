import { CircleAlert } from 'lucide-react';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import type { Problem } from '@/visualizer/core';

/** What could not be read, with where each bad value was. */
export function Problems({ error, problems }: { error: string; problems: Problem[] }) {
  return (
    <Alert variant="destructive">
      <CircleAlert aria-hidden />
      <AlertTitle>{problems.length > 0 ? 'Some values could not be read:' : error}</AlertTitle>
      {problems.length > 0 ? (
        <AlertDescription>
          <ul className="mt-1 list-disc pl-5 font-mono text-xs">
            {problems.map((p, i) => (
              <li key={i}>
                {p.line ? `line ${p.line}` : p.column ? `item ${p.column}` : 'value'}
                {p.line && p.column ? `, column ${p.column}` : ''}: &quot;{p.value}&quot; is{' '}
                {p.reason}
              </li>
            ))}
          </ul>
        </AlertDescription>
      ) : null}
    </Alert>
  );
}
