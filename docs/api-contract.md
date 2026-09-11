# StreamLab HTTP Contract

Executable contract for the current API. Generated IDs have prefixes (`vid-`,
`job-`, `session-`, and `event-`) followed by 16 hexadecimal characters; public
fixture IDs are `big-buck-bunny` and `abr-lab`. Timestamps are RFC 3339 in UTC.
Metadata, queue, telemetry, and catalog live in memory; local media and artifacts
live in the configured storage directory.

## Bases and Common Rules

When starting with Compose, use `http://localhost:3000/api`: NGINX removes `/api/`
before forwarding the request to the API. NGINX exposes `/healthz` without the
prefix. When running `go run ./apps/api` directly, the API listens on
`http://localhost:8080` and routes do not have `/api`.

- JSON usa `Content-Type: application/json`; upload usa `multipart/form-data`.
- Errors use `{ "error": "message" }`.
- The API does not implement authentication, signed URLs, `X-Request-ID`,
  pagination, `ETag`, or `If-Range`.
- The upload body limit is 2 GiB.
- Upload `POST` returns `202 Accepted`; the worker updates the status afterward.

## Endpoints

| Method | Route | Success | Use |
|---|---|---:|---|
| `GET` | `/health` or `/healthz` | `200` | API health |
| `GET` | `/metrics` | `200` | Prometheus metrics in text format |
| `GET` | `/videos` | `200` | List `{ "videos": [...] }` |
| `POST` | `/videos` | `201` or `202` | Create JSON metadata or receive multipart upload |
| `GET` | `/videos/{id}` | `200` | Video metadata |
| `POST` | `/videos/{id}/upload` | `202` | Multipart upload for a created video |
| `GET` | `/videos/{id}/status` | `200` | Status and queue jobs |
| `GET`/`HEAD` | `/videos/{id}/stream` | `200`/`206` | Local MP4 with Range support |
| `GET` | `/videos/{id}/playback` | `200` | Playback selection |
| `GET` | `/videos/{id}/playback/hls` | `200` | Local HLS manifest |
| `GET` | `/videos/{id}/playback/dash` | `200` | Local DASH manifest |
| `GET` | `/videos/{id}/playback/{hls\|dash}/{name}` | `200` | Local segment or artifact |
| `POST` | `/playback/sessions` or `/sessions` | `201` | Create telemetry session |
| `POST` | `/playback/events` or `/telemetry/playback` | `202` | Record playback event |

The routes `DELETE /videos/{id}` and `GET /readyz` do not exist in the current
implementation.

## Create and Upload a Video

There are two flows. Creating metadata does not write bytes or enqueue work; a
subsequent upload does. For the direct flow, send `file` (the `video` alias is
also accepted):

```bash
BASE=http://localhost:3000/api
```

```bash
curl -X POST "$BASE/videos" \
  -F 'title=Minha aula' \
  -F 'file=@samples/raw/big-buck-bunny-1080p-normal.mp4;type=video/mp4'
```

Direct response: `202` and a `Video` object with `status: "PROCESSING"`, along
with the `Location: /videos/{id}` header. The same flow through a separate
endpoint is:

```bash
created=$(curl -sS -X POST "$BASE/videos" \
  -H 'Content-Type: application/json' \
  -d '{"title":"Minha aula","source":{"filename":"aula.mov","contentType":"video/quicktime"}}')
id=$(printf '%s' "$created" | jq -r .id)
curl -X POST "$BASE/videos/$id/upload" \
  -F 'file=@samples/raw/big-buck-bunny-1080p-normal.mp4;type=video/mp4'
```

The creation JSON also accepts `filename` and `contentType` at the top level, or
`content_type` inside or outside `source`. The `201` response from this first
step resembles:

```json
{
  "id": "vid-0123456789abcdef",
  "title": "Minha aula",
  "filename": "aula.mov",
  "content_type": "video/quicktime",
  "status": "PROCESSING",
  "source": "upload",
  "variants": null,
  "manifests": null,
  "created_at": "2026-09-11T12:00:00Z",
  "updated_at": "2026-09-11T12:00:00Z"
}
```

## Status

`GET /videos/{id}/status` returns the video and the jobs found in the queue:

