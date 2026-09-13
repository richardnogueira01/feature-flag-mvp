# Task: 003-postgres-outbox

- Status: done
- Concluída em: 2026-09-07
- Branch: task/feature-flag-mvp-mvp

## Resultado

Implementados PostgreSQL como fonte de verdade, histórico, transactional outbox,
claim concorrente com `FOR UPDATE SKIP LOCKED`, retry e adapter do Control Plane.
Eventos publicados pelo outbox carregam snapshots completos, permitindo consumo
idempotente e ressincronização.

## Evidências

- `go test ./...`: OK
- `go vet ./...`: OK
- `go build ./...`: OK
- Smoke Compose: PostgreSQL migrou e o worker processou um evento com resultado `success`.
- Rollback é protegido pelo uso transacional em `Apply` e `Delete`.

## Limitação conhecida

Não foi executado `go test -race ./...` nem um teste automatizado externo de PostgreSQL;
o fluxo real foi validado pelo Compose.
