# 003 — ABR sob variação de rede

## Pergunta

O player troca de variante antes de esgotar o buffer, sem oscilar demais?

## Procedimento

Gere um pacote HLS master com `scripts/generate-hls.sh` e sirva o diretório por
um servidor HTTP estático, ou use o fixture remoto `abr-lab`. No caminho da API
local, use DevTools, `tc netem` ou um proxy para perfis de 10, 5, 1.5 e 0.5
Mbps. Alterne a banda durante a reprodução e repita cinco vezes por perfil.

```bash
scripts/smoke-http.sh --url http://localhost:3000 \
  --path /api/healthz
```

Para verificar apenas o manifest de um upload local `ID` que já esteja `READY`,
use `/api/videos/ID/playback/hls`:

```bash
scripts/smoke-http.sh --url http://localhost:3000 \
  --manifest /api/videos/ID/playback/hls --kind hls
```

A API publica uma playlist local, não uma rota `/video/ID/master.m3u8`; no
processador atual ela não é uma master ABR. A master com variantes pertence ao
pacote gerado pelos scripts. O fixture público
`abr-lab` deve ser testado pela URL externa retornada em
`/api/videos/abr-lab/playback`, não pelo endpoint de manifest local. O player
web usa `hls.js` quando o navegador não toca HLS nativamente.

## Medir

Startup, rebuffer ratio, buffer_seconds, bitrate médio, switches por minuto,
tempo de reação a queda e recuperação após retorno. Registrar a sequência de
variantes e throughput estimado por segmento.

## Hipótese e critério

Uma política conservadora deve reduzir qualidade antes do rebuffer e subir de
forma gradual. Oscilação repetida ou rebuffer durante uma mudança é falha a
investigar, não apenas um resultado de throughput.
