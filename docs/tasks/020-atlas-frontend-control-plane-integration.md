# Task 020 — Integração Atlas e frontend Control Plane

status: ativo
owner: developer + frontend

after: 013-configuration-scope-model, 017-push-reconciliation-distribution

## Objetivo

Integrar o feature-flag service ao Atlas sem expor PostgreSQL ou Control Plane administrativo diretamente ao navegador.

## Backend Atlas

Adicionar cliente interno configurável por `FEATURE_FLAG_CONTROL_PLANE_URL`, com timeout, logs sem secrets e endpoints autenticados do Atlas para o frontend:

```http
GET/POST /api/feature-flags/applications
GET/POST /api/feature-flags/applications/{application}/environments
GET/POST /api/feature-flags/.../configurations
GET/POST/PUT/PATCH/DELETE /api/feature-flags/.../flags
GET /api/feature-flags/.../status
```

O Agent de runtime não deve ser usado para operações administrativas.

## Frontend

Na área Configurações, criar fluxo Application → Environment → Configuration → Flags, com edição JSON, enabled, failure mode, revision e status do Agent. Usar same-origin `/api/...`, nunca URL direta do Control Plane.

## Segurança

Rotas administrativas exigem sessão e autorização do Atlas. O endpoint de avaliação pode permanecer interno. Não commitar URLs reais, tokens ou `.env.local`.

## Aceite

- usuário autenticado cria a hierarquia pelo frontend;
- CRUD de flags funciona com feedback de loading/erro;
- revision e status são exibidos;
- frontend passa lint/typecheck/test;
- Atlas passa build/test/vet;
- integração pode ser desabilitada sem quebrar o Atlas.