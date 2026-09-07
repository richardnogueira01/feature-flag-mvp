# Task: 005-observability-deployment

- Projeto: feature-flag-mvp
- Tipo: infra
- Status: in-progress
- Branch: task/feature-flag-mvp-mvp
- Dependências: 004-nats-sync-readiness

## Contexto

O MVP roda com PostgreSQL, NATS/JetStream e múltiplas instâncias. Liveness indica processo vivo;
readiness indica snapshot sincronizado. Nenhum secret pode ser versionado.

## Implementado

- Dockerfile multi-stage com imagem distroless nonroot.
- Compose com PostgreSQL 17, NATS 2.11 JetStream e app.
- Configuração por DATABASE_URL, AUTO_MIGRATE, NATS_URL, NATS_SUBJECT e NATS_DURABLE.
- Migration inicial embutida e execução opcional via AUTO_MIGRATE=true.
- Endpoints /healthz, /readyz, /internal/status e /metrics.
- Métricas de avaliações, latência, HTTP, gaps, resync e outbox.
- Worker da outbox iniciado quando PostgreSQL e NATS estão configurados.
- Labels limitados; key de flag não é label.
- POSTGRES_PASSWORD exigida externamente.

## Ainda necessário

- Adicionar smoke test com Compose real.
- Validar sincronização inicial antes do readiness.
- Garantir configuração do stream JetStream antes do subscriber.

## Evidências

- docker compose config --quiet com senha efêmera: OK.
- go test ./...: OK.
- go vet ./...: OK.
- go build ./...: OK.
- git diff --check: OK.

## Aceite

- [ ] Worker publica eventos no Compose.
- [ ] Smoke test com Compose passa.
- [ ] Sincronização inicial validada.
- [ ] Marcar done e mover para docs/history/.
