# ADR-007: H.264 como codec inicial

- **Status:** aceito
- **Data:** 2026-09-10

## Contexto

O primeiro caminho precisa ser decodificável por browsers e players sem
hardware específico, preservando a capacidade de comparar codecs depois.

## Decisão

A ladder inicial usa H.264/AVC com pixel format `yuv420p` e áudio AAC. HEVC,
VP9 e AV1 ficam em experimentos separados, com compatibilidade explicitamente
medida.

## Alternativas

- AV1: melhor eficiência potencial, encode e suporte mais variáveis.
- HEVC: eficiente, mas com questões de licenciamento e suporte.
- VP9: opção aberta, mas não mantém o mesmo perfil de compatibilidade inicial.

## Consequências

O bitrate não é comparável entre codecs sem uma métrica de qualidade. O
experimento registra tempo, tamanho, VMAF quando disponível e taxa de falha do
player.
