# Task: 008-large-payload-test-menu

- Projeto: feature-flag-mvp
- Tipo: test/tooling
- Status: in-progress
- Branch: task/feature-flag-mvp-mvp
- Dependências: 007-flexible-flag-values

## Objetivo

Adicionar um menu acessível pelo navegador para testar uma flag cujo `value` tenha
10 MiB (10.485.760 bytes), medindo tamanho e tempo de POST, GET e avaliação.

## Critérios

- [ ] Payload gerado no navegador, sem arquivo gigante versionado.
- [ ] Exibir tamanho em bytes, status HTTP e duração em ms.
- [ ] Testar criação/atualização, consulta administrativa e avaliação.
- [ ] Não armazenar secrets nem enviar dados externos.
- [ ] Testes Go, vet, build e smoke browser/API passam.
- [ ] Mover para `docs/history/` após validação.
