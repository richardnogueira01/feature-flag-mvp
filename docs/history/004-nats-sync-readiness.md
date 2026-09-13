# Task: 004-nats-sync-readiness

- Status: done
- Concluída em: 2026-09-07
- Branch: task/feature-flag-mvp-mvp

## Resultado

Implementados publisher JetStream, subscriber durable com manual ack, sincronizador
por revisão, detecção de gap com full resync via PostgreSQL e endpoints de readiness.
O stream é criado/configurado de forma idempotente no subscriber.

## Evidências

- Smoke Compose com NATS JetStream real: OK.
- `/readyz`: `READY` após snapshot inicial persistido.
- `/internal/status`: revisão `1`, `synchronized=true`.
- Evento de alteração percorreu outbox → NATS → subscriber sem erro no startup.
- Testes unitários de revisão antiga, gap, resync e status: OK.

## Limitação conhecida

Não foi executado um teste automatizado separado contra NATS externo; a validação real
foi feita no ambiente Compose.
