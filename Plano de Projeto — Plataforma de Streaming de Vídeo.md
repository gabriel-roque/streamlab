# Projeto: StreamLab

## 1. Objetivo

Construir uma plataforma de **Video on Demand (VOD)** capaz de receber vídeos enviados por usuários, processá-los em diferentes resoluções e bitrates e reproduzi-los através de streaming adaptativo.

O projeto deve servir como laboratório para estudar:

- Encoding e transcoding
- Codecs de vídeo e áudio
- Containers
- Bitrate
- GOP e Keyframes
- HLS
- MPEG-DASH
- Adaptive Bitrate Streaming
- FFmpeg
- Segmentação de vídeo
- CDN
- Object Storage
- Cache
- Range Requests
- Streaming progressivo
- Streaming adaptativo
- Filas
- Processamento assíncrono
- Workers
- Concorrência
- Escalabilidade horizontal
- Observabilidade
- Métricas de QoE
- DRM conceitualmente
- Arquiteturas utilizadas por grandes plataformas de streaming

---

# 2. Resultado final esperado

Ao final do projeto deverá ser possível acessar:

```text
http://localhost:3000
```

e visualizar uma biblioteca semelhante a:

```text
+---------------------------------------------------+
| StreamLab                                         |
+---------------------------------------------------+

   Meus vídeos

   +--------------+
   |              |
   | thumbnail    |
   |              |
   +--------------+
   Big Buck Bunny
   12:32
   READY

   +--------------+
   |              |
   | thumbnail    |
   |              |
   +--------------+
   Test Movie
   43:15
   PROCESSING
```

Ao selecionar um vídeo:

```text
+---------------------------------------------------+
|                                                   |
|                                                   |
|                VIDEO PLAYER                       |
|                                                   |
|                                                   |
+---------------------------------------------------+

Big Buck Bunny

Quality: AUTO

Current:
1920x1080
5.8 Mbps

Buffer:
23 seconds

Protocol:
HLS

Codec:
H.264
```

O player deverá alterar automaticamente a qualidade de reprodução de acordo com a velocidade da conexão.

Exemplo:

```text
Internet rápida

1080p
6000 kbps

        ↓

queda da conexão

720p
3000 kbps

        ↓

conexão ruim

480p
1200 kbps
```

sem interromper significativamente a reprodução.

---

# 3. Arquitetura conceitual

A arquitetura inicial será:

```text
                     ┌──────────────┐
                     │   Frontend   │
                     │ React Player │
                     └──────┬───────┘
                            │
                            │ HTTP
                            ▼
                     ┌──────────────┐
                     │     API      │
                     │   Gateway    │
                     └──────┬───────┘
                            │
              ┌─────────────┼──────────────┐
              │             │              │
              ▼             ▼              ▼
        ┌──────────┐   ┌──────────┐   ┌─────────┐
        │PostgreSQL│   │  Redis   │   │  MinIO  │
        └──────────┘   └──────────┘   └────┬────┘
                                           │
                                           │ vídeo original
                                           │
                                     ┌─────▼──────┐
                                     │   Queue    │
                                     │ RabbitMQ / │
                                     │   Kafka    │
                                     └─────┬──────┘
                                           │
                                           ▼
                                    ┌─────────────┐
                                    │ Transcoder  │
                                    │   Worker    │
                                    │             │
                                    │   FFmpeg    │
                                    └─────┬───────┘
                                          │
                                          ▼
                                      MinIO
                                          │
                                          │
                                  HLS / MPEG-DASH
                                          │
                                          ▼
                                      NGINX
                                          │
                                          ▼
                                       Player
```

---

# 4. Stack sugerida

Considerando uma arquitetura voltada para backend e sistemas distribuídos:

### Backend

```text
Go
```

Framework opcional:

```text
Gin
Fiber
Echo
Chi
```

Minha sugestão:

```text
Go + Chi
```

para manter o projeto próximo da biblioteca padrão.

---

### Banco

```text
PostgreSQL
```

Guardar:

