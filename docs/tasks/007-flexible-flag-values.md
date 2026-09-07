# Task: 007-flexible-flag-values

- Projeto: feature-flag-mvp
- Tipo: feature
- Status: planned
- Branch: task/feature-flag-mvp-mvp
- Dependências: 005-observability-deployment

## Objetivo

Permitir que uma flag carregue um valor JSON tipado, além do modo booleano legado.
Casos válidos incluem timeout numérico, string, objeto de configuração e listas.

## Contrato da API

### Criação

`POST /v1/flags`

```json
{"key":"checkout_timeout_ms","value":1500}
```

```json
{"key":"checkout_config","value":{"retries":3,"region":"sa-east-1"}}
```

O corpo legado `{"key":"checkout","enabled":true}` continua válido e deve ser
normalizado para `value: true`.

### Atualização

`PUT /v1/flags/{key}` recebe `{"value": ...}` ou, por compatibilidade,
`{"enabled": ...}`. O corpo deve conter exatamente um dos campos.

### Resposta e avaliação

Respostas retornam `key`, `value` e `revision`. `GET /v1/evaluate/{key}` retorna
`{"value": ..., "revision": ...}`. O campo `enabled` pode continuar presente
quando o valor for booleano, para clientes antigos.

## Implementação obrigatória

1. Alterar os modelos de control plane, snapshot e persistência para `json.RawMessage`.
2. Adicionar coluna PostgreSQL `value JSONB`; preservar `enabled` durante a transição
   e garantir migração idempotente em bases existentes.
3. Fazer o outbox e o snapshot NATS transportarem o valor completo, sem converter
   objetos ou números para string.
4. Manter `Evaluate(key)` compatível para testes/telemetria booleana e adicionar
   `EvaluateValue(key)` para valores arbitrários.
5. Atualizar Swagger com schemas `oneOf`/exemplos de boolean, número, string e objeto.
6. Rejeitar JSON nulo, corpo sem `value`/`enabled` e corpo com ambos os campos.
7. Adicionar testes unitários de cada tipo, compatibilidade legada, cópia defensiva,
   CRUD e serialização do outbox; adicionar integração PostgreSQL/Compose.

## Segurança

O MVP não deve armazenar chaves de API ou tokens em texto puro. Para esse caso, o
valor deve ser uma referência, por exemplo `{"secret_ref":"payments/api-key"}`, e a
resolução de secrets fica fora desta task.

## Aceite

- [ ] Booleano legado continua funcionando.
- [ ] String, número, objeto e lista podem ser cadastrados e avaliados.
- [ ] PostgreSQL, outbox e NATS preservam o tipo JSON.
- [ ] Swagger permite preencher todos os exemplos.
- [ ] Testes unitários, integração Compose, `go test ./...`, `go vet ./...` e build passam.
- [ ] Mover este documento para `docs/history/` somente após evidência do aceite.
