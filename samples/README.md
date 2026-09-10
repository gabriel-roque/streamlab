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

Os diretórios `raw/` e `generated/` são ignorados localmente. Para repetir um
experimento em outro ambiente, registre o URL, o hash SHA-256 opcional do
arquivo baixado, a versão do FFmpeg e os parâmetros do script no relatório.
