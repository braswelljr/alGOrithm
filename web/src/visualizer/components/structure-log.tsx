import type { Step } from '../core/schema';

/** The operations of a structure run, the current one marked. */
export function StructureLog({ steps, current }: { steps: Step[]; current: number }) {
  return (
    <ol
      className="av-counts"
      aria-label="Operations"
    >
      {steps.map((s, n) => (
        <li
          key={n}
          aria-current={n === current ? 'step' : undefined}
          style={n === current ? { color: 'var(--av-fg)', fontWeight: 600 } : undefined}
        >
          {n}. {s.op.name}
          {s.op.value ? ` ${s.op.value}` : ''}
        </li>
      ))}
    </ol>
  );
}
