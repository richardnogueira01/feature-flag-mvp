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
- docs/LOCAL-DEVELOPMENT.md.
- Endpoint /metrics e middleware HTTP.
- Métricas de avaliações, latência e requests HTTP.
- Coletores operacionais de gaps, resync e outbox, com labels limitados.
- POSTGRES_PASSWORD exigida externamente, sem senha versionada.

## Ainda necessário

- Emitir as métricas operacionais dentro de Syncer e Worker.
- Adicionar smoke test com dependências reais.
- Validar sincronização inicial antes do readiness.

## Evidências

- docker compose config --quiet com senha efêmera: OK.
- go test ./...: OK.
- go vet ./...: OK.
- go build ./...: OK.
- git diff --check: OK.

## Aceite

- [ ] Métricas de gaps/resync/outbox recebem eventos reais.
- [ ] Smoke test com Compose passa.
- [ ] Sincronização inicial validada.
- [ ] Marcar done e mover para docs/history/.
