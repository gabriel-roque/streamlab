# 001 — HTTP Range

## Question

Can the server provide seeking and resumption without transferring the entire
MP4?

## Procedure

First perform a local upload and save the returned `id`. With Compose, `BASE` is
`http://localhost:3000/api`; with the API run directly, use
`http://localhost:8080`:

```bash
curl -i -H 'Range: bytes=0-1023' "$BASE/videos/ID/stream" -o /tmp/range.bin
curl -i -H 'Range: bytes=1000000-1000999' "$BASE/videos/ID/stream" -o /tmp/range-2.bin
curl -I -H 'Range: bytes=0-0' "$BASE/videos/ID/stream"
```

Record `206 Partial Content`, `Accept-Ranges: bytes`, `Content-Range`,
`Content-Length`, and the actual number of bytes received. Repeat with an
invalid range, without a range, and with `HEAD`. The current API does not send
`ETag` or implement `If-Range`, so these headers are not part of the executable
experiment.

## Criteria and Metrics

Valid ranges must return exactly the requested bytes, and an invalid range must
return `416`. The `200` response without Range must remain reproducible. Measure
latency, bytes sent, and player seek time; do not attribute a cache hit to the
endpoint because the API does not define cache headers for this resource.

## Caveats

Do not confuse `206` with adaptive streaming: Range transports an object;
HLS/DASH selects representations and segments.
