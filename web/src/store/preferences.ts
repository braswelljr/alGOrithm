import { create } from 'zustand';
import { createJSONStorage, devtools, persist } from 'zustand/middleware';

/**
 * How a reader wants the site to look and behave (ported from
 * gsl-special-courses' preferences store).
 *
 * Kept in this browser's local storage and nowhere else: a choice applies the
 * moment it is made and the next visit is drawn the right way, with nothing
 * stored anywhere else. The theme is kept by next-themes, also in local
 * storage; it is here too so the settings screen reads one place.
 */
export type TextSize = 'standard' | 'large' | 'larger';
export type Density = 'compact' | 'standard' | 'comfortable';
export type MotionPreference = 'system' | 'reduce' | 'full';
export type Contrast = 'standard' | 'high';
export type Theme = 'system' | 'light' | 'dark';
export type TextSpacing = 'standard' | 'wide';
export type Focus = 'standard' | 'strong';

type PreferencesState = {
  theme: Theme;
  textSize: TextSize;
  density: Density;
  motion: MotionPreference;
  contrast: Contrast;
  underlineLinks: boolean;
  textSpacing: TextSpacing;
  focus: Focus;
  boldText: boolean;
  setTheme: (value: Theme) => void;
  setTextSize: (value: TextSize) => void;
  setDensity: (value: Density) => void;
  setMotion: (value: MotionPreference) => void;
  setContrast: (value: Contrast) => void;
  setUnderlineLinks: (value: boolean) => void;
  setTextSpacing: (value: TextSpacing) => void;
  setFocus: (value: Focus) => void;
  setBoldText: (value: boolean) => void;
  reset: () => void;
};

/** The preferences themselves, without the setters. */
export type Preferences = Pick<
  PreferencesState,
  | 'theme'
  | 'textSize'
  | 'density'
  | 'motion'
  | 'contrast'
  | 'underlineLinks'
  | 'textSpacing'
  | 'focus'
  | 'boldText'
>;

export const PREFERENCE_DEFAULTS: Preferences = {
  theme: 'system',
  textSize: 'standard',
  density: 'standard',
  motion: 'system',
  contrast: 'standard',
  underlineLinks: false,
  textSpacing: 'standard',
  focus: 'standard',
  boldText: false
};

export const usePreferencesStore = create<PreferencesState>()(
  devtools(
    persist(
      (set) => ({
        ...PREFERENCE_DEFAULTS,
        // Each choice lands here first. The theme is the exception: next-themes
        // stores it itself, and the site mirrors it in through `setTheme`.
        setTheme: (theme) => set({ theme }),
        setTextSize: (textSize) => set({ textSize }),
        setDensity: (density) => set({ density }),
        setMotion: (motion) => set({ motion }),
        setContrast: (contrast) => set({ contrast }),
        setUnderlineLinks: (underlineLinks) => set({ underlineLinks }),
        setTextSpacing: (textSpacing) => set({ textSpacing }),
        setFocus: (focus) => set({ focus }),
        setBoldText: (boldText) => set({ boldText }),
        reset: () => set(PREFERENCE_DEFAULTS)
      }),
      { name: 'algorithm:preferences', storage: createJSONStorage(() => localStorage) }
    ),
    { name: 'preferences' }
  )
);
