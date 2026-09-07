# Task: 005-observability-deployment

- Projeto: feature-flag-mvp
- Tipo: infra
- Status: planned
- Branch: task/feature-flag-mvp-observability-deployment
- Criada: 2026-09-07

## Descrição

Adicionar métricas, health checks, Docker Compose e documentação de execução local.

## Critérios de aceite

- [ ] Métricas de avaliação, revisão e sincronização.
- [ ] Health e readiness distinguem serviço vivo de instância sincronizada.
- [ ] Compose sobe PostgreSQL, NATS e serviço.
- [ ] Configuração não contém secrets.

## Dependências

004-nats-sync-readiness.
