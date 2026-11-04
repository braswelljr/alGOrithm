'use client';

import { useTheme } from 'next-themes';
import { useThemeToggle } from '@/components/animated/theme-toggle';
import { ChoiceGroup } from '@/components/settings/choice-group';
import { DensityPreview, ThemePreview } from '@/components/settings/previews';
import {
  Frame,
  FrameDescription,
  FrameHeader,
  FramePanel,
  FrameTitle
} from '@/components/ui/frame';
import { DENSITY_OPTIONS, THEME_OPTIONS } from '@/config/preferences';
import { usePreferencesStore } from '@/store/preferences';

/**
 * How the site looks: the theme, and how much fits on the screen.
 *
 * Kept in this browser only. The theme is set through
 * next-themes, which owns it and keeps it in local storage; density lives in the
 * preferences store beside the accessibility settings.
 */
export function AppearanceSettings() {
  const { theme } = useTheme();
  // Swept in on the diagonal the previews split on, unless motion is reduced.
  const { switchTo } = useThemeToggle({ variant: 'diagonal' });
  const { density, setDensity } = usePreferencesStore();
  const current = THEME_OPTIONS.find((option) => option.value === theme)?.value ?? 'system';

  return (
    <div className="flex w-full flex-col gap-5">
      <Frame>
        <FrameHeader>
          <FrameTitle>Theme</FrameTitle>
          <FrameDescription>Light, dark, or whichever the device is set to.</FrameDescription>
        </FrameHeader>
        <FramePanel>
          <ChoiceGroup
            legend="Theme"
            options={THEME_OPTIONS}
            value={current}
            onChange={switchTo}
            preview={(theme, selected) => (
              <ThemePreview
                theme={theme}
                selected={selected}
              />
            )}
          />
        </FramePanel>
      </Frame>

      <Frame>
        <FrameHeader>
          <FrameTitle>Density</FrameTitle>
          <FrameDescription>
            How tightly rows, controls and the sidebar are spaced.
          </FrameDescription>
        </FrameHeader>
        <FramePanel>
          <ChoiceGroup
            legend="Density"
            options={DENSITY_OPTIONS}
            value={density}
            onChange={setDensity}
            preview={(value) => <DensityPreview density={value} />}
          />
        </FramePanel>
      </Frame>
    </div>
  );
}
