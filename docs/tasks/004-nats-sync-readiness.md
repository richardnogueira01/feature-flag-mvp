# Task: 004-nats-sync-readiness

- Projeto: feature-flag-mvp
- Tipo: feature
- Status: planned
- Branch: task/feature-flag-mvp-nats-sync-readiness
- Criada: 2026-09-07

## Descrição

Distribuir revisões via NATS/JetStream, detectar gaps, executar full resync e controlar readiness.

## Critérios de aceite

- [ ] Atualização válida aplica snapshot atomicamente.
- [ ] Gap de revisão dispara full resync.
- [ ] Nova instância sincroniza antes de receber tráfego.
- [ ] Falha externa não interrompe avaliações com snapshot válido.

## Dependências

003-postgres-outbox.
