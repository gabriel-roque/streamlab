#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
COMPOSE=(docker compose -f "$ROOT_DIR/docker-compose.yml")
SKIP_DOWNLOAD=false
STACK_TIMEOUT="${STACK_TIMEOUT:-300}"
READY_TIMEOUT="${READY_TIMEOUT:-600}"
VIDEO_FILE="${VIDEO_FILE:-$ROOT_DIR/samples/raw/big-buck-bunny-1080p-normal.mp4}"
VIDEO_TITLE="${VIDEO_TITLE:-Big Buck Bunny quick-start}"
SELECTED_PORTS=""

usage() {
  cat <<'EOF'
Start the StreamLab Docker stack and upload a public Big Buck Bunny sample.

Usage: scripts/quick-start.sh [--skip-download]

Ports can be set explicitly with NGINX_PORT, PROMETHEUS_PORT, GRAFANA_PORT,
MINIO_PORT, MINIO_CONSOLE_PORT, POSTGRES_PORT, and REDIS_PORT. If unset, the
script uses the documented default when it is free and otherwise chooses the
next free port.

--skip-download  require the local Big Buck Bunny file and do not download it
EOF
}

die() {
  printf 'error: %s\n' "$*" >&2
  exit 1
}

log() {
  printf '\n==> %s\n' "$*"
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --skip-download) SKIP_DOWNLOAD=true; shift ;;
    -h|--help) usage; exit 0 ;;
    *) die "unknown argument: $1 (use --help for usage)" ;;
  esac
done

require_command() {
  command -v "$1" >/dev/null 2>&1 || die "required command not found: $1"
}

require_command docker
require_command curl
require_command jq

if ! docker info >/dev/null 2>&1; then
  die "Docker is installed but the daemon is not reachable; start Docker and try again"
fi
if ! docker compose version >/dev/null 2>&1; then
  die "Docker Compose v2 is required (docker compose)"
fi

for timeout_value in "$STACK_TIMEOUT" "$READY_TIMEOUT"; do
  [[ "$timeout_value" =~ ^[0-9]+$ ]] || die "timeouts must be whole seconds"
done

port_is_free() {
  local port="$1"
  if (exec 3<>"/dev/tcp/127.0.0.1/$port") 2>/dev/null; then
    exec 3>&-
    return 1
  fi
  return 0
}

