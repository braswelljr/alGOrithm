import Link from 'next/link';
import { Heading } from '@braswelljr/algorithm-ui/components/heading';

export default function NotFound() {
  return (
    <div className="py-16 text-center">
      <Heading
        level={1}
        className="text-2xl font-semibold"
      >
        Page not found
      </Heading>
      <p className="mt-2 text-muted">
        <Link
          className="underline"
          href="/"
        >
          See every algorithm
        </Link>
      </p>
    </div>
  );
}
