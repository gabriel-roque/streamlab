# 004 — CDN e cache

## Pergunta

Quanto tráfego deixa de chegar à origem quando manifests e segmentos são
cacheáveis?

## Procedimento

O Compose coloca NGINX entre o browser e a API, mas não implementa ainda um CDN
ou proxy de manifests/segmentos para MinIO. Use o fluxo abaixo para confirmar o
comportamento atual; depois coloque um cache reverso externo entre o cliente e
o endpoint e repita duas sessões para o mesmo manifest/segmentos.

```bash
scripts/smoke-http.sh --url http://localhost:3000 \
  --path /api/videos/ID/playback/hls
```

Substitua `ID` por um upload local em estado `READY`. O fixture remoto
`abr-lab` retorna a URL HLS pública em `/api/videos/abr-lab/playback`, mas não
possui manifest local para ser buscado por essa rota.

## Medir

`Age`, `ETag`, `Cache-Control`, hit/miss, latência p50/p95, bytes no origin,
requests ao origin e taxa de revalidação. No Compose atual, `/api/*` é proxy
para a API, que não envia `Age`, `ETag` ou `Cache-Control`; `/media/` é o único
path estático do NGINX e aponta para a mídia local, não para artifacts. Portanto
esses sinais só aparecem depois de adicionar/configurar o cache do experimento.
Em um VOD real, use manifests com TTL menor que segmentos e não cacheie resposta
privada sem considerar credenciais.

## Critério

A segunda leitura do mesmo segmento deve ser um HIT observável no cache
adicionado, e a API/origem deve receber menos bytes. Teste também invalidação,
`If-None-Match` e cache miss após expiração; não espere esses comportamentos do
NGINX padrão deste repositório.
