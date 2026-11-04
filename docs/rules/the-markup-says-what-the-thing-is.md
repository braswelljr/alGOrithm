# The markup says what the thing is

The website uses semantic elements and landmarks before ARIA, gives every control an accessible name, has one `h1` per page with no skipped heading levels, and never carries meaning by color alone. In a visualization, a compared, swapped, sorted or found element is told apart by shape, pattern or label as well as color, and the current step is described in text for screen readers.

## Why

Sorting visualizations traditionally mean "red is compared, green is sorted". About one in twelve men cannot tell those apart. A label, an outline or a pattern alongside the color makes the same picture readable for everyone, and a live text line ("Step 212: compare 7 and 3, swap") makes it usable without seeing it at all. Semantic markup is what keyboard and screen reader users navigate by.

## What it looks like

- A `<section aria-labelledby>` per panel (data, player, view, counts); `<button>`s with text or an `aria-label` for icon-only ones.
- Compared bars outlined, swapped bars marked with an arrow, sorted bars hatched, each also colored.
- An `aria-live="polite"` region under the view that describes the current event in words, throttled at high speed.

Do: pass a text description with every state the view draws.

Do not: a `div` with an `onClick`, a legend that is the only key to what colors mean, or a canvas with no text alternative.

## How it is checked

Review with the `web-design-guidelines` and `semantic-html` skills; planned automated accessibility checks (axe) in the `web/` tests.

## Related

- [Motion can be paused, stepped and reduced](motion-can-be-paused-stepped-and-reduced.md)
- The same rule in gnux: `docs/rules/the-markup-says-what-the-thing-is.md` there.
