# Task: 001-data-plane-snapshot

- Projeto: feature-flag-mvp
- Tipo: feature
- Status: planned
- Branch: task/feature-flag-mvp-data-plane-snapshot
- Criada: 2026-09-07

## Descrição

Implementar snapshot completo em memória, publicação atômica e avaliação sem I/O externo.

## Critérios de aceite

- [ ] Avaliação consulta somente RAM.
- [ ] Snapshot publicado é imutável para os leitores.
- [ ] Troca de revisão é atômica.
- [ ] Testes concorrentes cobrem publicação e leitura.

## Dependências

Nenhuma.
