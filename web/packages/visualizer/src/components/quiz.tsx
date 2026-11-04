import { useEffect, useId, useRef, type KeyboardEvent, type ReactNode } from 'react';
import type { Question } from '../core/quiz';

/**
 * One question, unstyled: the prompt, what the learner knows so far, and the
 * choices as native buttons (keys 1 to 4 pick one). Once answered it says in
 * words which choice was right and which was picked, never by color alone,
 * and announces the result in a live region. It only draws; the caller (with
 * gameReducer, or its own state) keeps the answers.
 *
 * Style it through `className` and `classNames` (utility classes for each
 * part; a choice's can depend on its state), or the data attributes:
 * data-quiz and data-state ("asking" or "answered") on the root, data-choice,
 * data-answer and data-picked on the buttons.
 */
export interface QuizClassNames {
  prompt?: string;
  context?: string;
  choices?: string;
  choice?: (s: { index: number; answer: boolean; picked: boolean; answered: boolean }) => string;
  result?: string;
}

export function Quiz({
  question,
  picked,
  answered = picked !== null,
  onAnswer,
  className,
  classNames = {},
  heading,
  children,
  renderChoice
}: {
  question: Question;
  /** The learner's choice, or null while asking (or after a timeout). */
  picked: number | null;
  /** Whether the answer is known; defaults to a choice having been picked. */
  answered?: boolean;
  onAnswer(choice: number): void;
  className?: string;
  classNames?: QuizClassNames;
  /** Shown above the prompt: "Question 3 of 10". */
  heading?: ReactNode;
  /** Shown after the choices: the explanation, a Next button. */
  children?: ReactNode;
  /** Draws a choice's content; the default is its text. */
  renderChoice?(choice: string, index: number): ReactNode;
}) {
  const id = useId();
  const first = useRef<HTMLButtonElement>(null);
  const right = answered && picked === question.answer;

  // Focus the first choice whenever a new question appears.
  useEffect(() => {
    if (!answered) first.current?.focus();
  }, [question, answered]);

  const onKeyDown = (e: KeyboardEvent) => {
    if (answered || e.altKey || e.ctrlKey || e.metaKey) return;
    const n = Number(e.key);
    if (Number.isInteger(n) && n >= 1 && n <= question.choices.length) {
      e.preventDefault();
      onAnswer(n - 1);
    }
  };

  return (
    <section
      data-quiz=""
      data-state={answered ? 'answered' : 'asking'}
      className={className}
      aria-labelledby={`${id}-prompt`}
      onKeyDown={onKeyDown}
    >
      {heading}
      <h2
        id={`${id}-prompt`}
        className={classNames.prompt}
      >
        {question.prompt}
      </h2>
      {question.context ? (
        <p
          data-context=""
          className={classNames.context}
        >
          {question.context}
        </p>
      ) : null}
      <ol
        data-choices=""
        className={classNames.choices}
      >
        {question.choices.map((c, i) => {
          const isAnswer = answered && i === question.answer;
          const isPick = answered && i === picked;
          return (
            <li key={c}>
              <button
                ref={i === 0 ? first : undefined}
                type="button"
                className={classNames.choice?.({
                  index: i,
                  answer: isAnswer,
                  picked: isPick,
                  answered
                })}
                data-choice={i}
                data-answer={isAnswer || undefined}
                data-picked={isPick || undefined}
                aria-disabled={answered || undefined}
                aria-keyshortcuts={String(i + 1)}
                onClick={() => {
                  if (!answered) onAnswer(i);
                }}
              >
                {renderChoice ? renderChoice(c, i) : c}
                {isAnswer ? <span data-tag="answer"> (the answer)</span> : null}
                {isPick && !isAnswer ? <span data-tag="picked"> (your answer)</span> : null}
              </button>
            </li>
          );
        })}
      </ol>
      <p
        aria-live="polite"
        className={classNames.result}
        data-result={answered ? (right ? 'right' : 'wrong') : undefined}
      >
        {answered
          ? `${right ? 'Right.' : picked === null ? 'Out of time.' : 'Not quite.'} ${question.explanation}`
          : ''}
      </p>
      {children}
    </section>
  );
}
