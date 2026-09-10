# 004 — CDN e cache

## Pergunta

Quanto tráfego deixa de chegar à origem quando manifests e segmentos são
cacheáveis?

## Procedimento

Coloque NGINX/CDN entre player e object storage. Faça duas sessões para o mesmo
manifest e os mesmos segmentos e depois uma sessão com outro vídeo.

```bash
scripts/smoke-http.sh --url http://localhost:8080 --path /video/ID/master.m3u8
```

## Medir

`Age`, `ETag`, `Cache-Control`, hit/miss, latência p50/p95, bytes no origin,
requests ao origin e taxa de revalidação. Use manifests com TTL menor que
segmentos em VOD; não cacheie resposta privada sem considerar o token.

## Critério

A segunda leitura do mesmo segmento deve ser um HIT observável e o origin
deve receber menos bytes. Teste também invalidação, `If-None-Match` e cache
miss após expiração.
