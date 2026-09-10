# 009 — VMAF

## Pergunta

Qual bitrate entrega qualidade perceptual aceitável em cada resolução?

## Procedimento

Use o original como referência alinhada e cada encode como distorted. Garanta
mesmo framerate, resolução, crop e duração; se necessário, normalize antes de
calcular. Com `libvmaf` disponível, execute uma medição por variante e salve o
JSON/CSV e a versão do modelo.

Exemplo conceitual:

```bash
ffmpeg -i encoded.mp4 -i original.mp4 -lavfi \
  "[0:v]setpts=PTS-STARTPTS[dist];[1:v]setpts=PTS-STARTPTS[ref];[dist][ref]libvmaf=log_fmt=json:log_path=vmaf.json" \
  -f null -
```

## Medir e interpretar

Compare VMAF médio e p5/p1 com bitrate, tamanho e tempo de encode. Procure
degraus da ladder, não um único número mágico. VMAF não mede startup, rebuffer,
artefatos de áudio ou compatibilidade do dispositivo.
