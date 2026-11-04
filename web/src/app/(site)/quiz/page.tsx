import { Suspense } from 'react';
import type { Metadata } from 'next';
import { QuizApp } from '@/components/quiz/quiz-app';
import { PageHeading } from '@/components/shared/page-header';
import { pageMetadata } from '@/lib/metadata';

export const metadata: Metadata = pageMetadata({
  title: 'Quiz',
  description:
    'Predict what an algorithm does next: which bars swap, which cell is explored, which process runs. Every answer is checked against the real run.',
  path: '/quiz/'
});

export default function QuizPage() {
  return (
    <>
      <PageHeading title="Quiz">
        Watch a real run stop just before a step and guess what happens next. Score points, build a
        streak, and beat your best.
      </PageHeading>
      <Suspense>
        <QuizApp />
      </Suspense>
    </>
  );
}
