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
- .dockerignore sem docs, Git ou artefatos.
- Compose com PostgreSQL 17, NATS 2.11 JetStream e app.
- PostgreSQL com healthcheck e depends_on condicionado à saúde.
- Configuração por DATABASE_URL, NATS_URL, NATS_SUBJECT e NATS_DURABLE.
- docs/LOCAL-DEVELOPMENT.md.
- Métricas Prometheus de avaliações, latência e requests HTTP.
- Endpoint /metrics e middleware HTTP.
- Labels limitados a result, method e status; key de flag não é label.
- POSTGRES_PASSWORD exigida externamente, sem senha versionada.

## Ainda necessário

- Aplicar migrações automaticamente ou documentar comando operacional completo.
- Adicionar métricas de revisão, gaps, resync e outbox.
- Adicionar smoke test com dependências reais.
- Configurar startup/readiness para sincronização inicial real.

## Evidências

- docker compose config --quiet com senha efêmera de processo: OK.
- go test ./...: OK.
- go vet ./...: OK.
- go build ./...: OK.
- git diff --check: OK.

## Aceite

- [ ] Métricas e smoke test completos.
- [ ] Migrações aplicáveis no Compose.
- [ ] Testes reais passam.
- [ ] Marcar done e mover para docs/history/.
