import { useEffect, useMemo } from 'react';
import {
  Controls,
  Handle,
  MarkerType,
  Position,
  ReactFlow,
  ReactFlowProvider,
  useReactFlow,
  type Edge,
  type Node,
  type NodeProps
} from '@xyflow/react';
import { layoutSnapshot } from '../core/layout';
import { diffSnapshots } from '../core/replay/structure';
import type { Snapshot } from '../core/schema';
import { useReducedMotion } from '../hooks/use-reduced-motion';

type ItemData = { label: string; role?: string; isNew: boolean; horizontal: boolean };
type ItemNode = Node<ItemData, 'item'>;

function ItemNodeView({ data }: NodeProps<ItemNode>) {
  const h = data.horizontal;
  return (
    <div
      className="av-node"
      data-role={data.role}
      data-new={data.isNew || undefined}
    >
      <Handle
        type="target"
        position={h ? Position.Left : Position.Top}
        isConnectable={false}
      />
      <span>{data.label}</span>
      {data.role ? <span className="av-node-role">{data.role}</span> : null}
      <Handle
        type="source"
        position={h ? Position.Right : Position.Bottom}
        isConnectable={false}
      />
    </div>
  );
}

const nodeTypes = { item: ItemNodeView };

function Fit({ count }: { count: number }) {
  const flow = useReactFlow();
  const reduced = useReducedMotion();
  useEffect(() => {
    const t = setTimeout(() => void flow.fitView({ padding: 0.2, duration: reduced ? 0 : 200 }), 0);
    return () => clearTimeout(t);
  }, [count, flow, reduced]);
  return null;
}

function describe(s: Snapshot): string {
  if (s.nodes.length < 1) return 'Empty.';
  const label = new Map(s.nodes.map((n) => [n.id, n.label]));
  const nodes = s.nodes.map((n) => (n.role ? `${n.label} (${n.role})` : n.label)).join(', ');
  const edges = s.edges
    .slice(0, 60)
    .map((e) => `${label.get(e.from)} to ${label.get(e.to)}${e.label ? ` (${e.label})` : ''}`)
    .join('; ');
  return `${s.nodes.length} nodes: ${nodes}.${edges ? ` Links: ${edges}.` : ''}`;
}

/**
 * A data structure's snapshot drawn with React Flow: rows and lists left to
 * right with arrows, trees and forests top down with a tidy layout, sets as a
 * grid. Nodes and links the last operation added are outlined and dashed
 * (pass the snapshot before it as `previous`). It is a picture, not an editor:
 * nothing can be dragged or connected; it pans, zooms with the buttons, and
 * comes with a text description for screen readers.
 */
export function StructureFlow({
  snapshot,
  previous,
  label = 'The data structure'
}: {
  snapshot: Snapshot;
  previous?: Snapshot;
  label?: string;
}) {
  const { nodes, edges } = useMemo(() => {
    const pos = layoutSnapshot(snapshot);
    const diff = diffSnapshots(previous, snapshot);
    const horizontal = snapshot.layout === 'row';
    const nodes: ItemNode[] = snapshot.nodes.map((n) => ({
      id: String(n.id),
      type: 'item',
      position: pos.get(n.id) ?? { x: 0, y: 0 },
      data: { label: n.label, role: n.role, isNew: diff.added.has(n.id), horizontal },
      draggable: false,
      connectable: false,
      selectable: false
    }));
    const edges: Edge[] = snapshot.edges.map((e) => ({
      id: `${e.from}-${e.to}`,
      source: String(e.from),
      target: String(e.to),
      label: e.label === 'left' || e.label === 'right' ? undefined : e.label,
      type: horizontal ? 'straight' : 'default',
      markerEnd: { type: MarkerType.ArrowClosed, width: 16, height: 16 },
      className: diff.newEdges.has(`${e.from}-${e.to}`) ? 'av-edge-new' : undefined,
      selectable: false,
      focusable: false
    }));
    return { nodes, edges };
  }, [snapshot, previous]);

  return (
    <figure
      style={{ margin: 0 }}
      aria-label={label}
    >
      <div
        data-flow-box
        className="av-flow"
      >
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
            panOnScroll={false}
            fitView
            minZoom={0.2}
            maxZoom={2}
            colorMode="system"
          >
            <Controls
              showInteractive={false}
              position="bottom-right"
            />
            <Fit count={nodes.length * 1000 + edges.length} />
          </ReactFlow>
        </ReactFlowProvider>
      </div>
      <figcaption className="av-sr-only">{describe(snapshot)}</figcaption>
    </figure>
  );
}
