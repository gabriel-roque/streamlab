# Amostras

Esta pasta contém somente instruções e metadados versionáveis. O Big Buck Bunny
é baixado sob demanda pela fonte oficial da Blender Foundation:

```text
https://download.blender.org/demo/movies/BBB/bbb_sunflower_1080p_30fps_normal.mp4.zip
```

O arquivo oficial é um ZIP; `scripts/download-bbb.sh` extrai o MP4 e chama
FFprobe. Nenhum vídeo, segmento ou manifesto gerado deve ser commitado:

```bash
scripts/download-bbb.sh
scripts/ffprobe-media.sh --input samples/raw/big-buck-bunny-1080p-normal.mp4
```

Para transformar a fonte em uma ladder H.264/AAC e empacotá-la nos dois
protocolos:

```bash
scripts/generate-ladder.sh --input samples/raw/big-buck-bunny-1080p-normal.mp4
scripts/generate-hls.sh --input samples/raw/big-buck-bunny-1080p-normal.mp4 \
  --segment-seconds 6 --output-dir samples/generated/hls --overwrite
scripts/generate-dash.sh --input samples/raw/big-buck-bunny-1080p-normal.mp4 \
  --segment-seconds 6 --output-dir samples/generated/dash --overwrite
scripts/validate-media.sh --input samples/generated/hls --kind hls
scripts/validate-media.sh --input samples/generated/dash --kind dash
```

O HLS gerado tem `master.m3u8`, playlists por resolução e segmentos MPEG-TS;
o DASH tem `manifest.mpd`, inicializações e chunks. `--segment-seconds` é um
alvo: confira `#EXTINF`, `Representation` e os arquivos com
`ffprobe-media.sh`. Os scripts aceitam FFmpeg local ou `MEDIA_TOOL=docker`.

Também existem dois fixtures públicos sem download automático quando a API
inicia: `big-buck-bunny` usa
`https://storage.googleapis.com/gtv-videos-bucket/sample/BigBuckBunny.mp4` como
MP4 remoto, e `abr-lab` usa
`https://test-streams.mux.dev/x36xhzz/x36xhzz.m3u8` como HLS multi-rendição.
Consulte-os com `GET /videos/{id}/playback`; eles não têm mídia ou manifest local
no `LocalStore`.

Com Compose, a API fica atrás do proxy em `http://localhost:3000/api`:

```bash
docker compose up --build
curl http://localhost:3000/api/videos/abr-lab/playback
curl -X POST http://localhost:3000/api/videos \
  -F 'title=BBB local' \
  -F 'file=@samples/raw/big-buck-bunny-1080p-normal.mp4;type=video/mp4'
```

O último comando retorna `202` e inicia o processamento assíncrono. Consulte
`/api/videos/{id}/status` até `READY`; depois use `/api/videos/{id}/playback`.
O container Compose instala FFmpeg/FFprobe para que esse fluxo produza HLS real e
um manifesto DASH didático. Se a API for executada localmente sem essas
ferramentas, ela usa o fixture didático; para comparar os dois caminhos e gerar
um pacote DASH real com variantes, rode os scripts de mídia explicitamente.

Os diretórios `raw/` e `generated/` são ignorados localmente. Para repetir um
experimento em outro ambiente, registre o URL, o hash SHA-256 opcional do
arquivo baixado, a versão do FFmpeg e os parâmetros do script no relatório.
