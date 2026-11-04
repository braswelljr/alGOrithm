/**
 * The shapes of everything the wasm build returns, as zod schemas. Every value
 * that crosses the boundary is parsed here before anything uses it
 * (docs/rules/data-from-outside-is-unknown-until-checked.md). The Go side is
 * cmd/ (the registry and Execute) and cmd/algorithm-wasm/main.go.
 */
import { z } from 'zod';

const int = z.number().int();

/** A whole number of any size: Go's int64 (math.MaxInt64 as "unreachable",
 * F(90), 20!) can exceed what JavaScript holds exactly, so these are not
 * required to be safe integers. */
const whole = z.number().refine(Number.isInteger, 'expected a whole number');

/** Go encodes a nil slice as null; read it as an empty array. */
function list<T extends z.ZodType>(item: T) {
  return z
    .array(item)
    .nullish()
    .transform((v) => v ?? []);
}

/** What an entry takes as input. */
export const inputKinds = [
  'ints',
  'sorted-ints',
  'any-ints',
  'text',
  'two-texts',
  'number',
  'two-numbers',
  'ops',
  'grid',
  'pattern',
  'amount',
  'processes',
  'pages',
  'tracks',
  'partitions',
  'banker',
  'simulation',
  'graph',
  'points'
] as const;
export type InputKind = (typeof inputKinds)[number];

/** Optional inputs an entry reads beyond its input kind. */
export const paramNames = ['target', 'shift', 'key', 'pattern', 'amount', 'start', 'end'] as const;
export type ParamName = (typeof paramNames)[number];

/** A numeric knob an entry reads from Input.options (quantum, frames, head, ...). */
export const optionSchema = z.object({
  name: z.string().min(1),
  label: z.string(),
  default: int,
  min: int.optional(),
  max: int.optional()
});
export type EntryOption = z.infer<typeof optionSchema>;

/** One real-world use of an entry; example is a path like examples/leaderboard. */
export const realWorldSchema = z.object({
  title: z.string(),
  where: z.string(),
  example: z.string().default('')
});
export type RealWorld = z.infer<typeof realWorldSchema>;

export const entrySchema = z.object({
  id: z.string().min(1),
  name: z.string(),
  category: z.string(),
  /** The Go import path, e.g. github.com/braswelljr/alGOrithm/pkg/algorithms/sorting. */
  import: z.string().default(''),
  /** The Go package name, e.g. sorting. */
  package: z.string().default(''),
  /** The exported function, e.g. Quick. */
  func: z.string().default(''),
  /** The source file, relative to the repository root. */
  file: z.string().default(''),
  summary: z.string(),
  time: z.string(),
  space: z.string(),
  /** Where the time comes from, in a sentence or two; empty when not written. */
  why: z.string().default(''),
  /** The shortcuts to the entry: a TL;DR, the key idea, when to use it and the
   *  pitfall (internal/core/notes.go); absent when not written. */
  notes: z
    .object({
      tldr: z.string(),
      keyIdea: z.string(),
      whenToUse: z.string(),
      pitfall: z.string()
    })
    .optional(),
  /** Where it shows up in real software, with a runnable program when there is one. */
  realWorld: list(realWorldSchema),
  input: z.enum(inputKinds),
  traced: z.boolean(),
  // Unknown names (from an older or newer build) are dropped, not fatal.
  params: list(z.string()).transform((ps) =>
    ps.filter((p): p is ParamName => (paramNames as readonly string[]).includes(p))
  ),
  options: list(optionSchema),
  ops: list(z.string()),
  limits: z.string().optional(),
  /** An input that shows the entry off: prefills forms and "Try it". */
  /** What the entry is: "algorithm", "data-structure", "simulation" or "utility". */
  kind: z.string().default('algorithm'),
  /** "beginner", "intermediate" or "advanced"; empty from an older build. */
  level: z.string().default(''),
  /** Tags from the controlled vocabulary (catalog().tags explains each). */
  tags: list(z.string()),
  /** The allowed Input.key names (a JS array method's callbacks), when fixed. */
  choices: list(z.string()),
  example: z.lazy(() => inputSchema).default({})
});
export type Entry = z.infer<typeof entrySchema>;

export const topics = [
  'Algorithms',
  'Mathematics',
  'Security',
  'Data structures',
  'Operating systems',
  'Machine learning',
  'Computing',
  'Systems'
] as const;
export type Topic = (typeof topics)[number];

