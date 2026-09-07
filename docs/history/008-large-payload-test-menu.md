# Task: 008-large-payload-test-menu

- Status: done
- Concluída em: 2026-09-07
- Branch: task/feature-flag-mvp-mvp

## Resultado

Adicionado menu navegável em `/test/large-payload`. Ele gera no navegador um valor
de exatamente 10 MiB (`10.485.760` bytes), envia a flag e mede tamanho, status HTTP
e duração do POST, GET administrativo e avaliação.

O payload é criado em runtime e não é versionado no repositório.

## Evidências

- Endpoint `/test/large-payload`: HTTP 200 no Compose.
- `go test ./...`, `go vet ./...`, `go build ./...` e `git diff --check`: OK.
- Menu usa `TextEncoder` para calcular os bytes reais enviados e `performance.now()`
  para medir latência no navegador.
