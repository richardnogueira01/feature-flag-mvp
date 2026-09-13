# Task: 006-performance-benchmark

- Projeto: feature-flag-mvp
- Tipo: qa
- Status: done
- Branch: task/feature-flag-mvp-mvp
- Concluída: 2026-09-07

## Resultado

Adicionados benchmark paralelo do evaluator, benchmark HTTP e teste com 1.000.000 de avaliações
concorrentes. O evaluator consulta somente o snapshot em memória.

## Evidências

- Cenário concorrente: 1.000.000 operações, 96 workers, aproximadamente 419.991.600 ops/s.
- p50/p95/p99: 0 ns neste Windows, abaixo da resolução observável do timer; máximo 1.003 ms.
- Evaluator: 62.60 ns/op, 15 B/op, 0 allocs/op.
- HTTP: 63.835 us/op, 22.490 B/op, 123 allocs/op.
- go test -run TestEvaluateMillionConcurrentOperations -v ./internal/snapshot: OK.
- go test -run ^$ -bench Evaluate -benchtime=1000x -benchmem ./internal/snapshot ./internal/httpapi: OK.
- go test ./...: OK.
- go vet ./...: OK.
- go build ./...: OK.

## Limitação

Repetir em Linux ou Windows com mecanismo de medição de maior resolução antes de usar p95/p99 como
SLO. Os números HTTP são baseline local, não garantia de produção.
