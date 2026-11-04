import Link from 'next/link';
import { Heading } from '@/components/ui/heading';

export default function NotFound() {
  return (
    <div className="py-16 text-center">
      <Heading
        level={1}
        className="text-2xl font-semibold"
      >
        Page not found
      </Heading>
      <p className="text-muted mt-2">
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