```text
videos
video_assets
encoding_jobs
video_variants
video_segments
playback_sessions
playback_events
```

---

### Object Storage

Local:

```text
MinIO
```

Cloud futuramente:

```text
AWS S3
```

---

### Processamento

```text
FFmpeg
FFprobe
```

---

### Mensageria

Primeira implementação:

```text
RabbitMQ
```

ou:

```text
Redis Streams
```

Depois evoluir para:

```text
Kafka
```

para estudar pipelines de mídia em larga escala.

---

### Frontend

```text
React
Vite
TypeScript
```

Player:

```text
hls.js
```

Posteriormente:

```text
Shaka Player
```

para estudar HLS + MPEG-DASH.

---

### Proxy / Media Server

```text
NGINX
```

---

### Infraestrutura

```text
Docker
Docker Compose
```

Posteriormente:

```text
Kubernetes
```

---

# 5. Estrutura do projeto

```text
streamlab/

├── apps/
│
│   ├── api/
│
│   ├── transcoder/
│
│   ├── metadata-worker/
│
│   └── frontend/
│
├── internal/
│
│   ├── video/
│   ├── encoding/
│   ├── storage/
│   ├── queue/
│   └── telemetry/
│
├── infrastructure/
│
│   ├── docker/
│   ├── nginx/
│   ├── prometheus/
│   ├── grafana/
│   └── minio/
│
├── scripts/
│
├── samples/
│
├── docs/
│
│   ├── architecture/
│   ├── adr/
│   └── experiments/
│
├── docker-compose.yml
│
└── README.md
```

---

# 6. Fase 0 — fundamentos

Antes da implementação principal, realizar alguns experimentos isolados.

Objetivo:

entender o que realmente existe dentro de um arquivo de vídeo.

Baixar um vídeo público de teste, por exemplo:

```text
Big Buck Bunny
```

Executar:

```bash
ffprobe video.mp4
```

Analisar:

```text
Container

MP4

Video Stream

codec: H264
resolution: 1920x1080
fps: 24
bitrate: 5 Mbps

Audio Stream

codec: AAC
sample rate: 48 kHz
channels: stereo
```

Conceitos que precisam ficar claros:

```text
Container != Codec
```

Exemplo:

```text
MP4
 ├── H.264 Video
 └── AAC Audio
```

Outros containers:

```text
MKV
WebM
MOV
TS
```

Codecs:

```text
H.264 / AVC
H.265 / HEVC
VP9
AV1
```

---

# 7. Experimento 1 — Streaming progressivo

Antes de implementar HLS.

Criar:

```text
GET /videos/{id}/stream
```

Inicialmente retornar:

```text
video.mp4
```

O player HTML5 deverá conseguir reproduzir:

```html
<video controls>
```

Depois implementar:

```text
HTTP Range Requests
```

Exemplo:

```text
Range: bytes=1000000-2000000
```

Servidor responde:

```text
HTTP 206 Partial Content
```

Com:

```text
Content-Range
Accept-Ranges
Content-Length
```

Essa etapa serve para entender como o navegador consegue fazer:

```text
seek
pause
resume
buffer
```

sem baixar todo o arquivo.

---

# 8. Fase 1 — Upload

Criar:

```text
POST /videos
```

Fluxo:

```text
Client
   │
   ▼
API
   │
   ▼
MinIO
```

Evitar enviar arquivos gigantes através da própria API futuramente.

Primeira versão:

```text
multipart/form-data
```

Segunda versão:

```text
Pre-Signed URL
```

Fluxo:

```text
Client
   │
   │ solicita upload
   ▼
API
   │
   │ presigned URL
   ▼
Client
   │
   │ upload direto
   ▼
MinIO
```

Isso reproduz um padrão comum de sistemas de mídia em cloud.

---

# 9. Modelo Video

Exemplo:

```json
{
  "id": "uuid",
  "title": "Big Buck Bunny",
  "status": "UPLOADING",
  "originalFile": "videos/{id}/original.mp4",
  "duration": null,
  "width": null,
  "height": null,
  "codec": null,
  "createdAt": ""
}
```

