'use client';

import { useTheme } from 'next-themes';
import { Toaster, ToasterProps } from 'sonner';

/**
 * The one toaster, mounted once in the providers.
 *
 * A toast is a short word on what just happened: a sentence of a few words
 * (`Template saved`), with the detail in `description`, and an `action` when
 * the change can be undone or followed. A failure is `toast.error` with a
 * description that says what to do about it. Three are seen at a time, so a
 * burst of saves does not stack a column up the screen.
 *
 * The toaster sets a system font of its own; `fontFamily: 'inherit'` gives it
 * the app's face back.
 */
const Sonner = ({ style, ...props }: ToasterProps) => {
  const { theme = 'system' } = useTheme();

  return (
    <Toaster
      theme={theme as ToasterProps['theme']}
      className="toaster group"
      visibleToasts={3}
      style={{ fontFamily: 'inherit', ...style }}
      {...props}
    />
  );
};

export { Sonner };
