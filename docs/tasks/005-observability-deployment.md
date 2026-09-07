# Task: 005-observability-deployment

- Projeto: feature-flag-mvp
- Tipo: infra
- Status: planned
- Branch: task/feature-flag-mvp-observability-deployment
- Dependências: 004-nats-sync-readiness

## Contexto

O MVP roda localmente com PostgreSQL, NATS/JetStream e múltiplas instâncias. Liveness indica
processo vivo; readiness indica snapshot sincronizado. Nenhum secret pode ser versionado.

## Objetivo e implementação

- Criar /healthz para liveness e /readyz para readiness.
- Expor métricas Prometheus de avaliação, latência, revisão, gaps, resync e outbox.
- Usar labels de baixa cardinalidade; nunca usar key como label.
- Criar Dockerfile multi-stage e docker-compose.yml com PostgreSQL, NATS JetStream e serviço.
- Documentar variáveis, portas, migrações, startup e shutdown.
- Adicionar logs estruturados sem payloads sensíveis.

## Testes e aceite

- [ ] docker compose config passa.
- [ ] Health/readiness refletem corretamente o estado do snapshot.
- [ ] Smoke test cria e avalia uma flag.
- [ ] go test ./..., go vet ./... e build da imagem passam.
- [ ] Marcar done, registrar evidências e mover para docs/history/.
