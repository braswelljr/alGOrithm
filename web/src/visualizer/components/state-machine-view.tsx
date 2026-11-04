import { useMemo } from 'react';
import {
  BaseEdge,
  EdgeLabelRenderer,
  Handle,
  MarkerType,
  Position,
  ReactFlow,
  ReactFlowProvider,
  type Edge,
  type EdgeProps,
  type Node,
  type NodeProps
} from '@xyflow/react';
import { layoutAutomaton } from '../core/replay/automaton';
import type { Automaton, AutomatonStep } from '../core/schema';

/** The radius of a state's circle, in pixels. */
const R = 28;
const EASE = 'cubic-bezier(0.23, 1, 0.32, 1)';

type StateData = { name: string; start: boolean; accept: boolean; active: boolean; quiet: boolean };
type StateNode = Node<StateData, 'state'>;
type MoveData = {
  label: string;
  active: boolean;
  /** The other direction exists too, so this edge bends to its own side. */
  paired: boolean;
  loop: boolean;
  quiet: boolean;
  /** Changes every step, so the token's animation restarts. */
  stepKey: number;
};
type MoveEdge = Edge<MoveData, 'move'>;

const centered = { left: '50%', top: '50%', opacity: 0, pointerEvents: 'none' as const };

/** One state: a circle, a double circle if it accepts, an arrow in if it
 * is the start, and a glow while the machine is in it. */
function StateView({ data }: NodeProps<StateNode>) {
  const label = `${data.name}${data.start ? ', start' : ''}${data.accept ? ', accepting' : ''}${data.active ? ', active' : ''}`;
  return (
    <div
      className="relative flex items-center justify-center"
      style={{ width: R * 2, height: R * 2 }}
      role="img"
      aria-label={`State ${label}`}
      data-active={data.active || undefined}
    >
      <Handle
        type="target"
        position={Position.Top}
        style={centered}
        isConnectable={false}
      />
      <Handle
        type="source"
        position={Position.Top}
        style={centered}
        isConnectable={false}
      />
      {data.start ? (
        <span
          aria-hidden
          className="text-muted absolute right-full mr-1 flex items-center text-[10px] font-medium tracking-wide uppercase"
        >
          start
          <svg
            width="18"
            height="10"
            viewBox="0 0 18 10"
            className="ml-0.5"
          >
            <path
              d="M0 5h14M10 1l4 4-4 4"
              fill="none"
              stroke="currentColor"
              strokeWidth="1.5"
            />
          </svg>
        </span>
      ) : null}
      <span
        aria-hidden
        className="border-fg bg-card data-[active=true]:border-accent data-[active=true]:bg-primary absolute inset-0 rounded-full border-2 shadow-sm motion-safe:transition-[background-color,border-color,box-shadow,transform] motion-safe:duration-200"
        data-active={data.active}
        style={{
          transitionTimingFunction: EASE,
          boxShadow: data.active
            ? '0 0 0 6px color-mix(in oklch, var(--av-accent) 25%, transparent)'
            : undefined,
          transform: data.active && !data.quiet ? 'scale(1.06)' : 'scale(1)'
        }}
      />
      {data.accept ? (
        <span
          aria-hidden
          className="border-fg data-[active=true]:border-accent-fg absolute inset-[5px] rounded-full border-2"
          data-active={data.active}
        />
      ) : null}
      <span
        className="text-fg data-[active=true]:text-accent-fg relative max-w-[3.25rem] truncate font-mono text-xs font-semibold"
        data-active={data.active}
        title={data.name}
      >
        {data.name}
      </span>
    </div>
  );
}

/** A transition: a curve between two state circles (a loop above a state
 * for a self-transition), its symbols as a label, and while it is being
 * taken a token that travels along it. */
function MoveView({
  id,
  sourceX,
  sourceY,
  targetX,
  targetY,
  data,
  markerEnd
}: EdgeProps<MoveEdge>) {
  const d = data!;
  let path: string;
  let lx: number;
  let ly: number;
  if (d.loop) {
    const x = sourceX;
    const y = sourceY - R;
    path = `M ${x - 12} ${y + 3} C ${x - 34} ${y - 52}, ${x + 34} ${y - 52}, ${x + 12} ${y + 3}`;
    lx = x;
    ly = y - 46;
  } else {
    const dx = targetX - sourceX;
    const dy = targetY - sourceY;
    const len = Math.hypot(dx, dy) || 1;
    const ux = dx / len;
    const uy = dy / len;
    // Bend paired edges apart; a lone edge bends a little, so labels on a
    // straight line of states do not sit on top of the line.
    const bend = d.paired ? 34 : 14;
    const mx = (sourceX + targetX) / 2 - uy * bend;
    const my = (sourceY + targetY) / 2 + ux * bend;
    const sx = sourceX + ux * R;
    const sy = sourceY + uy * R;
    const ex = targetX - ux * (R + 2);
    const ey = targetY - uy * (R + 2);
    path = `M ${sx} ${sy} Q ${mx} ${my} ${ex} ${ey}`;
    lx = (sx + 2 * mx + ex) / 4;
    ly = (sy + 2 * my + ey) / 4;
  }
  const stroke = d.active ? 'var(--av-accent)' : 'var(--av-muted)';
  return (
    <>
      <BaseEdge
        id={id}
        path={path}
        markerEnd={markerEnd}
        style={{
          stroke,
          strokeWidth: d.active ? 2.5 : 1.5,
          transition: d.quiet ? undefined : `stroke 200ms ${EASE}, stroke-width 200ms ${EASE}`
        }}
      />
      {d.active && !d.quiet ? (
        <circle
          key={d.stepKey}
          r="5"
          fill="var(--av-accent)"
          aria-hidden
        >
          <animateMotion
            dur="0.45s"
            path={path}
            fill="freeze"
            calcMode="spline"
            keyTimes="0;1"
            keySplines="0.23 1 0.32 1"
          />
        </circle>
      ) : null}
      <EdgeLabelRenderer>
        <span
          className="nodrag nopan border-border bg-card text-fg data-[active=true]:border-accent data-[active=true]:bg-primary data-[active=true]:text-accent-fg pointer-events-none absolute rounded-md border px-1.5 py-0.5 font-mono text-[11px] leading-none shadow-sm"
          data-active={d.active}
          style={{ transform: `translate(-50%, -50%) translate(${lx}px, ${ly}px)` }}
        >
          {d.label}
        </span>
      </EdgeLabelRenderer>
    </>
  );
}

