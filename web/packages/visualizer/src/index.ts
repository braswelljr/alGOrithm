/**
 * @braswelljr/algorithm-visualizer: watch real algorithms and data
 * structures run, step by step. Everything in ./core (no React) plus the
 * React components and hooks.
 */
export * from './core/index';

export { AlgorithmProvider, useAlgorithm, type AlgorithmState } from './hooks/use-algorithm';
export { usePlayer, type Player } from './hooks/use-player';
export { useReducedMotion } from './hooks/use-reduced-motion';

export { Visualizer } from './components/visualizer';
export { RunView } from './components/run-view';
export { frameFor, runKind, type Frame, type RunKind } from './components/run-frames';
export { TimelineView } from './components/timeline-view';
export { TrackView } from './components/track-view';
export { LinesView } from './components/lines-view';
export { SummaryList } from './components/summary-list';
export { StructureLog } from './components/structure-log';
export { BarsView } from './components/bars-view';
export { CellsView } from './components/cells-view';
export { GridView } from './components/grid-view';
export { TableView } from './components/table-view';
export { TextMatchView } from './components/text-match-view';
export { HanoiView } from './components/hanoi-view';
export { StructureFlow } from './components/lazy-flow';
export { GraphFlow } from './components/lazy-flow';
export { StateMachineView, InputStrip } from './components/lazy-flow';
export { CPUView } from './components/cpu-view';
export { PipelineView } from './components/pipeline-view';
export { VMView } from './components/vm-view';
export { TokensView } from './components/tokens-view';
export { HeapView } from './components/heap-view';
export { TapeView } from './components/tape-view';
export { StackView } from './components/stack-view';
export { CellBoardView } from './components/cell-board-view';
export { CircuitView } from './components/lazy-flow';
export { ArchitectureView } from './components/lazy-flow';
export { ArchitectureLegend, StatusMark } from './components/architecture-legend';
export { TruthTableView, KMapView } from './components/logic-view';
export { WaveformView } from './components/waveform-view';
export { BitsView } from './components/bits-view';
export { PointsView } from './components/points-view';
export { BoardView } from './components/board-view';
export { PlayerControls } from './components/player-controls';
export { Quiz, type QuizClassNames } from './components/quiz';
export { CodeView } from './components/code-view';
export { StepDescription } from './components/step-description';
export { CountsList } from './components/counts-list';
export { ResultSummary, summarize } from './components/result-summary';
export { Legend, MarkSymbol, Patterns, SwatchLegend, type MarkName } from './components/legend';
