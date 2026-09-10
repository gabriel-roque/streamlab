# ADR-003: FFmpeg como transcoder

- **Status:** aceito
- **Data:** 2026-09-10

## Contexto

O laboratório precisa suportar probe, filtros, codecs, HLS e DASH em uma cadeia
reproduzível.

## Decisão

Usar FFmpeg e FFprobe como ferramentas de referência. Os scripts aceitam
execução local ou `MEDIA_TOOL=docker` com `jrottenberg/ffmpeg:6.1-ubuntu`.

## Alternativas

- GStreamer: excelente composição, mas aumenta o custo inicial de operação.
- Serviço gerenciado de transcoding: esconde decisões importantes do estudo.
- Bibliotecas próprias: não são necessárias para o objetivo do laboratório.

## Consequências

Versão, flags e hardware precisam ser registrados nos experimentos. Presets
rápidos não são comparáveis a qualidade constante; toda comparação deve fixar
codec, resolução, duração e critério de qualidade.
