# Task: 010-enterprise-chaos-test

- Projeto: feature-flag-mvp
- Tipo: performance/chaos
- Status: in-progress
- Branch: task/feature-flag-mvp-mvp
- Dependências: 009-evaluate-load-1m-rps

## Objetivo

Simular uso empresarial com carga contínua de avaliação e falhas controladas em
dependências, medindo disponibilidade, p95, p99, throughput, erros e recuperação.

## Cenário realista

- O tráfego é distribuído entre múltiplos injetores k6 e instâncias da API.
- O snapshot é aquecido antes da carga.
- `evaluate` é somente leitura e deve continuar respondendo durante falhas de NATS
  e PostgreSQL.
- Alterações de flags são tráfego separado e podem falhar durante indisponibilidade
  do Control Plane, mas devem ser recuperadas por outbox/retry.
- O payload de 10 MiB é comprimido e representa um menu grande; em produção, CDN ou
  cache compartilhado deve evitar retransmiti-lo em toda avaliação.

## Execução local limitada

```powershell
./scripts/run-chaos-test.ps1 -TargetRps 50 -DurationSeconds 120
```

O script aquece a flag, inicia k6, reinicia NATS, reinicia PostgreSQL e recria o app
em fases diferentes. Não remove volumes e não executa `down -v`.

## Execução empresarial

Use `TARGET_RPS=1000000` com pelo menos 10 injetores k6, múltiplas réplicas da API,
telemetria centralizada e distribuição regional. Registre bytes comprimidos e não
comprimidos para evitar confundir requests/s com capacidade de rede.

## Aceite

- [ ] Erros de `evaluate` abaixo de 0,001% durante falhas de NATS/PostgreSQL.
- [ ] p95 e p99 registrados por fase, sem misturar recuperação com steady state.
- [ ] Nenhuma flag ativa retorna conteúdo incorreto após resync.
- [ ] Recuperação de NATS/PostgreSQL confirmada por métricas e status.
- [ ] Execução distribuída concluída antes de arquivar a task.