Status:

```text
UPLOADING
UPLOADED
ANALYZING
ENCODING
PACKAGING
READY
FAILED
```

---

# 10. Fase 2 — Media Probe

Depois do upload:

```text
VideoUploaded
```

deve gerar uma mensagem para uma fila.

```text
Queue:

video.analyze
```

Um worker executa:

```bash
ffprobe
```

Extrair:

```text
duration
width
height
fps
video codec
audio codec
bitrate
audio channels
sample rate
```

Persistir essas informações.

---

# 11. Fase 3 — Primeiro Transcoding

Pegar:

```text
1080p original
```

e gerar:

```text
720p
```

usando FFmpeg.

Exemplo conceitual:

```text
Original

1920x1080
12 Mbps

       ↓

FFmpeg

       ↓

1280x720
3 Mbps
```

Depois gerar várias versões.

---

# 12. Encoding Ladder

Criar uma ladder inicial:

| Qualidade | Resolução | Bitrate aproximado |
|---|---:|---:|
| 240p | 426x240 | 400 kbps |
| 360p | 640x360 | 800 kbps |
| 480p | 854x480 | 1.2 Mbps |
| 720p | 1280x720 | 2.5 Mbps |
| 1080p | 1920x1080 | 5 Mbps |

Não gerar resoluções superiores à original.

Se o original for:

```text
720p
```

gerar somente:

```text
240
360
480
720
```

---

# 13. Fase 4 — HLS

Esta é uma das etapas mais importantes.

Transformar:

```text
video.mp4
```

em:

```text
video/
│
├── master.m3u8
│
├── 1080p/
│   ├── playlist.m3u8
│   ├── segment000.ts
│   ├── segment001.ts
│   └── segment002.ts
│
├── 720p/
│
├── 480p/
│
└── 360p/
```

Cada arquivo:

```text
.ts
```

representa alguns segundos do vídeo.

Por exemplo:

```text
6 segundos
```

---

# 14. Manifest HLS

O arquivo:

```text
master.m3u8
```

pode apontar para:

```text
#EXTM3U

#EXT-X-STREAM-INF:BANDWIDTH=5000000,RESOLUTION=1920x1080
1080p/playlist.m3u8

#EXT-X-STREAM-INF:BANDWIDTH=2500000,RESOLUTION=1280x720
720p/playlist.m3u8

#EXT-X-STREAM-INF:BANDWIDTH=1200000,RESOLUTION=854x480
480p/playlist.m3u8
```

O player escolhe a qualidade.

---

# 15. Adaptive Bitrate Streaming

Simular uma conexão:

```text
10 Mbps
```

Player escolhe:

```text
1080p
```

Reduzir através do Chrome DevTools:

```text
1.5 Mbps
```

O player deverá mudar para:

```text
480p
```

sem reiniciar o vídeo.

Esse experimento demonstra:

```text
Adaptive Bitrate Streaming
```

ou:

```text
ABR
```

---

# 16. Conceito fundamental — segmentos alinhados

Estudar:

```text
GOP
I-frame
P-frame
B-frame
Keyframe
```

Os vídeos de diferentes resoluções precisam possuir keyframes compatíveis.

Exemplo:

```text
1080p

segment 01
0s -------- 6s

720p

segment 01
0s -------- 6s

480p

segment 01
0s -------- 6s
```

Isso permite trocar:

```text
1080p → 480p
```

no meio da reprodução.

---

# 17. Fase 5 — Transcoding distribuído

Inicialmente:

```text
1 worker
```

Depois:

```text
video uploaded
      │
      ▼
Encoding Job
      │
      ├── 1080p
      ├── 720p
      ├── 480p
      └── 360p
```

Cada qualidade poderá ser processada por um worker diferente.

Arquitetura:

```text
                   Queue

        ┌────────────┼────────────┐
        │            │            │
        ▼            ▼            ▼

     Worker 1     Worker 2     Worker 3

      1080p         720p         480p
```

