import { create } from 'zustand';
import { createJSONStorage, persist } from 'zustand/middleware';

/** The arena's games. */
export const GAMES = ['race', 'snake', 'bricks', 'swap-race', 'keys'] as const;
export type GameId = (typeof GAMES)[number];

/** Each game's settings, as its menu sets them. */
export type GameSettings = {
  race: { map: string; size: 'small' | 'medium' | 'large' };
  bricks: Record<string, never>;
  snake: { speed: 'calm' | 'brisk' | 'fast'; walls: boolean; pilot: string | null };
  'swap-race': { size: number };
  keys: { rounds: number };
};

export const DEFAULT_SETTINGS: GameSettings = {
  race: { map: 'weighted', size: 'medium' },
  bricks: {},
  snake: { speed: 'brisk', walls: true, pilot: null },
  'swap-race': { size: 8 },
  keys: { rounds: 10 }
};

/** One finished game, for the scores table. */
export type Score = { game: GameId; score: number; at: number; note?: string };

type ArenaState = {
  /** Every point earned, in every game. */
  points: number;
  best: Partial<Record<GameId, number>>;
  played: Partial<Record<GameId, number>>;
  /** The latest finished games, newest first, at most 100. */
  history: Score[];
  settings: GameSettings;
  /** Adds a finished game; says whether it beat the best. */
  record: (game: GameId, score: number, note?: string) => { best: boolean };
  /** Points earned outside a finished game (a right shortcut). */
  award: (points: number) => void;
  setSettings: <G extends GameId>(game: G, patch: Partial<GameSettings[G]>) => void;
  reset: () => void;
};

const empty = { points: 0, best: {}, played: {}, history: [], settings: DEFAULT_SETTINGS };

/**
 * The arena's points, bests, history and settings, kept in this browser
 * (`algorithm:arena`). Nothing is sent anywhere: there are no accounts and no
 * leaderboard, only the visitor's own record.
 */
export const useArena = create<ArenaState>()(
  persist(
    (set, get) => ({
      ...empty,
      record(game, score, note) {
        const before = get().best[game] ?? 0;
        const best = score > before;
        set((s) => ({
          points: s.points + Math.max(0, Math.round(score)),
          best: best ? { ...s.best, [game]: score } : s.best,
          played: { ...s.played, [game]: (s.played[game] ?? 0) + 1 },
          history: [{ game, score, at: Date.now(), note }, ...s.history].slice(0, 100)
        }));
        return { best: best && score > 0 };
      },
      award(points) {
        set((s) => ({ points: s.points + Math.max(0, Math.round(points)) }));
      },
      setSettings(game, patch) {
        set((s) => ({ settings: { ...s.settings, [game]: { ...s.settings[game], ...patch } } }));
      },
      reset() {
        set({ ...empty });
      }
    }),
    {
      name: 'algorithm:arena',
      version: 1,
      storage: createJSONStorage(() => localStorage),
      // A stored record missing a newer game's settings takes the defaults.
      merge: (stored, current) => {
        const s = (stored ?? {}) as Partial<ArenaState>;
        return {
          ...current,
          ...s,
          settings: {
            race: { ...DEFAULT_SETTINGS.race, ...s.settings?.race },
            bricks: {},
            snake: { ...DEFAULT_SETTINGS.snake, ...s.settings?.snake },
            'swap-race': { ...DEFAULT_SETTINGS['swap-race'], ...s.settings?.['swap-race'] },
            keys: { ...DEFAULT_SETTINGS.keys, ...s.settings?.keys }
          }
        };
      }
    }
  )
);
