# ADR-003: FFmpeg as the Transcoder

- **Status:** accepted
- **Data:** 2026-09-10

## Context

The lab needs to support probing, filters, codecs, HLS, and DASH in a
reproducible pipeline.

## Decision

Use FFmpeg and FFprobe as reference tools. The scripts support local execution or
`MEDIA_TOOL=docker` with `jrottenberg/ffmpeg:6.1-ubuntu`.

## Alternatives

- GStreamer: excellent composition, but increases the initial operational cost.
- Managed transcoding service: hides important decisions from the study.
- Custom libraries: unnecessary for the lab's objective.

## Consequences

Version, flags, and hardware must be recorded in experiments. Fast presets are
not comparable at constant quality; every comparison must fix the codec,
resolution, duration, and quality criterion.