Depois adicionar:

```text
Worker 4
Worker 5
Worker 6
```

e verificar ganho de throughput.

---

# 18. Encoding Job

Modelo:

```text
EncodingJob

id
video_id
profile
resolution
bitrate
codec

status

PENDING
RUNNING
COMPLETED
FAILED

attempts
worker_id
started_at
finished_at
```

Isso permitirá estudar:

```text
idempotência
retry
dead letter
worker crash
job recovery
concorrência
```

---

# 19. Fase 6 — MPEG-DASH

Depois que HLS estiver funcionando, implementar:

```text
MPEG-DASH
```

Manifest:

```text
manifest.mpd
```

Estrutura:

```text
video/

├── manifest.mpd

├── video_1080/
├── video_720/
├── video_480/
└── audio/
```

Comparar:

```text
HLS
vs
DASH
```

---

# 20. Fase 7 — Separar áudio e vídeo

Primeira implementação:

```text
video + audio
```

Depois separar:

```text
Video Track

1080p
720p
480p
```

e:

```text
Audio Track

AAC 128 kbps
AAC 256 kbps
```

O player deverá combinar ambos.

Essa etapa aproxima bastante o projeto de plataformas reais.

---

# 21. Fase 8 — Multi-áudio

Adicionar:

```text
Português
English
Español
```

Estrutura:

```text
audio/

pt/
en/
es/
```

O manifest deverá permitir selecionar idioma.

---

# 22. Fase 9 — Legendas

Suporte:

```text
WebVTT
```

Exemplo:

```text
subtitles/

pt-BR.vtt
en-US.vtt
```

Player:

```text
Português
English
Off
```

---

# 23. Fase 10 — Thumbnail

Worker deverá extrair uma imagem utilizando FFmpeg.

Por exemplo:

```text
30% da duração
```

Gerar:

```text
thumbnail.jpg
```

---

# 24. Preview durante seek

Gerar:

```text
sprite sheet
```

Exemplo:

```text
+-----+-----+-----+-----+
| 0s  | 10s | 20s | 30s |
+-----+-----+-----+-----+
| 40s | 50s | 60s | 70s |
+-----+-----+-----+-----+
```

Quando o usuário posicionar o mouse na timeline:

```text
01:32
[preview]
```

---

# 25. Fase 11 — CDN local

Adicionar:

```text
NGINX
```

em frente ao MinIO.

Fluxo:

```text
Player
   │
   ▼
NGINX
   │
   ▼
MinIO
```

Adicionar cache:

```text
segment001.ts
segment002.ts
```

Primeira chamada:

```text
MISS
```

Segunda:

```text
HIT
```

Estudar:

```text
Cache-Control
ETag
TTL
Cache Hit Ratio
```

---

# 26. Simulação de CDN

Criar três containers:

```text
cdn-node-1
cdn-node-2
cdn-node-3
```

Arquitetura:

```text
               Load Balancer
                     │
       ┌─────────────┼─────────────┐
       │             │             │
       ▼             ▼             ▼

    CDN 1          CDN 2          CDN 3

       │             │             │
       └─────────────┼─────────────┘
                     ▼

                  MinIO
```

Medir:

```text
cache hit
cache miss
bandwidth origin
latência
```

---

# 27. Fase 12 — Telemetria do Player

O frontend deverá enviar eventos:

```text
play
pause
seek
buffer_start
buffer_end
quality_change
playback_error
video_complete
```

Endpoint:

```text
POST /telemetry/playback
```

---

# 28. Quality of Experience

Criar métricas semelhantes às utilizadas por plataformas reais.

Exemplo:

```text
Startup Time

tempo entre:

click play
→
primeiro frame
```

---

### Rebuffer Ratio

```text
tempo bufferizando
-------------------
tempo reproduzindo
```

---

### Average Bitrate

Qualidade média assistida.

---

### Quality Switches

Quantidade de mudanças:

```text
1080
 ↓
720
 ↓
480
 ↑
720
```

---

# 29. Métricas

Adicionar:

