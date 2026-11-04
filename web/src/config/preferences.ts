import type {
  Contrast,
  Density,
  Focus,
  MotionPreference,
  TextSize,
  TextSpacing
} from '@/store/preferences';

/**
 * The choices the Appearance and Accessibility settings offer, each with the
 * sentence that says what it does (after gsl-special-courses).
 */

export const THEME_OPTIONS: { value: 'light' | 'dark' | 'system'; label: string; hint: string }[] =
  [
    { value: 'system', label: 'Match this device', hint: 'Light or dark, following the device.' },
    { value: 'light', label: 'Light', hint: 'Dark text on a light ground.' },
    { value: 'dark', label: 'Dark', hint: 'Light text on a dark ground.' }
  ];

export const DENSITY_OPTIONS: { value: Density; label: string; hint: string }[] = [
  { value: 'compact', label: 'Compact', hint: 'Tighter spacing, more rows on the screen.' },
  { value: 'standard', label: 'Standard', hint: 'The spacing the site was designed at.' },
  { value: 'comfortable', label: 'Comfortable', hint: 'More room around controls and rows.' }
];

export const TEXT_SIZE_OPTIONS: { value: TextSize; label: string; hint: string }[] = [
  { value: 'standard', label: 'Standard', hint: 'The size the site was designed at.' },
  { value: 'large', label: 'Large', hint: 'Text a step larger, with controls to match.' },
  { value: 'larger', label: 'Larger', hint: 'Text two steps larger, for reading at a distance.' }
];

export const MOTION_OPTIONS: { value: MotionPreference; label: string; hint: string }[] = [
  {
    value: 'system',
    label: 'Match this device',
    hint: 'Reduced when the device asks for less motion.'
  },
  { value: 'reduce', label: 'Reduce motion', hint: 'Things change in place instead of moving.' },
  { value: 'full', label: 'Full motion', hint: 'Transitions play whatever the device says.' }
];

export const CONTRAST_OPTIONS: { value: Contrast; label: string; hint: string }[] = [
  { value: 'standard', label: 'Standard', hint: 'Secondary text and borders as designed.' },
  {
    value: 'high',
    label: 'Higher contrast',
    hint: 'Secondary text and borders darker, so less fades.'
  }
];

export const TEXT_SPACING_OPTIONS: { value: TextSpacing; label: string; hint: string }[] = [
  { value: 'standard', label: 'Standard', hint: 'Letters and lines as designed.' },
  {
    value: 'wide',
    label: 'Wider spacing',
    hint: 'More room between letters, words and lines, which some readers track more easily.'
  }
];

export const FOCUS_OPTIONS: { value: Focus; label: string; hint: string }[] = [
  { value: 'standard', label: 'Standard', hint: 'A soft ring around what the keyboard is on.' },
  {
    value: 'strong',
    label: 'Strong',
    hint: 'A thick outline that shows on every surface, for finding your place.'
  }
];
