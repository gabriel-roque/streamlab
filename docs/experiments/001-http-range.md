# 001 — HTTP Range

## Pergunta

O servidor entrega seek e retomada sem transferir o MP4 inteiro?

## Procedimento

Faça um upload local primeiro e guarde o `id` retornado. Com Compose, `BASE` é
`http://localhost:3000/api`; com a API executada diretamente, use
`http://localhost:8080`:

```bash
curl -i -H 'Range: bytes=0-1023' "$BASE/videos/ID/stream" -o /tmp/range.bin
curl -i -H 'Range: bytes=1000000-1000999' "$BASE/videos/ID/stream" -o /tmp/range-2.bin
curl -I -H 'Range: bytes=0-0' "$BASE/videos/ID/stream"
```

Registrar `206 Partial Content`, `Accept-Ranges: bytes`, `Content-Range`,
`Content-Length` e o tamanho efetivamente recebido. Repetir com range inválido,
sem range e com `HEAD`. A API atual não envia `ETag` e não implementa
`If-Range`, então esses headers não fazem parte do experimento executável.

## Critério e métricas

Os ranges válidos devem retornar exatamente os bytes pedidos e o range inválido
deve retornar `416`. O `200` sem Range deve continuar sendo reprodutível. Medir
latência, bytes enviados e tempo de seek no player; não atribuir cache hit ao
endpoint, pois a API não define headers de cache para esse recurso.

## Cuidados

Não confundir `206` com streaming adaptativo: Range é transporte de um objeto;
HLS/DASH escolhe representações e segmentos.