existing_service_uses_port() {
  local service="$1"
  local internal_port="$2"
  local port="$3"
  local container_id
  local current_port

  container_id="$("${COMPOSE[@]}" ps -q "$service" 2>/dev/null || true)"
  [[ -n "$container_id" ]] || return 1
  current_port="$(docker inspect --format "{{(index (index .NetworkSettings.Ports \"$internal_port/tcp\") 0).HostPort}}" "$container_id" 2>/dev/null || true)"
  [[ "$current_port" == "$port" ]]
}

port_already_selected() {
  case " $SELECTED_PORTS " in
    *" $1 "*) return 0 ;;
    *) return 1 ;;
  esac
}

choose_port() {
  local variable="$1"
  local default_port="$2"
  local service="$3"
  local internal_port="$4"
  local requested="${!variable-}"
  local candidate

  if [[ -n "$requested" ]]; then
    [[ "$requested" =~ ^[0-9]+$ ]] || die "$variable must be a TCP port number"
    (( 10#$requested >= 1 && 10#$requested <= 65535 )) || die "$variable must be between 1 and 65535"
    requested="$((10#$requested))"
    port_already_selected "$requested" && die "port $requested is assigned more than once"
    if ! port_is_free "$requested" && ! existing_service_uses_port "$service" "$internal_port" "$requested"; then
      die "$variable=$requested is already in use; choose another port"
    fi
    candidate="$requested"
  else
    candidate="$default_port"
    while port_already_selected "$candidate" || { ! port_is_free "$candidate" && ! existing_service_uses_port "$service" "$internal_port" "$candidate"; }; do
      (( candidate < 65535 )) || die "could not find a free port for $variable"
      candidate=$((candidate + 1))
    done
  fi

  printf -v "$variable" '%s' "$candidate"
  export "$variable"
  SELECTED_PORTS+=" $candidate"
  printf '  %-20s %s\n' "$variable" "$candidate"
}

log "Checking host ports"
choose_port NGINX_PORT 3000 nginx 80
choose_port PROMETHEUS_PORT 9090 prometheus 9090
choose_port GRAFANA_PORT 3001 grafana 3000
choose_port MINIO_PORT 9000 minio 9000
choose_port MINIO_CONSOLE_PORT 9001 minio 9001
choose_port POSTGRES_PORT 5432 postgres 5432
choose_port REDIS_PORT 6379 redis 6379

log "Validating Docker Compose configuration"
"${COMPOSE[@]}" config >/dev/null

log "Starting the stack"
"${COMPOSE[@]}" up -d --build

wait_for_healthchecks() {
  local deadline=$((SECONDS + STACK_TIMEOUT))
  local service container_id health
  local all_healthy
  local -a services=(postgres redis minio api frontend nginx prometheus grafana)
  declare -A last_health=()

  log "Waiting for healthchecks (up to ${STACK_TIMEOUT}s)"
  while (( SECONDS < deadline )); do
    all_healthy=true
    for service in "${services[@]}"; do
      container_id="$("${COMPOSE[@]}" ps -q "$service" 2>/dev/null || true)"
      if [[ -z "$container_id" ]]; then
        health="missing"
      else
        health="$(docker inspect --format '{{if .State.Health}}{{.State.Health.Status}}{{else}}{{.State.Status}}{{end}}' "$container_id" 2>/dev/null || printf 'missing')"
      fi

      if [[ "${last_health[$service]-}" != "$health" ]]; then
        printf '  %-12s %s\n' "$service" "$health"
        last_health[$service]="$health"
      fi

      case "$health" in
        healthy) ;;
        starting|created|missing|running) all_healthy=false ;;
        *)
          "${COMPOSE[@]}" logs --tail=80 "$service" >&2 || true
          die "service $service is $health"
          ;;
      esac
    done

    [[ "$all_healthy" == true ]] && return 0
    sleep 2
  done

  "${COMPOSE[@]}" ps >&2 || true
  "${COMPOSE[@]}" logs --tail=80 >&2 || true
  die "the stack did not become healthy within ${STACK_TIMEOUT}s"
}

wait_for_healthchecks

BASE_URL="http://localhost:${NGINX_PORT}"
API_URL="$BASE_URL/api"

log "Checking the public NGINX entrypoint"
curl -fsS --retry 10 --retry-delay 1 "$BASE_URL/healthz" >/dev/null

if [[ "$SKIP_DOWNLOAD" == false && ! -s "$VIDEO_FILE" ]]; then
  log "Downloading Big Buck Bunny"
  (cd "$ROOT_DIR" && "$SCRIPT_DIR/download-bbb.sh")
elif [[ "$SKIP_DOWNLOAD" == true ]]; then
  printf 'Skipping download by request\n'
else
  printf 'Using existing Big Buck Bunny file: %s\n' "$VIDEO_FILE"
fi

[[ -s "$VIDEO_FILE" ]] || die "video file not found or empty: $VIDEO_FILE (remove --skip-download to download it)"

encoded_filename="$(basename "$VIDEO_FILE")"
videos_json="$(curl -fsS --retry 5 --retry-delay 1 "$API_URL/videos")" || die "could not list videos through NGINX"
video_id="$(jq -r --arg title "$VIDEO_TITLE" --arg filename "$encoded_filename" '[.videos[]? | select((.title // "") == $title and (.filename // "") == $filename and (.status // "") != "FAILED") | .id] | first // empty' <<<"$videos_json")"

if [[ -n "$video_id" ]]; then
  printf 'Reusing existing upload: %s\n' "$video_id"
