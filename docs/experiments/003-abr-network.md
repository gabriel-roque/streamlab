# 003 — ABR sob variação de rede

## Pergunta

O player troca de variante antes de esgotar o buffer, sem oscilar demais?

## Procedimento

Sirva o master HLS e use DevTools, `tc netem` ou um proxy para perfis de 10,
5, 1.5 e 0.5 Mbps. Alterne a banda durante a reprodução e repita cinco vezes
por perfil.

```bash
scripts/smoke-http.sh --url http://localhost:3000 --path /healthz \
  --manifest /video/ID/master.m3u8 --kind hls
```

## Medir

Startup, rebuffer ratio, buffer_seconds, bitrate médio, switches por minuto,
tempo de reação a queda e recuperação após retorno. Registrar a sequência de
variantes e throughput estimado por segmento.

## Hipótese e critério

Uma política conservadora deve reduzir qualidade antes do rebuffer e subir de
forma gradual. Oscilação repetida ou rebuffer durante uma mudança é falha a
investigar, não apenas um resultado de throughput.
