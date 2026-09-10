# ADR-005: PostgreSQL para metadados

- **Status:** aceito
- **Data:** 2026-09-10

## Contexto

Vídeos, variantes, jobs, tentativas e sessões têm relações e transições que
precisam de consulta e consistência.

## Decisão

Usar PostgreSQL como catálogo e store transacional de estado. Bytes não são
armazenados nele.

## Alternativas

- Redis como fonte principal: bom para estado efêmero, fraco como catálogo.
- Documento NoSQL: flexível, mas menos útil para invariantes relacionais da fase.
- Arquivos JSON: reproduzíveis, mas não concorrentes.

## Consequências

Transições devem ser atômicas e indexadas por `video_id`, `status` e `created_at`.
Métricas de fila não devem depender de scans frequentes do catálogo.
