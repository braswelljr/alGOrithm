import { lazy, Suspense, type ComponentProps } from 'react';
import { FlowFrame } from './flow-frame';

// The views drawn with React Flow, loaded on first use. React Flow is the
// package's largest dependency, and most runs (bars, cells, tables, text,
// timelines) never draw a graph, so a page pays for it only when it shows one.
// Every other module imports these wrappers, never the view files directly.

type GraphFlowProps = ComponentProps<typeof import('./graph-flow').GraphFlow>;
type StructureFlowProps = ComponentProps<typeof import('./structure-flow').StructureFlow>;
type StateMachineProps = ComponentProps<typeof import('./state-machine-view').StateMachineView>;
type InputStripProps = ComponentProps<typeof import('./state-machine-view').InputStrip>;
type CircuitProps = ComponentProps<typeof import('./circuit-view').CircuitView>;
type ArchitectureProps = ComponentProps<typeof import('./architecture-view').ArchitectureView>;

const LazyGraphFlow = lazy(() => import('./graph-flow').then((m) => ({ default: m.GraphFlow })));
const LazyStructureFlow = lazy(() =>
  import('./structure-flow').then((m) => ({ default: m.StructureFlow }))
);
const LazyStateMachine = lazy(() =>
  import('./state-machine-view').then((m) => ({ default: m.StateMachineView }))
);
const LazyInputStrip = lazy(() =>
  import('./state-machine-view').then((m) => ({ default: m.InputStrip }))
);
const LazyCircuit = lazy(() => import('./circuit-view').then((m) => ({ default: m.CircuitView })));
const LazyArchitecture = lazy(() =>
  import('./architecture-view').then((m) => ({ default: m.ArchitectureView }))
);

/** Holds the view's space while React Flow loads, so nothing shifts. */
function FlowSkeleton() {
  return (
    <div
      role="status"
      aria-label="Loading the view"
      className="bg-subtle min-h-80 w-full animate-pulse rounded-lg motion-reduce:animate-none"
    />
  );
}

export function GraphFlow(props: GraphFlowProps) {
  return (
    <FlowFrame>
      <Suspense fallback={<FlowSkeleton />}>
        <LazyGraphFlow {...props} />
      </Suspense>
    </FlowFrame>
  );
}

export function StructureFlow(props: StructureFlowProps) {
  return (
    <FlowFrame>
      <Suspense fallback={<FlowSkeleton />}>
        <LazyStructureFlow {...props} />
      </Suspense>
    </FlowFrame>
  );
}

export function StateMachineView(props: StateMachineProps) {
  return (
    <FlowFrame>
      <Suspense fallback={<FlowSkeleton />}>
        <LazyStateMachine {...props} />
      </Suspense>
    </FlowFrame>
  );
}

export function InputStrip(props: InputStripProps) {
  return (
    <Suspense fallback={null}>
      <LazyInputStrip {...props} />
    </Suspense>
  );
}

export function CircuitView(props: CircuitProps) {
  return (
    <FlowFrame>
      <Suspense fallback={<FlowSkeleton />}>
        <LazyCircuit {...props} />
      </Suspense>
    </FlowFrame>
  );
}

export function ArchitectureView(props: ArchitectureProps) {
  return (
    <FlowFrame>
      <Suspense fallback={<FlowSkeleton />}>
        <LazyArchitecture {...props} />
      </Suspense>
    </FlowFrame>
  );
}
