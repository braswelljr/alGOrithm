/**
 * The frame builder for a system design (Output.architecture): one step per
 * frame of the simulation, drawn by ArchitectureView, with the frame's note
 * and its requests as words and the running request totals as counts.
 */
import { architectureAt, describeArchitecture } from '../core/replay/architecture';
import type { RunResult } from '../core/schema';
import { ArchitectureLegend } from './architecture-legend';
import { ArchitectureView } from './lazy-flow';
import type { Frame } from './run-frames';

export function architectureFrame(r: RunResult): Frame {
  const a = r.output.architecture!;
  const total = a.frames.length;
  return {
    total,
    unit: 'Step',
    render(i, quiet) {
      const s = architectureAt(a, i);
      const counts: Record<string, number> = {};
      for (const [k, v] of Object.entries(s.totals)) if (v > 0) counts[k] = v;
      return {
        view: (
          <ArchitectureView
            architecture={a}
            state={s}
            quiet={quiet}
          />
        ),
        words: s.frame
          ? `Step ${s.step} of ${total}: ${describeArchitecture(a, s)}`
          : describeArchitecture(a, s),
        counts
      };
    },
    legend: <ArchitectureLegend />
  };
}
