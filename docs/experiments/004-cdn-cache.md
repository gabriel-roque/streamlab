# 004 — CDN and Caching

## Question

How much traffic no longer reaches the origin when manifests and segments are
cacheable?

## Procedure

Compose places NGINX between the browser and the API, but does not yet implement
a CDN or manifest/segment proxy for MinIO. Use the flow below to confirm the
current behavior; then place an external reverse cache between the client and
the endpoint and repeat two sessions for the same manifest/segments.

```bash
scripts/smoke-http.sh --url http://localhost:3000 \
  --path /api/videos/ID/playback/hls
```

Replace `ID` with a local upload in `READY` state. The remote `abr-lab` fixture
returns the public HLS URL at `/api/videos/abr-lab/playback`, but has no local
manifest to fetch through this route.

## Measure

`Age`, `ETag`, `Cache-Control`, hit/miss, p50/p95 latency, origin bytes, origin
requests, and revalidation rate. In the current Compose setup, `/api/*` is a
proxy to the API, which does not send `Age`, `ETag`, or `Cache-Control`; `/media/`
is NGINX's only static path and points to local media, not artifacts. Therefore,
these signals appear only after adding/configuring the experiment's cache. In
real VOD, use manifests with a shorter TTL than segments and do not cache a
private response without considering credentials.

## Criteria

The second read of the same segment should be an observable HIT in the added
cache, and the API/origin should receive fewer bytes. Also test invalidation,
`If-None-Match`, and a cache miss after expiration; do not expect these behaviors
from this repository's default NGINX.
