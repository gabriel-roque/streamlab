# 011 — Chaos testing

## Objetivo

Demonstrar comportamento previsível quando uma dependência falha, sem confundir
indisponibilidade deliberada com incidente não observado.

## Matriz

| Falha | Sinal esperado | Recuperação |
|---|---|---|
| matar worker/API | fila em memória pode perder trabalho | restart e novo upload |
| parar storage local | uploads/packaging falham | erro HTTP, sem garantia de retry |
| parar API | novos pedidos falham | player usa cache já publicado |
| parar Redis/PostgreSQL/MinIO | sem efeito no código atual | validar que são dependências futuras |
| aumentar latência | p95 e rebuffer sobem | timeout e circuit breaker |
| limitar banda | ABR reduz variante | sem loop de switches |
| corromper vídeo | FFmpeg pode falhar; fixture pode mascarar | observar job, não assumir DLQ |

## Procedimento

Defina baseline, duração, blast radius e rollback. Execute uma falha por vez;
correlacione por job ID (a API atual não gera `X-Request-ID`) e capture status,
error rate, retries, DLQ, startup, rebuffer e cache. Em Compose, os nomes dos
serviços são `api`, `nginx`, `postgres`, `redis`, `minio`, `prometheus` e
`grafana`; lembre que os quatro últimos não são dependências lidas pela API
atual. Nunca injete caos em um ambiente com dados de usuário.

## Critério

Toda falha tem detector, limite de impacto, comportamento documentado e
recuperação verificável. Um sistema que apenas retorna 500 rapidamente não é
considerado resiliente.
