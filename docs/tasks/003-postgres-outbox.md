# Task: 003-postgres-outbox

- Projeto: feature-flag-mvp
- Tipo: feature
- Status: in-progress
- Branch: task/feature-flag-mvp-mvp
- Dependências: 002-control-plane-api

## Contexto

PostgreSQL é a fonte de verdade. Estado, histórico e evento devem ser gravados na mesma transação.
O Data Plane não consulta PostgreSQL durante Evaluate.

## Implementado

- pgx v5 e pgxpool.
- Store PostgreSQL com Apply, Delete, Get e List.
- Migrações para revision_counter, feature_flags, flag_history e outbox_events.
- Worker com claim FOR UPDATE SKIP LOCKED, tentativas e MarkPublished.
- Porta control.Repository e ControlAdapter.
- PersistentService no Control Plane.
- Handler HTTP aceita serviço em memória ou persistente.
- main usa PostgreSQL quando DATABASE_URL existe e fallback explícito em memória.

## Ainda necessário

- Executar migrações automaticamente ou documentar comando operacional.
- Implementar publisher NATS na task 004.
- Adicionar testes com PostgreSQL real para commit, rollback, concorrência e retry.
- Propagar e observar falhas do worker sem perder eventos.
- Revisar a semântica de criação concorrente no adapter.

## Aceite

- [ ] Testes de integração PostgreSQL passam.
- [ ] Rollback não deixa registros parciais.
- [ ] Retry não duplica efeito lógico.
- [ ] go test ./..., go test -race ./... e go vet ./... passam.
- [ ] Marcar done e mover para docs/history/.

## Evidências

- gofmt -w internal: OK
- go test ./...: OK
- go vet ./...: OK
- go build ./...: OK
- git diff --check: OK
