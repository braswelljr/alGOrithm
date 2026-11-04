import { useMemo } from 'react';
import {
  BaseEdge,
  getSmoothStepPath,
  Handle,
  Position,
  ReactFlow,
  ReactFlowProvider,
  type Edge,
  type EdgeProps,
  type Node,
  type NodeProps
} from '@xyflow/react';
import { cn } from '@/lib/utils';
import type { CircuitState } from '../core/replay/circuit';
import type { Circuit, Level } from '../core/schema';
import { useReducedMotion } from '../hooks/use-reduced-motion';

const W = 60;
const H = 40;
const COL = 130;
const ROW = 72;

const levelText = (l: Level) => (l === 1 ? '1' : l === 0 ? '0' : 'X');
const levelName = (l: Level) => (l === 1 ? 'high' : l === 0 ? 'low' : 'unknown');
const strokeFor = (l: Level) =>
  l === 1 ? 'stroke-high' : l === 0 ? 'stroke-low' : 'stroke-unknown';

/** The body of each gate, in a 60 x 40 box, inputs on the left. */
const bodies: Record<string, string> = {
  and: 'M8 4 H30 A16 16 0 0 1 30 36 H8 Z',
  or: 'M6 4 Q30 4 50 20 Q30 36 6 36 Q16 20 6 4 Z',
  not: 'M8 6 L44 20 L8 34 Z',
  buf: 'M8 6 L48 20 L8 34 Z'
};
const base: Record<string, keyof typeof bodies> = {
  and: 'and',
  nand: 'and',
  or: 'or',
  nor: 'or',
  xor: 'or',
  xnor: 'or',
  not: 'not',
  buf: 'buf',
  output: 'buf'
};
const bubbled = new Set(['not', 'nand', 'nor', 'xnor']);
const doubled = new Set(['xor', 'xnor']);
/** Where the output leaves each shape. */
const tipX = (kind: string) =>
  bubbled.has(kind)
    ? base[kind] === 'and'
      ? 54
      : base[kind] === 'or'
        ? 58
        : 52
    : base[kind] === 'and'
      ? 46
      : 50;

function GateSymbol({ kind, level }: { kind: string; level: Level }) {
  const b = bodies[base[kind] ?? 'buf'];
  const bubbleX = base[kind] === 'and' ? 50 : base[kind] === 'or' ? 54 : 48;
  return (
    <svg
      width={W}
      height={H}
      viewBox={`0 0 ${W} ${H}`}
      aria-hidden="true"
      className="overflow-visible"
    >
      {doubled.has(kind) ? (
        <path
          d="M1 4 Q11 20 1 36"
          className="stroke-fg fill-none"
          strokeWidth={2}
        />
      ) : null}
      <path
        d={b}
        strokeWidth={2.25}
        className={cn('fill-card transition-[stroke] duration-200', strokeFor(level))}
      />
      {bubbled.has(kind) ? (
        <circle
          cx={bubbleX}
          cy={20}
          r={4}
          strokeWidth={2}
          className={cn('fill-card', strokeFor(level))}
        />
      ) : null}
    </svg>
  );
}

type GateData = {
  name: string;
  kind: string;
  inputs: number;
  level: Level;
  changed: boolean;
  onToggle?: () => void;
};
type GateNode = Node<GateData, 'gate'>;

function inputHandles(n: number) {
  return Array.from({ length: n }, (_, k) => (
    <Handle
      key={k}
      id={`in${k}`}
      type="target"
      position={Position.Left}
      isConnectable={false}
      style={{ top: `${((k + 1) / (n + 1)) * 100}%` }}
      className="bg-muted! size-1.5! min-w-0! border-0!"
    />
  ));
}

const out = (
  <Handle
    type="source"
    position={Position.Right}
    isConnectable={false}
    className="bg-muted! size-1.5! min-w-0! border-0!"
  />
);

