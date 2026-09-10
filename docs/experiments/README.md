# Experimentos

Todos os resultados devem registrar data, commit, hardware, versão do FFmpeg,
fonte, parâmetros, métricas brutas e conclusão. Nunca compare execuções com
ladder, duração ou preset diferentes.

| ID | Tema | Resultado principal |
|---|---|---|
| 001 | HTTP Range | comportamento 200/206 e seek |
| 002 | HLS | duração de segmento e startup |
| 003 | ABR | rebuffer e qualidade sob throttling |
| 004 | CDN | hit ratio, origem e latência |
| 005 | Métricas | definições e coleta de QoE |
| 006 | Retries/DLQ | recuperação e poison jobs |
| 007 | CPU/GPU | tempo, custo e qualidade |
| 008 | Codecs | eficiência e compatibilidade |
| 009 | VMAF | qualidade objetiva vs bitrate |
| 010 | Per-title | ladder baseada em conteúdo |
| 011 | Chaos | blast radius e recuperação |

Os comandos de mídia funcionam com FFmpeg instalado ou com `MEDIA_TOOL=docker`.
Veja `docs/architecture/README.md` para o fluxo completo.
