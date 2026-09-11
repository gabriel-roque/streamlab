# ADR-004: Encoding assíncrono

- **Status:** aceito
- **Data:** 2026-09-10

## Contexto

Transcoding é CPU/GPU intensive, pode durar mais que o timeout HTTP e falha
independentemente do pedido de upload.

## Decisão

O upload grava o arquivo e enfileira um job `transcode` em uma fila em memória.
Um worker processa análise/empacotamento no mesmo processo, com estados
`QUEUED`, `RUNNING`, `RETRYING`, `SUCCEEDED` e `DLQ`, até três tentativas. Não há
broker, lease, evento `VideoUploaded` ou chave de idempotência na implementação
atual; esses itens ficam para a evolução do laboratório.

## Alternativas

- Encode síncrono na API: feedback imediato, mas bloqueia recursos e escala mal.
- Cron polling no banco: simples, porém aumenta contenção e latência.
- Orquestrador externo: possível depois, mas excessivo para a fase inicial.

## Consequências

O cliente consulta `/videos/{id}/status` e não presume que upload implica
`READY`. O retry atual usa atraso linear curto em memória e a DLQ não tem rota
HTTP de replay; reiniciar a API perde a fila. Backoff, lease e DLQ persistente
continuam sendo exercícios de arquitetura.
