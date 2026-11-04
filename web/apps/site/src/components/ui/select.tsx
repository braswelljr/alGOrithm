import type * as React from 'react';
import { NativeSelect } from '@braswelljr/algorithm-ui/components/native-select';

/** A native select (keyboard and screen readers work as they should): the library's NativeSelect. */
export function Select(props: React.ComponentProps<typeof NativeSelect>) {
  return <NativeSelect {...props} />;
}