/** An input: a switch the reader can flip, labeled with its name and value. */
function InputNode({ data }: { data: GateData }) {
  const on = data.level === 1;
  return (
    <div className="flex items-center gap-2">
      <span className="text-fg font-mono text-xs font-semibold">{data.name}</span>
      <button
        type="button"
        role="switch"
        aria-checked={on}
        aria-label={`Input ${data.name}, ${levelName(data.level)}${data.onToggle ? ', press to flip' : ''}`}
        disabled={!data.onToggle}
        onClick={data.onToggle}
        className={cn(
          'nodrag nopan border-border relative h-6 w-11 rounded-full border transition-[background-color,transform] duration-150 active:scale-[0.97] disabled:cursor-default',
          on ? 'bg-high' : 'bg-subtle',
          data.onToggle ? 'cursor-pointer' : ''
        )}
      >
        <span
          className={cn(
            'bg-card text-fg ease-out-strong absolute top-0.5 left-0.5 grid size-[1.125rem] place-items-center rounded-full font-mono text-[0.625rem] font-bold shadow-sm transition-transform duration-200',
            on ? 'translate-x-5' : 'translate-x-0'
          )}
        >
          {levelText(data.level)}
        </span>
      </button>
      {out}
    </div>
  );
}

/** An output: an LED that lights when high, with its name and value as text. */
function OutputNode({ data }: { data: GateData }) {
  const lit = data.level === 1;
  return (
    <div className="flex items-center gap-2">
      {inputHandles(1)}
      <span
        role="img"
        aria-label={`Output ${data.name}, ${levelName(data.level)}`}
        className={cn(
          'grid size-7 place-items-center rounded-full border-2 font-mono text-xs font-bold transition-[background-color,box-shadow,border-color] duration-200',
          lit
            ? 'border-high bg-high text-card shadow-[0_0_12px_var(--av-high)]'
            : data.level === 0
              ? 'border-low bg-subtle text-muted'
              : 'border-unknown bg-subtle text-fg border-dashed'
        )}
      >
        {levelText(data.level)}
      </span>
      <span className="text-fg font-mono text-xs font-semibold">{data.name}</span>
    </div>
  );
}

function ConstNode({ data }: { data: GateData }) {
  return (
    <div className="border-border bg-subtle text-muted grid h-6 w-8 place-items-center rounded border font-mono text-xs">
      {data.kind === 'const1' ? '1' : '0'}
      {out}
    </div>
  );
}

function GateView({ data }: NodeProps<GateNode>) {
  if (data.kind === 'input') return <InputNode data={data} />;
  if (data.kind === 'output') return <OutputNode data={data} />;
  if (data.kind === 'const0' || data.kind === 'const1') return <ConstNode data={data} />;
  return (
    <div
      className={cn(
        'ease-out-strong relative transition-transform duration-200',
        data.changed ? 'scale-[1.06]' : 'scale-100'
      )}
      title={`${data.name}: ${data.kind.toUpperCase()}, output ${levelName(data.level)}`}
    >
      {inputHandles(data.inputs)}
      <GateSymbol
        kind={data.kind}
        level={data.level}
      />
      <span className="text-muted pointer-events-none absolute inset-x-0 -bottom-4 text-center font-mono text-[0.625rem]">
        {data.kind.toUpperCase()}
      </span>
      <Handle
        type="source"
        position={Position.Right}
        isConnectable={false}
        style={{ left: tipX(data.kind) }}
        className="bg-muted! size-1.5! min-w-0! border-0!"
      />
    </div>
  );
}

const nodeTypes = { gate: GateView };

type WireData = { level: Level; fresh: boolean; motion: boolean };
type WireEdge = Edge<WireData, 'wire'>;

/** A wire, colored by its level; a high wire carries a moving pulse. */
function WireView(props: EdgeProps<WireEdge>) {
  const [path] = getSmoothStepPath({ ...props, borderRadius: 10 });
  const l = props.data?.level ?? -1;
  return (
    <>
      <BaseEdge
        id={props.id}
        path={path}
        className={cn('transition-[stroke] duration-200', strokeFor(l))}
        style={{ strokeWidth: l === 1 ? 3 : 2, strokeDasharray: l === -1 ? '4 4' : undefined }}
      />
      {l === 1 && props.data?.motion ? (
        <path
          d={path}
          fill="none"
          className="animate-wire-pulse stroke-card pointer-events-none"
          style={{
            strokeWidth: 1.5,
            strokeDasharray: '2 14',
            opacity: props.data.fresh ? 0.95 : 0.6
          }}
        />
      ) : null}
    </>
  );
}

