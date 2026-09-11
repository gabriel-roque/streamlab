# ADR-002: Object storage para bytes de mídia

- **Status:** aceito
- **Data:** 2026-09-10

## Contexto

Originais e segmentos são grandes, imutáveis e têm padrão de leitura diferente
dos metadados transacionais.

## Decisão

Manter filesystem local como adapter atual: `media/` guarda originais e
`artifacts/{video_id}/` guarda manifests/segmentos. O Compose também sobe MinIO
com interface compatível com S3 para o próximo estágio, mas a API atual não se
conecta a ele nem entrega upload por URL pré-assinada.

## Alternativas

- MinIO/S3 desde o início: representa object storage, mas ainda não é usado pelo
  código executável.
- Guardar bytes no PostgreSQL: acopla escala de banco ao throughput de mídia.
- S3 desde o início: realista, mas menos reproduzível offline.

## Consequências

O adapter local não oferece retenção, versionamento, URLs assinadas ou
publicação atômica de prefixos. Essas são preocupações do estágio MinIO/S3; por
enquanto, reiniciar a API perde metadados, mas os bytes do diretório persistem.
