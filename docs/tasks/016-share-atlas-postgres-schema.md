# Task 016 — Compartilhar PostgreSQL do Atlas com schema isolado

status: ativo

## Decisão

Reutilizar o PostgreSQL/database do Atlas sem reutilizar tabelas do Atlas. Todas as tabelas do serviço ficam no schema `feature_flags`.

## Homelab

O override `docker-compose.atlas.yml` conecta o serviço à rede externa `atlas-go_default`, desativa o Postgres local e exige uma URL fornecida fora do Git:

```powershell
$env:ATLAS_FEATURE_FLAGS_DATABASE_URL = 'postgres://atlas:<senha>@atlas-postgres:5432/atlas_prod?sslmode=disable&options=-csearch_path%3Dfeature_flags%2Cpublic'
docker compose -f docker-compose.yml -f docker-compose.atlas.yml up -d --build app agent
```

## Migração

A migration cria `feature_flags` e move tabelas legadas do schema `public` somente quando não existe tabela equivalente no novo schema. Não executa DROP, `down -v` ou alteração de tabelas do Atlas.

## Aceite

- serviço conecta ao `atlas_prod` pela rede `atlas-go_default`;
- tabelas ficam em `feature_flags.*`;
- Atlas continua iniciando sem alteração no próprio schema;
- testes Go e migration de teste passam;
- URL e credenciais não entram no Git ou logs.