export const categorySchema = z.union([
  z.object({
    id: z.string().min(1),
    name: z.string(),
    summary: z.string().default(''),
    doc: z.string().default(''),
    import: z.string().default(''),
    topic: z.enum(topics).default('Algorithms')
  }),
  // An older wasm build lists bare ids.
  z
    .string()
    .min(1)
    .transform((id) => ({
      id,
      name: id,
      summary: '',
      doc: '',
      import: '',
      topic: 'Algorithms' as Topic
    }))
]);
export type Category = z.infer<typeof categorySchema>;

export const catalogSchema = z.object({
  ok: z.literal(true),
  version: z.string(),
  categories: z.array(categorySchema),
  entries: z.array(entrySchema),
  datasets: list(z.string()),
  gridKinds: list(z.string()),
  /** The tag vocabulary: what each tag means and how many entries carry it. */
  tags: list(
    z.object({
      tag: z.string().min(1),
      meaning: z.string().default(''),
      count: int.min(0).default(0)
    })
  ),
  maxValues: int.positive()
});
export type Catalog = z.infer<typeof catalogSchema>;

/** The kinds of event, in the order the Go side numbers them. */
export const kinds = [
  'compare',
  'swap',
  'set',
  'visit',
  'mark',
  'push',
  'pop',
  'link',
  'unlink'
] as const;
export type KindName = (typeof kinds)[number];

export const eventSchema = z.object({
  kind: int.min(0).max(kinds.length - 1),
  indices: list(int),
  values: list(int),
  note: z.string().optional()
});
export type TraceEvent = z.infer<typeof eventSchema>;

export const nodeSchema = z.object({
  id: int,
  label: z.string(),
  role: z.string().optional(),
  parent: int
});
export type SnapshotNode = z.infer<typeof nodeSchema>;

export const edgeSchema = z.object({
  from: int,
  to: int,
  label: z.string().optional()
});
export type SnapshotEdge = z.infer<typeof edgeSchema>;

export const snapshotSchema = z.object({
  layout: z.enum(['row', 'tree', 'set']),
  nodes: list(nodeSchema),
  edges: list(edgeSchema)
});
export type Snapshot = z.infer<typeof snapshotSchema>;

export const opSchema = z.object({ name: z.string(), value: z.string().optional() });
export type Op = z.infer<typeof opSchema>;

export const stepSchema = z.object({
  op: opSchema,
  result: z.string().optional(),
  error: z.string().optional(),
  snapshot: snapshotSchema
});
export type Step = z.infer<typeof stepSchema>;

export const gridSchema = z
  .object({
    rows: int.positive(),
    cols: int.positive(),
    cost: z.array(int.min(0).max(9)),
    start: int.min(0),
    end: int.min(0),
    /** A label per cell, row-major like cost (word search letters). */
    labels: list(z.string())
  })
  .refine((g) => g.cost.length === g.rows * g.cols, 'the grid has the wrong number of cells')
  .refine(
    (g) => g.labels.length === 0 || g.labels.length === g.cost.length,
    'the grid has the wrong number of labels'
  );
export type Grid = z.infer<typeof gridSchema>;

export const tableSchema = z
  .object({
    rows: int.positive(),
    cols: int.positive(),
    cells: z.array(whole),
    rowLabels: list(z.string()),
    colLabels: list(z.string()),
    /** A cell equal to blank means "nothing here" (an empty frame, an unreachable pair). */
    blank: whole.optional()
  })
  .refine((t) => t.cells.length === t.rows * t.cols, 'the table has the wrong number of cells');
export type Table = z.infer<typeof tableSchema>;

/** One bar of a timeline: a process on the CPU, a thread at a tick. */
export const spanSchema = z.object({
  lane: z.string(),
  label: z.string().default(''),
  start: z.number().min(0),
  end: z.number().min(0),
  state: z.string().optional()
});
export type Span = z.infer<typeof spanSchema>;

/** A disk head's path over tracks 0..size-1. */
export const trackSchema = z.object({ size: int.positive(), path: z.array(int.min(0)) });
export type Track = z.infer<typeof trackSchema>;

/** A graph: node names, and weighted edges between node indexes. */
export const graphSchema = z
  .object({
    directed: z.boolean().default(false),
    nodes: list(z.string()),
    edges: list(z.object({ from: int.min(0), to: int.min(0), weight: int.default(1) }))
  })
  .refine(
    (g) => g.edges.every((e) => e.from < g.nodes.length && e.to < g.nodes.length),
    'an edge names a node that does not exist'
  );