```text
Prometheus
Grafana
```

Métricas do backend:

```text
videos_uploaded_total

encoding_jobs_total

encoding_job_duration_seconds

encoding_failures_total

worker_active_jobs

queue_depth
```

Player:

```text
playback_startup_seconds

playback_rebuffers_total

playback_buffer_seconds

playback_quality_changes_total
```

---

# 30. Dashboard

Criar dashboard:

```text
STREAMING PLATFORM

Uploads/min

Encoding jobs
███████████ 42

Queue depth
████ 12

Average encoding time
2m 13s

Playback sessions
142

Startup latency
1.3s

Rebuffer ratio
0.8%

CDN cache hit
87%
```

---

# 31. Fase 13 — Falhas

Matar um worker enquanto processa:

```bash
docker kill transcoder-1
```

O sistema deverá detectar:

```text
job incompleto
```

e colocar novamente na fila.

Estudar:

```text
at-least-once delivery

idempotency

retry

visibility timeout

dead letter queue
```

---

# 32. Poison Job

Enviar arquivo:

```text
corrupted.mp4
```

O sistema deve tentar:

```text
attempt 1
attempt 2
attempt 3
```

e depois:

```text
DLQ
```

---

# 33. Fase 14 — Backpressure

Enviar:

```text
100 vídeos
```

simultaneamente.

Workers:

```text
3
```

Observar crescimento:

```text
Queue Depth

0
12
36
72
97
```

Adicionar autoscaling posteriormente.

---

# 34. Kubernetes

Migrar:

```text
Docker Compose
```

para:

```text
Kubernetes
```

Componentes:

```text
api
worker
frontend
nginx
postgres
redis
minio
```

---

# 35. Autoscaling

Escalar transcoding baseado em:

```text
queue depth
```

Exemplo:

```text
queue < 10

2 workers

queue > 20

5 workers

queue > 50

10 workers
```

Essa etapa demonstra processamento distribuído real.

---

# 36. Experimento CPU vs GPU

Executar encoding utilizando:

```text
libx264
```

CPU.

Depois, caso tenha NVIDIA:

```text
NVENC
```

Comparar:

```text
encoding time

CPU usage

GPU usage

quality

file size
```

Exemplo:

```text
Encoding 10 min video

CPU

03:42

NVENC

00:58
```

---

# 37. Codecs

Executar o mesmo vídeo utilizando:

```text
H.264

H.265

VP9

AV1
```

Comparar:

| Codec | Tamanho | Encoding | Qualidade | Compatibilidade |
|---|---:|---:|---:|---|
| H.264 | maior | rápido | boa | excelente |
| H.265 | menor | médio | ótima | boa |
| VP9 | menor | lento | ótima | boa |
| AV1 | muito menor | lento | excelente | crescente |

Não tratar esses valores como absolutos.

O experimento deve produzir dados próprios.

---

# 38. Estudo de VMAF

Adicionar posteriormente:

```text
VMAF
```

Comparar:

```text
Original
vs
encoded
```

Gerar:

```text
VMAF Score
```

Exemplo:

```text
1080p / 8 Mbps

96

1080p / 5 Mbps

93

1080p / 2 Mbps

78
```

Encontrar o melhor equilíbrio:

```text
qualidade
   x
bandwidth
```

---

# 39. Per-Title Encoding

Esta é uma das fases mais interessantes.

Inicialmente existe uma ladder fixa:

```text
1080 → 5 Mbps
720  → 2.5 Mbps
480  → 1.2 Mbps
```

Mas vídeos diferentes possuem necessidades diferentes.

Exemplo:

```text
Desenho animado

1080p
2.5 Mbps
```

pode ser suficiente.

Enquanto:

```text
filme de ação

1080p
6 Mbps
```

pode ser necessário.

Criar experimentalmente:

```text
Per-Title Encoding
```

Analisar o vídeo e gerar uma ladder específica.

---

# 40. Content Complexity

Criar um serviço:

```text
VideoAnalyzer
```

Classificar aproximadamente:

