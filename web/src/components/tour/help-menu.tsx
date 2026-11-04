'use client';

/**
 * The top bar's Help (after gsl-certificates): the guides, the keyboard
 * shortcuts, the tour of this page when it has one, every tour, and the game
 * arena. The shortcuts also open on `?`, and the page's tour on Shift+T.
 */
import Link from 'next/link';
import { Book1, Category, Command, Game, Map as MapIcon, MessageQuestion } from 'iconsax-reactjs';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuShortcut,
  DropdownMenuTrigger
} from '@/components/ui/dropdown-menu';
import { repoUrl } from '@/config/site';
import { useShortcuts } from '@/context/use-shortcuts';
import { useTour } from '@/lib/tour/use-tour';

export function HelpMenu() {
  const { openHelp } = useShortcuts();
  const { pageTours, tours, start, openList } = useTour();
  const tour = pageTours[0];

  return (
    <DropdownMenu>
      <DropdownMenuTrigger
        render={
          <Button
            variant="ghost"
            size="icon-sm"
            data-tour="help-menu"
            aria-label="Help"
            className="text-muted hover:text-fg"
          />
        }
      >
        <MessageQuestion
          aria-hidden
          size="1.1em"
        />
      </DropdownMenuTrigger>
      <DropdownMenuContent
        align="end"
        className="w-64"
      >
        <DropdownMenuGroup>
          <DropdownMenuLabel>Help</DropdownMenuLabel>
          <DropdownMenuItem render={<a href={`${repoUrl}/tree/main/docs/guides`} />}>
            <Book1
              aria-hidden
              className="size-4"
            />
            <span>The guides</span>
          </DropdownMenuItem>
          <DropdownMenuItem onClick={openHelp}>
            <Command
              aria-hidden
              className="size-4"
            />
            <span>Keyboard shortcuts</span>
            <DropdownMenuShortcut>?</DropdownMenuShortcut>
          </DropdownMenuItem>
          <DropdownMenuItem render={<Link href="/shortcuts/" />}>
            <Game
              aria-hidden
              className="size-4"
            />
            <span>The game arena</span>
          </DropdownMenuItem>
        </DropdownMenuGroup>
        {tour || tours.length > 0 ? <DropdownMenuSeparator /> : null}
        <DropdownMenuGroup>
          {tour ? (
            <DropdownMenuItem onClick={() => start(tour.id)}>
              <MapIcon
                aria-hidden
                className="size-4"
              />
              <span>Tour of this page</span>
              <DropdownMenuShortcut>⇧T</DropdownMenuShortcut>
            </DropdownMenuItem>
          ) : null}
          {tours.length > 0 ? (
            <DropdownMenuItem onClick={openList}>
              <Category
                aria-hidden
                className="size-4"
              />
              <span>All tours</span>
            </DropdownMenuItem>
          ) : null}
        </DropdownMenuGroup>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
