# 005 — Métricas de playback

## Contrato de medição

- `startup_time`: primeiro frame menos intenção de play.
- `rebuffer_ratio`: segundos em buffer dividido por segundos reproduzidos.
- `average_bitrate`: bytes de mídia consumidos dividido pelo tempo de mídia.
- `quality_switches`: contagem de mudanças de representação.
- `error_rate`: sessões com erro dividido por sessões iniciadas.

## Procedimento

Execute sessões com rede estável e variável. Crie uma sessão em
`POST /playback/sessions` e envie eventos com os campos `video_id`, `session_id`,
`type`, `position` e `payload` para o contrato em `docs/api-contract.md`. O
frontend atual envia, entre outros, `play`, `pause`, `seek`, `playing`,
`rebuffer_start`, `quality_change` e `audio_change`. O servidor aceita qualquer
`type`, não implementa `sequence` nem deduplicação e guarda os eventos somente
em memória.

## Análise

Agrupe p50/p95/p99 por vídeo, variante, navegador, região e rede. Não publique
média global sem tamanho de amostra. Correlacione `startup_time` com cache miss
quando houver um cache externo, mas não trate correlação como causalidade. Use
`/metrics` para confirmar a contagem agregada de eventos aceitos, não para obter
p50/p95 de QoE: esses percentis precisam ser calculados a partir dos eventos
coletados pelo experimento.