const edgeTypes = { wire: WireView };

/**
 * A gate-level circuit with React Flow: gates in columns by depth, drawn as
 * their standard symbols; wires green when high, gray when low, dashed amber
 * while unknown, each also told by the 1, 0 or X on the gate it leaves. High
 * wires carry a moving pulse unless the reader asks for reduced motion. Pass
 * onToggle to let the reader flip an input.
 */
export function CircuitView({
  circuit,
  state,
  onToggle,
  quiet
}: {
  circuit: Circuit;
  state: CircuitState;
  onToggle?: (gate: number) => void;
  quiet?: boolean;
}) {
  const reduced = useReducedMotion();
  const motion = !reduced && !quiet;
  const { nodes, edges } = useMemo(() => {
    const count = new Map<number, number>();
    for (const g of circuit.gates) count.set(g.layer, (count.get(g.layer) ?? 0) + 1);
    const nodes: GateNode[] = circuit.gates.map((g, i) => {
      const n = count.get(g.layer) ?? 1;
      return {
        id: String(i),
        type: 'gate',
        position: { x: g.layer * COL, y: (g.row - (n - 1) / 2) * ROW },
        data: {
          name: g.name,
          kind: g.kind,
          inputs: g.in.length,
          level: state.levels[i] ?? -1,
          changed: state.changed.has(i) && state.delay > 0,
          onToggle: g.kind === 'input' && onToggle ? () => onToggle(i) : undefined
        },
        draggable: false,
        connectable: false,
        selectable: false
      };
    });
    const edges: WireEdge[] = circuit.gates.flatMap((g, i) =>
      g.in.map((from, k) => ({
        id: `w${i}.${k}`,
        source: String(from),
        target: String(i),
        targetHandle: `in${k}`,
        type: 'wire' as const,
        data: { level: state.levels[from] ?? -1, fresh: state.changed.has(from), motion },
        selectable: false,
        focusable: false
      }))
    );
    return { nodes, edges };
  }, [circuit, state, onToggle, motion]);

  const outputs = circuit.gates
    .map((g, i) => (g.kind === 'output' ? `${g.name} is ${levelName(state.levels[i] ?? -1)}` : ''))
    .filter(Boolean)
    .join(', ');
  return (
    <figure
      className="m-0"
      aria-label={`The circuit${circuit.name ? `: ${circuit.name}` : ''}`}
    >
      <div
        data-flow-box
        className="av-flow h-[min(28rem,60vh)]"
      >
        <ReactFlowProvider>
          <ReactFlow
            nodes={nodes}
            edges={edges}
            nodeTypes={nodeTypes}
            edgeTypes={edgeTypes}
            nodesDraggable={false}
            nodesConnectable={false}
            elementsSelectable={false}
            nodesFocusable={false}
            edgesFocusable={false}
            zoomOnScroll={false}
            preventScrolling={false}
            fitView
            fitViewOptions={{ padding: 0.15 }}
            minZoom={0.2}
            maxZoom={2}
            colorMode="system"
            proOptions={{ hideAttribution: true }}
          />
        </ReactFlowProvider>
      </div>
      <figcaption className="text-muted mt-2 flex flex-wrap items-center gap-x-4 gap-y-1 text-xs">
        <span className="inline-flex items-center gap-1.5">
          <span className="bg-high inline-block h-0.5 w-5 rounded" /> high (1)
        </span>
        <span className="inline-flex items-center gap-1.5">
          <span className="bg-low inline-block h-0.5 w-5 rounded" /> low (0)
        </span>
        <span className="inline-flex items-center gap-1.5">
          <span className="border-unknown inline-block h-0.5 w-5 border-t-2 border-dashed" />{' '}
          unknown (X)
        </span>
        <span className="av-sr-only">{outputs}.</span>
      </figcaption>
    </figure>
  );
}
