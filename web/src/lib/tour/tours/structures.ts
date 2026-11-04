import { defineTour } from '@/lib/tour/define';

/** The data structure playground. */
export const structuresTour = defineTour({
  id: 'structures',
  title: 'Tour of the structure playground',
  route: '/structures',
  steps: [
    {
      target: '[data-tour="entry-picker"]',
      title: 'Pick a structure',
      body: 'Stacks, queues, heaps, trees, tries, caches and more, each the real Go code.'
    },
    {
      target: '[data-tour="configure"]',
      title: 'What it starts with',
      body: 'The values it holds before your first operation, and its options. The cog opens the form.'
    },
    {
      target: '[data-tour="operations"]',
      title: 'Operations',
      body: 'Queue push, pop, insert, search and the rest, take any one back, then apply them; the structure plays them in order.'
    },
    {
      target: '[data-tour="run-view"]',
      title: 'The structure, as it changes',
      body: 'Each operation redraws it. Step back through them to see every change, or show it full screen.'
    }
  ]
});
