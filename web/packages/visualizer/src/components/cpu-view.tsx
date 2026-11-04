import type { Exec, ExecFrame } from '../core/schema';
import { Listing, Panel, Value } from './exec-parts';

const PHASES = ['fetch', 'decode', 'execute', 'writeback'] as const;
const PHASE_LABEL: Record<string, string> = {
  fetch: 'Fetch',
  decode: 'Decode',
  execute: 'Execute',
  writeback: 'Write back'
};

/** The instruction cycle as a stepper: done phases filled, the current one
 * marked, the rest outlined. */
function PhaseStepper({ phase }: { phase?: string }) {
  const at = PHASES.indexOf((phase ?? '') as (typeof PHASES)[number]);
  return (
    <ol
      className="av:m-0 av:flex av:list-none av:items-center av:gap-1 av:p-0"
      aria-label="Instruction cycle"
    >
      {PHASES.map((p, i) => (
        <li
          key={p}
          aria-current={i === at ? 'step' : undefined}
          data-done={i < at}
          className="av:flex av:items-center av:gap-1"
        >
          <span
            className="av:rounded-full av:border av:border-border av:px-3 av:py-1 av:text-xs av:font-medium av:text-muted av:transition-[background-color,color,border-color] av:duration-200 av:data-[done=true]:bg-subtle av:data-[current=true]:border-accent av:data-[current=true]:bg-accent av:data-[current=true]:text-accent-fg"
            data-done={i < at}
            data-current={i === at}
          >
            {PHASE_LABEL[p]}
          </span>
          {i < PHASES.length - 1 ? (
            <span
              aria-hidden
              className="av:text-muted"
            >
              →
            </span>
          ) : null}
        </li>
      ))}
    </ol>
  );
}

/**
 * A tiny CPU at one phase of one instruction: the listing with the current
 * instruction marked, the registers and program counter, the flags, memory
 * and output. Whatever this phase changed is outlined and rolls in.
 */
export function CPUView({
  exec,
  frame,
  quiet = false
}: {
  exec: Exec;
  frame: ExecFrame;
  quiet?: boolean;
}) {
  const changed = new Set(frame.changed);
  return (
    <div className="av:flex av:flex-col av:gap-3">
      <PhaseStepper phase={frame.phase} />
      <div className="av:grid av:gap-3 av:md:grid-cols-[minmax(0,1fr)_minmax(0,1.2fr)]">
        <Listing
          code={exec.code}
          pc={frame.pc}
          quiet={quiet}
          label="Program"
        />
        <div className="av:flex av:min-w-0 av:flex-col av:gap-3">
          <Panel title="Registers">
            <div className="av:grid av:grid-cols-3 av:gap-1.5 av:sm:grid-cols-5">
              <Value
                name="pc"
                value={String(frame.pc)}
                changed={changed.has('pc')}
                quiet={quiet}
              />
              {frame.registers.map((v, i) => (
                <Value
                  key={i}
                  name={i === 7 ? 'r7 (sp)' : `r${i}`}
                  value={String(v)}
                  changed={changed.has(`r${i}`)}
                  quiet={quiet}
                />
              ))}
            </div>
            <p className="av:m-0 av:flex av:items-center av:gap-2 av:text-xs av:text-muted">
              Flags
              {['Z', 'N'].map((f) => {
                const on = frame.flags.includes(f);
                return (
                  <span
                    key={f}
                    data-on={on}
                    className="av:rounded av:border av:border-border av:px-1.5 av:font-mono av:data-[on=true]:border-accent av:data-[on=true]:text-accent"
                  >
                    {f} {on ? 'set' : 'clear'}
                  </span>
                );
              })}
            </p>
          </Panel>
          {frame.memory.length ? (
            <Panel title="Memory">
              <div className="av:grid av:grid-cols-4 av:gap-1.5 av:sm:grid-cols-8">
                {frame.memory.map((v, i) => (
                  <Value
                    key={i}
                    name={`[${i}]`}
                    value={String(v)}
                    changed={changed.has(`m${i}`)}
                    quiet={quiet}
                  />
                ))}
              </div>
            </Panel>
          ) : null}
          {frame.output.length ? (
            <Panel title={exec.outputLabel ?? 'Output'}>
              <p className="av:m-0 av:font-mono av:text-sm">{frame.output.join(' ')}</p>
            </Panel>
          ) : null}
        </div>
      </div>
    </div>
  );
}
