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
  l === 1 ? 'av:stroke-high' : l === 0 ? 'av:stroke-low' : 'av:stroke-unknown';

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
      className="av:overflow-visible"
    >
      {doubled.has(kind) ? (
        <path
          d="M1 4 Q11 20 1 36"
          className="av:fill-none av:stroke-fg"
          strokeWidth={2}
        />
      ) : null}
      <path
        d={b}
        strokeWidth={2.25}
        className={`av:fill-card av:transition-[stroke] av:duration-200 ${strokeFor(level)}`}
      />
      {bubbled.has(kind) ? (
        <circle
          cx={bubbleX}
          cy={20}
          r={4}
          strokeWidth={2}
          className={`av:fill-card ${strokeFor(level)}`}
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
      className="av:size-1.5! av:min-w-0! av:border-0! av:bg-muted!"
    />
  ));
}

const out = (
  <Handle
    type="source"
    position={Position.Right}
    isConnectable={false}
    className="av:size-1.5! av:min-w-0! av:border-0! av:bg-muted!"
  />
);

/** An input: a switch the reader can flip, labeled with its name and value. */
function InputNode({ data }: { data: GateData }) {
  const on = data.level === 1;
  return (
    <div className="av:flex av:items-center av:gap-2">
      <span className="av:font-mono av:text-xs av:font-semibold av:text-fg">{data.name}</span>
      <button
        type="button"
        role="switch"
        aria-checked={on}
        aria-label={`Input ${data.name}, ${levelName(data.level)}${data.onToggle ? ', press to flip' : ''}`}
        disabled={!data.onToggle}
        onClick={data.onToggle}
        className={`nodrag nopan av:relative av:h-6 av:w-11 av:rounded-full av:border av:border-border av:transition-[background-color,transform] av:duration-150 av:active:scale-[0.97] av:disabled:cursor-default ${
          on ? 'av:bg-high' : 'av:bg-subtle'
        } ${data.onToggle ? 'av:cursor-pointer' : ''}`}
      >
        <span
          className={`av:absolute av:top-0.5 av:left-0.5 av:grid av:size-[1.125rem] av:place-items-center av:rounded-full av:bg-card av:font-mono av:text-[0.625rem] av:font-bold av:text-fg av:shadow-sm av:transition-transform av:duration-200 av:ease-out-strong ${
            on ? 'av:translate-x-5' : 'av:translate-x-0'
          }`}
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
    <div className="av:flex av:items-center av:gap-2">
      {inputHandles(1)}
      <span
        role="img"
        aria-label={`Output ${data.name}, ${levelName(data.level)}`}
        className={`av:grid av:size-7 av:place-items-center av:rounded-full av:border-2 av:font-mono av:text-xs av:font-bold av:transition-[background-color,box-shadow,border-color] av:duration-200 ${
          lit
            ? 'av:border-high av:bg-high av:text-card av:shadow-[0_0_12px_var(--av-high)]'
            : data.level === 0
              ? 'av:border-low av:bg-subtle av:text-muted'
              : 'av:border-unknown av:border-dashed av:bg-subtle av:text-fg'
        }`}
      >
        {levelText(data.level)}
      </span>
      <span className="av:font-mono av:text-xs av:font-semibold av:text-fg">{data.name}</span>
    </div>
  );
}

function ConstNode({ data }: { data: GateData }) {
  return (
    <div className="av:grid av:h-6 av:w-8 av:place-items-center av:rounded av:border av:border-border av:bg-subtle av:font-mono av:text-xs av:text-muted">
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
      className={`av:relative av:transition-transform av:duration-200 av:ease-out-strong ${
        data.changed ? 'av:scale-[1.06]' : 'av:scale-100'
      }`}
      title={`${data.name}: ${data.kind.toUpperCase()}, output ${levelName(data.level)}`}
    >
      {inputHandles(data.inputs)}
      <GateSymbol
        kind={data.kind}
        level={data.level}
      />
      <span className="av:pointer-events-none av:absolute av:inset-x-0 av:-bottom-4 av:text-center av:font-mono av:text-[0.625rem] av:text-muted">
        {data.kind.toUpperCase()}
      </span>
      <Handle
        type="source"
        position={Position.Right}
        isConnectable={false}
        style={{ left: tipX(data.kind) }}
        className="av:size-1.5! av:min-w-0! av:border-0! av:bg-muted!"
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
        className={`av:transition-[stroke] av:duration-200 ${strokeFor(l)}`}
        style={{ strokeWidth: l === 1 ? 3 : 2, strokeDasharray: l === -1 ? '4 4' : undefined }}
      />
      {l === 1 && props.data?.motion ? (
        <path
          d={path}
          fill="none"
          className="av:pointer-events-none av:stroke-card av:animate-wire-pulse"
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
      className="av:m-0"
      aria-label={`The circuit${circuit.name ? `: ${circuit.name}` : ''}`}
    >
      <div className="av-flow av:h-[min(28rem,60vh)]">
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
      <figcaption className="av:mt-2 av:flex av:flex-wrap av:items-center av:gap-x-4 av:gap-y-1 av:text-xs av:text-muted">
        <span className="av:inline-flex av:items-center av:gap-1.5">
          <span className="av:inline-block av:h-0.5 av:w-5 av:rounded av:bg-high" /> high (1)
        </span>
        <span className="av:inline-flex av:items-center av:gap-1.5">
          <span className="av:inline-block av:h-0.5 av:w-5 av:rounded av:bg-low" /> low (0)
        </span>
        <span className="av:inline-flex av:items-center av:gap-1.5">
          <span className="av:inline-block av:h-0.5 av:w-5 av:border-t-2 av:border-dashed av:border-unknown" />{' '}
          unknown (X)
        </span>
        <span className="av-sr-only">{outputs}.</span>
      </figcaption>
    </figure>
  );
}