```text
LOW MOTION

MEDIUM MOTION

HIGH MOTION
```

Utilizar métricas do FFmpeg ou análise de frames.

Depois escolher bitrate.

---

# 41. Content-Aware Encoding

Evolução:

```text
Vídeo
   │
   ▼
Content Analyzer
   │
   ▼
Encoding Profile Generator
   │
   ▼
Transcoder
```

Resultado:

```text
encoding ladder dinâmica
```

Essa fase aproxima bastante o laboratório de técnicas utilizadas por serviços comerciais.

---

# 42. Fase 15 — segurança

Implementar:

```text
signed playback URL
```

Exemplo:

```text
/video/123/master.m3u8
```

não deve ser completamente público.

Gerar:

```text
token

exp=...

signature=...
```

---

# 43. DRM

Não implementar inicialmente DRM real.

Primeiro estudar:

```text
Widevine

FairPlay

PlayReady
```

Entender arquitetura:

```text
Player
   │
   ▼
DRM License Server
   │
   ▼
License
   │
   ▼
Player

Encrypted media
   │
   ▼
Decrypt
```

Posteriormente experimentar:

```text
ClearKey
```

como implementação didática.

---

# 44. APIs

API final aproximada:

```text
POST
/videos

POST
/videos/{id}/upload

GET
/videos

GET
/videos/{id}

DELETE
/videos/{id}

GET
/videos/{id}/playback

GET
/videos/{id}/status

POST
/playback/sessions

POST
/playback/events
```

---

# 45. Endpoint Playback

Exemplo:

```text
GET /videos/{id}/playback
```

Resposta:

```json
{
  "videoId": "123",
  "protocol": "HLS",
  "manifest": "https://cdn.streamlab/video/123/master.m3u8",
  "expiresAt": "...",
  "tracks": {
    "audio": [
      "pt-BR",
      "en-US"
    ],
    "subtitles": [
      "pt-BR",
      "en-US"
    ]
  }
}
```

---

# 46. Eventos internos

Utilizar eventos como:

```text
VideoUploaded

VideoAnalyzed

EncodingRequested

EncodingStarted

EncodingCompleted

EncodingFailed

PackagingStarted

PackagingCompleted

VideoReady
```

Pipeline:

```text
VideoUploaded
      │
      ▼
VideoAnalyzed
      │
      ▼
EncodingRequested
      │
      ▼
EncodingCompleted
      │
      ▼
PackagingCompleted
      │
      ▼
VideoReady
```

---

# 47. Machine State

Modelar o vídeo como máquina de estados:

```text
UPLOADED
    │
    ▼
ANALYZING
    │
    ▼
ENCODING
    │
    ▼
PACKAGING
    │
    ▼
READY
```

Falha:

```text
ANY STATE
    │
    ▼
FAILED
```

---

# 48. Testes

Criar:

```text
unit tests

integration tests

end-to-end tests
```

Além dos tradicionais, criar:

```text
media validation tests
```

Depois do encoding executar:

```text
ffprobe
```

e validar automaticamente:

```text
resolution

codec

duration

bitrate

audio
```

---

# 49. Teste de integridade

Comparar:

```text
Original duration

600.03 sec
```

com:

```text
Encoded duration

600.01 sec
```

Definir tolerância.

---

# 50. Load Test

Utilizar:

```text
K6
```

Simular:

```text
100

500

1000

5000
```

players.

Não é necessário reproduzir o vídeo completo.

Solicitar:

```text
manifest

segment1

segment2

segment3
```

simulando comportamento de clientes.

---

# 51. Chaos Testing

Executar experimentos como:

```text
matar transcoder

matar MinIO

matar API

derrubar Redis

aumentar latência

limitar banda

corromper vídeo
```

Documentar comportamento.

---

# 52. ADRs

Criar registros de decisões arquiteturais.

Exemplo:

```text
ADR-001 — HLS como protocolo inicial

ADR-002 — MinIO para object storage

ADR-003 — FFmpeg como transcoder

ADR-004 — encoding assíncrono

ADR-005 — PostgreSQL como metadata store

ADR-006 — segmentos de 6 segundos

ADR-007 — H.264 como codec inicial
```

