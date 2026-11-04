import type { ReactNode } from 'react';
import { Heading } from '@braswelljr/algorithm-ui/components/heading';

/** The page's one h1, with an optional lead paragraph. */
export function PageHeading({ title, children }: { title: string; children?: ReactNode }) {
  return (
    <div className="mb-6">
      <Heading
        level={1}
        className="text-2xl font-semibold"
      >
        {title}
      </Heading>
      {children ? <p className="mt-2 max-w-3xl text-muted">{children}</p> : null}
    </div>
  );
}
