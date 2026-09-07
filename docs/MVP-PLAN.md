# Feature Flag Service — Plano do MVP

## Objetivo

Entregar um serviço Go de Feature Flags que separa Control Plane e Data Plane. O Data Plane deve
avaliar flags somente em memória, enquanto o Control Plane persiste alterações e as distribui.

## Escopo

Incluído: CRUD de flags, revisões monotônicas, histórico, Transactional Outbox, PostgreSQL,
NATS/JetStream, snapshots imutáveis, troca atômica, detecção de gaps, full resync, readiness,
métricas, Docker Compose e benchmark.

Fora do MVP: targeting complexo, A/B testing, analytics, multi-região active-active, SDKs externos,
UI administrativa sofisticada, Redis obrigatório e adapters específicos de cloud.

## Invariantes

1. Avaliação normal não faz I/O externo.
2. PostgreSQL é fonte de verdade, não serving layer.
3. Snapshots publicados são imutáveis.
4. Atualizações trocam snapshots completos atomicamente.
5. Toda configuração possui revisão monotônica.
6. Gaps de revisão são detectáveis e causam resync.
7. Avaliações continuam durante falha do Control Plane.
8. Instância sem snapshot válido não fica pronta.

## Ordem de execução

001 núcleo do snapshot → 002 API do Control Plane → 003 PostgreSQL e Outbox → 004 NATS,
resync e readiness → 005 observabilidade e deployment → 006 performance e benchmark.

Cada task deve ser executada em sua própria branch/worktree. Ao passar pelo quality gate, mover a
spec de docs/tasks para docs/history e manter no backlog somente tarefas planejadas.

## Quality gate

Executar, conforme o projeto evoluir: gofmt, go test ./..., go test -race ./..., go vet ./...,
testes de integração com PostgreSQL/NATS, benchmark paralelo e revisão do diff.
