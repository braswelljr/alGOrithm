import { create } from 'zustand';
import { createJSONStorage, persist } from 'zustand/middleware';

/** The editor's settings, as its settings menu shows them. */
export type EditorSettings = {
  /** Format a file when it is saved: gofmt for Go, tidy whitespace for all. */
  formatOnSave: boolean;
  /** Check files as they are edited, flag problems and list them. */
  lint: boolean;
  /** Refuse to run a file that has errors, and show them instead. */
  stopOnErrors: boolean;
  /** Mark trailing spaces and a missing final newline too. */
  styleHints: boolean;
  wordWrap: boolean;
  fontSize: 12 | 13 | 14 | 16;
};

export const DEFAULT_EDITOR_SETTINGS: EditorSettings = {
  formatOnSave: true,
  lint: true,
  stopOnErrors: true,
  styleHints: true,
  wordWrap: false,
  fontSize: 13
};

type State = EditorSettings & {
  set: (patch: Partial<EditorSettings>) => void;
  reset: () => void;
};

/** Kept in this browser (`algorithm:editor-settings`), as a convenience. */
export const useEditorSettings = create<State>()(
  persist(
    (set) => ({
      ...DEFAULT_EDITOR_SETTINGS,
      set: (patch) => set(patch),
      reset: () => set(DEFAULT_EDITOR_SETTINGS)
    }),
    {
      name: 'algorithm:editor-settings',
      version: 1,
      storage: createJSONStorage(() => localStorage),
      partialize: ({ set: _set, reset: _reset, ...rest }) => rest
    }
  )
);
