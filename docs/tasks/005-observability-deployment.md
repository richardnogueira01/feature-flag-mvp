# Task: 005-observability-deployment

- Projeto: feature-flag-mvp
- Tipo: infra
- Status: in-progress
- Branch: task/feature-flag-mvp-mvp
- Dependências: 004-nats-sync-readiness

## Contexto

O MVP roda localmente com PostgreSQL, NATS/JetStream e múltiplas instâncias. Liveness indica
processo vivo; readiness indica snapshot sincronizado. Nenhum secret pode ser versionado.

## Implementado

- Dockerfile multi-stage com imagem final distroless nonroot.
- .dockerignore sem docs, Git ou artefatos.
- docker-compose.yml com PostgreSQL 17, NATS 2.11 JetStream e app.
- Healthcheck do PostgreSQL e depends_on condicionado à saúde.
- DATABASE_URL, NATS_URL, NATS_SUBJECT e NATS_DURABLE documentados.
- docs/LOCAL-DEVELOPMENT.md com execução e validação.
- Compose exige POSTGRES_PASSWORD externo, sem senha versionada.

## Ainda necessário

- Aplicar migrações automaticamente ou documentar comando operacional completo.
- Adicionar métricas Prometheus sem labels de alta cardinalidade.
- Adicionar smoke test com dependências reais.
- Configurar startup/readiness para sincronização inicial real.

## Evidências

- docker compose config --quiet com senha efêmera de processo: OK.
- go test ./...: OK.
- go vet ./...: OK.
- go build ./...: OK.
- git diff --check: OK.

## Aceite

- [ ] Métricas e smoke test implementados.
- [ ] Migrações aplicáveis no ambiente Compose.
- [ ] Testes reais passam.
- [ ] Marcar done e mover para docs/history/.
