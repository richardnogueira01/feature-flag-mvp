# Task: 007-flexible-flag-values

- Status: done
- Concluída em: 2026-09-07
- Branch: task/feature-flag-mvp-mvp

## Resultado

Flags agora aceitam `value` JSON arbitrário (boolean, número, string, objeto e lista),
mantendo `enabled` como formato legado. O valor é persistido em PostgreSQL `JSONB`,
incluído no snapshot, outbox e avaliação do Data Plane.

## Evidências

- `POST /v1/flags` com `{"key":"timeout_ms","value":1500}` retornou o número.
- `POST /v1/flags` com objeto de configuração retornou o objeto preservando seus tipos.
- Avaliação HTTP retornou `value` e revisão correta.
- `go test ./...`, `go vet ./...`, `go build ./...` e `git diff --check`: OK.
- Compose reconstruído e fluxo PostgreSQL/outbox/NATS validado.

## Segurança

Chaves de API não devem ser armazenadas diretamente; use `{"secret_ref":"..."}`.
