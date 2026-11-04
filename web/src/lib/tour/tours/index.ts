import type { Tour } from '@/lib/tour/define';
import { arenaTour } from './arena';
import { docsEntryTour } from './docs-entry';
import { quizTour } from './quiz';
import { shellTour } from './shell';
import { structuresTour } from './structures';
import { systemDesignTour } from './system-design';
import { tagsTour } from './tags';
import { visualizeTour } from './visualize';

/**
 * Every guided tour, in the order the list of tours shows them. A new tour
 * file in this folder is listed here; test/tour.test.ts fails on one that is
 * not, on a target that is not a `data-tour` name, and on a name no page carries.
 */
export const TOURS: Tour[] = [
  shellTour,
  visualizeTour,
  structuresTour,
  systemDesignTour,
  quizTour,
  arenaTour,
  docsEntryTour,
  tagsTour
];
