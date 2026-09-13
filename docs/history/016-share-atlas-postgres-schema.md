# Task: 016-share-atlas-postgres-schema

- Status: done
- Implementada em: 2026-09-07 (branch task/feature-flag-mvp-mvp)
- Fechada/confirmada em: 2026-09-13 (junto com a task 013)

## Resultado

`docker-compose.atlas.yml` conecta o serviço à rede externa `atlas-go_default`, desativa o
Postgres local (`profiles: ["local"]`) e exige `ATLAS_FEATURE_FLAGS_DATABASE_URL` fora do Git. A
migration `000000_feature_flags_schema.up.sql` cria o schema `feature_flags` e move qualquer
tabela legada de `public` para lá, sem DROP e sem tocar em tabelas do Atlas. O modo padrão
(`docker-compose.yml`, Postgres próprio) continua funcionando sem esse override — o schema
`feature_flags` é usado nos dois casos, então o produto não fica acoplado ao Atlas.

## Confirmação feita na task 013

Ao implementar a task 013 foi encontrado e corrigido um bug que também afeta esta task: o
`search_path` da migration só valia para a conexão que a executou, não para o pool inteiro (ver
`docs/history/013-configuration-scope-model.md`). Sem essa correção
(`persistence.NewPool`/`AfterConnect`), o override `docker-compose.atlas.yml` funcionaria de
forma instável mesmo com o schema isolado corretamente criado. Com a correção, testes de
integração contra um PostgreSQL real (fora do container) confirmam que qualquer conexão do pool
resolve as tabelas em `feature_flags` corretamente.

## Pendências para o dono

- Não validado neste ambiente (sem Docker nem acesso à rede `atlas-go_default` real): subir
  `docker compose -f docker-compose.yml -f docker-compose.atlas.yml up -d --build app agent` no
  homelab apontando para o Postgres real do Atlas, com a URL definida fora do Git.
