# 009 — VMAF

## Question

What bitrate delivers acceptable perceptual quality at each resolution?

## Procedure

Use the original as the aligned reference and each encode as distorted. Ensure
the same framerate, resolution, crop, and duration; normalize before calculating
if necessary. With `libvmaf` available, run one measurement per variant and save
the JSON/CSV and model version.

Conceptual example:

```bash
ffmpeg -i encoded.mp4 -i original.mp4 -lavfi \
  "[0:v]setpts=PTS-STARTPTS[dist];[1:v]setpts=PTS-STARTPTS[ref];[dist][ref]libvmaf=log_fmt=json:log_path=vmaf.json" \
  -f null -
```

## Measure and Interpret

Compare mean VMAF and p5/p1 with bitrate, size, and encoding time. Look for
ladder steps, not a single magic number. VMAF does not measure startup,
rebuffering, audio artifacts, or device compatibility.
