# Task: 003-postgres-outbox

- Projeto: feature-flag-mvp
- Tipo: feature
- Status: in-progress
- Branch: task/feature-flag-mvp-mvp
- Dependências: 002-control-plane-api

## Contexto

PostgreSQL é a fonte de verdade. Estado, histórico e evento de distribuição devem ser gravados na
mesma transação. O Data Plane não pode consultar PostgreSQL durante Evaluate.

## Implementado

- pgx v5 e pgxpool em go.mod/go.sum.
- internal/persistence com Apply e Delete transacionais.
- migrations/000001_initial.up.sql e down.sql.
- Tabelas revision_counter, feature_flags, flag_history e outbox_events.
- internal/persistence/outbox.go com ClaimPending usando FOR UPDATE SKIP LOCKED.
- Incremento de attempts durante claim.
- Publisher injetável, Worker.RunOnce e Worker.Run.
- MarkPublished só marca eventos ainda não publicados.

## Ainda necessário

- Adaptar control.Service para uma interface de persistência e usar o PostgreSQL em produção.
- Implementar o publisher NATS na task 004.
- Criar testes de integração com PostgreSQL real para commit, rollback, concorrência e retry.
- Validar que erros do worker sejam observáveis sem perder eventos.

## Aceite

- [ ] Control Plane usa repository PostgreSQL.
- [ ] Mutação, histórico e outbox confirmam na mesma transação.
- [ ] Rollback não deixa registros parciais.
- [ ] Retry não duplica efeito lógico.
- [ ] Testes de integração passam.
- [ ] go test ./..., go test -race ./... e go vet ./... passam.
- [ ] Marcar done e mover para docs/history/.

## Evidências atuais

- go get github.com/jackc/pgx/v5/pgxpool@v5.7.6: OK
- gofmt -w internal: OK
- go test ./...: OK
- go vet ./...: OK
- go build ./...: OK
- git diff --check: OK
