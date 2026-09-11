# ADR-006: Six-Second Segments

- **Status:** accepted
- **Data:** 2026-09-10

## Context

Short segments reduce ABR reaction time, but increase requests, manifest
overhead, and pressure on the CDN.

## Decision

Use six seconds as the VOD baseline: it is the default for `generate-hls.sh`,
`generate-dash.sh`, and the local processor's FFmpeg command. Vary it to 2, 4,
10, and 12 seconds in the HLS experiment. The HLS fixture without FFmpeg has
exactly one six-second segment; in real packages, `hls_time` is a target and the
effective duration must be read from the manifest.

## Alternatives

- 2 seconds: fast switching, high request cost.
- 10-12 seconds: lower overhead, slower adaptation.
- Variable segmentation: can improve content handling, but makes the initial
  comparison harder.

## Consequences

The number of segments, startup, rebuffer, and cache hit ratio must be measured
together; no single metric determines the optimal size.
