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
          className="av:m-0 av:flex av:list-none av:flex-col-reverse av:gap-1 av:p-0"
          aria-label={`${label}, bottom first: ${items.join(', ')}`}
        >
          {items.map((v, i) => {
            const isTop = i === items.length - 1;
            return (
              <li
                key={`${i}-${v}`}
                ref={isTop ? top : undefined}
                data-top={isTop}
                className="av:flex av:items-center av:justify-between av:gap-3 av:rounded-md av:border av:border-border av:bg-subtle av:px-3 av:py-1.5 av:font-mono av:text-sm av:data-[top=true]:border-accent av:data-[top=true]:bg-accent/10"
              >
                <span className="av:truncate">{v}</span>
                {isTop ? (
                  <span className="av:text-[10px] av:text-accent av:uppercase">top</span>
                ) : null}
              </li>
            );
          })}
        </ol>
      ) : (
        <p className="av:m-0 av:text-sm av:text-muted">empty</p>
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
        <dl className="av:m-0 av:grid av:grid-cols-[auto_1fr] av:gap-x-3 av:gap-y-1 av:font-mono av:text-sm">
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
        <p className="av:m-0 av:text-sm av:text-muted">none yet</p>
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
      <dt className="av:text-muted">{name}</dt>
      <dd
        ref={ref}
        data-changed={changed}
        className="av:m-0 av:font-semibold av:tabular-nums av:data-[changed=true]:text-accent"
      >
        {value}
        {changed ? <span className="av:sr-only"> (changed)</span> : null}
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
    <div className="av:grid av:gap-3 av:md:grid-cols-[minmax(0,1.3fr)_minmax(0,1fr)]">
      {exec.code.length ? (
        <Listing
          code={exec.code}
          pc={frame.pc}
          quiet={quiet}
          label="Code"
        />
      ) : null}
      <div className="av:flex av:min-w-0 av:flex-col av:gap-3">
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
              <ol className="av:m-0 av:flex av:list-none av:flex-wrap av:gap-1.5 av:p-0 av:font-mono av:text-sm">
                {frame.output.map((v, i) => (
                  <li
                    key={i}
                    className="av:rounded av:bg-subtle av:px-2 av:py-0.5 av:transition-opacity av:duration-200 av:starting:opacity-0"
                  >
                    {v}
                  </li>
                ))}
              </ol>
            ) : (
              <p className="av:m-0 av:text-sm av:text-muted">nothing yet</p>
            )}
          </Panel>
        ) : null}
      </div>
    </div>
  );
}
