# Task: 002-control-plane-api

- Projeto: feature-flag-mvp
- Tipo: feature
- Status: planned
- Branch: task/feature-flag-mvp-control-plane-api
- Dependências: 001-data-plane-snapshot

## Contexto

A API administra flags sem participar do hot path. Use net/http ou router mínimo. A API é
versionada em /v1 e o domínio depende de interfaces, não de banco, mensageria ou tipos HTTP.

## Objetivo

Implementar CRUD de flags em memória, revisões monotônicas e publicação no store do Data Plane.
Persistência será adicionada na task 003.

## Contrato

- POST /v1/flags com {"key":"checkout","enabled":true} retorna 201, flag e revision.
- GET /v1/flags retorna 200 e lista.
- GET /v1/flags/{key} retorna 200 ou 404.
- PUT /v1/flags/{key} com {"enabled":false} retorna 200 e nova revision.
- DELETE /v1/flags/{key} retorna 204 ou 404.
- JSON inválido, key vazia ou key maior que 128 retorna 400.
- Método não suportado retorna 405; erro usa {"error":"mensagem"}.

## Implementação e aceite

- [ ] Separar domínio, serviço e handlers.
- [ ] Não expor mapas internos nas respostas.
- [ ] Garantir revisão única e crescente sob mutações concorrentes.
- [ ] Publicar snapshot completo após mutação.
- [ ] Testar endpoints, validações, erros e concorrência.
- [ ] Passar gofmt, go test ./..., go test -race ./... e go vet ./....
- [ ] Marcar done e mover a spec para docs/history/.
