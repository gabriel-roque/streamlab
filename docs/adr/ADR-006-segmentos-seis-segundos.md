# ADR-006: Segmentos de seis segundos

- **Status:** aceito
- **Data:** 2026-09-10

## Contexto

Segmentos curtos reduzem o tempo de reação do ABR, mas aumentam requests,
overhead de manifest e pressão no CDN.

## Decisão

Usar seis segundos como baseline VOD: é o default de `generate-hls.sh`,
`generate-dash.sh` e do comando FFmpeg do processador local. Variar para 2, 4,
10 e 12 segundos no experimento de HLS. O fixture HLS sem FFmpeg tem exatamente
um segmento de seis segundos; nos pacotes reais, `hls_time` é um alvo e a
duração efetiva deve ser lida do manifest.

## Alternativas

- 2 segundos: troca rápida, custo alto de requests.
- 10-12 segundos: menor overhead, adaptação mais lenta.
- Segmentação variável: pode melhorar conteúdo, mas dificulta a comparação inicial.

## Consequências

O número de segmentos, startup, rebuffer e cache hit ratio devem ser medidos
juntos; nenhuma métrica isolada decide o tamanho ótimo.
