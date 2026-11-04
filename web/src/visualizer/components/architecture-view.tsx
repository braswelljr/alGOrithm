import { useMemo, type ReactNode } from 'react';
import {
  BaseEdge,
  EdgeLabelRenderer,
  getBezierPath,
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
import {
  layoutArchitecture,
  STATUS_WORDS,
  type ArchitectureState
} from '../core/replay/architecture';
import type { Architecture, ComponentKind, RequestStatus } from '../core/schema';
import { ArchitectureLegend, StatusMark } from './architecture-legend';

const EASE = 'cubic-bezier(0.23, 1, 0.32, 1)';
const W = 168;
const H = 78;

const svg = {
  width: 18,
  height: 18,
  viewBox: '0 0 24 24',
  fill: 'none',
  stroke: 'currentColor',
  strokeWidth: 2,
  strokeLinecap: 'round' as const,
  strokeLinejoin: 'round' as const,
  'aria-hidden': true
};

/** One small line icon per kind of component (drawn after lucide's). */
const KIND_ICON: Record<ComponentKind, ReactNode> = {
  client: (
    <svg {...svg}>
      <rect
        x="3"
        y="4"
        width="18"
        height="12"
        rx="2"
      />
      <path d="M8 20h8M12 16v4" />
    </svg>
  ),
  lb: (
    <svg {...svg}>
      <path d="M12 3v6M12 9l-7 6M12 9l7 6M12 9v6" />
      <circle
        cx="5"
        cy="18"
        r="2"
      />
      <circle
        cx="12"
        cy="18"
        r="2"
      />
      <circle
        cx="19"
        cy="18"
        r="2"
      />
    </svg>
  ),
  service: (
    <svg {...svg}>
      <rect
        x="3"
        y="3"
        width="18"
        height="7"
        rx="1.5"
      />
      <rect
        x="3"
        y="14"
        width="18"
        height="7"
        rx="1.5"
      />
      <path d="M7 6.5h.01M7 17.5h.01" />
    </svg>
  ),
  cache: (
    <svg {...svg}>
      <path d="M13 2 4 14h7l-1 8 9-12h-7z" />
    </svg>
  ),
  db: (
    <svg {...svg}>
      <ellipse
        cx="12"
        cy="5"
        rx="8"
        ry="3"
      />
      <path d="M4 5v14c0 1.7 3.6 3 8 3s8-1.3 8-3V5M4 12c0 1.7 3.6 3 8 3s8-1.3 8-3" />
    </svg>
  ),
  queue: (
    <svg {...svg}>
      <rect
        x="2"
        y="7"
        width="5"
        height="10"
        rx="1"
      />
      <rect
        x="9.5"
        y="7"
        width="5"
        height="10"
        rx="1"
      />
      <rect
        x="17"
        y="7"
        width="5"
        height="10"
        rx="1"
      />
    </svg>
  ),
  cdn: (
    <svg {...svg}>
      <circle
        cx="12"
        cy="12"
        r="9"
      />
      <path d="M3 12h18M12 3a14 14 0 0 1 0 18M12 3a14 14 0 0 0 0 18" />
    </svg>
  ),
  blob: (
    <svg {...svg}>
      <path d="M21 8v13H3V8M1 3h22v5H1zM10 12h4" />
    </svg>
  ),
  search: (
    <svg {...svg}>
      <circle
        cx="11"
        cy="11"
        r="7"
      />
      <path d="m20 20-4-4" />
    </svg>
  ),
  worker: (
    <svg {...svg}>
      <circle
        cx="12"
        cy="12"
        r="3"
      />
      <path d="M12 2v3M12 19v3M4.9 4.9 7 7M17 17l2.1 2.1M2 12h3M19 12h3M4.9 19.1 7 17M17 7l2.1-2.1" />
    </svg>
  )
};

const KIND_WORDS: Record<ComponentKind, string> = {
  client: 'client',
  lb: 'load balancer',
  service: 'service',
  cache: 'cache',
  db: 'database',
  queue: 'queue',
  cdn: 'CDN',
  blob: 'blob store',
  search: 'search index',
  worker: 'worker'
};

type BoxData = {
  label: string;
  kind: ComponentKind;
  notes?: string;
  load?: { load: number; queue: number; hitRate: number };
  busy: boolean;
  quiet: boolean;
};
type BoxNode = Node<BoxData, 'box'>;

const hidden = { opacity: 0, pointerEvents: 'none' as const };

/** One component: its kind's icon, its label and notes, and its latest
 * metrics as small badges. It lifts while requests touch it. */
function BoxView({ data }: NodeProps<BoxNode>) {
  const m = data.load;
  const metrics = m
    ? [
        `load ${m.load}`,
        ...(m.queue > 0 ? [`queue ${m.queue}`] : []),
        ...(m.hitRate >= 0 ? [`hit ${m.hitRate}%`] : [])
      ]
    : [];
  return (
    <div
      className="border-border bg-card text-fg data-[busy=true]:border-accent flex flex-col gap-1 rounded-[var(--av-radius)] border px-2.5 py-2 shadow-sm motion-safe:transition-[border-color,transform,box-shadow] motion-safe:duration-200"
      style={{
        width: W,
        minHeight: H,
        transitionTimingFunction: EASE,
        transform: data.busy && !data.quiet ? 'translateY(-2px)' : undefined
      }}
      data-busy={data.busy}
      role="img"
      aria-label={`${data.label}, ${KIND_WORDS[data.kind]}${data.notes ? `, ${data.notes}` : ''}${metrics.length ? `, ${metrics.join(', ')}` : ''}`}
    >
      <Handle
        type="target"
        position={Position.Left}
        style={hidden}
        isConnectable={false}
      />
      <Handle
        type="source"
        position={Position.Right}
        style={hidden}
        isConnectable={false}
      />
      <span className="flex items-center gap-1.5">
        <span className="text-accent">{KIND_ICON[data.kind]}</span>
        <span className="truncate text-xs font-semibold">{data.label}</span>
      </span>
      <span className="text-muted text-[10px] tracking-wide uppercase">
        {KIND_WORDS[data.kind]}
      </span>
      {data.notes ? (
        <span className="text-muted line-clamp-2 text-[11px] leading-snug">{data.notes}</span>
      ) : null}
      {metrics.length ? (
        <span className="flex flex-wrap gap-1">
          {metrics.map((t) => (
            <span
              key={t}
              className="bg-subtle text-fg rounded px-1 font-mono text-[10px]"
            >
              {t}
            </span>
          ))}
        </span>
      ) : null}
    </div>
  );
}

type WireData = {
  label?: string;
  requests: { id: number; status: RequestStatus }[];
  quiet: boolean;
  stepKey: number;
};
type WireEdge = Edge<WireData, 'wire'>;

/** A link: a curve with its label, and each request on it this step as a
 * mark that travels from one end to the other (staggered when several share
 * the link). With reduced motion the marks sit still at the middle. */
function WireView({
  id,
  sourceX,
  sourceY,
  targetX,
  targetY,
  data,
  markerEnd
}: EdgeProps<WireEdge>) {
  const d = data!;
  const [path, lx, ly] = getBezierPath({
    sourceX,
    sourceY,
    targetX,
    targetY,
    sourcePosition: Position.Right,
    targetPosition: Position.Left
  });
  const active = d.requests.length > 0;
  return (
    <>
      <BaseEdge
        id={id}
        path={path}
        markerEnd={markerEnd}
        style={{
          stroke: active ? 'var(--av-accent)' : 'var(--av-border, var(--av-muted))',
          strokeWidth: active ? 2.25 : 1.25,
          transition: d.quiet ? undefined : `stroke 200ms ${EASE}`
        }}
      />
      {d.requests.slice(0, 6).map((q, k) =>
        d.quiet ? (
          <g
            key={`${d.stepKey}-${k}`}
            transform={`translate(${lx + k * 14 - ((Math.min(d.requests.length, 6) - 1) * 14) / 2}, ${ly + 14})`}
          >
            <StatusMark status={q.status} />
          </g>
        ) : (
          <g key={`${d.stepKey}-${k}`}>
            <StatusMark status={q.status} />
            <animateMotion
              dur="0.7s"
              begin={`${k * 0.12}s`}
              path={path}
              fill="freeze"
              calcMode="spline"
              keyTimes="0;1"
              keySplines="0.23 1 0.32 1"
            />
          </g>
        )
      )}
      {d.label ? (
        <EdgeLabelRenderer>
          <span
            className="nodrag nopan bg-bg text-muted pointer-events-none absolute rounded px-1 text-[10px] leading-none"
            style={{ transform: `translate(-50%, -150%) translate(${lx}px, ${ly}px)` }}
          >
            {d.label}
          </span>
        </EdgeLabelRenderer>
      ) : null}
    </>
  );
}

const nodeTypes = { box: BoxView };
const edgeTypes = { wire: WireView };

/**
 * A system design drawn with React Flow: components as boxes with an icon,
 * a label and their live metrics, links as curves, and each step's requests
 * as marks traveling along the links, shaped by status (hit, miss, queued,
 * error, dropped). With reduced motion nothing travels: the marks rest on
 * their links and the step is listed in words below.
 */
export function ArchitectureView({
  architecture,
  state,
  quiet = false,
  label = 'The system design'
}: {
  architecture: Architecture;
  state: ArchitectureState;
  quiet?: boolean;
  label?: string;
}) {
  const pos = useMemo(() => layoutArchitecture(architecture), [architecture]);
  const touched = new Set<string>();
  const onLink = new Map<number, { id: number; status: RequestStatus }[]>();
  for (const q of state.frame?.requests ?? []) {
    const l = architecture.links[q.link];
    if (!l) continue;
    touched.add(l.from);
    touched.add(l.to);
    onLink.set(q.link, [...(onLink.get(q.link) ?? []), { id: q.id, status: q.status }]);
  }
  const nodes: BoxNode[] = architecture.components.map((c) => ({
    id: c.id,
    type: 'box',
    position: pos.get(c.id) ?? { x: 0, y: 0 },
    data: {
      label: c.label,
      kind: c.kind,
      notes: c.notes,
      load: state.loads.get(c.id),
      busy: touched.has(c.id),
      quiet
    },
    draggable: false,
    connectable: false,
    selectable: false
  }));
  const edges: WireEdge[] = architecture.links.map((l, k) => ({
    id: `l${k}`,
    source: l.from,
    target: l.to,
    type: 'wire',
    markerEnd: { type: MarkerType.ArrowClosed, width: 14, height: 14, color: 'var(--av-muted)' },
    data: { label: l.label, requests: onLink.get(k) ?? [], quiet, stepKey: state.step }
  }));
  const requests = state.frame?.requests ?? [];
  const name = (id: string) => architecture.components.find((c) => c.id === id)?.label ?? id;
  return (
    <figure
      className="m-0 flex flex-col gap-3"
      aria-label={label}
    >
      <div
        data-flow-box
        className="border-border bg-bg h-[26rem] w-full overflow-hidden rounded-[var(--av-radius)] border"
      >
        <ReactFlowProvider>
          <ReactFlow
            nodes={nodes}
            edges={edges}
            nodeTypes={nodeTypes}
            edgeTypes={edgeTypes}
            fitView
            fitViewOptions={{ padding: 0.15 }}
            nodesDraggable={false}
            nodesConnectable={false}
            elementsSelectable={false}
            panOnScroll={false}
            zoomOnScroll={false}
            proOptions={{ hideAttribution: true }}
          />
        </ReactFlowProvider>
      </div>
      <ArchitectureLegend />
      {quiet && requests.length ? (
        <ol
          className="text-fg m-0 flex flex-col gap-1 pl-5 text-sm"
          aria-label="Requests this step"
        >
          {requests.map((q, k) => {
            const l = architecture.links[q.link]!;
            return (
              <li key={k}>
                request {q.id}: {name(l.from)} to {name(l.to)}, {STATUS_WORDS[q.status]}
              </li>
            );
          })}
        </ol>
      ) : null}
      <figcaption className="sr-only">
        {architecture.components.length} components:{' '}
        {architecture.components.map((c) => `${c.label} (${KIND_WORDS[c.kind]})`).join(', ')}.
        Links:{' '}
        {architecture.links
          .map((l) => `${name(l.from)} to ${name(l.to)}${l.label ? ` (${l.label})` : ''}`)
          .join('; ')}
        .
      </figcaption>
    </figure>
  );
}
