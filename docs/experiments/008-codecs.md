# 008 — Codec Comparison

## Question

How much bitrate does each codec need for the same quality, and what is the
encode/decode cost?

## Procedure

Encode the same clip with H.264, HEVC, VP9, and AV1, keeping resolution,
framerate, audio, duration, and quality target constant. Use the available
encoder and record its version/build; an unavailable encoder is a compatibility
result, not a reason to invent numbers.

## Measure

Size, bitrate, encoding fps, time, CPU/GPU, VMAF when a reference is available,
startup time, and player support. Report bitrate by quality, not an absolute
ranking.

## Risks

Licenses, hardware, presets, and tuning change the result. A single Big Buck
Bunny scene does not represent the entire catalog; repeat with animation, low
light, and high motion.
