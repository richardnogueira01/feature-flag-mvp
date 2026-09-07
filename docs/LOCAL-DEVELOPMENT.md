# Desenvolvimento local

## Pré-requisitos

- Docker com Compose v2.
- Go 1.27 para execução fora do container.

## Execução

1. Defina POSTGRES_PASSWORD fora do repositório.
2. Execute docker compose up --build.
3. A API fica em http://localhost:8080.
4. Liveness: GET /healthz; readiness: GET /readyz; status: GET /internal/status.
5. Métricas: GET /metrics.

## Configuração

- DATABASE_URL: URL PostgreSQL. Se ausente, o serviço usa memória.
- AUTO_MIGRATE: use true para aplicar a migration inicial após conectar ao PostgreSQL.
- NATS_URL: URL NATS. Se ausente, distribuição é desativada.
- NATS_SUBJECT: subject de eventos; padrão feature-flags.events.
- NATS_DURABLE: durable consumer; padrão feature-flag-mvp.

Não versione secrets reais. O Compose exige POSTGRES_PASSWORD no ambiente do processo.

## Validação

Use docker compose config, go test ./..., go vet ./... e go build ./....
