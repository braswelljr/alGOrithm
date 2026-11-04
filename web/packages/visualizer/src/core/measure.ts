/**
 * Measured complexity, as the wasm build's measure() returns it, and the
 * growth classes it fits, so a page can draw each class's curve beside the
 * measured counts. The classes and their order match internal/core/bench.go.
 */
import { z } from 'zod';
import { failureSchema } from './schema';

const int = z.number().int();

export const measureSampleSchema = z.object({
  n: int.positive(),
  ops: int.min(0),
  nanos: int.min(0)
});
export type MeasureSample = z.infer<typeof measureSampleSchema>;

export const growthFitSchema = z.object({
  class: z.string(),
  /** ops is about scale * f(n). */
  scale: z.number(),
  /** The root mean square relative error; -1 when the class overflowed. */
  error: z.number()
});
export type GrowthFit = z.infer<typeof growthFitSchema>;

export const measurementSchema = z.object({
  id: z.string(),
  name: z.string(),
  order: z.string(),
  stated: z.string(),
  samples: z.array(measureSampleSchema).min(1),
  byTime: z.boolean().default(false),
  fits: z.array(growthFitSchema).min(1),
  best: z.string(),
  error: z.number(),
  quality: z.enum(['good', 'rough', 'poor']),
  /** Absent when the stated cost uses a size the measurement holds fixed. */
  agrees: z.boolean().optional(),
  note: z.string().default('')
});
export type Measurement = z.infer<typeof measurementSchema>;

export const measureResponse = z.discriminatedUnion('ok', [
  z.object({ ok: z.literal(true), measurement: measurementSchema }),
  failureSchema
]);

/** The growth classes, slowest first, with the function each one names. */
export const growthClasses = [
  { class: 'O(1)', f: () => 1 },
  { class: 'O(log n)', f: (n: number) => Math.log2(n) },
  { class: 'O(sqrt n)', f: (n: number) => Math.sqrt(n) },
  { class: 'O(n)', f: (n: number) => n },
  { class: 'O(n log n)', f: (n: number) => n * Math.log2(n) },
  { class: 'O(n^2)', f: (n: number) => n * n },
  { class: 'O(n^3)', f: (n: number) => n * n * n },
  { class: 'O(2^n)', f: (n: number) => 2 ** n }
] as const;

export type GrowthClass = (typeof growthClasses)[number]['class'];

/** What a fit predicts at size n: scale * f(n), or NaN for an unknown class. */
export function predict(fit: GrowthFit, n: number): number {
  const g = growthClasses.find((c) => c.class === fit.class);
  return g ? fit.scale * g.f(n) : Number.NaN;
}

/** f(n) for a class by name, or NaN when the name is unknown. */
export function growth(name: string, n: number): number {
  const g = growthClasses.find((c) => c.class === name);
  return g ? g.f(n) : Number.NaN;
}
