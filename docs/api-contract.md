# Contrato HTTP do StreamLab

Contrato mínimo para o laboratório. IDs são UUIDs; timestamps são RFC 3339 em
UTC. A API não serve bytes de segmentos.

## Regras comuns

- `Content-Type: application/json` para requests/responses JSON.
- `X-Request-ID` é aceito e devolvido; a API gera um ID quando ausente.
- Erros usam `{ "code": "machine_code", "message": "safe message", "requestId": "uuid" }`.
- `POST` que cria trabalho retorna `202 Accepted` quando o processamento é assíncrono.
- Clientes devem tratar `429` e `5xx` com backoff e não repetir uploads sem idempotency key.

## Endpoints

| Método | Rota | Sucesso | Uso |
|---|---|---:|---|
| `POST` | `/videos` | `201` | Cria metadados e inicia upload |
| `POST` | `/videos/{id}/upload` | `202` | Upload multipart na fase inicial |
| `GET` | `/videos` | `200` | Lista vídeos paginados |
| `GET` | `/videos/{id}` | `200` | Metadados e estado |
| `GET` | `/videos/{id}/status` | `200` | Estado e progresso do pipeline |
| `GET` | `/videos/{id}/playback` | `200` | URL assinada de playback |
| `DELETE` | `/videos/{id}` | `204` | Solicita remoção lógica/assíncrona |
| `POST` | `/playback/sessions` | `201` | Abre sessão de QoE |
| `POST` | `/playback/events` | `202` | Envia evento de playback |
| `GET` | `/healthz` | `200` | Liveness sem dependências pesadas |
| `GET` | `/readyz` | `200` | Readiness de dependências |

## Criar vídeo

`POST /videos`

```json
{
  "title": "Big Buck Bunny",
  "source": { "filename": "big-buck-bunny.avi", "contentType": "video/x-msvideo" }
}
```

Resposta `201`:

```json
{
  "id": "4a0f5a4d-0cc5-4d26-8d34-8f59e8e948d0",
  "title": "Big Buck Bunny",
  "status": "UPLOADING",
  "upload": { "method": "PUT", "url": "https://object.example/upload-token", "expiresAt": "2026-09-10T12:00:00Z" },
  "createdAt": "2026-09-10T11:00:00Z"
}
```

## Status

`GET /videos/{id}/status`

```json
{
  "videoId": "4a0f5a4d-0cc5-4d26-8d34-8f59e8e948d0",
  "status": "ENCODING",
  "progress": 0.62,
  "jobs": [
    { "profile": "720p", "status": "COMPLETED", "attempts": 1 },
    { "profile": "1080p", "status": "RUNNING", "attempts": 1 }
  ],
  "updatedAt": "2026-09-10T11:03:00Z"
}
```

## Playback

`GET /videos/{id}/playback`

```json
{
  "videoId": "4a0f5a4d-0cc5-4d26-8d34-8f59e8e948d0",
  "protocol": "HLS",
  "manifest": "https://cdn.example/video/4a0f.../master.m3u8?token=...",
  "expiresAt": "2026-09-10T12:15:00Z",
  "tracks": {
    "audio": [{ "language": "en-US", "role": "main" }],
    "subtitles": [{ "language": "pt-BR", "format": "webvtt" }]
  }
}
```

Só vídeos `READY` podem retornar playback. O token cobre manifest e segmentos,
tem expiração curta e não deve ser registrado em logs.

## Telemetria

`POST /playback/events`

```json
{
  "sessionId": "session-uuid",
  "videoId": "video-uuid",
  "sequence": 17,
  "event": "quality_change",
  "at": "2026-09-10T11:04:00.120Z",
  "positionSeconds": 42.5,
  "variant": { "height": 720, "bitrate": 2500000 },
  "network": { "throughputBps": 4100000, "bufferSeconds": 18.2 }
}
```

Eventos aceitos: `play`, `pause`, `seek`, `buffer_start`, `buffer_end`,
`quality_change`, `playback_error`, `video_complete`. `sequence` permite
deduplicação; o servidor deve aceitar reordenação limitada.
