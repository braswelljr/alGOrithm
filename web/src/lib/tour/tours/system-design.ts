import { defineTour } from '@/lib/tour/define';

/** The system design hub. */
export const systemDesignTour = defineTour({
  id: 'system-design',
  title: 'Tour of system design',
  route: '/system-design',
  steps: [
    {
      target: '[data-tour="item-list"]',
      title: 'Case studies and patterns',
      body: 'Every case study, data pattern and sizing tool. Search them, and choose one to open it beside the list.',
      placement: 'right'
    },
    {
      target: '[data-tour="item-detail"]',
      title: 'The design',
      body: 'Its requirements, estimate, trade-offs and failure modes, and a live architecture you can change with knobs. It scrolls on its own; the list stays put.',
      placement: 'left'
    }
  ]
});
