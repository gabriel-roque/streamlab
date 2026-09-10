# 010 — Per-title encoding

## Pergunta

Uma ladder baseada na complexidade do título economiza bytes sem degradar QoE?

## Procedimento

Separe títulos por baixa, média e alta complexidade usando movimento, cortes,
textura e resolução. Faça uma análise curta, gere curvas bitrate x VMAF e
selecione pontos com qualidade mínima e distância útil entre representações.

Compare com a ladder fixa do script:

```bash
scripts/generate-ladder.sh --input samples/raw/big-buck-bunny-1080p-normal.mp4
```

## Critério

Para a mesma meta de qualidade, a ladder per-title deve reduzir bitrate/bytes ou
melhorar qualidade sem aumentar rebuffer. Verifique que ainda há uma variante
compatível com redes lentas e que os segmentos continuam alinhados.

## Cuidados

O custo do encode adicional, cache fragmentation, número de variantes e tempo
de decisão fazem parte do resultado econômico.
