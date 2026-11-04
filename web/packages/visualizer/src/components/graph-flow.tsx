import { useMemo } from 'react';
import {
  Controls,
  Handle,
  MarkerType,
  Position,
  ReactFlow,
  ReactFlowProvider,
  type Edge,
  type Node,
  type NodeProps
} from '@xyflow/react';
import { layoutGraph } from '../core/graph-layout';
import type { GraphState } from '../core/replay/graph';
import type { Graph } from '../core/schema';

type GNodeData = {
  label: string;
  state: string[];
  distance?: number;
  order?: number;
  group?: number;
};
type GNode = Node<GNodeData, 'gnode'>;

/** The node's states, in words and as data attributes. */
function nodeStates(i: number, s?: GraphState): string[] {
  if (!s) return [];
  const out: string[] = [];
  if (s.path.includes(i)) out.push('path');
  if (s.cycle.has(i)) out.push('cycle');
  if (s.current === i) out.push('current');
  if (s.visited.has(i)) out.push('visited');
  else if (s.frontier.has(i)) out.push('frontier');
  return out;
}

function GraphNodeView({ data }: NodeProps<GNode>) {
  return (
    <div
      className="av-gnode"
      data-state={data.state.join(' ') || undefined}
      data-group={data.group}
    >
      <Handle
        type="target"
        position={Position.Top}
        isConnectable={false}
      />
      <span className="av-gnode-label">{data.label}</span>
      {data.distance !== undefined ? (
        <span
          className="av-gnode-badge"
          title="distance"
        >
          d={data.distance}
        </span>
      ) : null}
      {data.order !== undefined ? (
        <span
          className="av-gnode-badge"
          title="order"
        >
          #{data.order}
        </span>
      ) : null}
      {data.group !== undefined ? (
        <span
          className="av-gnode-badge"
          title="group"
        >
          g{data.group}
        </span>
      ) : null}
      <Handle
        type="source"
        position={Position.Bottom}
        isConnectable={false}
      />
    </div>
  );
}

const nodeTypes = { gnode: GraphNodeView };

function describe(g: Graph, s?: GraphState): string {
  const parts = [
    `${g.nodes.length} nodes and ${g.edges.length} ${g.directed ? 'directed ' : ''}edges.`
  ];
  if (s?.path.length) parts.push(`Path: ${s.path.map((k) => g.nodes[k]).join(', ')}.`);
  if (s?.tree.size) parts.push(`${s.tree.size} edges chosen.`);
  if (s?.order.size)
    parts.push(
      `Order: ${[...s.order.entries()]
        .sort((a, b) => a[1] - b[1])
        .map(([k]) => g.nodes[k])
        .join(', ')}.`
    );
  if (s?.distance.size)
    parts.push(
      `Distances: ${[...s.distance.entries()].map(([k, d]) => `${g.nodes[k]} ${d}`).join(', ')}.`
    );
  return parts.join(' ');
}

/**
 * A graph drawn with React Flow: a directed acyclic graph in layers, any other
 * graph on a circle; weights on the edges, arrows when directed. As events
 * play, the frontier, visited nodes, distances, the path, the order, groups,
 * chosen edges and flow appear; every state has a label or pattern as well as
 * a color, and a text description follows for screen readers.
 */
export function GraphFlow({
  graph,
  state,
  label = 'The graph'
}: {
  graph: Graph;
  state?: GraphState;
  label?: string;
}) {
  const { nodes, edges } = useMemo(() => {
    const pos = layoutGraph(graph);
    const nodes: GNode[] = graph.nodes.map((name, i) => ({
      id: String(i),
      type: 'gnode',
      position: pos[i] ?? { x: 0, y: 0 },
      data: {
        label: name,
        state: nodeStates(i, state),
        distance: state?.distance.get(i),
        order: state?.order.get(i),
        group: state?.group.get(i)
      },
      draggable: false,
      connectable: false,
      selectable: false
    }));
    const onPath = new Set<string>();
    const p = state?.path ?? [];
    for (let k = 1; k < p.length; k++) {
      onPath.add(`${p[k - 1]}-${p[k]}`);
      if (!graph.directed) onPath.add(`${p[k]}-${p[k - 1]}`);
    }
    const edges: Edge[] = graph.edges.map((e, k) => {
      const flow = state?.flow.get(k);
      const chosen = state?.tree.has(k) || onPath.has(`${e.from}-${e.to}`);
      return {
        id: `e${k}`,
        source: String(e.from),
        target: String(e.to),
        label: flow !== undefined ? `${flow}/${e.weight}` : String(e.weight),
        type: 'straight',
        markerEnd: graph.directed
          ? { type: MarkerType.ArrowClosed, width: 16, height: 16 }
          : undefined,
        className: chosen
          ? 'av-edge-chosen'
          : state?.considering === k
            ? 'av-edge-considered'
            : undefined,
        selectable: false,
        focusable: false
      };
    });
    return { nodes, edges };
  }, [graph, state]);

  return (
    <figure
      style={{ margin: 0 }}
      aria-label={label}
    >
      <div className="av-flow">
        <ReactFlowProvider>
          <ReactFlow
            nodes={nodes}
            edges={edges}
            nodeTypes={nodeTypes}
            nodesDraggable={false}
            nodesConnectable={false}
            elementsSelectable={false}
            nodesFocusable={false}
            edgesFocusable={false}
            zoomOnScroll={false}
            preventScrolling={false}
            fitView
            minZoom={0.2}
            maxZoom={2}
            colorMode="system"
          >
            <Controls
              showInteractive={false}
              position="bottom-right"
            />
          </ReactFlow>
        </ReactFlowProvider>
      </div>
      <figcaption className="av-sr-only">{describe(graph, state)}</figcaption>
    </figure>
  );
}
