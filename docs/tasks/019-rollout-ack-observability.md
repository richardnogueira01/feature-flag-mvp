# Task 019 — Rollout, ACK e observabilidade de frota

status: ativo
owner: architect + devops

## Objetivo

Permitir rollout gradual e tornar visível qual revision cada Agent está usando.

## Deployment

Uma configuração publicada deve suportar estratégia `linear` ou `exponential`, percentual alvo, intervalo de bake e rollback manual/automático.

## ACK

Cada Agent publica heartbeat/ACK contendo:

```json
{
  "agent_id": "atlas-production-01",
  "application": "atlas",
  "environment": "production",
  "configuration": "mobile-menu",
  "revision": 42,
  "snapshot_hash": "sha256:...",
  "observed_at": "..."
}
```

O Control Plane agrega `requested`, `acknowledged`, `stale`, `failed` e percentis de propagação.

## Métricas

- `feature_flag_agent_snapshot_revision`;
- `feature_flag_agent_snapshot_age_seconds`;
- `feature_flag_agent_updates_total{source,result}`;
- `feature_flag_agent_update_duration_seconds`;
- `feature_flag_agent_ack_total`;
- `feature_flag_deployment_ack_ratio`;
- `feature_flag_deployment_propagation_seconds`.

## Aceite

- dashboard mostra revision/hash por escopo;
- rollout não expõe configuração a 100% antes da etapa;
- rollback troca para a última revision válida;
- métricas existem sem depender de Grafana instalado.