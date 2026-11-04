'use client';

import { CheckIcon, XIcon } from 'lucide-react';
import { QuestionnaireChoice } from '@/components/ui/questionnaire';
import { cn } from '@/lib/utils';

export type ChoiceResult = 'idle' | 'correct' | 'wrong' | 'answer';

/**
 * One answer in a quiz question: a big card in a Questionnaire, with its
 * number key (1-4, from the Questionnaire's shortcuts) and, once answered, its
 * result as an icon and a word as well as a color: "Correct", "Wrong", or
 * "The answer" on the right one when the pick was wrong. A wrong pick shakes
 * (not under reduced motion).
 *
 *   <QuestionnaireChoices>{choices.map((c) => <QuizChoiceCard key={c} value={c} result={...}>{c}</QuizChoiceCard>)}</QuestionnaireChoices>
 */
export function QuizChoiceCard({
  value,
  result = 'idle',
  disabled,
  onSelect,
  children,
  className
}: {
  value: string;
  result?: ChoiceResult;
  disabled?: boolean;
  /** Called when this choice is picked, by pointer, keyboard or its number key. */
  onSelect?: () => void;
  children: React.ReactNode;
  className?: string;
}) {
  const word = { idle: '', correct: 'Correct', wrong: 'Wrong', answer: 'The answer' }[result];
  return (
    <QuestionnaireChoice
      value={value}
      disabled={disabled}
      onChange={(e) => {
        if (e.target.checked) onSelect?.();
      }}
      data-result={result}
      className={cn(
        'min-h-14 items-center text-base transition-[transform,border-color,background-color] motion-safe:active:scale-[0.99]',
        'data-[result=correct]:border-success data-[result=correct]:bg-success/10',
        'data-[result=answer]:border-success data-[result=answer]:border-dashed',
        'data-[result=wrong]:border-danger data-[result=wrong]:bg-danger/10 motion-safe:data-[result=wrong]:animate-[shake_0.35s_ease-in-out]',
        className
      )}
    >
      <span className="flex items-center justify-between gap-3">
        <span className="font-mono">{children}</span>
        {word ? (
          <span
            className={cn(
              'inline-flex shrink-0 items-center gap-1 text-sm font-medium',
              result === 'wrong' ? 'text-danger' : 'text-success'
            )}
          >
            {result === 'wrong' ? (
              <XIcon
                aria-hidden
                className="size-4"
              />
            ) : (
              <CheckIcon
                aria-hidden
                className="size-4"
              />
            )}
            {word}
          </span>
        ) : null}
      </span>
    </QuestionnaireChoice>
  );
}
