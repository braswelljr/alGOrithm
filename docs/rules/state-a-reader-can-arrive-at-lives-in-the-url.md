# State a reader can arrive at lives in the URL

Anything a user would want to share or come back to is in the URL: the entry, the dataset (or its generator, size and seed), the speed, and the step the player is on. It is read and written through `nuqs`, and opening a link shows exactly what the sender saw, paused, without starting to play.

## Why

The point of a visualizer is often to show someone else a moment: "look at step 212, that is where quick sort picks a bad pivot". A link that opens the same algorithm on the same data at the same step does that in one click. Starting paused on arrival respects the reader, who may not expect motion.

## What it looks like

```text
/visualize/quick-sort?data=nearly-sorted&n=60&seed=7&step=212&speed=4
/compare?ids=quick-sort,merge-sort,heap-sort&data=random&n=5000&seed=1
```

Do: generated datasets by kind, size and seed; small typed datasets encoded in the URL; parse every parameter as unknown input ([rule](data-from-outside-is-unknown-until-checked.md)).

Do not: keep the step or the dataset only in React state, or autoplay when a link is opened.

## How it is checked

Review, and planned `web/` tests that a link round-trips: open it, read the state, write the URL again, and get the same link.

## Related

- [The user's data never leaves the browser](the-users-data-never-leaves-the-browser.md), for datasets too large for a URL.