export type Graph = z.infer<typeof graphSchema>;
export type GraphEdge = Graph['edges'][number];

/** A point in the plane, for geometry. */
export const pointSchema = z.object({ x: int, y: int });
export type PlanePoint = z.infer<typeof pointSchema>;

/** An automaton: states, transitions merged by state pair, the input it
 * read, and one step per move (a Turing machine's tape, a PDA's stack, a
 * transducer's output when the machine has one). */
export const machineSchema = z
  .object({
    states: list(
      z.object({
        name: z.string(),
        start: z.boolean().default(false),
        accept: z.boolean().default(false)
      })
    ),
    edges: list(z.object({ from: int.min(0), to: int.min(0), label: z.string().default('') })),
    input: z.string().default(''),
    steps: list(
      z.object({
        active: list(int.min(0)),
        edges: list(int.min(0)),
        head: int.min(0).default(0),
        tape: z.string().default(''),
        tapeHead: int.default(0),
        stack: z.string().default(''),
        output: z.string().default(''),
        note: z.string().default('')
      })
    ),
    accepted: z.boolean().optional()
  })
  .refine(
    (m) =>
      m.edges.every((e) => e.from < m.states.length && e.to < m.states.length) &&
      m.steps.every(
        (s) =>
          s.active.every((a) => a < m.states.length) && s.edges.every((e) => e < m.edges.length)
      ),
    'the machine names a state or transition that does not exist'
  );
export type Automaton = z.infer<typeof machineSchema>;
export type AutomatonStep = Automaton['steps'][number];

/** A wire's level: 1 high, 0 low, -1 unknown (nothing has driven it yet). */
const level = z.union([z.literal(1), z.literal(0), z.literal(-1)]);

/** A gate-level circuit and every gate's output level after each gate delay. */
export const circuitSchema = z.object({
  name: z.string().default(''),
  gates: z.array(
    z.object({
      name: z.string(),
      kind: z.string(),
      in: list(int.min(0)),
      layer: int.min(0),
      row: int.min(0)
    })
  ),
  frames: z.array(z.array(level)),
  phases: list(z.object({ frame: int.min(0), label: z.string() }))
});
export type Circuit = z.infer<typeof circuitSchema>;
export type CircuitGate = Circuit['gates'][number];
export type Level = z.infer<typeof level>;

/** A Karnaugh map: cells row by row (1, 0, or 2 for "does not matter"). */
export const kmapSchema = z.object({
  rows: int.min(1),
  cols: int.min(1),
  rowVars: z.string(),
  colVars: z.string(),
  rowLabels: z.array(z.string()),
  colLabels: z.array(z.string()),
  cells: z.array(int),
  minterm: z.array(int),
  groups: list(z.array(int.min(0))),
  terms: list(z.string()),
  sop: z.string()
});
export type KMap = z.infer<typeof kmapSchema>;

/** A truth table (out[i] is row i's output: 1, 0, or 2 for "does not
 * matter"), and a Karnaugh map when the entry draws one. */
export const logicSchema = z.object({
  vars: z.array(z.string()),
  out: z.array(int),
  expression: z.string().optional(),
  kmap: kmapSchema.optional()
});
export type LogicTable = z.infer<typeof logicSchema>;

/** A row of bits split into named fields with roles (sign, exponent, ...). */
export const bitRowSchema = z.object({
  label: z.string(),
  fields: list(z.object({ name: z.string(), bits: z.string(), role: z.string().default('value') })),
  note: z.string().optional()
});
export type BitRow = z.infer<typeof bitRowSchema>;

/** One step of a running machine: a CPU, a pipeline, a VM, a lexer, a call
 * stack or a garbage-collected heap. pc indexes the listing and focus the
 * token, object or row in focus (-1 for none); the rest applies by view. */
export const execFrameSchema = z.object({
  note: z.string(),
  phase: z.string().optional(),
  pc: int.min(-1),
  focus: int.min(-1),
  cycle: int.min(0).default(0),
  registers: list(whole),
  flags: list(z.string()),
  memory: list(whole),
  changed: list(z.string()),
  stack: list(z.string()),
  vars: z
    .record(z.string(), z.string())
    .nullish()
    .transform((v) => v ?? {}),
  output: list(z.string()),
  colors: list(z.enum(['white', 'gray', 'black', 'free'])),
  refs: list(list(int.min(0))),
  roots: list(z.boolean()),
  counts: list(int),
  toSpace: list(int.min(0))
});
export type ExecFrame = z.infer<typeof execFrameSchema>;

