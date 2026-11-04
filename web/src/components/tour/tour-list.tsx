'use client';

import { usePathname } from 'next/navigation';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogBody,
  DialogDescription,
  DialogHeader,
  DialogPopup,
  DialogTitle
} from '@/components/ui/dialog';
import { tourMatches, tourRouteIsConcrete, type Tour } from '@/lib/tour/define';
import { useTour } from '@/lib/tour/use-tour';

/** Where a tour can be taken from, in words. */
function whereFrom(tour: Tour, here: boolean): string {
  if (here) return 'On this page';
  if (!tour.route) return 'Offered on its own page';
  if (tourRouteIsConcrete(tour)) return 'Opens its page first';
  return 'Open one first, then choose the tour of this page from Help';
}

/**
 * Every tour, the ones for this page first. Choosing one goes to its page
 * when it is somewhere else; a tour of one entry's page is taken from it.
 */
export function TourList({
  open,
  onOpenChange
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}) {
  const { tours, start, seen } = useTour();
  const pathname = usePathname() ?? '/';
  const sorted = [...tours].sort(
    (a, b) => Number(tourMatches(b, pathname)) - Number(tourMatches(a, pathname))
  );

  return (
    <Dialog
      open={open}
      onOpenChange={onOpenChange}
    >
      <DialogPopup className="max-w-lg">
        <DialogHeader>
          <DialogTitle>Tours</DialogTitle>
          <DialogDescription>
            Short walks through a page, a step at a time. Press Escape to leave one at any time.
          </DialogDescription>
        </DialogHeader>
        <DialogBody>
          {sorted.length < 1 ? (
            <p className="text-muted text-sm">There are no tours yet.</p>
          ) : (
            <ul className="divide-border divide-y">
              {sorted.map((tour) => {
                const here = tourMatches(tour, pathname);
                const startable = here || tourRouteIsConcrete(tour);
                return (
                  <li
                    key={tour.id}
                    className="flex items-center justify-between gap-4 py-2.5"
                  >
                    <div className="min-w-0">
                      <p className="text-sm font-medium">{tour.title}</p>
                      <p className="text-muted text-sm">
                        {whereFrom(tour, here)}
                        {seen(tour.id) ? ', taken before' : ''}
                      </p>
                    </div>
                    {startable ? (
                      <Button
                        type="button"
                        variant="secondary"
                        size="sm"
                        aria-label={`Start the ${tour.title.toLowerCase()}`}
                        onClick={() => start(tour.id)}
                      >
                        Start
                      </Button>
                    ) : null}
                  </li>
                );
              })}
            </ul>
          )}
        </DialogBody>
      </DialogPopup>
    </Dialog>
  );
}
