# 006 — Retries e DLQ

## Pergunta

O pipeline recupera falhas transitórias sem reprocessar infinitamente um job
corrompido?

## Procedimento

1. Mate um worker durante um encode e confirme lease expirado e redelivery.
2. Faça o worker falhar duas vezes e suceder na terceira.
3. Envie um arquivo corrompido e observe o limite de tentativas.

Use backoff exponencial com jitter, `max_attempts=3`, e envie à DLQ com razão,
primeiro erro, última tentativa e correlation ID. A chave de saída deve ser
determinística por `video_id/profile/version`.

## Critério

Não deve haver dois artefatos públicos concorrentes. Um crash não pode perder o
job; um poison job não pode ocupar workers indefinidamente. O replay da DLQ
deve ser uma operação explícita e auditável.
