'use client';

import { createContext, useContext } from 'react';
import type { Tour, TourStep } from './define';

export type TourControls = {
  /** Starts a tour by id, going to its page first when it is somewhere else. */
  start: (id: string) => void;
  /** Ends the tour, marking it seen. */
  stop: () => void;
  next: () => void;
  back: () => void;
  /** The tour running now, or null. */
  active: Tour | null;
  /** The step showing now, or null while none is (between steps, or no tour). */
  step: TourStep | null;
  /** The step's place among the steps found on the page, from 0, and how many there are. */
  index: number;
  total: number;
  /** Whether this browser has seen the tour. */
  seen: (id: string) => boolean;
  /** Forgets that one tour, or every tour, was seen. */
  reset: (id?: string) => void;
  /** Every tour, wherever it is. */
  tours: Tour[];
  /** The tours of the page open now. */
  pageTours: Tour[];
  /** Opens the list of every tour. */
  openList: () => void;
};

const noop = () => {};

/** Outside a provider (a test) the tours do nothing. */
export const TourContext = createContext<TourControls>({
  start: noop,
  stop: noop,
  next: noop,
  back: noop,
  active: null,
  step: null,
  index: -1,
  total: 0,
  seen: () => false,
  reset: noop,
  tours: [],
  pageTours: [],
  openList: noop
});

/** The tours: start one, move through it, and ask which have been seen. */
export function useTour(): TourControls {
  return useContext(TourContext);
}
