# 001 — HTTP Range

## Pergunta

O servidor entrega seek e retomada sem transferir o MP4 inteiro?

## Procedimento

Com um arquivo servido por um endpoint de mídia, executar:

```bash
curl -i -H 'Range: bytes=0-1023' http://localhost:3000/videos/ID/stream -o /tmp/range.bin
curl -i -H 'Range: bytes=1000000-1000999' http://localhost:3000/videos/ID/stream -o /tmp/range-2.bin
```

Registrar `206 Partial Content`, `Accept-Ranges: bytes`, `Content-Range`,
`Content-Length` e o tamanho efetivamente recebido. Repetir com range inválido,
sem range e com `If-Range`/`ETag`.

## Critério e métricas

Os ranges válidos devem retornar exatamente os bytes pedidos e o range inválido
deve retornar `416`. O `200` sem Range deve continuar sendo reprodutível. Medir
latência, bytes enviados, cache hit e tempo de seek no player.

## Cuidados

Não confundir `206` com streaming adaptativo: Range é transporte de um objeto;
HLS/DASH escolhe representações e segmentos.
