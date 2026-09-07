# Task: 003-postgres-outbox

- Projeto: feature-flag-mvp
- Tipo: feature
- Status: planned
- Branch: task/feature-flag-mvp-postgres-outbox
- Criada: 2026-09-07

## Descrição

Persistir flags, histórico e eventos em PostgreSQL usando Transactional Outbox.

## Critérios de aceite

- [ ] Migrações para feature_flags, flag_history e outbox_events.
- [ ] Mutação e evento confirmados na mesma transação.
- [ ] Histórico é auditável por revisão.
- [ ] Falha de publicação permite retry idempotente.

## Dependências

002-control-plane-api.
