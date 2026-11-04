'use client';

import { PrinterIcon } from 'lucide-react';
import { Button } from '@braswelljr/algorithm-ui/components/button';

/** Prints the cheat sheets; the print stylesheet leaves only the tables. */
export function PrintButton() {
  return (
    <Button
      variant="primary"
      onClick={() => window.print()}
    >
      <PrinterIcon aria-hidden />
      Print
    </Button>
  );
}
