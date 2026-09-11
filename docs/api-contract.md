# Contrato HTTP do StreamLab

Contrato executável da API atual. IDs gerados têm prefixos (`vid-`, `job-`,
`session-` e `event-`) seguidos por 16 caracteres hexadecimais; os IDs dos
fixtures públicos são `big-buck-bunny` e `abr-lab`. Timestamps são RFC 3339 em
UTC. Metadados, fila, telemetria e catálogo vivem em memória; mídia e artefatos
locais vivem no diretório configurado de storage.

## Bases e regras comuns

Ao subir com Compose, use `http://localhost:3000/api`: o NGINX remove `/api/`
antes de encaminhar a requisição para a API. O NGINX expõe `/healthz` sem o
prefixo. Ao executar `go run ./apps/api` diretamente, a API escuta
`http://localhost:8080` e as rotas não têm `/api`.

- JSON usa `Content-Type: application/json`; upload usa `multipart/form-data`.
- Erros usam `{ "error": "mensagem" }`.
- A API não implementa autenticação, URLs assinadas, `X-Request-ID`, paginação,
  `ETag` ou `If-Range`.
- O limite do corpo de upload é 2 GiB.
- `POST` de upload retorna `202 Accepted`; o worker atualiza o status depois.

## Endpoints

| Método | Rota | Sucesso | Uso |
|---|---|---:|---|
| `GET` | `/health` ou `/healthz` | `200` | Saúde da API |
| `GET` | `/metrics` | `200` | Métricas Prometheus em texto |
| `GET` | `/videos` | `200` | Lista `{ "videos": [...] }` |
| `POST` | `/videos` | `201` ou `202` | Cria metadados JSON ou recebe upload multipart |
| `GET` | `/videos/{id}` | `200` | Metadados do vídeo |
| `POST` | `/videos/{id}/upload` | `202` | Upload multipart para vídeo criado |
| `GET` | `/videos/{id}/status` | `200` | Status e jobs da fila |
| `GET`/`HEAD` | `/videos/{id}/stream` | `200`/`206` | MP4 local com suporte a Range |
| `GET` | `/videos/{id}/playback` | `200` | Seleção de playback |
| `GET` | `/videos/{id}/playback/hls` | `200` | Manifest HLS local |
| `GET` | `/videos/{id}/playback/dash` | `200` | Manifest DASH local |
| `GET` | `/videos/{id}/playback/{hls\|dash}/{name}` | `200` | Segmento ou artefato local |
| `POST` | `/playback/sessions` ou `/sessions` | `201` | Cria sessão de telemetria |
| `POST` | `/playback/events` ou `/telemetry/playback` | `202` | Registra evento de playback |

Não existem as rotas `DELETE /videos/{id}` ou `GET /readyz` na implementação
atual.

## Criar e enviar vídeo

Há dois fluxos. Criar metadados não grava bytes nem enfileira trabalho; o upload
posterior faz isso. Para o fluxo direto, envie `file` (o alias `video` também é
aceito):

```bash
BASE=http://localhost:3000/api
```

```bash
curl -X POST "$BASE/videos" \
  -F 'title=Minha aula' \
  -F 'file=@samples/raw/big-buck-bunny-1080p-normal.mp4;type=video/mp4'
```

Resposta direta: `202` e um objeto `Video` com `status: "PROCESSING"`, além do
header `Location: /videos/{id}`. O mesmo fluxo via endpoint separado é:

```bash
created=$(curl -sS -X POST "$BASE/videos" \
  -H 'Content-Type: application/json' \
  -d '{"title":"Minha aula","source":{"filename":"aula.mov","contentType":"video/quicktime"}}')
id=$(printf '%s' "$created" | jq -r .id)
curl -X POST "$BASE/videos/$id/upload" \
  -F 'file=@samples/raw/big-buck-bunny-1080p-normal.mp4;type=video/mp4'
```

