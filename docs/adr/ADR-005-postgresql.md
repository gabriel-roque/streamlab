# ADR-005: PostgreSQL para metadados

- **Status:** aceito
- **Data:** 2026-09-10

## Contexto

Vídeos, variantes, jobs, tentativas e sessões têm relações e transições que
precisam de consulta e consistência.

## Decisão

O Compose provisiona PostgreSQL para o estágio de catálogo, mas a implementação
atual mantém vídeos e jobs no `LocalStore`/`MemoryQueue` em memória. A API não
abre conexão PostgreSQL e não persiste bytes no banco.

## Alternativas

- Redis como fonte principal: bom para estado efêmero, fraco como catálogo.
- Documento NoSQL: flexível, mas menos útil para invariantes relacionais da fase.
- Arquivos JSON: reproduzíveis, mas não concorrentes.

## Consequências

Ao migrar para PostgreSQL, transições deverão ser atômicas e indexadas por
`video_id`, `status` e `created_at`. Hoje essas garantias não existem após um
restart e métricas são contadores da instância em memória.
