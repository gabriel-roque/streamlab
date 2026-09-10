# 011 — Chaos testing

## Objetivo

Demonstrar comportamento previsível quando uma dependência falha, sem confundir
indisponibilidade deliberada com incidente não observado.

## Matriz

| Falha | Sinal esperado | Recuperação |
|---|---|---|
| matar transcoder | lease expira, job retorna | retry ou DLQ |
| parar object storage | uploads/packaging falham | backoff, sem publicar parcial |
| parar API | novos pedidos falham | player usa cache já publicado |
| derrubar broker | fila fica indisponível | alarme, sem perda após retorno |
| aumentar latência | p95 e rebuffer sobem | timeout e circuit breaker |
| limitar banda | ABR reduz variante | sem loop de switches |
| corromper vídeo | probe/encode falha | poison job na DLQ |

## Procedimento

Defina baseline, duração, blast radius e rollback. Execute uma falha por vez;
correlacione logs por `X-Request-ID`/job ID e capture queue depth, error rate,
retries, DLQ, startup, rebuffer e cache. Nunca injete caos em um ambiente com
dados de usuário.

## Critério

Toda falha tem detector, limite de impacto, comportamento documentado e
recuperação verificável. Um sistema que apenas retorna 500 rapidamente não é
considerado resiliente.
