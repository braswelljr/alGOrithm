// Keeps the wasm build for repeat visits and offline use, and nothing else.
// Only versioned requests (algorithm.wasm?v=HASH, wasm_exec.js?v=HASH) are
// cached, keyed by their full URL, so a new build is a new key and the page
// never gets an old wasm with new JavaScript; older versions are deleted when
// a new one is stored. Pages, scripts and data always come from the network.
const CACHE = 'algorithm-wasm';
const FILES = /\/(algorithm\.wasm|wasm_exec\.js)$/;

self.addEventListener('install', () => self.skipWaiting());
self.addEventListener('activate', (e) => e.waitUntil(self.clients.claim()));

self.addEventListener('fetch', (e) => {
  const url = new URL(e.request.url);
  if (e.request.method !== 'GET' || url.origin !== location.origin) return;
  if (!FILES.test(url.pathname) || !url.searchParams.has('v')) return;
  e.respondWith(
    (async () => {
      const cache = await caches.open(CACHE);
      const hit = await cache.match(e.request);
      if (hit) return hit;
      const res = await fetch(e.request);
      if (res.ok) {
        for (const key of await cache.keys()) {
          const k = new URL(key.url);
          if (k.pathname === url.pathname && k.search !== url.search) await cache.delete(key);
        }
        await cache.put(e.request, res.clone());
      }
      return res;
    })()
  );
});