O JSON de criação também aceita `filename` e `contentType` no nível superior,
ou `content_type` dentro/fora de `source`. A resposta `201` desse primeiro
passo é semelhante a:

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

`GET /videos/{id}/status` retorna o vídeo e os jobs encontrados na fila:

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

Vídeos usam `PROCESSING`, `READY` ou `FAILED` (o processador atual publica
`READY` após gerar os artefatos; uma falha de job pode deixar o vídeo em
`PROCESSING` enquanto o job vai para `DLQ`). Jobs usam `QUEUED`, `RUNNING`,
`RETRYING`, `SUCCEEDED` ou `DLQ`. Não há campo de progresso percentual.

## Range e playback

Após um upload local ficar `READY`, o MP4 original está em
`/videos/{id}/stream`:

```bash
curl -i -H 'Range: bytes=0-1023' "$BASE/videos/$id/stream" -o /tmp/range.bin
curl -I -H 'Range: bytes=1024-' "$BASE/videos/$id/stream"
```

Sem `Range`, a resposta é `200`; com Range válido, `206 Partial Content`,
`Accept-Ranges: bytes`, `Content-Range` e `Content-Length`. O servidor aceita
intervalos únicos `bytes=start-end`, `bytes=start-` e `bytes=-suffix`; intervalo
inválido retorna `416` com `Content-Range: bytes */tamanho`. `HEAD` mantém os
headers e não envia o corpo. Isso é transporte de um arquivo, não ABR.

Vídeos locais `READY` retornam playback semelhante a:

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

O processador local publica os dois manifests. Sem `ffmpeg`, ele usa um fixture
HLS de uma playlist de seis segundos e um manifest DASH didático; com `ffmpeg`
disponível, gera HLS real a partir do arquivo. Os scripts independentes em
`scripts/` geram pacotes HLS master com variantes e um pacote DASH completo para
estudo.

Só vídeos `READY` servem manifestos e artefatos. Para os fixtures sem bytes
locais, `/videos/big-buck-bunny/playback` aponta para o MP4 público do Google e
`/videos/abr-lab/playback` aponta para o HLS público do Mux; essas URLs não são
assinadas pela API.

## Telemetria

Crie uma sessão com snake_case:

```bash
curl -sS -X POST "$BASE/playback/sessions" \
  -H 'Content-Type: application/json' \
  -d '{"video_id":"big-buck-bunny","user_id":"student-1"}'
```

Envie o `id` retornado em eventos:

```bash
curl -i -X POST "$BASE/playback/events" \
  -H 'Content-Type: application/json' \
  -d '{"video_id":"big-buck-bunny","session_id":"session-...","type":"quality_change","position":42.5,"payload":{"height":720,"bitrate":2500000}}'
```

`video_id`, `session_id`, `position` e `payload` são opcionais; somente `type` é
obrigatório. Se `session_id` for informado, a sessão precisa existir. O servidor
aceita qualquer string de tipo, retorna `202` com `{ "accepted": true,
"event_id": "..." }` e mantém sessões/eventos somente em memória. Não há
`sequence`, deduplicação ou lista fechada de eventos. `/metrics` expõe
`streamlab_http_requests_total`, `streamlab_playback_events_total` e
`streamlab_playback_events_stored` em formato Prometheus.

## Aprendizado prático

Para separar as camadas, compare `GET /stream` com o manifest HLS: o primeiro
usa Range sobre um objeto; o segundo referencia segmentos. Gere a mesma fonte
com `scripts/generate-ladder.sh`, `scripts/generate-hls.sh` e
`scripts/generate-dash.sh`, valide com `scripts/validate-media.sh` e observe
`#EXTINF`, `Representation`, codec, resolução e bitrate com
`scripts/ffprobe-media.sh`. Registre sempre versão do FFmpeg, preset, GOP,
duração de segmento e tamanho dos artefatos antes de comparar resultados.
