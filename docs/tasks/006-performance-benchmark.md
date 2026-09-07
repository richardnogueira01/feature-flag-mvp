# Task: 006-performance-benchmark

- Projeto: feature-flag-mvp
- Tipo: qa
- Status: planned
- Branch: task/feature-flag-mvp-performance-benchmark
- Criada: 2026-09-07

## Descrição

Medir o hot path com milhões de avaliações concorrentes, incluindo throughput, p95 e p99.

## Critérios de aceite

- [ ] Benchmark usa RunParallel.
- [ ] Cenário executa no mínimo 1.000.000 avaliações.
- [ ] Teste registra p50, p95, p99, máximo e throughput.
- [ ] Resultado é reproduzível e identifica CPU, Go e parâmetros.
- [ ] Existe cenário do evaluator e cenário HTTP.

## Dependências

001-data-plane-snapshot e 005-observability-deployment.
