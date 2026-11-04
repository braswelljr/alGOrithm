/**
 * The React hooks alone, without any view: a page that only needs the loaded
 * API (a catalog, a quiz, a measurement) imports this entry and so never
 * downloads React Flow or the views. The views come from the package root.
 */
export { AlgorithmProvider, useAlgorithm, type AlgorithmState } from './hooks/use-algorithm';
