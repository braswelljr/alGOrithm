import { defineTour } from '@/lib/tour/define';

/** The quiz's start screen. */
export const quizTour = defineTour({
  id: 'quiz',
  title: 'Tour of the quiz',
  route: '/quiz',
  steps: [
    {
      target: '[data-tour="quiz-topic"]',
      title: 'What to be asked about',
      body: 'One domain, or a mix of everything.'
    },
    {
      target: '[data-tour="quiz-difficulty"]',
      title: 'How hard',
      body: 'Beginner, intermediate or advanced entries, or any.'
    },
    {
      target: '[data-tour="quiz-mode"]',
      title: 'How to play',
      body: 'Ten questions and three lives, as many as you can in a minute, or until the first miss. Your best in each stays in this browser.'
    },
    {
      target: '[data-tour="quiz-start"]',
      title: 'Start',
      body: 'Each question stops a real run just before a step; guess what happens next. Keys 1 to 4 answer.'
    }
  ]
});