```json
{
  "video_id": "vid-0123456789abcdef",
  "videoId": "vid-0123456789abcdef",
  "status": "READY",
  "video": { "id": "vid-0123456789abcdef", "status": "READY" },
  "jobs": [
    { "video_id": "vid-0123456789abcdef", "kind": "transcode", "state": "SUCCEEDED", "attempts": 1, "max_retries": 3 }
  ],
  "updated_at": "2026-09-11T12:00:10Z",
  "updatedAt": "2026-09-11T12:00:10Z"
}
```

Videos use `PROCESSING`, `READY`, or `FAILED` (the current processor publishes
`READY` after generating artifacts; a job failure can leave the video in
`PROCESSING` while the job moves to `DLQ`). Jobs use `QUEUED`, `RUNNING`,
`RETRYING`, `SUCCEEDED`, or `DLQ`. There is no percentage progress field.

## Range and Playback

After a local upload becomes `READY`, the original MP4 is available at
`/videos/{id}/stream`:

```bash
curl -i -H 'Range: bytes=0-1023' "$BASE/videos/$id/stream" -o /tmp/range.bin
curl -I -H 'Range: bytes=1024-' "$BASE/videos/$id/stream"
```

Without `Range`, the response is `200`; with a valid Range, it is `206 Partial
Content`, `Accept-Ranges: bytes`, `Content-Range`, and `Content-Length`. The
server accepts single ranges `bytes=start-end`, `bytes=start-`, and
`bytes=-suffix`; an invalid range returns `416` with `Content-Range: bytes
*/tamanho`. `HEAD` retains the headers and sends no body. This is file transport,
not ABR.

Local `READY` videos return playback similar to:

```json
{
  "video_id": "vid-0123456789abcdef",
  "videoId": "vid-0123456789abcdef",
  "status": "READY",
  "protocol": "HLS",
  "manifest": "/videos/vid-0123456789abcdef/playback/hls",
  "protocols": {
    "hls": "/videos/vid-0123456789abcdef/playback/hls",
    "dash": "/videos/vid-0123456789abcdef/playback/dash"
  },
  "manifests": { "hls": "hls.m3u8", "dash": "manifest.mpd" },
  "variants": [{ "name": "720p", "width": 1280, "height": 720, "bitrate": 3000000, "video_codec": "h264", "audio_codec": "aac" }]
}
```

The local processor publishes both manifests. Without `ffmpeg`, it uses an HLS
fixture with a six-second playlist and an educational DASH manifest; with
`ffmpeg` available, it generates real HLS from the file. The independent scripts
in `scripts/` generate HLS master packages with variants and a complete DASH
package for study.

Only `READY` videos serve manifests and artifacts. For fixtures without local
bytes, `/videos/big-buck-bunny/playback` points to Google's public MP4 and
`/videos/abr-lab/playback` points to Mux's public HLS; these URLs are not signed
by the API.

## Telemetry

Create a session with snake_case:

```bash
curl -sS -X POST "$BASE/playback/sessions" \
  -H 'Content-Type: application/json' \
  -d '{"video_id":"big-buck-bunny","user_id":"student-1"}'
```

Send the returned `id` in events:

```bash
curl -i -X POST "$BASE/playback/events" \
  -H 'Content-Type: application/json' \
  -d '{"video_id":"big-buck-bunny","session_id":"session-...","type":"quality_change","position":42.5,"payload":{"height":720,"bitrate":2500000}}'
```

`video_id`, `session_id`, `position`, and `payload` are optional; only `type` is
required. If `session_id` is provided, the session must exist. The server accepts
any type string, returns `202` with `{ "accepted": true, "event_id": "..." }`,
and keeps sessions/events only in memory. There is no `sequence`, deduplication,
or closed event list. `/metrics` exposes `streamlab_http_requests_total`,
`streamlab_playback_events_total`, and `streamlab_playback_events_stored` in
Prometheus format.

## Practical Learning

To separate the layers, compare `GET /stream` with the HLS manifest: the former
uses Range on an object; the latter references segments. Generate the same source
with `scripts/generate-ladder.sh`, `scripts/generate-hls.sh`, and
`scripts/generate-dash.sh`, validate with `scripts/validate-media.sh`, and
inspect `#EXTINF`, `Representation`, codec, resolution, and bitrate with
`scripts/ffprobe-media.sh`. Always record the FFmpeg version, preset, GOP,
segment duration, and artifact size before comparing results.
