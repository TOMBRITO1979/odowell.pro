const CACHE_NAME = 'drcrwell-v5-odowell'; // Incrementado para forçar atualização
const urlsToCache = [
  // Apenas arquivos estáticos essenciais - NÃO cachear JS/CSS
  '/manifest.json',
];

// Install event - cache assets
self.addEventListener('install', (event) => {
  event.waitUntil(
    caches.open(CACHE_NAME)
      .then((cache) => cache.addAll(urlsToCache))
      .then(() => self.skipWaiting())
  );
});

// Activate event - clean up old caches
self.addEventListener('activate', (event) => {
  event.waitUntil(
    caches.keys().then((cacheNames) => {
      return Promise.all(
        cacheNames.map((cacheName) => {
          if (cacheName !== CACHE_NAME) {
            return caches.delete(cacheName);
          }
        })
      );
    }).then(() => self.clients.claim())
  );
});

// Fetch event - NÃO cachear JS/CSS para garantir código atualizado
self.addEventListener('fetch', (event) => {
  // Skip non-GET requests
  if (event.request.method !== 'GET') return;

  // Skip API requests (let them go to network)
  if (event.request.url.includes('/api/')) return;

  // Skip chrome-extension requests
  if (event.request.url.startsWith('chrome-extension://')) return;

  // NÃO cachear arquivos JS, CSS e HTML - sempre buscar da rede
  const url = event.request.url;
  if (url.endsWith('.js') || url.endsWith('.css') || url.endsWith('.html') || url.includes('/assets/')) {
    return; // Let browser handle normally without SW interference
  }

  // Apenas cachear recursos estáticos (imagens, fontes)
  event.respondWith(
    fetch(event.request).catch(() => {
      // Network failed, try cache only for static assets
      return caches.match(event.request);
    })
  );
});
