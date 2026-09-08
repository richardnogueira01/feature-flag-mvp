# Task 014 — Agent near-cache por aplicação e ambiente

status: ativo
owner: developer

after: 013-configuration-scope-model

## Objetivo

Configurar o Agent para sincronizar apenas uma configuração, sem polling por avaliação.

## Variáveis

```env
APP_CONFIG_APPLICATION=atlas
APP_CONFIG_ENVIRONMENT=production
APP_CONFIG_CONFIGURATION=mobile-menu
APP_CONFIG_POLL_INTERVAL=15s
APP_CONFIG_UPSTREAM_URL=http://feature-flag-mvp:8080
```

## Fluxo

1. `GET /v1/.../configurations/{configuration}` com `If-None-Match`.
2. `304`: manter snapshot atual.
3. `200`: validar escopo/revision, publicar snapshot atômico em memória.
4. falha: continuar servindo última versão e expor estado degradado.

O Agent deve servir `GET /v1/evaluate/{key}` localmente e nunca consultar PostgreSQL durante avaliação.

## Métricas obrigatórias

- `feature_flag_agent_poll_total{result}`;
- `feature_flag_agent_snapshot_revision`;
- `feature_flag_agent_snapshot_age_seconds`;
- `feature_flag_agent_poll_duration_seconds`;
- `feature_flag_agent_upstream_errors_total`.

## Aceite

- dois Agents com escopos diferentes não enxergam flags um do outro;
- polling sem alteração retorna 304;
- evaluation continua durante parada do upstream;
- `GET /metrics` expõe as métricas acima;
- teste de carga local mede p95/p99 da avaliação de 10 MB.