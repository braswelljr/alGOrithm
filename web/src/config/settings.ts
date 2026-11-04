/**
 * The Settings tabs, in order. Each is a route under /settings/, so a reader
 * can be sent straight to one ("turn on larger text in Settings,
 * Accessibility"), as in gsl-certificates.
 */
export type SettingsSection = { to: string; label: string; description: string };

export const SETTINGS_SECTIONS: SettingsSection[] = [
  {
    to: 'appearance',
    label: 'Appearance',
    description: 'The theme, and how much fits on the screen. Kept in this browser.'
  },
  {
    to: 'accessibility',
    label: 'Accessibility',
    description: 'Text size, spacing, focus, motion and contrast. Kept in this browser.'
  }
];

/** The tab /settings/ opens on. */
export const SETTINGS_FIRST = SETTINGS_SECTIONS[0].to;
