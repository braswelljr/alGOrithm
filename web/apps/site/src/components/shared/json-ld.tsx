/**
 * Structured data for search engines, as one JSON-LD script. The value is
 * serialized here, with "<" escaped so text from an entry can never close the
 * script element.
 */
export function JsonLd({ data }: { data: Record<string, unknown> | Record<string, unknown>[] }) {
  return (
    <script
      type="application/ld+json"
      dangerouslySetInnerHTML={{ __html: JSON.stringify(data).replace(/</g, '\\u003c') }}
    />
  );
}
