# Motion can be paused, stepped and reduced

Every animation can be paused, stepped one event at a time in both directions, and scrubbed. When the user's system asks for reduced motion (`prefers-reduced-motion: reduce`), the visualizer starts paused, moves between states without tweening, and never autoplays. Nothing flashes more than three times a second at any speed.

## Why

An animation people cannot stop is one they cannot learn from: the interesting step is gone before it is understood. Stepping is how a learner checks their own prediction ("the next swap is 7 and 3"). Reduced motion is an accessibility setting people rely on to avoid nausea and distraction, and fast sorting animations are exactly the motion it exists for. The flash limit (WCAG 2.3.1) matters because a sort at full speed can flicker highlights many times a second.

## What it looks like

- The player has play/pause, step back, step forward, a scrubber and a speed control, all reachable by keyboard (space, left and right arrows).
- At high speeds the view draws every Nth state instead of flashing every highlight.
- Under reduced motion, positions change instantly and the player opens paused.

Do: drive every view from the player's index, so pausing and stepping work for every view without extra code.

Do not: start an animation on page load, or animate with a timer the player does not control.

## How it is checked

Review, and planned `web/` tests that the player starts paused with reduced motion set, and that stepping forward then back returns the same state.

## Related

- [The markup says what the thing is](the-markup-says-what-the-thing-is.md)
- The `emil-design-eng`, `gsap` and `web-design-guidelines` skills.
