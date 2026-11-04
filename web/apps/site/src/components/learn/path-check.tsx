'use client';

import * as React from 'react';
import { CheckCircle2Icon } from 'lucide-react';
import { Button } from '@braswelljr/algorithm-ui/components/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@braswelljr/algorithm-ui/components/card';
import {
  Questionnaire,
  QuestionnaireActions,
  QuestionnaireChoice,
  QuestionnaireChoices,
  QuestionnaireItem,
  QuestionnaireNext,
  QuestionnairePrevious,
  QuestionnaireProgress,
  QuestionnaireSubmit,
  QuestionnaireTitle
} from '@braswelljr/algorithm-ui/components/questionnaire';
import { markDone, readDone } from '@/lib/learn-progress';

/**
 * The end of a learning path: three questions to look back on it honestly,
 * then the path is marked done (in this browser). Answers are not scored;
 * they point back to the steps worth another run.
 */
export function PathCheck({ id, title }: { id: string; title: string }) {
  const [done, setDone] = React.useState(false);
  const [open, setOpen] = React.useState(false);
  React.useEffect(() => setDone(readDone().includes(id)), [id]);

  if (done) {
    return (
      <Card variant="selected" size="md" className="flex-row flex-wrap items-center justify-between gap-3 px-(--card-spacing)">
        <span className="flex items-center gap-2 font-medium">
          <CheckCircle2Icon aria-hidden className="size-5 text-success" />
          You finished {title}.
        </span>
        <Button
          variant="ghost"
          onClick={() => {
            markDone(id, false);
            setDone(false);
          }}
        >
          Mark as not done
        </Button>
      </Card>
    );
  }
  if (!open) {
    return (
      <Card variant="ghost" size="md" className="flex-row flex-wrap items-center justify-between gap-3 px-(--card-spacing)">
        <span>Ran every step? Check what stuck, then mark the path done.</span>
        <Button variant="primary" onClick={() => setOpen(true)}>
          Finish this path
        </Button>
      </Card>
    );
  }
  return (
    <Card size="lg">
      <CardHeader>
        <CardTitle>Looking back on {title}</CardTitle>
        <CardDescription>Be honest; nothing is scored.</CardDescription>
      </CardHeader>
      <CardContent>
        <Questionnaire
          shortcuts="numbers"
          onSubmit={(e) => {
            e.preventDefault();
            markDone(id);
            setDone(true);
          }}
        >
          <QuestionnaireProgress />
          <QuestionnaireItem name="ran" required>
            <QuestionnaireTitle>Did you run every step?</QuestionnaireTitle>
            <QuestionnaireChoices>
              <QuestionnaireChoice value="all">Every one</QuestionnaireChoice>
              <QuestionnaireChoice value="most">Most of them</QuestionnaireChoice>
              <QuestionnaireChoice value="few">A few</QuestionnaireChoice>
            </QuestionnaireChoices>
          </QuestionnaireItem>
          <QuestionnaireItem name="explain" required>
            <QuestionnaireTitle>Could you explain each step&apos;s key idea to a friend?</QuestionnaireTitle>
            <QuestionnaireChoices>
              <QuestionnaireChoice value="yes">Yes, all of them</QuestionnaireChoice>
              <QuestionnaireChoice value="some">Some of them</QuestionnaireChoice>
              <QuestionnaireChoice value="not-yet">Not yet</QuestionnaireChoice>
            </QuestionnaireChoices>
          </QuestionnaireItem>
          <QuestionnaireItem name="cost">
            <QuestionnaireTitle>Could you say what each one costs, and why?</QuestionnaireTitle>
            <QuestionnaireChoices>
              <QuestionnaireChoice value="yes">Yes</QuestionnaireChoice>
              <QuestionnaireChoice value="roughly">Roughly</QuestionnaireChoice>
              <QuestionnaireChoice value="no">Not yet: the Big O page will help</QuestionnaireChoice>
            </QuestionnaireChoices>
          </QuestionnaireItem>
          <QuestionnaireActions>
            <QuestionnairePrevious variant="ghost" />
            <QuestionnaireNext />
            <QuestionnaireSubmit>Mark the path done</QuestionnaireSubmit>
          </QuestionnaireActions>
        </Questionnaire>
      </CardContent>
    </Card>
  );
}