export const execViews = ['cpu', 'pipeline', 'vm', 'tokens', 'heap'] as const;
export type ExecView = (typeof execViews)[number];

/** A machine running a program, one frame per step, and what its view draws
 * besides: a listing, a lexer's tokens, a heap's objects, a pipeline's
 * stage table and forwards. */
export const execSchema = z
  .object({
    view: z.enum(execViews),
    code: list(z.string()),
    source: z.string().optional(),
    tokens: list(
      z.object({
        kind: z.string(),
        text: z.string(),
        start: int.min(0),
        end: int.min(0),
        line: int.min(1),
        col: int.min(1)
      })
    ),
    objects: list(z.string()),
    rows: list(z.string()),
    flushed: list(z.boolean()),
    cells: list(list(z.string())),
    forwards: list(
      z.object({ from: int.min(0), to: int.min(0), cycle: int.min(0), reg: int.min(0) })
    ),
    stackLabel: z.string().optional(),
    varsLabel: z.string().optional(),
    outputLabel: z.string().optional(),
    frames: list(execFrameSchema)
  })
  .refine(
    (x) =>
      x.frames.every(
        (f) =>
          f.pc < Math.max(x.code.length, 1) &&
          f.refs.every((r) => r.every((k) => k < x.objects.length)) &&
          f.toSpace.every((k) => k < x.objects.length)
      ) &&
      x.forwards.every((f) => f.from < x.rows.length && f.to < x.rows.length) &&
      x.cells.length === x.rows.length,
    'the run points at a line, object or row that does not exist'
  );
export type Exec = z.infer<typeof execSchema>;

export const componentKinds = [
  'client',
  'lb',
  'service',
  'cache',
  'db',
  'queue',
  'cdn',
  'blob',
  'search',
  'worker'
] as const;
export type ComponentKind = (typeof componentKinds)[number];
export const requestStatuses = ['ok', 'hit', 'miss', 'error', 'queued', 'dropped'] as const;
export type RequestStatus = (typeof requestStatuses)[number];

/** A system design: its components (boxes), the links requests travel
 * along, and one frame per step with the requests on each link and each
 * component's load (hitRate is a percent, -1 when not a cache). */
export const architectureSchema = z
  .object({
    components: list(
      z.object({
        id: z.string().min(1),
        kind: z.enum(componentKinds),
        label: z.string(),
        notes: z.string().optional()
      })
    ),
    links: list(z.object({ from: z.string(), to: z.string(), label: z.string().optional() })),
    frames: list(
      z.object({
        note: z.string(),
        requests: list(z.object({ link: int.min(0), id: int, status: z.enum(requestStatuses) })),
        loads: list(
          z.object({
            component: z.string(),
            load: int,
            queue: int,
            hitRate: int.min(-1).max(100)
          })
        )
      })
    )
  })
  .refine((a) => {
    const ids = new Set(a.components.map((c) => c.id));
    return (
      a.links.every((l) => ids.has(l.from) && ids.has(l.to)) &&
      a.frames.every(
        (f) =>
          f.requests.every((q) => q.link < a.links.length) &&
          f.loads.every((l) => ids.has(l.component))
      )
    );
  }, 'the design points at a component or link that does not exist');
export type Architecture = z.infer<typeof architectureSchema>;
export type ArchitectureFrame = Architecture['frames'][number];

export const outputSchema = z.object({
  initial: list(int),
  values: list(whole),
  index: int.optional(),
  text: z.string().optional(),
  number: whole.optional(),
  bool: z.boolean().optional(),
  lines: list(z.string()),
  steps: list(stepSchema),
  grid: gridSchema.optional(),
  path: list(int),
  table: tableSchema.optional(),
  matches: z.array(int).nullish(),
  timeline: list(spanSchema),
  track: trackSchema.optional(),
  graph: graphSchema.optional(),
  points: list(pointSchema),
  /** Pairs of point indexes to draw as line segments. */
  segments: list(z.tuple([int.min(0), int.min(0)])),
  summary: z.record(z.string(), z.number()).optional(),
  machine: machineSchema.optional(),
  circuit: circuitSchema.optional(),
  logic: logicSchema.optional(),
  bits: list(bitRowSchema),
  exec: execSchema.optional(),
  architecture: architectureSchema.optional()
});
export type Output = z.infer<typeof outputSchema>;

