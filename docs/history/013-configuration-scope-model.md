# Task: 013-configuration-scope-model

- Status: done
- Concluída em: 2026-09-13
- Branch: task/013-configuration-scope-model

## Resultado

Hierarquia `application -> environment -> configuration -> flag` adicionada sem remover a API
plana existente. Toda flag pré-existente foi migrada para uma configuração implícita
`legacy/default`, criada e preenchida pela própria migration.

Endpoints novos, todos sob `/v1/applications`:

```
POST/GET /v1/applications
POST/GET /v1/applications/{application}/environments
POST/GET /v1/applications/{application}/environments/{environment}/configurations
GET      /v1/applications/{application}/environments/{environment}/configurations/{configuration}
POST/GET /v1/applications/{application}/environments/{environment}/configurations/{configuration}/flags
GET/PUT/PATCH/DELETE .../flags/{key}
```

O CRUD de flags por configuração (`.../flags[/{key}]`) não estava na lista literal da spec, mas
é necessário para os critérios de aceite ("atualização grava outbox e snapshot do escopo numa
transação") e para as tasks 015/020, que dependem de poder criar/editar/remover flags dentro de
uma configuração pelo HTTP.

## Bug crítico encontrado e corrigido (fora do escopo original da task, mas bloqueante)

`cmd/feature-flag-mvp/main.go` abria o pool com `pgxpool.New`, sem configurar `search_path` por
conexão. A migration `000000` roda `SET search_path TO feature_flags, public`, mas isso só valia
para a *conexão específica* que executou a migration — qualquer outra conexão do pool (o caso
normal sob concorrência real) falhava com `relation "feature_flags" does not exist`. Reproduzido
e confirmado com 8 conexões simultâneas contra um PostgreSQL real. Corrigido com
`persistence.NewPool`, que usa `pgxpool.Config.AfterConnect` para aplicar o `search_path` a toda
conexão nova do pool. Isso afetava a camada de persistência inteira (não só as tabelas novas),
tanto no modo standalone quanto no override `docker-compose.atlas.yml` da task 016.

## Evidências

- `gofmt -l .`, `go vet ./...`, `go build ./...`: limpos.
- `go test ./...` e `go test -race ./...`: OK (inclui testes novos de `internal/control` e
  `internal/httpapi` com um `ScopeRepository` fake).
- Testes de integração contra PostgreSQL real (guardados por `TEST_DATABASE_URL`, pulados por
  padrão): `TestMigrateBackfillsLegacyConfiguration` (seed pré-013 → migration → backfill →
  idempotência em 3 execuções) e `TestScopeStoreAgainstRealPostgres` (hierarquia completa, mesma
  key em duas configurations, toggle, revision, delete) — ambos passaram contra um PostgreSQL 16
  local.
- Não validado: `docker compose config` (Docker indisponível neste ambiente de execução) — a
  sintaxe dos arquivos de compose não foi alterada por esta task, só o código Go.

## Pendências para o dono

- Rodar `docker compose up --build` num ambiente com Docker para confirmar o fluxo completo em
  container (a lógica já foi validada com PostgreSQL real, só falta o empacotamento).
