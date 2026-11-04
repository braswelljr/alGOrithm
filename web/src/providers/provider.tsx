'use client';

import type { ReactNode } from 'react';
import { ThemeProvider } from 'next-themes';
import { TourProvider } from '@/components/tour/tour-provider';
import { Sonner } from '@/components/ui/sonner';
import { TooltipProvider } from '@/components/ui/tooltip';
import { ShortcutsProvider } from '@/context/use-shortcuts';

/**
 * What every page sits inside: light and dark (a class on <html>, the system
 * preference by default), tooltips, the one toaster, and the keyboard layer
 * (shortcuts, the command palette and the shortcut help), and the guided tours.
 */
export function Provider({ children }: { children: ReactNode }) {
  return (
    <ThemeProvider
      attribute="class"
      defaultTheme="system"
      enableSystem
      disableTransitionOnChange
    >
      <TooltipProvider>
        <ShortcutsProvider>
          {/* The guided tours; the tour of the site starts once by itself. */}
          <TourProvider welcome="shell">{children}</TourProvider>
        </ShortcutsProvider>
        <Sonner position="bottom-right" />
      </TooltipProvider>
    </ThemeProvider>
  );
}