const nodeTypes = { state: StateView };
const edgeTypes = { move: MoveView };

/** The input with a read head: read characters dimmed, the next one under
 * the head. The head slides, unless motion is reduced. */
export function InputStrip({
  input,
  head,
  quiet
}: {
  input: string;
  head: number;
  quiet?: boolean;
}) {
  const chars = [...input];
  if (chars.length < 1) return null;
  return (
    <div
      className="flex flex-col items-start gap-1"
      aria-label={`Input ${input}, ${Math.min(head, chars.length)} of ${chars.length} characters read`}
    >
      <span className="text-muted text-[11px] font-medium tracking-wide uppercase">Input</span>
      <div className="relative">
        <ol className="flex font-mono text-sm">
          {chars.map((c, i) => (
            <li
              key={i}
              className="border-border data-[next=true]:text-accent data-[read=true]:bg-subtle data-[read=true]:text-muted flex h-8 w-8 items-center justify-center border-y border-r first:rounded-l-md first:border-l last:rounded-r-md data-[next=true]:font-bold"
              data-read={i < head}
              data-next={i === head}
            >
              {c}
            </li>
          ))}
        </ol>
        <span
          aria-hidden
          className="border-b-accent absolute -bottom-2 left-0 h-0 w-0 border-x-[6px] border-b-[7px] border-x-transparent"
          style={{
            transform: `translateX(${Math.min(head, chars.length) * 32 + 10}px)`,
            transition: quiet ? undefined : `transform 200ms ${EASE}`
          }}
        />
      </div>
    </div>
  );
}

/**
 * An automaton drawn with React Flow: states as circles (double for
 * accepting, an arrow into the start), transitions as labeled curves and
 * loops. At each step the states the machine is in glow, the transitions it
 * just took light up, and a token travels along them; with reduced motion
 * nothing moves, the highlight alone shows the step.
 */
export function StateMachineView({
  machine,
  step,
  stepIndex = 0,
  quiet = false,
  label = 'The automaton'
}: {
  machine: Automaton;
  step?: AutomatonStep;
  stepIndex?: number;
  quiet?: boolean;
  label?: string;
}) {
  const pos = useMemo(() => layoutAutomaton(machine), [machine]);
  const active = useMemo(() => new Set(step?.active ?? []), [step]);
  const taken = useMemo(() => new Set(step?.edges ?? []), [step]);
  const nodes: StateNode[] = machine.states.map((s, i) => ({
    id: String(i),
    type: 'state',
    position: { x: (pos[i]?.x ?? 0) - R, y: (pos[i]?.y ?? 0) - R },
    data: { name: s.name, start: s.start, accept: s.accept, active: active.has(i), quiet },
    draggable: false,
    connectable: false,
    selectable: false
  }));
  const pairs = new Set(machine.edges.map((e) => `${e.from}-${e.to}`));
  const edges: MoveEdge[] = machine.edges.map((e, k) => ({
    id: `m${k}`,
    source: String(e.from),
    target: String(e.to),
    type: 'move',
    markerEnd: {
      type: MarkerType.ArrowClosed,
      width: 16,
      height: 16,
      color: taken.has(k) ? 'var(--av-accent)' : 'var(--av-muted)'
    },
    data: {
      label: e.label,
      active: taken.has(k),
      paired: e.from !== e.to && pairs.has(`${e.to}-${e.from}`),
      loop: e.from === e.to,
      quiet,
      stepKey: stepIndex
    }
  }));
  return (
    <figure
      className="m-0 flex flex-col gap-3"
      aria-label={label}
    >
      <div
        data-flow-box
        className="border-border bg-bg h-[22rem] w-full overflow-hidden rounded-[var(--av-radius)] border"
      >
        <ReactFlowProvider>
          <ReactFlow
            nodes={nodes}
            edges={edges}
            nodeTypes={nodeTypes}
            edgeTypes={edgeTypes}
            fitView
            fitViewOptions={{ padding: 0.25 }}
            nodesDraggable={false}
            nodesConnectable={false}
            elementsSelectable={false}
            panOnScroll={false}
            zoomOnScroll={false}
            proOptions={{ hideAttribution: true }}
          />
        </ReactFlowProvider>
      </div>
      <figcaption className="sr-only">
        {machine.states.length} states:{' '}
        {machine.states
          .map((s) => `${s.name}${s.start ? ' (start)' : ''}${s.accept ? ' (accepting)' : ''}`)
          .join(', ')}
        . Transitions:{' '}
        {machine.edges
          .map(
            (e) => `${machine.states[e.from]?.name} to ${machine.states[e.to]?.name} on ${e.label}`
          )
          .join('; ')}
        .
      </figcaption>
    </figure>
  );
}
