'use client';

import type { ReactNode } from 'react';
import { ChoiceGroup } from '@/components/settings/choice-group';
import {
  ContrastPreview,
  FocusPreview,
  MotionPreview,
  TextSizePreview,
  TextSpacingPreview
} from '@/components/settings/previews';
import { Button } from '@/components/ui/button';
import { Field, FieldDescription, FieldLabel } from '@/components/ui/field';
import {
  Frame,
  FrameDescription,
  FrameHeader,
  FramePanel,
  FrameTitle
} from '@/components/ui/frame';
import { Switch } from '@/components/ui/switch';
import {
  CONTRAST_OPTIONS,
  FOCUS_OPTIONS,
  MOTION_OPTIONS,
  TEXT_SIZE_OPTIONS,
  TEXT_SPACING_OPTIONS
} from '@/config/preferences';
import { cn } from '@/lib/utils';
import { PREFERENCE_DEFAULTS, usePreferencesStore } from '@/store/preferences';

/** A preference that is on or off, with a sample of what turning it on does. */
function ToggleRow({
  id,
  label,
  description,
  sample,
  checked,
  onCheckedChange
}: {
  id: string;
  label: string;
  description: string;
  sample: ReactNode;
  checked: boolean;
  onCheckedChange: (checked: boolean) => void;
}) {
  return (
    <Field
      orientation="horizontal"
      className="items-center justify-between gap-4 rounded-lg border border-stone-200 p-3 dark:border-stone-800"
    >
      <span className="flex min-w-0 flex-col gap-1">
        <FieldLabel htmlFor={id}>{label}</FieldLabel>
        <FieldDescription>{description}</FieldDescription>
        <span
          aria-hidden="true"
          className="text-sm text-stone-900 dark:text-stone-100"
        >
          {sample}
        </span>
      </span>
      <Switch
        id={id}
        checked={checked}
        onCheckedChange={onCheckedChange}
      />
    </Field>
  );
}

/**
 * How the site reads: text, focus, motion, contrast and links.
 *
 * Each applies the moment it is chosen and is kept in this browser's local
 * storage (src/store/preferences.ts), so the next visit reads the same way. Badges, status tags and filters keep their
 * size whatever is chosen, so a table still reads as a table.
 */
export function AccessibilitySettings() {
  const preferences = usePreferencesStore();

  return (
    <div className="flex w-full flex-col gap-5">
      <Frame>
        <FrameHeader>
          <FrameTitle>Text size</FrameTitle>
          <FrameDescription>
            Text, controls and the sidebar grow together, so nothing is cramped around the larger
            text.
          </FrameDescription>
        </FrameHeader>
        <FramePanel>
          <ChoiceGroup
            legend="Text size"
            options={TEXT_SIZE_OPTIONS}
            value={preferences.textSize}
            onChange={preferences.setTextSize}
            preview={(size) => <TextSizePreview size={size} />}
          />
        </FramePanel>
      </Frame>

      <Frame>
        <FrameHeader>
          <FrameTitle>Reading</FrameTitle>
          <FrameDescription>
            The space around letters, words and lines, and how heavy the text is drawn.
          </FrameDescription>
        </FrameHeader>
        <FramePanel className="flex flex-col gap-5">
          <ChoiceGroup
            legend="Text spacing"
            options={TEXT_SPACING_OPTIONS}
            value={preferences.textSpacing}
            onChange={preferences.setTextSpacing}
            preview={(spacing) => <TextSpacingPreview spacing={spacing} />}
          />
          <ToggleRow
            id="bold-text"
            label="Heavier text"
            description="Every weight a step heavier, for screens where thin strokes break up."
            sample={
              <>
                <span>Regular</span> <span className="font-semibold">becomes</span>{' '}
                <span className="font-bold">heavier</span>
              </>
            }
            checked={preferences.boldText}
            onCheckedChange={preferences.setBoldText}
          />
        </FramePanel>
      </Frame>

      <Frame>
        <FrameHeader>
          <FrameTitle>Keyboard focus</FrameTitle>
          <FrameDescription>
            How the field, button or link the keyboard is on is marked.
          </FrameDescription>
        </FrameHeader>
        <FramePanel>
          <ChoiceGroup
            legend="Keyboard focus"
            options={FOCUS_OPTIONS}
            value={preferences.focus}
            onChange={preferences.setFocus}
            preview={(focus) => <FocusPreview focus={focus} />}
          />
        </FramePanel>
      </Frame>

      <Frame>
        <FrameHeader>
          <FrameTitle>Motion</FrameTitle>
          <FrameDescription>Transitions, sliding panels and animated changes.</FrameDescription>
        </FrameHeader>
        <FramePanel>
          <ChoiceGroup
            legend="Motion"
            options={MOTION_OPTIONS}
            value={preferences.motion}
            onChange={preferences.setMotion}
            preview={(motion) => <MotionPreview motion={motion} />}
          />
        </FramePanel>
      </Frame>

      <Frame>
        <FrameHeader>
          <FrameTitle>Contrast and links</FrameTitle>
          <FrameDescription>
            For reading in bright light, on a dim screen, or without relying on color.
          </FrameDescription>
        </FrameHeader>
        <FramePanel className="flex flex-col gap-5">
          <ChoiceGroup
            legend="Contrast"
            options={CONTRAST_OPTIONS}
            value={preferences.contrast}
            onChange={preferences.setContrast}
            preview={(contrast) => <ContrastPreview contrast={contrast} />}
          />
          <ToggleRow
            id="underline-links"
            label="Underline links"
            description="So a link is told from text by its shape, not only its color."
            sample={
              <>
                Read the{' '}
                <span
                  className={cn(
                    'text-primary-600 dark:text-primary-400',
                    preferences.underlineLinks && 'underline underline-offset-[0.2em]'
                  )}
                >
                  Big O guide
                </span>{' '}
                before you start.
              </>
            }
            checked={preferences.underlineLinks}
            onCheckedChange={preferences.setUnderlineLinks}
          />
        </FramePanel>
      </Frame>

      <div>
        <Button
          variant="ghost"
          // This tab's own, not the theme or the density, which are
          // Appearance's and a reader here did not ask to change.
          onClick={() => {
            preferences.setTextSize(PREFERENCE_DEFAULTS.textSize);
            preferences.setTextSpacing(PREFERENCE_DEFAULTS.textSpacing);
            preferences.setBoldText(PREFERENCE_DEFAULTS.boldText);
            preferences.setFocus(PREFERENCE_DEFAULTS.focus);
            preferences.setMotion(PREFERENCE_DEFAULTS.motion);
            preferences.setContrast(PREFERENCE_DEFAULTS.contrast);
            preferences.setUnderlineLinks(PREFERENCE_DEFAULTS.underlineLinks);
          }}
        >
          Restore the defaults
        </Button>
      </div>
    </div>
  );
}
