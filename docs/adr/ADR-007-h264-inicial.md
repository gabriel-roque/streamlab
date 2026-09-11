# ADR-007: H.264 as the Initial Codec

- **Status:** accepted
- **Data:** 2026-09-10

## Context

The first path must be decodable by browsers and players without specific
hardware, while preserving the ability to compare codecs later.

## Decision

The initial ladder uses H.264/AVC with `yuv420p` pixel format and AAC audio.
HEVC, VP9, and AV1 are covered in separate experiments, with compatibility
explicitly measured.

## Alternatives

- AV1: potentially more efficient, with more variable encoding and support.
- HEVC: efficient, but with licensing and support concerns.
- VP9: an open option, but it does not maintain the same initial compatibility
  profile.

## Consequences

Bitrate is not comparable across codecs without a quality metric. The experiment
records time, size, VMAF when available, and player failure rate.
