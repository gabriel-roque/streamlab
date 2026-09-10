# 008 — Comparação de codecs

## Pergunta

Quanto bitrate cada codec precisa para a mesma qualidade e qual é o custo de
encode/decode?

## Procedimento

Codifique o mesmo trecho com H.264, HEVC, VP9 e AV1, mantendo resolução,
framerate, áudio, duração e alvo de qualidade. Use o encoder disponível e
registre versão/build; ausência de encoder é resultado de compatibilidade, não
motivo para inventar números.

## Medir

Tamanho, bitrate, fps de encode, tempo, CPU/GPU, VMAF quando houver referência,
tempo de startup e suporte do player. Reporte bitrate por qualidade e não uma
ordenação absoluta.

## Riscos

Licenças, hardware, presets e tuning mudam o resultado. Uma única cena de
Big Buck Bunny não representa todo catálogo; repita em animação, baixa luz e
alto movimento.
