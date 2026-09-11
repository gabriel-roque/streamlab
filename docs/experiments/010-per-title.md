# 010 — Per-title encoding

## Question

Can a ladder based on title complexity save bytes without degrading QoE?

## Procedure

Separate titles into low, medium, and high complexity using motion, cuts,
texture, and resolution. Perform a short analysis, generate bitrate x VMAF
curves, and select points with minimum quality and useful distance between
representations.

Compare with the script's fixed ladder:

```bash
scripts/generate-ladder.sh --input samples/raw/big-buck-bunny-1080p-normal.mp4
```

## Criteria

For the same quality target, the per-title ladder should reduce bitrate/bytes or
improve quality without increasing rebuffering. Verify that a variant remains
compatible with slow networks and that segments remain aligned.

## Caveats

The cost of additional encoding, cache fragmentation, number of variants, and
decision time are part of the economic result.
