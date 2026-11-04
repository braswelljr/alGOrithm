import { defineTour } from '@/lib/tour/define';

/** The game arena's select screen. */
export const arenaTour = defineTour({
  id: 'arena',
  title: 'Tour of the arena',
  route: '/shortcuts',
  steps: [
    {
      target: '[data-tour="game-select"]',
      title: 'Choose a game',
      body: 'Each game is played against the real algorithms. Arrows choose, Enter opens its menu: play, how to play, settings and scores. Your points and rank stay in this browser.'
    },
    {
      target: '[data-tour="game-records"]',
      title: 'Your record',
      body: 'Each game\u2019s high score, how often you played it and your last score, with the latest games below. It all stays in this browser.'
    }
  ]
});
