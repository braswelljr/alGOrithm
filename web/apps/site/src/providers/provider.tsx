'use client';

import type { ReactNode } from 'react';
import { Sonner } from '@braswelljr/algorithm-ui/components/sonner';
import { TooltipProvider } from '@braswelljr/algorithm-ui/components/tooltip';
import { ThemeProvider } from 'next-themes';
import { ShortcutsProvider } from '@/context/use-shortcuts';

/**
 * What every page sits inside: light and dark (a class on <html>, the system
 * preference by default), tooltips, the one toaster, and the keyboard layer
 * (shortcuts, the command palette and the shortcut help).
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
        <ShortcutsProvider>{children}</ShortcutsProvider>
        <Sonner position="bottom-right" />
      </TooltipProvider>
    </ThemeProvider>
  );
}
