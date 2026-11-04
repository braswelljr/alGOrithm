import type React from 'react';
import type { RowData } from '@tanstack/react-table';
/**
 * niko-table, created by Semir N. (Semkoo, https://github.com/Semkoo) with AI assistance.
 *
 * Before reporting anything: please check the changelog first.
 *  - In-repo: ./CHANGELOG.md
 *  - Docs site: https://niko-table.com/changelog
 *
 * Found a bug or have a fix? Open an issue or PR on GitHub so other
 * users (and future LLMs reading this code) benefit:
 * https://github.com/Semkoo/niko-table-registry
 */

import type { AppColumn } from '@/lib/table';

export const getCommonPinningStyles = <TData extends RowData>(
  column: AppColumn<TData>,
  isHeader: boolean = false
): React.CSSProperties => {
  const isPinned = column.getIsPinned();
  if (!isPinned) return {};

  const isLeft = isPinned === 'start';
  const columnSize = column.getSize();

  return {
    position: 'sticky',
    top: isHeader ? 0 : undefined,
    left: isLeft ? `${column.getStart('start')}px` : undefined,
    right: !isLeft ? `${column.getAfter('end')}px` : undefined,
    opacity: 1,
    width: columnSize,
    minWidth: columnSize, // Prevent column from shrinking
    maxWidth: columnSize, // Prevent column from growing
    flexShrink: 0, // Prevent flex shrinking
    // Headers: z-20 to stay above other headers and body.
    // Body: z-10 to stay above other body cells.
    zIndex: isHeader ? 20 : 10,
    // No background here on purpose. A pinned cell does need an opaque surface,
    // or the rows scrolling under it show through, but it is painted by classes
    // on the cell itself: those also carry the row's hover and selected shades,
    // and an inline colour would win over all of them and leave the frozen
    // column dead while the rest of its row lit up.
    //
    // Pinned-column separator. Drawn as an INSET box-shadow (part of the sticky
    // cell's own paint) rather than a plain border: a border on a sticky cell
    // gets covered by scrolled-under content in some browsers, so it vanishes
    // once you scroll horizontally. An inset shadow stays put, and a soft OUTER
    // shadow adds depth. The line is the same weight and colour as the rule
    // between rows, so the header and the body agree on where the seam is.
    ...(isLeft
      ? {
          boxShadow:
            'inset -1px 0 0 0 var(--table-pinned-divider), 4px 0 8px -4px rgb(0 0 0 / 0.12)'
        }
      : {
          boxShadow:
            'inset 1px 0 0 0 var(--table-pinned-divider), -4px 0 8px -4px rgb(0 0 0 / 0.12)'
        })
  };
};