else
  log "Uploading multipart media through NGINX"
  upload_response="$(mktemp)"
  if ! upload_status="$(curl -sS --retry 3 --retry-delay 2 -o "$upload_response" -w '%{http_code}' \
    -F "file=@${VIDEO_FILE};type=video/mp4" \
    -F "title=${VIDEO_TITLE}" \
    "$API_URL/videos")"; then
    printf 'Upload response:\n' >&2
    jq . "$upload_response" >&2 2>/dev/null || true
    rm -f "$upload_response"
    die "multipart upload failed"
  fi
  if [[ "$upload_status" != 202 && "$upload_status" != 201 && "$upload_status" != 200 ]]; then
    printf 'Upload returned HTTP %s:\n' "$upload_status" >&2
    jq . "$upload_response" >&2 2>/dev/null || true
    rm -f "$upload_response"
    die "multipart upload was not accepted"
  fi
  video_id="$(jq -r '.id // .video_id // empty' "$upload_response")"
  rm -f "$upload_response"
  [[ -n "$video_id" ]] || die "upload response did not contain a video id"
  printf 'Uploaded video: %s\n' "$video_id"
fi

encoded_id="$(jq -nr --arg value "$video_id" '$value | @uri')"
status_url="$API_URL/videos/$encoded_id/status"
status_json=""
last_status=""
current_status=""
ready_deadline=$((SECONDS + READY_TIMEOUT))

log "Waiting for video $video_id to reach READY (up to ${READY_TIMEOUT}s)"
while (( SECONDS < ready_deadline )); do
  if ! status_json="$(curl -fsS --retry 3 --retry-delay 1 "$status_url")"; then
    die "could not read video status through NGINX"
  fi
  current_status="$(jq -r '.status // .video.status // empty' <<<"$status_json")"
  progress="$(jq -r '.progress // .video.progress // empty' <<<"$status_json")"
  if [[ "$current_status" != "$last_status" ]]; then
    if [[ -n "$progress" ]]; then
      printf '  status: %s (progress %s)\n' "$current_status" "$progress"
    else
      printf '  status: %s\n' "$current_status"
    fi
    last_status="$current_status"
  fi
  case "$current_status" in
    READY) break ;;
    FAILED)
      printf '%s\n' "$status_json" | jq . >&2
      die "video processing failed"
      ;;
    '') die "video status response did not contain a status" ;;
  esac
  sleep 3
done

[[ "$current_status" == READY ]] || die "video did not reach READY within ${READY_TIMEOUT}s"

playback_json="$(curl -fsS "$API_URL/videos/$encoded_id/playback")" || die "video is READY but playback metadata could not be read"
hls_url="$API_URL/videos/$encoded_id/playback/hls"
mp4_url="$API_URL/videos/$encoded_id/stream"

log "StreamLab is ready"
printf '  Library:       %s/\n' "$BASE_URL"
printf '  Asset screen:  %s/ (select %s)\n' "$BASE_URL" "$VIDEO_TITLE"
printf '  Asset API:     %s/videos/%s\n' "$API_URL" "$encoded_id"
printf '  Status API:    %s\n' "$status_url"
printf '  Playback API:  %s/videos/%s/playback\n' "$API_URL" "$encoded_id"
printf '  HLS playback:  %s\n' "$hls_url"
printf '  MP4 playback:  %s\n' "$mp4_url"
printf '  Grafana:       http://localhost:%s\n' "$GRAFANA_PORT"
printf '  Prometheus:    http://localhost:%s\n' "$PROMETHEUS_PORT"
printf '  MinIO console: http://localhost:%s\n' "$MINIO_CONSOLE_PORT"
printf '  Protocol:      %s\n' "$(jq -r '.protocol // "unknown"' <<<"$playback_json")"

cat <<EOF

  ${COMPOSE[*]} ps
  curl -fsS "$API_URL/videos" | jq
  curl -fsS "$status_url" | jq
  curl -fsS "$API_URL/videos/$encoded_id/playback" | jq
  curl -fsS "$hls_url"
  curl -fL -r 0-1023 -o /tmp/streamlab-sample.mp4 "$mp4_url"
  ${COMPOSE[*]} logs -f api
EOF
