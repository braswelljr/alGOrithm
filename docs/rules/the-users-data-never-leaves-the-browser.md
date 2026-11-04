# The user's data never leaves the browser

Datasets the user pastes, uploads or generates are processed in their browser by the wasm build and stored, if at all, in their browser. The site has no server that receives them, no analytics that records them, and no "upload" in the network sense. A URL import is fetched by the user's own browser, directly from the server they named.

## Why

People will try the visualizer on real data: a column of salaries, a list of customer IDs, exam scores. They should not have to wonder where it went. Running the real Go code as WebAssembly makes the promise easy to keep, because there is nothing to send it to, and stating it as a rule stops a later feature (sharing, saving, telemetry) from quietly breaking it.

## What it looks like

- The file input reads the file with the `File` API and passes its text to `alGOrithm.parse`.
- A shared link carries a seed or a small dataset encoded in the URL; a large dataset stays in IndexedDB and the share dialog says the link will not include it.
- The data page says, beside the import controls: "Your data stays in this browser."

Do: keep every dataset operation in the page.

Do not: add an API route that accepts a dataset, put dataset contents in analytics events, or proxy URL imports through the site (that would send the URL and its contents to our server).

## How it is checked

Review, and a planned test in `web/` that no request leaves the page while a file is imported and animated.

## Related

- [Data from outside is unknown until checked](data-from-outside-is-unknown-until-checked.md)
- [State a reader can arrive at lives in the URL](state-a-reader-can-arrive-at-lives-in-the-url.md)
