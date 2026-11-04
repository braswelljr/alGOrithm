# Use the component library

The site is built from `@braswelljr/algorithm-ui` (`web/packages/ui`): if the library has a component for it, use it, and if it looks like a card, it is a `Card`. Extend a component with a cva variant rather than restyling it with one-off classes, and add what is missing to the library rather than to a page.

## Why

One component per idea is one place to make it accessible, themeable and good-looking: a `Card` gets the elevation scale, squircle corners, the focus ring and dark mode; a `Heading` keeps the outline and the type scale apart; a `Field` ties the label, the description and the error to the control. A hand-drawn copy gets none of that, drifts from the rest, and has to be fixed again in every place it was copied. Variants (`Card variant="interactive"`, `Button variant="ghost"`) keep choices to a few names a reviewer can check.

## What it looks like

| If you are making | Use |
|---|---|
| A card, panel, tile or callout | `Card` and its parts (`CardHeader`, `CardTitle`, `CardDescription`, `CardContent`, `CardFooter`, `CardAction`); `variant`: default, ghost, interactive, selected, accent; `size`: sm, md, default, lg |
| A framed region (a stage, a code panel, a preview) | `Frame`, `FramePanel`, `FrameHeader` |
| A page or section title | `Heading` (`level` for the outline, `size` for the look); one level 1 per page |
| A number that matters (counts, scores, metrics) | `Stat` (`variant="plain"` inline), `NumberTicker` for counts that change |
| A list row (entries, steps, results) | `Item`, `ItemGroup` |
| Nothing to show, or loading | `Empty`; `Skeleton`, `Spinner` |
| A label for a level, tag or cost | `Badge` (`beginner`, `intermediate`, `advanced`, `cost`, `secondary`), `LevelBadge`, `ComplexityBadge`, `TagChip` |
| A key | `Kbd`, `KeyCombo` (from the shortcuts registry) |
| A divider | `Separator` |
| A form control | `Field` + `FieldLabel` with `Input`, `InputGroup`, `Textarea`, `NativeSelect`, `Select`, `Combobox`, `NumberField`, `Checkbox`, `Switch`, `RadioGroup` |
| A segmented control or filter chips | `ToggleGroup`, `Toggle` (`variant="chip"`) |
| Tabs | `SpringTabs`, or `Tabs` from the library |
| A hint or a menu | `Tooltip`, `Popover`, `HoverCard`, `DropdownMenu` |
| An overlay | `Dialog`, `Sheet`, `Drawer`, `CommandDialog`, `MorphingDialog` |
| A note or an error | `Alert` |
| Code | `CodeSample` (or `CodeBlock` parts), `CodeTabs` |
| A run's steps | `TraceView`, `PlayerControls` |
| A question flow | `Questionnaire` with `QuizChoiceCard` |

Which button variant:

| Variant | For |
|---|---|
| `primary` | The one main action of a view: Run, Start quiz, Next question, Measure |
| `ghost` | Everything secondary: toolbars, player controls, icon buttons, card actions, filters, close |
| `outline` | Rare: a bordered control that must stand off a busy background, with a `// ui-ok:` reason |
| `link`, `destructive` | Inline actions in text; deleting or resetting |

Do: `<Card variant="interactive" render={<Link href={href} />}>`, `<Heading level={2}>`, `<Button variant="ghost">`.

Do not: `<div className="rounded-xl border bg-card p-4">`, `<h2 className="text-lg font-semibold">`, `<button className="rounded-md border ...">`, two primary buttons in one view.

## How it is checked

`web/apps/site/test/markup.test.ts` scans the site and the library's own algorithm components and fails, with the file and line, on:

- a class string that draws a card (rounded, outlined, and filled or shadowed) outside `Card`;
- a raw `<button>`, `<input>`, `<select>`, `<textarea>` or `<h1>`-`<h6>` in the site;
- `variant="outline"` without a reason.

It warns when a file has more than one primary button. An exception carries its reason on the same or the previous line: `// card-ok: why` or `// ui-ok: why`. The published visualizer package (`web/packages/visualizer`) does not depend on the library and is exempt; the site wraps its views in `Card` or `Frame`.

## Related

- [The markup says what the thing is](the-markup-says-what-the-thing-is.md): the library's components carry the semantics.
- [Motion can be paused, stepped and reduced](motion-can-be-paused-stepped-and-reduced.md).
- `web/packages/ui/SYNC.md`: where each component came from.
