'use client';

/**
 * The arena's games, as the select screen and the menus know them: a name, a
 * line, a mark, how to play, and what the game itself is. Each game is its own
 * component; this is the one list of them.
 */
import type { ComponentType, ReactNode } from 'react';
import { LuArrowDownUp, LuBlocks, LuFlag, LuKeyboard, LuWorm } from 'react-icons/lu';
import type { GameId } from '@/lib/arena/store';
import type { AlgorithmApi, Catalog } from '@/visualizer/core';
import { BricksGame } from './bricks-game';
import { KeysGame } from './keys-game';
import { PathRace } from './path-race';
import { SnakeGame } from './snake-game';
import { SwapRace } from './swap-race';

/** What a game is handed by the play screen. */
export type GameProps = {
  api: AlgorithmApi;
  catalog: Catalog;
  /** The play screen's pause: the game holds still while it is true. */
  paused: boolean;
  /** The score so far, for the play screen's bar. */
  onScore: (score: number) => void;
  /** The game is over: its score, and a line about it for the scores table. */
  onOver: (score: number, note?: string) => void;
};

export type GameInfo = {
  id: GameId;
  name: string;
  tagline: string;
  icon: ComponentType<{ className?: string; 'aria-hidden'?: boolean }>;
  /** The cartridge's color: a class for its mark. */
  tone: string;
  howTo: ReactNode;
  Game: ComponentType<GameProps>;
};

export const GAME_INFO: GameInfo[] = [
  {
    id: 'race',
    name: 'Path race',
    tagline: 'Drive from S to the flag before every pathfinder in the catalog.',
    icon: LuFlag,
    tone: 'text-accent',
    howTo: (
      <>
        <p>
          Drive with the arrow keys or W A S D, or press and drag toward the way you want to go.
          Entering a cell costs its weight: a plain cell is quick, a 9 is nine times as slow.
        </p>
        <p>
          Every pathfinder in the catalog races you, driving the path it found as the real Go code.
          Each first waits for the time it spent thinking, the cells it looked at, so a careful
          algorithm starts late on a better road. The first three places score 400, 250 and 150
          points, and each rival you beat adds 20.
        </p>
      </>
    ),
    Game: PathRace
  },
  {
    id: 'snake',
    name: 'Snake',
    tagline: 'Eat, grow, and race a pathfinder at its own game.',
    icon: LuWorm,
    tone: 'text-success',
    howTo: (
      <>
        <p>
          Steer with the arrow keys or W A S D (or the pad on a touch screen). Each piece of food is
          worth 10 points, more at a higher speed; running into a wall or yourself ends the run.
        </p>
        <p>
          In Settings, hand the snake to an <strong>autopilot</strong>: a pathfinder from the
          catalog, running as the real Go code, plans every move from the head to the food. Watch
          which one lasts longest. A piloted run scores a quarter of the points.
        </p>
      </>
    ),
    Game: SnakeGame
  },
  {
    id: 'bricks',
    name: 'Bricks',
    tagline: 'Falling blocks: fill rows to clear them, faster every level.',
    icon: LuBlocks,
    tone: 'text-viz-write',
    howTo: (
      <>
        <p>
          Left and right move the piece; up, X or R turns it clockwise as it falls and Z turns it
          back; down drops it a row and Space drops it to the bottom. With a mouse, move over the
          well to slide the piece, click to turn it and right click to drop it.
        </p>
        <p>
          A full row clears: 100 points for one, 300 for two, 500 for three and 800 for four at
          once, times the level. Every ten lines the level goes up and the pieces fall faster.
          Pieces come from a shuffled bag of all seven, so none is missing for long.
        </p>
      </>
    ),
    Game: BricksGame
  },
  {
    id: 'swap-race',
    name: 'Swap race',
    tagline: 'Sort the bars in fewer swaps than the sorting algorithms.',
    icon: LuArrowDownUp,
    tone: 'text-warning',
    howTo: (
      <>
        <p>
          Pick two bars to swap them (a click each, or the arrow keys and Space). Put them in order,
          shortest first, in as few swaps as you can.
        </p>
        <p>
          When the bars are sorted, every sort in the catalog runs on the same values, as the real
          Go code, and you see how many swaps each needed. Points for sorting at all, more the
          closer you come to the fewest swaps possible, and a bonus for each algorithm you beat.
        </p>
      </>
    ),
    Game: SwapRace
  },
  {
    id: 'keys',
    name: 'Keys',
    tagline: 'Learn the site’s shortcuts until your fingers know them.',
    icon: LuKeyboard,
    tone: 'text-accent',
    howTo: (
      <>
        <p>
          The game names an action; press its keys. The keyboard lights up as you go. A right answer
          scores 10 points, more as your streak grows; a wrong one breaks the streak.
        </p>
        <p>The round ends after the number of right answers set in Settings.</p>
      </>
    ),
    Game: KeysGame
  }
];

export const gameInfo = (id: string) => GAME_INFO.find((g) => g.id === id);
