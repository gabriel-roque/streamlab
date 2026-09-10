# 007 — CPU versus GPU

## Pergunta

Qual é o trade-off entre throughput, custo, consumo e qualidade no hardware
disponível?

## Procedimento

Fixe fonte, resolução, duração, bitrate e saída. Compare `libx264` com NVENC
quando uma GPU NVIDIA e runtime compatível estiverem disponíveis. Não compare
`preset` CPU com `preset` GPU sem documentar o mapeamento.

```bash
time MEDIA_TOOL=local scripts/generate-ladder.sh --input samples/raw/big-buck-bunny-1080p-normal.mp4 --output-dir samples/generated/cpu
```

Repita usando uma imagem/runtime que exponha NVENC; o fallback Docker genérico
não garante acesso à GPU.

## Medir

Wall time, fps de encode, CPU%, GPU%, memória, energia/custo estimado, tamanho,
bitrate efetivo e VMAF. Registre warm-up e concorrência. Aceleração não é
melhoria se a qualidade ou a compatibilidade cair fora do limite.
