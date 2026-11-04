import { defineTour } from '@/lib/tour/define';

/** The visualizer: pick an entry, set its data, and play the run. */
export const visualizeTour = defineTour({
  id: 'visualize',
  title: 'Tour of the visualizer',
  route: '/visualize',
  steps: [
    {
      target: '[data-tour="entry-title"]',
      title: 'What is running',
      body: 'The entry this page runs, with its kind, its level and its tags.'
    },
    {
      target: '[data-tour="entry-picker"]',
      title: 'Pick another',
      body: 'Choose any algorithm here. Its cost and a link to its docs sit beside it.'
    },
    {
      target: '[data-tour="configure"]',
      title: 'The data and settings',
      body: 'What the run works on. The cog opens a form to generate data, paste your own, load a file or address, and set the options; nothing changes until you apply.'
    },
    {
      target: '[data-tour="run-view"]',
      title: 'The run',
      body: 'Every comparison, swap, visit and step, drawn as it happens. A graph can be shown full screen.'
    },
    {
      target: '[data-tour="player-controls"]',
      title: 'Play, pause and step',
      body: 'Play the run, step it a move at a time, or scrub to any moment. Space plays and pauses; the arrow keys step.'
    },
    {
      target: '[data-tour="step-list"]',
      title: 'Every step, in words',
      body: 'The run as a list. Hide a kind of event with its chip, and select a step to jump to it.'
    },
    {
      target: '[data-tour="run-actions"]',
      title: 'Share and reuse',
      body: 'Copy a link to this exact step, the command that runs it in a terminal, or the Go that calls it; export the run, or embed it in another page.'
    }
  ]
});
