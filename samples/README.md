# Samples

This directory contains only versionable instructions and metadata. Big Buck Bunny
is downloaded on demand from the official Blender Foundation source:

```text
https://download.blender.org/demo/movies/BBB/bbb_sunflower_1080p_30fps_normal.mp4.zip
```

The official file is a ZIP; `scripts/download-bbb.sh` extracts the MP4 and calls
FFprobe. No video, segment, or generated manifest should be committed:

```bash
scripts/download-bbb.sh
scripts/ffprobe-media.sh --input samples/raw/big-buck-bunny-1080p-normal.mp4
```

To transform the source into an H.264/AAC ladder and package it for both
protocols:

```bash
scripts/generate-ladder.sh --input samples/raw/big-buck-bunny-1080p-normal.mp4
scripts/generate-hls.sh --input samples/raw/big-buck-bunny-1080p-normal.mp4 \
  --segment-seconds 6 --output-dir samples/generated/hls --overwrite
scripts/generate-dash.sh --input samples/raw/big-buck-bunny-1080p-normal.mp4 \
  --segment-seconds 6 --output-dir samples/generated/dash --overwrite
scripts/validate-media.sh --input samples/generated/hls --kind hls
scripts/validate-media.sh --input samples/generated/dash --kind dash
```

The generated HLS has `master.m3u8`, resolution-specific playlists, and MPEG-TS
segments; DASH has `manifest.mpd`, initialization files, and chunks.
`--segment-seconds` is a target: check `#EXTINF`, `Representation`, and the
files with `ffprobe-media.sh`. The scripts support local FFmpeg or
`MEDIA_TOOL=docker`.

There are also two public fixtures that are not downloaded automatically when the
API starts: `big-buck-bunny` uses
`https://storage.googleapis.com/gtv-videos-bucket/sample/BigBuckBunny.mp4` as
a remote MP4, and `abr-lab` uses
`https://test-streams.mux.dev/x36xhzz/x36xhzz.m3u8` as a multi-variant HLS.
Query them with `GET /videos/{id}/playback`; they have no local media or manifest
in `LocalStore`.

With Compose, the API is behind the proxy at `http://localhost:3000/api`:

```bash
docker compose up --build
curl http://localhost:3000/api/videos/abr-lab/playback
curl -X POST http://localhost:3000/api/videos \
  -F 'title=BBB local' \
  -F 'file=@samples/raw/big-buck-bunny-1080p-normal.mp4;type=video/mp4'
```

The last command returns `202` and starts asynchronous processing. Check
`/api/videos/{id}/status` until `READY`; then use `/api/videos/{id}/playback`.
The Compose container installs FFmpeg/FFprobe so this flow produces real HLS and
an educational DASH manifest. If the API is run locally without these tools, it
uses the educational fixture; to compare both paths and generate a real DASH
package with variants, run the media scripts explicitly.

The `raw/` and `generated/` directories are ignored locally. To repeat an
experiment in another environment, record the URL, the optional SHA-256 hash of
the downloaded file, the FFmpeg version, and the script parameters in the report.
