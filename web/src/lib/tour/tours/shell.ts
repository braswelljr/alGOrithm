import { defineTour } from '@/lib/tour/define';

/** The site's frame and the catalog: started once, by itself, on a browser's first visit. */
export const shellTour = defineTour({
  id: 'shell',
  title: 'Tour of the site',
  route: '/',
  steps: [
    {
      target: '[data-tour="sidebar"]',
      title: 'The menu',
      body: 'Every part of the site, grouped by what it is for: learn, run, concepts and reference. Press Ctrl+B (Cmd+B on a Mac) to fold it away.',
      placement: 'right'
    },
    {
      target: '[data-tour="search"]',
      title: 'Search and go anywhere',
      body: 'Find any algorithm, data structure or page, or start a task. Press Ctrl+K (Cmd+K on a Mac) from anywhere.',
      placement: 'bottom'
    },
    {
      target: '[data-tour="catalog-filters"]',
      title: 'Narrow the catalog',
      body: 'Filter every entry by kind, level and tag. The filters stay in the link, so a filtered view can be shared.'
    },
    {
      target: '[data-tour="catalog-entries"]',
      title: 'Every entry',
      body: 'Each card opens the entry in the visualizer, running the real Go code in your browser. Its badges say what it is and how hard.'
    },
    {
      target: '[data-tour="theme"]',
      title: 'Light or dark',
      body: 'Switch the theme here, or press T. Text size, density, contrast and motion are in Settings.',
      placement: 'bottom'
    },
    {
      target: '[data-tour="help-menu"]',
      title: 'Help',
      body: 'The guides, the keyboard shortcuts, the game arena and a tour of whichever page you are on. Press ? for the shortcuts.',
      placement: 'bottom'
    }
  ]
});
