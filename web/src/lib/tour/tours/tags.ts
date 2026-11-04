import { defineTour } from '@/lib/tour/define';

/** The tags page of the docs. */
export const tagsTour = defineTour({
  id: 'tags',
  title: 'Tour of the tags',
  route: '/docs/tags',
  steps: [
    {
      target: '[data-tour="item-list"]',
      title: 'Every tag',
      body: 'What each tag means and how many entries carry it. Search by name or meaning.',
      placement: 'right'
    },
    {
      target: '[data-tour="item-detail"]',
      title: 'Its entries',
      body: 'Every entry carrying the chosen tag, each linking to its docs; or open them all in the catalog.',
      placement: 'left'
    }
  ]
});
