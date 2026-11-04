# Where the components came from

`@braswelljr/algorithm-ui` is the site's component library: Base UI primitives (`@base-ui/react` 1.9.0, no Radix anywhere) in the shadcn **base-vega** style, Tailwind 4 only. Some components were synced from two sibling projects, a few came from the shadcn registry, and the rest are written here. To sync again, copy the file from the source path, then redo the adaptations listed for it.

## The design system (the shadcn questionnaire's answers)

Set by hand in `components.json` and `src/styles/theme.css`, because `shadcn init` (or `create`, the same command) would overwrite the tuned theme and config. To reproduce them in a new project: `pnpm dlx shadcn@latest create`, then answer as below, or edit `components.json` to match.

| Question | Answer |
|---|---|
| Component library | Base UI |
| Style | base-vega |
| Base color | zinc, mapped onto the site's own OKLCH neutrals (`src/lib/colors.ts`: slate by default, zinc and stone selectable) |
| Theme (accent) | indigo by default; blue, violet, emerald, amber and rose selectable |
| Icon library | lucide |
| Font | Ysabeau Office for text, JetBrains Mono for code, both through `next/font/google` in `apps/site/src/lib/fonts.ts` |
| Radius | 0.625rem (`--radius`), with sm, md, lg, xl and 2xl derived from it |
| CSS variables | yes, as the site's tokens (`--bg`, `--fg`, `--muted`, `--card`, `--border`, `--subtle`, `--accent`, `--control`, ...), plus shadcn's names mapped onto them in `@theme inline` |

The font: Ysabeau Office is a humanist sans with calligraphic stroke modulation, gently flared terminals, open round bowls and a generous x-height, which keeps small UI text readable. It was chosen over Signika (lower contrast, narrower) and Mukta (drawn for Devanagari first, more geometric). It is one variable font, latin subset, `display: swap`, with next/font's `adjustFontFallback` sizing the system fallback so swapping causes no layout shift.

## From gsl-certificates

Source: `/Users/braswelljr/projects/gsl/gsl-certificates/web/src/components/ui/` (Vite, `rsc: false`) and its `hooks/use-isomorphic-layout-effect.ts`, `hooks/use-media-query.ts`, `utils/utils.ts` (now `src/lib/utils.ts`).

Files: alert-dialog alert autocomplete badge breadcrumb button-group button card carousel checkbox collapsible combobox command context-menu dialog drawer dropdown-menu empty field frame heading hover-card input-group input item kbd label magnetic marquee meter native-select number-field overflow-row pagination popover progress radio-group resizable-split-view resizable scroll-area select separator sheet skeleton sonner spinner stat switch table tabs textarea timeline toggle-group toggle tooltip 

Adaptations on every file:

- Imports go through the package (`@braswelljr/algorithm-ui/components/...`, `/lib/utils`, `/hooks/...`).
- `'use client'` where a file uses state, effects, handlers or a Base UI component.
- Hard-coded grays (`zinc-*`, `white`) became the theme's tokens (`bg`, `card`, `subtle`, `border`, `control`, `muted`, `fg`), and their `dark:` twins were dropped, so every component follows the neutral, the accent and dark mode.
- Fixed-color fills keep fixed dark text (`text-zinc-950`).
- The gold `primary` became the chosen accent with `accent-fg` text.

Changes to particular files:

- `button`: variants restated on the tokens. `primary` is the accent; `ghost` has a visible hover, pressed and focus state.
- `badge`: gained `beginner`, `intermediate`, `advanced` and `cost` variants.
- `card`: gained cva `variant` (default, outline, ghost, interactive, selected, accent) and `size` (sm, md, default, lg), plus squircle corners.
- `stat`: gained `variant` (tile, plain) and lost the Council's adinkra watermark.
- `toggle`: pressed state uses `accent-fg` for contrast, and it gained a `chip` variant.
- `input`: the forwarded-ref write moved out of render (React compiler lint).
- `timeline`: dark text on the light dark-mode dots.
- `marquee`: replaced by gnux's, which hides the repeats from assistive technology and stands still under reduced motion.

Not synced (app-specific): minimal-tiptap, rich-text, calendar, chat, message, bubble, attachment, phone-number-input, signature-input, sortable, data-table, facet-filter, bulk-action-bar, refresh-button, confirm-dialog, adinkra, texture-overlay, sidebar, chart (would add recharts; the site draws its own SVG charts).

## From gnux

Source: `/Users/braswelljr/projects/go/gnux/web/src/components/ui/` and `shared/`, and its `hooks/` (`use-reduced-motion`, `use-typewriter`, `use-in-view`, `use-platform`, `use-mounted`, `use-copy-to-clipboard`).

Files: action-swap, animated-background, caret, copy-button, gradient-text, keyboard, keycap, marquee, reveal, section, terminal-window, typing-code, typing-terminal, theme-toggle (from `shared/`), json-ld (from `shared/`; the site's own `components/shared/json-ld.tsx` is BRAND's and stays the one in use).

Adaptations:

- Fumadocs tokens (`fd-*`), `line`, `faint` and `lit` are mapped onto the site's tokens in `theme.css`.
- The keyboard's orange accent colorway became the site's accent.
- The utilities the components rely on (`keycap`, `plate`, `lifted`, `tick`, `hud-label`, the animations, the theme toggle's view-transition reveal) are copied into `theme.css`.
- gnux's `code-block` was merged into shadcn's (below), keeping its title and copy button.

## From the shadcn registry

- `code-block` (`@shadcn/code-block`, base-vega) and `use-highlighted-code` with `lib/highlighter` (`@shadcn/use-highlighted-code`).
  - The `IconPlaceholder` became lucide icons, and muted surfaces became `subtle`.
  - Go was added to the highlighter.
  - It gained `CodeSample`, gnux's one-piece sample: a title, a copy button, line numbers and highlighted lines.
  - Highlighting is shiki on a JavaScript regex engine, loaded in the browser, so it works in the static export.
- `trace` (`@shadcn/trace`, base-vega): the icon placeholder became lucide. `TraceView` is built on it.
- `questionnaire` (`@shadcn/questionnaire`, base-vega, on `@shadcn/react/questionnaire`): checked states use the accent and `variant="default"` became `primary`. `QuizChoiceCard`, the quiz, the placement test and the end-of-path check are built on it.

No item was pulled from @beui, @soralabs or @atelier; they are configured in `components.json` and used as references. TextScramble, NumberTicker, SpringTabs, MorphingDialog and the tilt on EntryCard are written here in their spirit. Before pulling one, check its license (Atelier's pro items are paid and must never be pulled; Sora UI's license was not verified), record it here with its notice, and swap any Radix import for Base UI.

## Written here (`src/algorithm/`, `src/lib/`)

- **Algorithm components:**
  - EntryCard, CategoryCard
  - LevelBadge, ComplexityBadge, TagChip
  - PathStepper, PlayerControls, TraceView, TLDRCard
  - CommandPalette (with `command-palette-filter`), ShortcutOverlay, KeyCombo
  - CheatSheetTable, QuizChoiceCard, CodeTabs
- **Motion:** NumberTicker, TextScramble, SpringTabs, MorphingDialog.
- **Theme:** ThemePicker (with `useThemePrefs`).
- **Libraries:**
  - `lib/colors.ts`: the color system, the one source of `styles/colors.css` (written by `scripts/colors.ts`, checked by `test/colors.test.ts`).
  - `lib/theme-prefs.ts`, `lib/shortcuts.ts`, `lib/motion.ts`.
