# Task: 005-observability-deployment

- Status: done
- Concluída em: 2026-09-07
- Branch: task/feature-flag-mvp-mvp

## Resultado

Entregues Dockerfile multi-stage distroless nonroot, Compose com PostgreSQL/NATS/app,
migração automática opt-in, health/readiness, métricas HTTP, avaliação, gaps, resync
e outbox. `POSTGRES_PASSWORD` permanece obrigatório e externo.

## Evidências

- `docker compose up -d --build`: OK.
- Smoke HTTP de health, readiness, CRUD e avaliação: OK.
- `feature_flag_outbox_events_total{result="success"}` observado: `1`.
- `git diff --check`, `go test ./...`, `go vet ./...` e `go build ./...`: OK.

## Limitação conhecida

`go test -race ./...` não foi executado neste fechamento.
