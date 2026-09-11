# 002 — HLS e duração de segmento

## Pergunta

Qual tamanho equilibra startup, adaptação e overhead de requests?

## Procedimento

Gerar pacotes com 2, 4, 6, 10 e 12 segundos, mantendo fonte, codec, ladder e
GOP constantes. Use uma pasta de saída diferente para cada execução:

```bash
scripts/generate-hls.sh --input samples/raw/big-buck-bunny-1080p-normal.mp4 \
  --segment-seconds 6 --output-dir samples/generated/hls-6s --overwrite
scripts/validate-media.sh --input samples/generated/hls-6s --kind hls
```

Executar o mesmo fluxo para cada diretório de saída. Conte segmentos, tamanho
médio e duração indicada por `#EXTINF`.

## Medir

Startup time, requests por minuto, tempo até primeira troca de qualidade,
rebuffer ratio, tamanho do manifest e cache hit ratio. A aceitação é uma
conclusão baseada no conjunto, não um número universal.

## Cuidados

Sem keyframes alinhados, `hls_time` é apenas uma intenção e os cortes podem
escapar da duração nominal. O script de ladder fixa H.264/AAC, `yuv420p` e GOP
de 48 frames; confirme no manifest e em `ffprobe`, em vez de assumir que cada
segmento tem exatamente o valor solicitado.
