# ADR-001: HLS como protocolo inicial

- **Status:** aceito
- **Data:** 2026-09-10

## Contexto

O primeiro player precisa tocar VOD adaptativo em navegadores comuns e permitir
inspeção simples de manifests e segmentos.

## Decisão

Começar com HLS VOD, playlist master e variantes com segmentos MPEG-TS. DASH
será gerado em paralelo como experimento posterior, não como dependência do
primeiro fluxo.

## Alternativas

- Entregar MP4 progressivo: simples, mas sem ABR e com pior controle de cache.
- Começar com MPEG-DASH: bom modelo adaptativo, porém adiciona uma segunda
  cadeia de compatibilidade antes de validar o pipeline básico.

## Consequências

O contrato de playback expõe `application/vnd.apple.mpegurl` e exige keyframes
alinhados. HLS não elimina a necessidade de validar Range, cache e QoE.
