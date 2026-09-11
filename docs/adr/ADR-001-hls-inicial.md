# ADR-001: HLS como protocolo inicial

- **Status:** aceito
- **Data:** 2026-09-10

## Contexto

O primeiro player precisa tocar VOD adaptativo em navegadores comuns e permitir
inspeção simples de manifests e segmentos.

## Decisão

Começar com HLS VOD e segmentos MPEG-TS. O processador local também publica um
manifest DASH para comparação. A ladder master com variantes é gerada pelos
scripts de laboratório; o fixture usado quando FFmpeg não está disponível é
uma playlist HLS de mídia única, não uma master.

## Alternativas

- Entregar MP4 progressivo: simples, mas sem ABR e com pior controle de cache.
- Começar com MPEG-DASH: bom modelo adaptativo, porém adiciona uma segunda
  cadeia de compatibilidade antes de validar o pipeline básico.

## Consequências

O playback local expõe HLS em `/videos/{id}/playback/hls` e DASH em
`/videos/{id}/playback/dash`; o endpoint de manifest HLS usa
`application/vnd.apple.mpegurl`. HLS não elimina a necessidade de validar Range,
cache e QoE. Os fixtures públicos são remotos e não passam pelo empacotador
local.
