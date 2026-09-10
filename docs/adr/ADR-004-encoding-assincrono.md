# ADR-004: Encoding assíncrono

- **Status:** aceito
- **Data:** 2026-09-10

## Contexto

Transcoding é CPU/GPU intensive, pode durar mais que o timeout HTTP e falha
independentemente do pedido de upload.

## Decisão

Upload publica `VideoUploaded`; workers consomem jobs de análise, encode e
package. Cada job tem estado, tentativa, lease, erro e chave de idempotência.

## Alternativas

- Encode síncrono na API: feedback imediato, mas bloqueia recursos e escala mal.
- Cron polling no banco: simples, porém aumenta contenção e latência.
- Orquestrador externo: possível depois, mas excessivo para a fase inicial.

## Consequências

O cliente consulta status e não presume que upload implica `READY`. Retries
precisam ser limitados, com backoff e DLQ para poison jobs.
