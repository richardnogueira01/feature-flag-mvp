# Task: 003-postgres-outbox

- Projeto: feature-flag-mvp
- Tipo: feature
- Status: in-progress
- Branch: task/feature-flag-mvp-mvp
- Dependências: 002-control-plane-api

## Contexto

PostgreSQL é a fonte de verdade. Cada alteração deve persistir estado, histórico auditável e
evento de distribuição na mesma transação. Publicação posterior pode falhar, mas o evento não pode
ser perdido; o worker usa Transactional Outbox.

## Objetivo

Trocar o armazenamento temporário por PostgreSQL sem colocar banco em Evaluate.

## Implementado nesta etapa

- Dependência pgx v5 e pgxpool.
- Adapter internal/persistence com Apply e Delete transacionais.
- Tabelas revision_counter, feature_flags, flag_history e outbox_events.
- Migrações migrations/000001_initial.up.sql e down.sql.
- Histórico e outbox são gravados junto da alteração.

## Ainda necessário

- Integrar o Control Plane HTTP ao adapter por uma interface de persistência.
- Implementar worker que busca eventos pendentes, publica idempotentemente e marca published_at.
- Adicionar testes de integração com PostgreSQL real e testar rollback/retry.

## Aceite

- [ ] Mutação e evento confirmados na mesma transação.
- [ ] Rollback não altera estado, histórico nem outbox.
- [ ] Retry não duplica efeito lógico.
- [ ] Testes de integração passam.
- [ ] go test ./..., go test -race ./... e go vet ./... passam.
- [ ] Só então marcar done e mover esta spec para docs/history/.

## Execução

- go get github.com/jackc/pgx/v5/pgxpool@v5.7.6: OK
- gofmt -w internal: OK
- go test ./...: OK
- go vet ./...: OK
- go build ./...: OK
- git diff --check: OK
