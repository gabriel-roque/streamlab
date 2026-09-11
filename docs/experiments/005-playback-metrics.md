# 005 — Playback Metrics

## Measurement Contract

- `startup_time`: first frame minus play intent.
- `rebuffer_ratio`: time stalled due to missing data divided by the time actually
  played. The current player sends `rebuffer_start`, but does not yet calculate
  this interval or send `rebuffer_end`; treat the formula as an experimental
  contract, not as an application-ready metric.
- `average_bitrate`: consumed media bytes divided by media time.
- `quality_switches`: count of representation changes.
- `error_rate`: sessions with an error divided by started sessions.

## Procedure

Run sessions with stable and variable networks. Create a session at
`POST /playback/sessions` and send events with the `video_id`, `session_id`,
`type`, `position`, and `payload` fields according to the contract in
`docs/api-contract.md`. The current frontend sends, among others, `play`,
`pause`, `seek`, `playing`, `rebuffer_start`, `quality_change`, and `audio_change`.
The server accepts any `type`, does not implement `sequence` or deduplication,
and stores events only in memory.

## Analysis

Group p50/p95/p99 by video, variant, browser, region, and network. Do not publish
a global average without a sample size. Correlate `startup_time` with cache miss
when an external cache is present, but do not treat correlation as causation. Use
`/metrics` to confirm the aggregate count of accepted events, not to obtain QoE
p50/p95: these percentiles must be calculated from the events collected by the
experiment.
