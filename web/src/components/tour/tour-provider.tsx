'use client';

import { useCallback, useEffect, useMemo, useRef, useState, type ReactNode } from 'react';
import { usePathname, useRouter } from 'next/navigation';
import { useHotkeys } from 'react-hotkeys-hook';
import { TourList } from '@/components/tour/tour-list';
import { TourOverlay } from '@/components/tour/tour-overlay';
import { keysOf } from '@/config/shortcuts';
import { isTypingTarget } from '@/lib/shortcuts';
import { tourMatches, tourRouteIsConcrete, type Tour, type TourStep } from '@/lib/tour/define';
import { readSeen, writeSeen } from '@/lib/tour/seen';
import { TOURS } from '@/lib/tour/tours';
import { TourContext, type TourControls } from '@/lib/tour/use-tour';

/** How long a tour waits for its page's parts to appear, and a step for its own. */
export const FIRST_WAIT_MS = 4000;
export const STEP_WAIT_MS = 1000;
const POLL_MS = 100;

/** The element a step points at, when it is on the page and shown. */
export function findTarget(selector: string): Element | null {
  let el: Element | null;
  try {
    el = document.querySelector(selector);
  } catch {
    return null;
  }
  if (!el || !el.isConnected) return null;
  if (el.closest('[hidden], [inert], [aria-hidden="true"]')) return null;
  const check = (el as Element & { checkVisibility?: () => boolean }).checkVisibility;
  if (typeof check === 'function' && !check.call(el)) return null;
  return el;
}

type Running = {
  tour: Tour;
  /** The steps whose parts were found, once the page has been looked at. */
  steps: TourStep[] | null;
  index: number;
  /** Which way the reader is going, so a step that vanished is passed in that direction. */
  direction: 1 | -1;
  /** When it was asked for, so a page that never arrives gives up. */
  startedAt: number;
};

/** The step on screen and the element it points at, swapped together so the card glides. */
type Shown = { index: number; target: Element };

type Props = {
  children: ReactNode;
  /** The tours on offer; every tour in `lib/tour/tours` unless a test says otherwise. */
  tours?: Tour[];
  /** A tour started once, automatically, on its own page, on a browser's first visit. */
  welcome?: string;
};

/**
 * Runs the guided tours for everything inside it: one tour at a time, its
 * steps over the page, the list of every tour, and which tours this browser
 * has seen (after gsl-certificates). Mounted once, by the root providers, so
 * the site layout and the docs layout share it.
 */
