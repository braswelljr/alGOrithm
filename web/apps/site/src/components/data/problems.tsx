import type { Problem } from '@braswelljr/algorithm-visualizer/core';

/** What could not be read, with where each bad value was. */
export function Problems({ error, problems }: { error: string; problems: Problem[] }) {
  return (
    <div
      role="alert"
      className="rounded-md border border-danger/40 bg-danger/5 p-3 text-sm"
    >
      <p className="font-medium text-danger">
        {problems.length > 0 ? 'Some values could not be read:' : error}
      </p>
      {problems.length > 0 ? (
        <ul className="mt-1 list-disc pl-5 font-mono text-xs">
          {problems.map((p, i) => (
            <li key={i}>
              {p.line ? `line ${p.line}` : p.column ? `item ${p.column}` : 'value'}
              {p.line && p.column ? `, column ${p.column}` : ''}: &quot;{p.value}&quot; is{' '}
              {p.reason}
            </li>
          ))}
        </ul>
      ) : null}
    </div>
  );
}
