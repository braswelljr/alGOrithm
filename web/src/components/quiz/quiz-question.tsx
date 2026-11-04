'use client';

import type { ReactNode } from 'react';
import { QuizChoiceCard, type ChoiceResult } from '@/components/algorithm/quiz-choice-card';
import {
  Questionnaire,
  QuestionnaireChoices,
  QuestionnaireDescription,
  QuestionnaireItem,
  QuestionnaireTitle
} from '@/components/ui/questionnaire';
import type { Question } from '@/visualizer/core';

/**
 * One quiz question on the shadcn Questionnaire: the prompt, the context, and
 * the choices as QuizChoiceCards with their number keys (1-4, the
 * Questionnaire's own shortcuts). Picking a choice answers; after that the
 * choices lock and show right, wrong and the answer in words and icons, and
 * the explanation is announced.
 */
export function QuizQuestion({
  question,
  index,
  picked,
  answered,
  onAnswer,
  children
}: {
  question: Question;
  /** The question's place in the game, for the Questionnaire item's name. */
  index: number;
  picked: number | null;
  answered: boolean;
  onAnswer: (choice: number) => void;
  /** Shown under the choices once answered: the Next button. */
  children?: ReactNode;
}) {
  const right = picked === question.answer;
  const result = (i: number): ChoiceResult => {
    if (!answered) return 'idle';
    if (i === question.answer) return picked === i ? 'correct' : 'answer';
    return i === picked ? 'wrong' : 'idle';
  };
  const name = `q${index}`;
  return (
    <Questionnaire
      key={name}
      shortcuts="numbers"
      onSubmit={(e) => e.preventDefault()}
      className="gap-4"
    >
      <QuestionnaireItem name={name}>
        <QuestionnaireTitle className="text-lg">{question.prompt}</QuestionnaireTitle>
        {question.context ? (
          <QuestionnaireDescription className="whitespace-pre-line">
            {question.context}
          </QuestionnaireDescription>
        ) : null}
        <QuestionnaireChoices>
          {question.choices.map((c, i) => (
            <QuizChoiceCard
              key={`${name}-${i}`}
              value={String(i)}
              result={result(i)}
              disabled={answered}
              onSelect={() => {
                if (!answered) onAnswer(i);
              }}
            >
              {c}
            </QuizChoiceCard>
          ))}
        </QuestionnaireChoices>
      </QuestionnaireItem>
      <p
        aria-live="polite"
        className="min-h-6 text-[0.9375rem]"
        data-result={answered ? (right ? 'right' : 'wrong') : undefined}
      >
        {answered
          ? `${right ? 'Right.' : picked === null ? 'Out of time.' : 'Not quite.'} ${question.explanation}`
          : ''}
      </p>
      {children}
    </Questionnaire>
  );
}
