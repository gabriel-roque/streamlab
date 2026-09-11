# Experiments

All results must record the date, commit, hardware, FFmpeg version, source,
parameters, raw metrics, and conclusion. Never compare runs with different
ladders, durations, or presets.

| ID | Topic | Main result |
|---|---|---|
| 001 | HTTP Range | 200/206 behavior and seeking |
| 002 | HLS | segment duration and startup |
| 003 | ABR | rebuffering and quality under throttling |
| 004 | CDN | hit ratio, origin, and latency |
| 005 | Metrics | QoE definitions and collection |
| 006 | Retries/DLQ | recovery and poison jobs |
| 007 | CPU/GPU | time, cost, and quality |
| 008 | Codecs | efficiency and compatibility |
| 009 | VMAF | objective quality vs. bitrate |
| 010 | Per-title | content-based ladder |
| 011 | Chaos | blast radius and recovery |

Media commands work with FFmpeg installed or with `MEDIA_TOOL=docker`. See
`docs/architecture/README.md` for the complete flow.