export const resultSchema = z.object({
  id: z.string(),
  output: outputSchema,
  events: list(eventSchema),
  truncated: z.boolean().optional(),
  counts: z.record(z.string(), int).optional(),
  nanos: int.min(0)
});
export type RunResult = z.infer<typeof resultSchema>;

export const problemSchema = z.object({
  line: int.optional(),
  column: int.optional(),
  value: z.string(),
  reason: z.string()
});
export type Problem = z.infer<typeof problemSchema>;

export const failureSchema = z.object({
  ok: z.literal(false),
  error: z.string(),
  problems: list(problemSchema)
});
export type Failure = z.infer<typeof failureSchema>;

function response<T extends z.ZodRawShape>(shape: T) {
  return z.discriminatedUnion('ok', [z.object({ ok: z.literal(true), ...shape }), failureSchema]);
}

export const runResponse = response({ result: resultSchema });
export const compareResponse = response({ results: z.array(resultSchema) });
export const valuesResponse = response({ values: list(int) });
export const textResponse = response({ text: z.string() });
export const previewResponse = response({
  header: list(z.string()),
  rows: list(list(z.string())),
  width: int.min(0),
  count: int.min(0)
});
export type Preview = Extract<z.infer<typeof previewResponse>, { ok: true }>;

export const sourceResponse = response({
  file: z.string(),
  code: z.string(),
  doc: z.string().default(''),
  packageDoc: z.string().default('')
});
export type Source = Extract<z.infer<typeof sourceResponse>, { ok: true }>;

export const exampleResponse = response({
  path: z.string(),
  code: z.string(),
  command: z.string()
});
export type ExampleProgram = Extract<z.infer<typeof exampleResponse>, { ok: true }>;

export const snippetResponse = response({ go: z.string(), cli: z.string() });
export type Snippet = Extract<z.infer<typeof snippetResponse>, { ok: true }>;

/** What is sent as an entry's input. `target` is also the amount for coin
 * change and the capacity for the LRU cache. */
export const inputSchema = z.object({
  values: z.array(int).optional(),
  target: int.optional(),
  text: z.string().optional(),
  other: z.string().optional(),
  key: z.string().optional(),
  shift: int.optional(),
  ops: z.array(opSchema).optional(),
  /** Numeric knobs, by Entry.options name. */
  options: z.record(z.string(), int).optional()
});
export type Input = z.infer<typeof inputSchema>;

/** A run saved to a file or the run history: what ran, on what, and the result. */
export const savedRunSchema = z.object({
  format: z.literal('algorithm-run'),
  version: z.literal(1),
  entry: z.string().min(1),
  input: inputSchema,
  savedAt: z.string(),
  result: resultSchema.optional()
});
export type SavedRun = z.infer<typeof savedRunSchema>;

/** One step of a learning path: an entry to run and what to notice. A step
 * without an input runs the entry's example. */
export const pathStepSchema = z.object({
  entry: z.string().min(1),
  note: z.string(),
  input: inputSchema.optional(),
  name: z.string(),
  command: z.string()
});
export type PathStep = z.infer<typeof pathStepSchema>;

/** A learning path: ordered entries that build one idea on the last. */
export const learningPathSchema = z.object({
  id: z.string().min(1),
  title: z.string(),
  summary: z.string(),
  level: z.string(),
  audience: z.string(),
  prerequisites: z.array(z.string()).default([]),
  guide: z.string().optional(),
  steps: z.array(pathStepSchema)
});
export type LearningPath = z.infer<typeof learningPathSchema>;

export const pathsResponse = response({ paths: z.array(learningPathSchema) });

/** A system design case study (pkg/systems/design Studies): requirements,
 * the estimate that sizes it, trade-offs, failure modes and entries to run
 * next. id is the entry that simulates it. */
export const designStudySchema = z.object({
  id: z.string().min(1),
  title: z.string(),
  level: z.enum(['beginner', 'intermediate', 'advanced']),
  pitch: z.string(),
  functional: z.array(z.string()),
  nonFunctional: z.array(z.string()),
  estimate: z.record(z.string(), int),
  tradeoffs: z.array(z.object({ choice: z.string(), gain: z.string(), cost: z.string() })),
  failures: z.array(z.object({ what: z.string(), handling: z.string() })),
  related: z.array(z.string())
});
export type DesignStudy = z.infer<typeof designStudySchema>;

export const designsResponse = response({
  studies: z.array(designStudySchema),
  patterns: z.array(z.string()),
  foundations: z.array(z.string())
});
export type Designs = Omit<z.infer<typeof designsResponse> & { ok: true }, 'ok'>;