export function TourProvider({ children, tours = TOURS, welcome }: Props) {
  const router = useRouter();
  const pathname = usePathname() ?? '/';

  const [running, setRunning] = useState<Running | null>(null);
  const [shown, setShown] = useState<Shown | null>(null);
  const [listOpen, setListOpen] = useState(false);
  // Read after mounting: the exported HTML has no storage to read.
  const [seenIds, setSeenIds] = useState<Set<string>>(() => new Set());
  const [ready, setReady] = useState(false);
  const returnFocus = useRef<HTMLElement | null>(null);
  const runningRef = useRef(running);
  const pathRef = useRef(pathname);
  useEffect(() => {
    runningRef.current = running;
    pathRef.current = pathname;
  });

  useEffect(() => {
    setSeenIds(readSeen());
    setReady(true);
  }, []);

  const available = tours;
  const pageTours = useMemo(
    () => available.filter((tour) => tourMatches(tour, pathname)),
    [available, pathname]
  );

  const markSeen = useCallback((id: string) => {
    setSeenIds((current) => {
      if (current.has(id)) return current;
      const next = new Set(current).add(id);
      writeSeen(next);
      return next;
    });
  }, []);

  const finish = useCallback(() => {
    const current = runningRef.current;
    if (current) markSeen(current.tour.id);
    setRunning(null);
    setShown(null);
    const back = returnFocus.current;
    returnFocus.current = null;
    // After the card has gone, so focus is not dropped on the body.
    window.setTimeout(() => {
      if (back?.isConnected) back.focus({ preventScroll: true });
      else document.getElementById('main')?.focus({ preventScroll: true });
    }, 0);
  }, [markSeen]);

  const start = useCallback(
    (id: string) => {
      const tour = tours.find((t) => t.id === id);
      if (!tour) return;
      setListOpen(false);
      if (tour.route && !tourMatches(tour, pathRef.current)) {
        if (!tourRouteIsConcrete(tour)) return;
        router.push(`${tour.route.replace(/\/?$/, '/')}`);
      }
      returnFocus.current = null;
      setShown(null);
      setRunning({ tour, steps: null, index: 0, direction: 1, startedAt: Date.now() });
    },
    [tours, router]
  );

  const next = useCallback(() => {
    setRunning((current) =>
      current ? { ...current, index: current.index + 1, direction: 1 } : current
    );
  }, []);

  const back = useCallback(() => {
    setRunning((current) =>
      current && current.index > 0
        ? { ...current, index: current.index - 1, direction: -1 }
        : current
    );
  }, []);

  // Look at the page: keep the steps whose parts are there. Waits for the page
  // to arrive, and a moment more for parts drawn after it.
  useEffect(() => {
    if (!running || running.steps) return;
    // Started from elsewhere: wait for its page, not this one's parts.
    if (running.tour.route && !tourMatches(running.tour, pathname)) {
      const left = FIRST_WAIT_MS - (Date.now() - running.startedAt);
      const timer = window.setTimeout(
        () => setRunning((current) => (current === running ? null : current)),
        Math.max(left, 0)
      );
      return () => window.clearTimeout(timer);
    }
    const began = Date.now();
    const look = () => {
      const found = running.tour.steps.filter((s) => findTarget(s.target));
      const waited = Date.now() - began;
      if (found.length === running.tour.steps.length || (found.length > 0 && waited >= 600)) {
        setRunning((current) => (current === running ? { ...current, steps: found } : current));
        return true;
      }
      if (waited >= FIRST_WAIT_MS) {
        // Nothing of it on this page: leave quietly, unseen.
        setRunning((current) => (current === running ? null : current));
        return true;
      }
      return false;
    };
    if (look()) return;
    const timer = window.setInterval(() => {
      if (look()) window.clearInterval(timer);
    }, POLL_MS);
    return () => window.clearInterval(timer);
  }, [running, pathname]);

  // Going to another page ends the tour of this one.
  useEffect(() => {
    const current = runningRef.current;
    if (current?.steps && current.tour.route && !tourMatches(current.tour, pathname)) finish();
  }, [pathname, finish]);

  // Find this step's part; one that has gone is waited for briefly, then passed.
  useEffect(() => {
    if (!running?.steps) return;
    const { steps, index, direction } = running;
    if (index >= steps.length) {
      finish();
      return;
    }
    const step = steps[index];
    const began = Date.now();
    const look = () => {
      const el = findTarget(step.target);
      if (el) {
        if (!returnFocus.current) {
          const active = document.activeElement;
          returnFocus.current =
            active instanceof HTMLElement && active !== document.body ? active : null;
        }
        setShown({ index, target: el });
        return true;
      }
      if (Date.now() - began >= STEP_WAIT_MS) {
        const to = index + direction;
        if (to < 0) setRunning((c) => (c === running ? { ...c, index: 0, direction: 1 } : c));
        else if (to >= steps.length) finish();
        else setRunning((c) => (c === running ? { ...c, index: to } : c));
        return true;
      }
      return false;
    };
    if (look()) return;
    const timer = window.setInterval(() => {
      if (look()) window.clearInterval(timer);
    }, POLL_MS);
    return () => window.clearInterval(timer);
  }, [running, finish]);

  // The welcome tour, once per browser, on its own page and never again by
  // itself. Not in an embed (?embed=1), where the site is someone else's.
  useEffect(() => {
    if (!welcome || !ready || running) return;
    if (document.documentElement.dataset.embed === '1') return;
    const tour = tours.find((t) => t.id === welcome);
    if (!tour || seenIds.has(tour.id) || !tourMatches(tour, pathname)) return;
    const timer = window.setTimeout(() => {
      markSeen(tour.id);
      start(tour.id);
    }, 800);
    return () => window.clearTimeout(timer);
  }, [welcome, ready, running, tours, seenIds, pathname, markSeen, start]);

  const rememberFocus = useCallback((el: HTMLElement) => {
    if (el !== document.body && !el.closest('[data-tour-overlay]')) returnFocus.current = el;
  }, []);

  const reset = useCallback((id?: string) => {
    setSeenIds((current) => {
      const next = id ? new Set([...current].filter((s) => s !== id)) : new Set<string>();
      writeSeen(next);
      return next;
    });
  }, []);

  // Shift+T: the tour of this page, from anywhere on it.
  useHotkeys(
    keysOf('tour.page'),
    () => {
      if (pageTours[0]) start(pageTours[0].id);
    },
    { enableOnFormTags: false, ignoreEventWhen: (e) => isTypingTarget(e.target) },
    [pageTours, start]
  );

  const steps = running?.steps;
  const step = steps && shown ? (steps[shown.index] ?? null) : null;

  const controls = useMemo<TourControls>(
    () => ({
      start,
      stop: finish,
      next,
      back,
      active: running?.tour ?? null,
      step,
      index: step && shown ? shown.index : -1,
      total: steps?.length ?? 0,
      seen: (id) => seenIds.has(id),
      reset,
      tours: available,
      pageTours,
      openList: () => setListOpen(true)
    }),
    [start, finish, next, back, running, step, shown, steps, seenIds, reset, available, pageTours]
  );

  return (
    <TourContext.Provider value={controls}>
      {children}
      {step && shown && steps ? (
        <TourOverlay
          target={shown.target}
          step={step}
          index={shown.index}
          total={steps.length}
          onNext={next}
          onBack={back}
          onSkip={finish}
          onFocusOutside={rememberFocus}
        />
      ) : null}
      <TourList
        open={listOpen}
        onOpenChange={setListOpen}
      />
    </TourContext.Provider>
  );
}
