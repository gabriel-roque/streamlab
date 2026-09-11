# 002 — HLS and Segment Duration

## Question

What size balances startup, adaptation, and request overhead?

## Procedure

Generate packages with 2, 4, 6, 10, and 12 seconds, keeping the source, codec,
ladder, and GOP constant. Use a different output directory for each run:

```bash
scripts/generate-hls.sh --input samples/raw/big-buck-bunny-1080p-normal.mp4 \
  --segment-seconds 6 --output-dir samples/generated/hls-6s --overwrite
scripts/validate-media.sh --input samples/generated/hls-6s --kind hls
```

Run the same flow for each output directory. Count segments, average size, and
the duration indicated by `#EXTINF`.

## Measure

Startup time, requests per minute, time to the first quality switch, rebuffer
ratio, manifest size, and cache hit ratio. Acceptance is a conclusion based on
the complete set, not a universal number.

## Caveats

Without aligned keyframes, `hls_time` is only an intention and cuts may deviate
from the nominal duration. The ladder script fixes H.264/AAC, `yuv420p`, and a
48-frame GOP; confirm this in the manifest and with `ffprobe` instead of
assuming that every segment has exactly the requested value.