Cada ADR deve registrar:

```text
Context

Decision

Alternatives

Consequences
```

---

# 53. Experimentos documentados

Criar:

```text
docs/experiments/
```

Exemplo:

```text
001-http-range-request.md

002-hls-segment-duration.md

003-bitrate-comparison.md

004-h264-vs-h265.md

005-cpu-vs-nvenc.md

006-worker-scalability.md

007-cdn-cache.md

008-abr-network-throttling.md

009-vmaf.md

010-per-title-encoding.md
```

Isso transforma o repositório em um verdadeiro estudo de engenharia.

---

# 54. Roadmap

Executar aproximadamente nesta ordem:

```text
FASE 1
Video fundamentals

        ↓

FASE 2
Upload

        ↓

FASE 3
HTTP Range Streaming

        ↓

FASE 4
FFprobe

        ↓

FASE 5
FFmpeg Transcoding

        ↓

FASE 6
Encoding Ladder

        ↓

FASE 7
HLS

        ↓

FASE 8
Adaptive Bitrate

        ↓

FASE 9
Async Transcoding

        ↓

FASE 10
Distributed Workers

        ↓

FASE 11
MinIO

        ↓

FASE 12
NGINX Cache

        ↓

FASE 13
Playback Telemetry

        ↓

FASE 14
Prometheus + Grafana

        ↓

FASE 15
MPEG-DASH

        ↓

FASE 16
Audio Tracks

        ↓

FASE 17
Subtitles

        ↓

FASE 18
Thumbnail / Preview

        ↓

FASE 19
Kubernetes

        ↓

FASE 20
Autoscaling

        ↓

FASE 21
CPU vs GPU encoding

        ↓

FASE 22
H264/H265/VP9/AV1

        ↓

FASE 23
VMAF

        ↓

FASE 24
Per-title Encoding

        ↓

FASE 25
Content-aware Encoding

        ↓

FASE 26
Signed playback

        ↓

FASE 27
DRM concepts
```

---

# 55. Escala hipotética

Depois que tudo estiver funcionando, fazer um exercício arquitetural.

Suponha:

```text
10 milhões usuários

1 milhão simultâneos

Catálogo

100.000 vídeos
```

Se cada usuário consumir:

```text
5 Mbps
```

então:

```text
1.000.000 × 5 Mbps

=

5 Tbps
```

A discussão arquitetural muda completamente.

O backend da aplicação deixa de ser o problema principal.

O principal desafio passa a ser:

```text
distribuição de mídia
```

Por isso plataformas grandes dependem fortemente de:

```text
CDN
edge caching
multi-region
traffic engineering
```

---

# 56. Arquitetura de referência final

```text
                       USERS
                         │
                         ▼
                  ┌─────────────┐
                  │ Global DNS  │
                  └──────┬──────┘
                         │
                         ▼

               ┌───────────────────┐
               │       CDN         │
               │                   │
               │ Edge Cache        │
               └─────────┬─────────┘
                         │
                         │ cache miss
                         ▼

                 ┌───────────────┐
                 │ Object Storage│
                 │               │
                 │ HLS / DASH    │
                 └───────────────┘


UPLOAD PIPELINE


Client
  │
  ▼
Upload API
  │
  ▼
Object Storage
  │
  ▼
Event
  │
  ▼
Message Broker
  │
  ├─────────────┐
  │             │
  ▼             ▼
Analyzer     Transcoder
                │
       ┌────────┼────────┐
       │        │        │
       ▼        ▼        ▼
     1080p     720p     480p
       │        │        │
       └────────┼────────┘
                │
                ▼
             Packager
                │
                ▼
          Object Storage
                │
                ▼
               CDN
```

---

# 57. O que você deverá conseguir explicar depois do projeto

Ao terminar o StreamLab, você deverá conseguir responder tecnicamente:

