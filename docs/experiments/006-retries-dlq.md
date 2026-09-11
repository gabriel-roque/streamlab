# 006 — Retries e DLQ

## Pergunta

O pipeline recupera falhas transitórias sem reprocessar infinitamente um job
corrompido?

## Procedimento

1. Observe o job criado pelo upload em `GET /videos/{id}/status`.
2. Em testes de código, faça o handler falhar e observe `RETRYING`, até três
   retries (quatro execuções totais contando a primeira), e depois `DLQ`.
3. Registre que a fila atual não expõe uma rota HTTP para injetar falha, replay
   ou consultar a DLQ; um arquivo arbitrário pode seguir pelo fixture e não é
   um teste confiável de erro de transcodificação.

O comportamento implementado usa `max_retries=3` e atraso curto linear em
memória, sem jitter, lease, correlation ID ou persistência. Para o exercício de
arquitetura, substitua-o por backoff exponencial com jitter e registre razão,
primeiro erro, última tentativa e correlation ID. A chave de saída desejada
continua determinística por `video_id/profile/version`.

## Critério

Na fila atual, valide que o job chega a `DLQ` após o limite e que o vídeo pode
permanecer `PROCESSING`; não há garantia de recuperação após crash, pois a fila
é em memória. Na arquitetura-alvo, não deve haver dois artefatos públicos
concorrentes, um crash não pode perder o job e o replay da DLQ deve ser explícito
e auditável.
