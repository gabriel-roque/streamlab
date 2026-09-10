# 002 — HLS e duração de segmento

## Pergunta

Qual tamanho equilibra startup, adaptação e overhead de requests?

## Procedimento

Gerar pacotes com 2, 4, 6, 10 e 12 segundos, mantendo fonte, codec, ladder e
GOP constantes:

```bash
scripts/generate-hls.sh --input samples/raw/big-buck-bunny-1080p-normal.mp4 --segment-seconds 6 --overwrite
scripts/validate-media.sh --input samples/generated/hls --kind hls
```

Executar o mesmo fluxo para cada diretório de saída. Conte segmentos, tamanho
médio e duração indicada por `#EXTINF`.

## Medir

Startup time, requests por minuto, tempo até primeira troca de qualidade,
rebuffer ratio, tamanho do manifest e cache hit ratio. A aceitação é uma
conclusão baseada no conjunto, não um número universal.

## Cuidados

Sem keyframes alinhados, `hls_time` é apenas uma intenção e os cortes podem
escapar da duração nominal.
