# ADR-002: Object storage para bytes de mídia

- **Status:** aceito
- **Data:** 2026-09-10

## Contexto

Originais e segmentos são grandes, imutáveis e têm padrão de leitura diferente
dos metadados transacionais.

## Decisão

Usar MinIO localmente, com uma interface compatível com S3, para originais,
variantes e manifests. A API entrega upload direto por URL pré-assinada quando
o fluxo estiver pronto.

## Alternativas

- Filesystem local: útil no primeiro teste, mas não representa object storage.
- Guardar bytes no PostgreSQL: acopla escala de banco ao throughput de mídia.
- S3 desde o início: realista, mas menos reproduzível offline.

## Consequências

É necessário tratar consistência/publicação de prefixos, retenção e URLs
assinadas. A origem deve continuar acessível para validar um cache miss.
