# Task 012 — Agent local near-cache

status: concluída

## Objetivo

Permitir que aplicações consultem flags localmente, sem polling por requisição e sem dependência do PostgreSQL durante a avaliação.

## Uso

O Compose sobe o Agent na porta `2772`:

```text
GET http://localhost:2772/v1/evaluate/<flag>
```

O Agent consulta `GET /v1/flags` a cada `AGENT_POLL_INTERVAL` (padrão: `15s`), envia `If-None-Match`, publica o snapshot atomicamente em memória e mantém a última versão se o upstream estiver indisponível.

## Variáveis

- `AGENT_UPSTREAM_URL`: API central; padrão `http://app:8080`.
- `AGENT_POLL_INTERVAL`: intervalo Go, por exemplo `15s` ou `30s`.
- `AGENT_ADDR`: porta/endereço local; padrão `2772`.

## Critérios validados

- Binário independente em `cmd/feature-flag-agent`.
- Endpoint local usa o mesmo contrato de evaluate, inclusive payload JSON de 10 MB, `enabled`, gzip e ETag.
- `GET /v1/flags` suporta ETag e retorna `304` quando não há mudança.
- Falha do upstream não apaga o snapshot anterior.
- `go test ./...` e `go vet ./...` devem passar.

## Limitação

Em produção, o Agent deve ser sidecar da aplicação consumidora ou daemon local por nó. O container Compose é uma forma simples de executar o near-cache localmente e não representa alta disponibilidade por si só.