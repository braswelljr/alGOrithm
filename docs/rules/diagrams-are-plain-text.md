# Diagrams are plain text

A diagram in the repository's documentation (markdown, doc comments) is drawn with plain characters (`+`, `-`, `|`, `>`, `v`) in a fenced `text` block, never with Mermaid or another diagram language that needs a renderer.

The website is different: it has a renderer, so its pages may draw diagrams with Mermaid through `MermaidDiagram` (`web/src/components/algorithm/mermaid-diagram.tsx`), which loads Mermaid only on a page that draws one, renders in strict mode with the theme's colors, and carries a text label for screen readers. The learning-path map on `/learn` is one.

## Why

A Mermaid block is a picture only where something renders it. On GitHub it draws; in a terminal, in `cat`, in an editor without a plugin, in a man page, in pkg.go.dev or in an agent's context it is a block of syntax a reader has to decode. A plain-text diagram reads the same everywhere these docs are read, and it can be edited with any editor and checked in review line by line.

## What it looks like

```text
dataset  ------>  internal/core: Entries(), Execute()  ------>  sorting, ...
                          |                                           |
                          v                                           | record events
                   cmd/algorithm-wasm  --->  web/                             v
                                                              trace: Event, Recorder
```

Do: keep diagrams narrow (under about 90 columns), with arrows (`-->`, `v`) that say which way things flow and a label on an arrow when it carries something specific.

Do not: write a ` ```mermaid ` block, or embed an image of a diagram that has no text version.

## How it is checked

Review. This prints any Mermaid block left in the docs:

```bash
grepx -rn '```mermaid' --include='*.md' . | grepx -v '^./.agents/'
```

## Related

- [No em dashes, and US spelling](no-em-dashes-and-us-spelling.md), the other rule about how the docs are written.
- [Visualizer design](../design/visualizer.md), whose architecture diagram follows this rule.
