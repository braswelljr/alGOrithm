import { defineTour } from '@/lib/tour/define';

/** The editor: files, code, the problems and terminals, Run and the settings. */
export const editorTour = defineTour({
  id: 'editor',
  title: 'Tour of the editor',
  route: '/editor',
  steps: [
    {
      target: '[data-tour="editor-explorer"]',
      title: 'Your files and the library',
      body: 'Your home folder and every package of alGOrithm. Right-click for new, rename, duplicate, delete, copy path, download and open in a terminal.'
    },
    {
      target: '[data-tour="editor-code"]',
      title: 'The code',
      body: 'Go, shell, Makefiles and more, highlighted. Problems are flagged as you type; Ctrl+S saves (and formats), Shift+Alt+F formats.'
    },
    {
      target: '[data-tour="editor-terminal"]',
      title: 'Problems and terminals',
      body: 'The problems in the open files, and as many terminals as you like, each with gnux, go run, bash and make.'
    },
    {
      target: '[data-tour="editor-run"]',
      title: 'Run',
      body: 'Saves and runs the file: Go on the Go Playground, scripts and Makefiles here. With errors, it shows them instead.'
    },
    {
      target: '[data-tour="editor-settings"]',
      title: 'Settings',
      body: 'Format on save, stop on errors, problem flags, style hints, word wrap and text size, kept in this browser.'
    }
  ]
});
