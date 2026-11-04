'use client';

import * as React from 'react';
import { CompassIcon, RotateCcwIcon } from 'lucide-react';
import type { LearningPath } from '@braswelljr/algorithm-visualizer/core';
import { LevelBadge } from '@braswelljr/algorithm-ui/algorithm/level-badge';
import { QuizChoiceCard } from '@braswelljr/algorithm-ui/algorithm/quiz-choice-card';
import { Button } from '@braswelljr/algorithm-ui/components/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@braswelljr/algorithm-ui/components/card';
import {
  Questionnaire,
  QuestionnaireActions,
  QuestionnaireChoices,
  QuestionnaireDescription,
  QuestionnaireItem,
  QuestionnaireNext,
  QuestionnairePrevious,
  QuestionnaireProgress,
  QuestionnaireSubmit,
  QuestionnaireTitle
} from '@braswelljr/algorithm-ui/components/questionnaire';
import { readPlacement, savePlacement, type Placement } from '@/lib/learn-progress';
import { pathHref } from '@/lib/learn';

type Q = {
  id: string;
  level: Placement['level'];
  prompt: string;
  hint?: string;
  choices: string[];
  answer: number;
};

/**
 * The placement questions: concepts the catalog's entries teach, from the
 * fundamentals up. The harder ones only appear when the easier ones were
 * answered right (conditional questions on the Questionnaire).
 */
const questions: Q[] = [
  {
    id: 'lifo-order',
    level: 'beginner',
    prompt: 'You push 1, 2 and 3 onto a stack and pop once. What comes out?',
    choices: ['1', '2', '3', 'Nothing: stacks do not pop'],
    answer: 2
  },
  {
    id: 'halving',
    level: 'beginner',
    prompt: 'Binary search needs the list to be...',
    choices: ['Sorted', 'Short', 'Free of duplicates', 'Stored in a hash table'],
    answer: 0
  },
  {
    id: 'growth',
    level: 'intermediate',
    prompt: 'Doubling the input of an O(n^2) algorithm makes it take about...',
    hint: 'Think of the square.',
    choices: ['The same time', 'Twice as long', 'Four times as long', 'n times as long'],
    answer: 2
  },
  {
    id: 'negative-edges',
    level: 'intermediate',
    prompt: 'Which graph breaks Dijkstra’s algorithm?',
    choices: ['One with cycles', 'One with a negative edge', 'A directed one', 'One with ten thousand nodes'],
    answer: 1
  },
  {
    id: 'bad-pivot',
    level: 'advanced',
    prompt: 'Quick sort that always picks the first value as its pivot meets sorted input. Its cost is...',
    choices: ['O(n)', 'O(n log n)', 'O(n^2)', 'O(2^n)'],
    answer: 2
  },
  {
    id: 'majorities',
    level: 'advanced',
    prompt: 'Why does Raft commit an entry once a majority stores it?',
    choices: [
      'A majority is faster to reach',
      'Any two majorities overlap, so a new leader has it',
      'The leader stops waiting after half',
      'Disks fail less often in a majority'
    ],
    answer: 1
  }
];

const order: Placement['level'][] = ['beginner', 'intermediate', 'advanced'];

/** The level the answers place a learner at: the last level fully answered right, plus one. */
function place(answers: Record<string, string>): Placement['level'] {
  const right = (q: Q) => answers[q.id] === String(q.answer);
  const passed = (level: Placement['level']) => questions.filter((q) => q.level === level).every(right);
  if (!passed('beginner')) return 'beginner';
  if (!passed('intermediate')) return 'intermediate';
  return passed('advanced') ? 'advanced' : 'intermediate';
}

/**
 * "Where should I start?": a short placement test on the shadcn Questionnaire,
 * one question on screen, the harder questions only when the easier ones were
 * right. It recommends a level and the paths for it, and remembers the result
 * in this browser so the paths below are marked.
 */
export function PlacementTest({
  paths,
  onPlaced
}: {
  paths: LearningPath[];
  onPlaced?: (p: Placement) => void;
}) {
  const [answers, setAnswers] = React.useState<Record<string, string>>({});
  const [result, setResult] = React.useState<Placement | null>(null);
  const [open, setOpen] = React.useState(false);
  React.useEffect(() => setResult(readPlacement()), []);

  const reached = (level: Placement['level']) => {
    const i = order.indexOf(level);
    return order
      .slice(0, i)
      .every((l) => questions.filter((q) => q.level === l).every((q) => answers[q.id] === String(q.answer)));
  };
  const shown = questions.filter((q) => reached(q.level));
  const recommended = result ? paths.filter((p) => p.level === result.level).slice(0, 3) : [];

  if (result && !open) {
    return (
      <Card variant="accent" size="md" className="px-(--card-spacing)">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <p className="flex items-center gap-2">
            <CompassIcon aria-hidden className="size-5 text-accent" />
            Start at <LevelBadge level={result.level} />
          </p>
          <Button
            variant="ghost"
            onClick={() => {
              setAnswers({});
              setOpen(true);
            }}
          >
            <RotateCcwIcon aria-hidden />
            Take the test again
          </Button>
        </div>
        {recommended.length ? (
          <ul className="flex flex-wrap gap-2 text-sm">
            {recommended.map((p) => (
              <li key={p.id}>
                <a href={pathHref(p.id)} className="text-accent underline-offset-4 hover:underline">
                  {p.title}
                </a>
              </li>
            ))}
          </ul>
        ) : null}
      </Card>
    );
  }

  if (!open) {
    return (
      <Card variant="accent" size="md" className="flex-row flex-wrap items-center justify-between gap-3 px-(--card-spacing)">
        <span>
          <span className="font-medium">Where should I start?</span>{' '}
          <span className="text-muted">Six quick questions place you at a level and pick your first paths.</span>
        </span>
        <Button variant="primary" onClick={() => setOpen(true)}>
          <CompassIcon aria-hidden />
          Take the placement test
        </Button>
      </Card>
    );
  }

  return (
    <Card size="lg">
      <CardHeader>
        <CardTitle>Where should I start?</CardTitle>
        <CardDescription>Answer what you can. Number keys pick an answer; Enter moves on.</CardDescription>
      </CardHeader>
      <CardContent>
        <Questionnaire
          shortcuts="numbers"
          onSubmit={(e) => {
            e.preventDefault();
            const level = place(answers);
            savePlacement(level);
            const p = { level, at: new Date().toISOString() };
            setResult(p);
            setOpen(false);
            onPlaced?.(p);
          }}
        >
          <QuestionnaireProgress />
          {shown.map((q) => (
            <QuestionnaireItem key={q.id} name={q.id} required>
              <QuestionnaireTitle>{q.prompt}</QuestionnaireTitle>
              {q.hint ? <QuestionnaireDescription>{q.hint}</QuestionnaireDescription> : null}
              <QuestionnaireChoices>
                {q.choices.map((c, i) => (
                  <QuizChoiceCard
                    key={i}
                    value={String(i)}
                    onSelect={() => setAnswers((a) => ({ ...a, [q.id]: String(i) }))}
                  >
                    {c}
                  </QuizChoiceCard>
                ))}
              </QuestionnaireChoices>
            </QuestionnaireItem>
          ))}
          <QuestionnaireActions>
            <QuestionnairePrevious variant="ghost" />
            <QuestionnaireNext />
            <QuestionnaireSubmit>See my level</QuestionnaireSubmit>
          </QuestionnaireActions>
        </Questionnaire>
      </CardContent>
    </Card>
  );
}
