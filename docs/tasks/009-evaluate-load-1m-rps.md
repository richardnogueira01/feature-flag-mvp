# Task: 009-evaluate-load-1m-rps

- Projeto: feature-flag-mvp
- Tipo: performance
- Status: in-progress
- Branch: task/feature-flag-mvp-mvp
- Dependências: 008-large-payload-test-menu

## Objetivo

Executar carga sustentada de 1.000.000 requests/s no endpoint de avaliação de uma
flag com menu de 10 MiB, mantendo erro abaixo de 0,001% e registrando p95/p99.

## Execução

```bash
k6 run -e BASE_URL=http://localhost:8080 -e FLAG_KEY=menu_itau_mobile scripts/evaluate-load.js
```

O script usa `constant-arrival-rate`, aumenta VUs conforme necessário e falha se a
taxa de erros for `>= 0,001%`. O alvo pode ser reduzido com `TARGET_RPS` para validar
localmente antes de usar um ambiente distribuído.

## Limitação física

Uma resposta de 10 MiB a 1 milhão/s exige aproximadamente 10 TB/s de saída, além de
CPU, conexões e memória distribuídas. Portanto, esse cenário não é um teste realista
para um único Docker Desktop; o arquivo é um plano de carga para injetores distribuídos.

## Aceite

- [ ] Executar em infraestrutura distribuída com `TARGET_RPS=1000000`.
- [ ] Erros abaixo de 0,001%.
- [ ] Relatar p95, p99, throughput e bytes transferidos.
- [ ] Mover para `docs/history/` somente após execução em ambiente compatível.
