# 007 — CPU versus GPU

## Question

What is the trade-off between throughput, cost, consumption, and quality on the
available hardware?

## Procedure

Fix the source, resolution, duration, bitrate, and output. Compare `libx264` with
NVENC when an NVIDIA GPU and compatible runtime are available. Do not compare a
CPU `preset` with a GPU `preset` without documenting the mapping.

```bash
time MEDIA_TOOL=local scripts/generate-ladder.sh --input samples/raw/big-buck-bunny-1080p-normal.mp4 --output-dir samples/generated/cpu
```

Repeat using an image/runtime that exposes NVENC; the generic Docker fallback
does not guarantee GPU access.

## Measure

Wall time, encoding fps, CPU%, GPU%, memory, estimated energy/cost, size,
effective bitrate, and VMAF. Record warm-up and concurrency. Acceleration is not
an improvement if quality or compatibility falls outside the limit.
