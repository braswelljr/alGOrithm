'use client';

import { useState } from 'react';
import { PointsView } from '@braswelljr/algorithm-visualizer';
import { Button } from '@/components/ui/button';
import { formatPoints, parsePoints, randomPoints } from '@/lib/points-text';

/**
 * Add points by clicking the plane, or start from random ones; the points
 * are the text field's text, so editing either changes both.
 */
export function PointsPicker({ text, onText }: { text: string; onText: (t: string) => void }) {
  const [seed, setSeed] = useState(1);
  const points = parsePoints(text);
  return (
    <div className="flex flex-col gap-2 sm:col-span-2">
      <PointsView
        points={points}
        label="Click to add a point"
        onAdd={(p) => onText(formatPoints([...points, p]))}
      />
      <div className="flex flex-wrap gap-2">
        <Button
          type="button"
          variant="secondary"
          onClick={() => {
            onText(formatPoints(randomPoints(20, seed)));
            setSeed(seed + 1);
          }}
        >
          Random points
        </Button>
        <Button
          type="button"
          variant="secondary"
          onClick={() => onText('')}
          disabled={points.length === 0}
        >
          Clear
        </Button>
        <p className="text-xs text-muted">{points.length} points. Click the plane to add one.</p>
      </div>
    </div>
  );
}