```text
Por que não simplesmente entregar um MP4?

Como HTTP Range funciona?

Qual é a diferença entre codec e container?

O que é bitrate?

O que é bitrate variável?

O que é transcoding?

O que é encoding?

O que é HLS?

O que é MPEG-DASH?

O que é um manifest?

O que é um media segment?

Como um player troca de resolução sem reiniciar?

O que é Adaptive Bitrate?

O que são I/P/B frames?

O que é GOP?

Por que alinhar keyframes?

Por que vídeo é dividido em segmentos?

Por que diferentes resoluções precisam de bitrates diferentes?

Como escolher uma encoding ladder?

Como FFmpeg funciona?

Por que transcoding é CPU intensive?

Quando utilizar GPU encoding?

Por que object storage é utilizado?

Por que um CDN é fundamental?

Como cache reduz bandwidth?

Como medir qualidade de playback?

O que é rebuffer ratio?

O que é startup latency?

Como escalar transcoding?

Como recuperar encoding jobs após crashes?

Como implementar idempotência em workers?

Como lidar com milhões de usuários simultâneos?

Por que Netflix não simplesmente coloca seus vídeos no S3?
```

Se você souber explicar e demonstrar cada um desses pontos no próprio sistema, já terá aprendido uma quantidade considerável de engenharia de streaming.

---

# 58. Objetivo adicional de arquitetura

Não tente criar uma cópia visual da Netflix.

O valor do projeto está em construir:

```text
Upload Pipeline
       +
Media Processing Pipeline
       +
Encoding Pipeline
       +
Streaming Protocol
       +
ABR Player
       +
CDN
       +
Distributed Processing
       +
Observability
```

O frontend pode permanecer relativamente simples.

A complexidade e o aprendizado devem ficar na infraestrutura e no backend.

---

# 59. Definition of Done

O projeto estará considerado completo quando:

```text
[ ] Usuário consegue subir um MP4

[ ] Upload ocorre diretamente para object storage

[ ] Sistema detecta novo vídeo

[ ] FFprobe extrai metadata

[ ] Encoding ocorre assincronamente

[ ] São produzidas múltiplas resoluções

[ ] Jobs podem executar paralelamente

[ ] HLS é gerado

[ ] MPEG-DASH é gerado

[ ] Player reproduz HLS

[ ] Player possui ABR

[ ] Troca de qualidade ocorre sem interrupção significativa

[ ] Áudio pode ser alterado

[ ] Legendas podem ser alteradas

[ ] Thumbnail é gerado

[ ] Preview da timeline funciona

[ ] NGINX funciona como cache

[ ] Métricas do CDN são coletadas

[ ] Métricas do player são coletadas

[ ] Grafana possui dashboard

[ ] Encoding jobs possuem retry

[ ] DLQ existe

[ ] Workers são idempotentes

[ ] Sistema sobrevive ao crash de worker

[ ] K6 consegue simular players concorrentes

[ ] Workers escalam horizontalmente

[ ] Projeto roda em Kubernetes

[ ] Existe benchmark CPU vs GPU

[ ] Existe benchmark H264/H265/VP9/AV1

[ ] Existe análise VMAF

[ ] Existe experimento de Per-Title Encoding

[ ] ADRs estão documentados

[ ] Experimentos estão documentados

[ ] Toda arquitetura está documentada
```

---

# 60. Meta final

O objetivo não é simplesmente conseguir reproduzir:

```text
movie.mp4
```

O objetivo é conseguir olhar para algo como:

```text
Netflix
Prime Video
Max
YouTube
```

e compreender o caminho aproximado:

```text
arquivo original
      ↓
ingest
      ↓
media analysis
      ↓
encoding ladder
      ↓
transcoding
      ↓
packaging
      ↓
HLS / DASH
      ↓
object storage
      ↓
CDN
      ↓
edge
      ↓
player
      ↓
ABR algorithm
      ↓
frame exibido
```

e conseguir implementar experimentalmente cada uma dessas etapas.

Esse é o verdadeiro objetivo do **StreamLab**.