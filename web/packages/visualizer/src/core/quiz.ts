/**
 * Predict mode's questions, as the wasm build's quiz() returns them. Each one
 * pauses the player at `step` (in `unit`, which matches the run's frame) and
 * asks what happens next; `answer` is the index of the right choice, taken
 * from the run's real events.
 */
import { z } from 'zod';
import { failureSchema } from './schema';

const int = z.number().int();

export const questionSchema = z
  .object({
    kind: z.string(),
    prompt: z.string(),
    step: int.min(0),
    unit: z.enum(['step', 'time', 'seek', 'operation']),
    shown: z
      .array(int)
      .nullish()
      .transform((v) => v ?? []),
    context: z.string().default(''),
    choices: z.array(z.string()).min(2),
    answer: int.min(0),
    explanation: z.string()
  })
  .refine((q) => q.answer < q.choices.length, 'the answer is not one of the choices');
export type Question = z.infer<typeof questionSchema>;

export const quizResponse = z.discriminatedUnion('ok', [
  z.object({ ok: z.literal(true), questions: z.array(questionSchema) }),
  failureSchema
]);

/** The questions that fit a frame: the same unit, within its steps. */
export function questionsFor(questions: Question[], unit: string, total: number): Question[] {
  const u = unit.toLowerCase();
  return questions.filter((q) => q.unit === u && q.step <= total);
}
