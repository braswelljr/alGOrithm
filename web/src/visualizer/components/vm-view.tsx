import type { Exec, ExecFrame } from '../core/schema';
import { Listing, Panel, SPRING, useEnter } from './exec-parts';

/** A stack drawn top first; a new top springs down into place. */
function StackColumn({ items, label, quiet }: { items: string[]; label: string; quiet: boolean }) {
  const top = useEnter<HTMLLIElement>(
    `${items.length}:${items[items.length - 1] ?? ''}`,
    quiet,
    [
      { transform: 'translateY(-14px) scale(0.96)', opacity: 0 },
      { transform: 'translateY(0) scale(1)', opacity: 1 }
    ],
    { duration: 260, easing: SPRING }
  );
  return (
    <Panel title={label}>
      {items.length ? (
        <ol
          className="m-0 flex list-none flex-col-reverse gap-1 p-0"
          aria-label={`${label}, bottom first: ${items.join(', ')}`}
        >
          {items.map((v, i) => {
            const isTop = i === items.length - 1;
            return (
              <li
                key={`${i}-${v}`}
                ref={isTop ? top : undefined}
                data-top={isTop}
                className="border-border bg-subtle data-[top=true]:border-accent data-[top=true]:bg-accent/10 flex items-center justify-between gap-3 rounded-md border px-3 py-1.5 font-mono text-sm"
              >
                <span className="truncate">{v}</span>
                {isTop ? <span className="text-accent text-[10px] uppercase">top</span> : null}
              </li>
            );
          })}
        </ol>
      ) : (
        <p className="text-muted m-0 text-sm">empty</p>
      )}
    </Panel>
  );
}

function Variables({
  vars,
  previous,
  label,
  quiet
}: {
  vars: Record<string, string>;
  previous?: Record<string, string>;
  label: string;
  quiet: boolean;
}) {
  const names = Object.keys(vars).sort();
  return (
    <Panel title={label}>
      {names.length ? (
        <dl className="m-0 grid grid-cols-[auto_1fr] gap-x-3 gap-y-1 font-mono text-sm">
          {names.map((k) => (
            <VarRow
              key={k}
              name={k}
              value={vars[k]}
              changed={previous !== undefined && previous[k] !== vars[k]}
              quiet={quiet}
            />
          ))}
        </dl>
      ) : (
        <p className="text-muted m-0 text-sm">none yet</p>
      )}
    </Panel>
  );
}

function VarRow({
  name,
  value,
  changed,
  quiet
}: {
  name: string;
  value: string;
  changed: boolean;
  quiet: boolean;
}) {
  const ref = useEnter<HTMLElement>(changed ? value : null, quiet || !changed, [
    { transform: 'translateY(-60%)', opacity: 0 },
    { transform: 'translateY(0)', opacity: 1 }
  ]);
  return (
    <>
      <dt className="text-muted">{name}</dt>
      <dd
        ref={ref}
        data-changed={changed}
        className="data-[changed=true]:text-accent m-0 font-semibold tabular-nums"
      >
        {value}
        {changed ? <span className="sr-only"> (changed)</span> : null}
      </dd>
    </>
  );
}

/**
 * A machine with a stack: a bytecode VM (listing, operand stack, frame
 * variables, output), an interpreter (source lines, variables, output), the
 * shunting-yard algorithm (operator stack, output queue) or a call stack.
 * Panels appear only for what the run has.
 */
export function VMView({
  exec,
  frame,
  previous,
  quiet = false
}: {
  exec: Exec;
  frame: ExecFrame;
  previous?: ExecFrame;
  quiet?: boolean;
}) {
  const hasStack = exec.stackLabel !== undefined || frame.stack.length > 0;
  const hasVars = exec.varsLabel !== undefined;
  const hasOutput = exec.outputLabel !== undefined;
  return (
    <div className="grid gap-3 md:grid-cols-[minmax(0,1.3fr)_minmax(0,1fr)]">
      {exec.code.length ? (
        <Listing
          code={exec.code}
          pc={frame.pc}
          quiet={quiet}
          label="Code"
        />
      ) : null}
      <div className="flex min-w-0 flex-col gap-3">
        {hasStack ? (
          <StackColumn
            items={frame.stack}
            label={exec.stackLabel ?? 'Stack'}
            quiet={quiet}
          />
        ) : null}
        {hasVars ? (
          <Variables
            vars={frame.vars}
            previous={previous?.vars}
            label={exec.varsLabel ?? 'Variables'}
            quiet={quiet}
          />
        ) : null}
        {hasOutput ? (
          <Panel title={exec.outputLabel ?? 'Output'}>
            {frame.output.length ? (
              <ol className="m-0 flex list-none flex-wrap gap-1.5 p-0 font-mono text-sm">
                {frame.output.map((v, i) => (
                  <li
                    key={i}
                    className="bg-subtle rounded px-2 py-0.5 transition-opacity duration-200 starting:opacity-0"
                  >
                    {v}
                  </li>
                ))}
              </ol>
            ) : (
              <p className="text-muted m-0 text-sm">nothing yet</p>
            )}
          </Panel>
        ) : null}
      </div>
    </div>
  );
}
