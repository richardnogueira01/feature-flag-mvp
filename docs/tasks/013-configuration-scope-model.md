# Task 013 — Modelo Application/Environment/Configuration

status: ativo
owner: architect + developer

## Objetivo

Substituir o escopo global de flags por uma hierarquia explícita e compatível com o modelo do AWS AppConfig:

`application -> environment -> configuration -> flag`.

## Contrato

Cada flag deve ser única por `(application, environment, configuration, key)`. Todos os identificadores aceitam 1–128 caracteres, sem espaços nas extremidades.

Endpoints novos:

```http
POST /v1/applications
GET  /v1/applications
POST /v1/applications/{application}/environments
GET  /v1/applications/{application}/environments
POST /v1/applications/{application}/environments/{environment}/configurations
GET  /v1/applications/{application}/environments/{environment}/configurations
GET  /v1/applications/{application}/environments/{environment}/configurations/{configuration}
```

O último endpoint retorna `{application, environment, configuration, revision, flags}` e suporta `ETag`/`304`.

## Persistência

Adicionar migration sem apagar dados legados:

- `applications(id, name, created_at, updated_at)`;
- `environments(id, application_id, name, created_at, updated_at)`;
- `configurations(id, environment_id, name, revision, created_at, updated_at)`;
- adicionar `configuration_id` em `feature_flags`;
- preservar a API plana existente como compatibilidade usando uma configuração `legacy/default`.

## Aceite

- mesma key pode existir em aplicações/ambientes/configurações diferentes;
- revision é monotônica por configuração;
- bulk response nunca mistura escopos;
- atualização grava outbox e snapshot do escopo numa transação;
- `go test ./...`, `go vet ./...` e migration em Postgres real passam.