import { defineTour } from '@/lib/tour/define';

/** One entry's documentation page. */
export const docsEntryTour = defineTour({
  id: 'docs-entry',
  title: 'Tour of an entry’s docs',
  route: '/docs/:id',
  steps: [
    {
      target: '[data-tour="sidebar"]',
      title: 'Every entry',
      body: 'The docs of every domain, category and entry; the one open now is filled in.',
      placement: 'right'
    },
    {
      target: '[data-tour="entry-actions"]',
      title: 'Try it, quiz it, measure it',
      body: 'Open the entry in the visualizer, be quizzed on it, or measure its cost against its stated Big O.'
    }
  ]
});
