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
      className="m-0 flex list-none items-center gap-1 p-0"
      aria-label="Instruction cycle"
    >
      {PHASES.map((p, i) => (
        <li
          key={p}
          aria-current={i === at ? 'step' : undefined}
          data-done={i < at}
          className="flex items-center gap-1"
        >
          <span
            className="border-border text-muted data-[current=true]:border-accent data-[current=true]:bg-primary data-[current=true]:text-accent-fg data-[done=true]:bg-subtle rounded-full border px-3 py-1 text-xs font-medium transition-[background-color,color,border-color] duration-200"
            data-done={i < at}
            data-current={i === at}
          >
            {PHASE_LABEL[p]}
          </span>
          {i < PHASES.length - 1 ? (
            <span
              aria-hidden
              className="text-muted"
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
    <div className="flex flex-col gap-3">
      <PhaseStepper phase={frame.phase} />
      <div className="grid gap-3 md:grid-cols-[minmax(0,1fr)_minmax(0,1.2fr)]">
        <Listing
          code={exec.code}
          pc={frame.pc}
          quiet={quiet}
          label="Program"
        />
        <div className="flex min-w-0 flex-col gap-3">
          <Panel title="Registers">
            <div className="grid grid-cols-3 gap-1.5 sm:grid-cols-5">
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
            <p className="text-muted m-0 flex items-center gap-2 text-xs">
              Flags
              {['Z', 'N'].map((f) => {
                const on = frame.flags.includes(f);
                return (
                  <span
                    key={f}
                    data-on={on}
                    className="border-border data-[on=true]:border-accent data-[on=true]:text-accent rounded border px-1.5 font-mono"
                  >
                    {f} {on ? 'set' : 'clear'}
                  </span>
                );
              })}
            </p>
          </Panel>
          {frame.memory.length ? (
            <Panel title="Memory">
              <div className="grid grid-cols-4 gap-1.5 sm:grid-cols-8">
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
              <p className="m-0 font-mono text-sm">{frame.output.join(' ')}</p>
            </Panel>
          ) : null}
        </div>
      </div>
    </div>
  );
}
