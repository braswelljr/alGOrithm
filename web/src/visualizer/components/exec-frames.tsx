/**
 * The frame builder for a running machine (Output.exec): one step per
 * frame, drawn by the view the run names, with the frame's note as words.
 */
import type { Exec, RunResult } from '../core/schema';
import { CPUView } from './cpu-view';
import { HeapView } from './heap-view';
import { PipelineView } from './pipeline-view';
import type { Frame } from './run-frames';
import { TokensView } from './tokens-view';
import { VMView } from './vm-view';

const UNIT: Record<Exec['view'], string> = {
  cpu: 'Phase',
  pipeline: 'Cycle',
  vm: 'Step',
  tokens: 'Token',
  heap: 'Step'
};

/** The step index i shows frame i - 1; step 0 is the machine before it
 * starts, drawn as the first frame without its note. */
export function execFrame(r: RunResult): Frame {
  const exec = r.output.exec!;
  const frames = exec.frames;
  const total = frames.length;
  return {
    total,
    unit: UNIT[exec.view],
    render(i, quiet) {
      const k = Math.max(0, Math.min(i, total) - 1);
      const frame = frames[k];
      if (!frame) return { view: null, words: '', counts: {} };
      const previous = k > 0 ? frames[k - 1] : undefined;
      const started = i > 0;
      const words = started
        ? `${UNIT[exec.view]} ${Math.min(i, total)} of ${total}: ${frame.note}`
        : 'Ready to start.';
      let view;
      switch (exec.view) {
        case 'cpu':
          view = (
            <CPUView
              exec={exec}
              frame={started ? frame : { ...frame, changed: [] }}
              quiet={quiet}
            />
          );
          break;
        case 'pipeline':
          view = (
            <PipelineView
              exec={exec}
              cycle={started ? frame.cycle : -1}
              quiet={quiet}
            />
          );
          break;
        case 'tokens':
          view = (
            <TokensView
              exec={exec}
              focus={started ? frame.focus : -1}
              quiet={quiet}
            />
          );
          break;
        case 'heap':
          view = (
            <HeapView
              exec={exec}
              frame={frame}
            />
          );
          break;
        case 'vm':
          view = (
            <VMView
              exec={exec}
              frame={frame}
              previous={previous}
              quiet={quiet}
            />
          );
          break;
      }
      return { view, words, counts: {} };
    }
  };
}
