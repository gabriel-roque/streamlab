# 005 — Métricas de playback

## Contrato de medição

- `startup_time`: primeiro frame menos intenção de play.
- `rebuffer_ratio`: segundos em buffer dividido por segundos reproduzidos.
- `average_bitrate`: bytes de mídia consumidos dividido pelo tempo de mídia.
- `quality_switches`: contagem de mudanças de representação.
- `error_rate`: sessões com erro dividido por sessões iniciadas.

## Procedimento

Execute sessões com rede estável e variável. Envie `play`, `buffer_start`,
`buffer_end`, `quality_change`, `playback_error` e `video_complete` para o
contrato em `docs/api-contract.md`. Use `sequence` para tornar o envio
idempotente.

## Análise

Agrupe p50/p95/p99 por vídeo, variante, navegador, região e rede. Não publique
média global sem tamanho de amostra. Correlacione `startup_time` com cache miss,
mas não trate correlação como causalidade.
