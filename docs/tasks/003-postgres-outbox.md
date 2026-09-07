# Task: 003-postgres-outbox

- Projeto: feature-flag-mvp
- Tipo: feature
- Status: planned
- Branch: task/feature-flag-mvp-postgres-outbox
- Dependências: 002-control-plane-api

## Contexto

PostgreSQL é a fonte de verdade. Cada alteração deve persistir estado, histórico e evento de
distribuição na mesma transação. Publicação posterior pode falhar, mas o evento não pode ser perdido.

## Objetivo

Trocar o armazenamento temporário por PostgreSQL sem alterar HTTP nem colocar banco em Evaluate.

## Implementação

- Criar migrações para feature_flags, flag_history e outbox_events.
- Usar pgx e ports/adapters.
- feature_flags: key única, enabled, revision, created_at e updated_at.
- flag_history: revision, key, operation, payload e timestamp.
- outbox_events: id, revision, event_type, payload, published_at e attempts.
- Na transação, atualizar revisão, salvar histórico e inserir outbox.
- Worker busca pendentes, publica idempotentemente e marca sucesso; falha mantém retry.

## Testes e aceite

- [ ] Testes com PostgreSQL real.
- [ ] Rollback não altera estado, histórico nem outbox.
- [ ] Retry não duplica efeito lógico.
- [ ] Passar go test ./..., go test -race ./... e go vet ./....
- [ ] Registrar migrações e comandos, marcar done e mover para docs/history/.
