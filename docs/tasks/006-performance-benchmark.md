# Task: 006-performance-benchmark

- Projeto: feature-flag-mvp
- Tipo: qa
- Status: planned
- Branch: task/feature-flag-mvp-performance-benchmark
- Dependências: 001-data-plane-snapshot e 005-observability-deployment

## Contexto

O requisito central é avaliação rápida sem I/O externo. Medir evaluator puro e HTTP separadamente,
informando CPU, Go, GOMAXPROCS e parâmetros. Percentis usam latências individuais, não médias.

## Objetivo e implementação

- Criar BenchmarkEvaluateParallel usando testing.B.RunParallel.
- Criar cenário com no mínimo 1.000.000 avaliações concorrentes.
- Coletar latência por operação com relógio monotônico.
- Ordenar amostras e calcular p50, p95 e p99 com floor((n-1)*p).
- Registrar total, throughput, máximo, GOMAXPROCS, CPU e tamanho.
- Criar benchmark HTTP separado com servidor local e cliente concorrente.
- Não impor threshold fixo dependente da máquina; documentar baseline e regressões.

## Comandos e aceite

- [ ] go test -run TestEvaluateMillionConcurrentOperations -v ./....
- [ ] go test -bench Evaluate -benchmem -count=5 ./....
- [ ] go test -race ./... passa nos testes funcionais.
- [ ] Relatório inclui p95 e p99 de evaluator e HTTP.
- [ ] Atualizar resultados, marcar done e mover para docs/history/.
