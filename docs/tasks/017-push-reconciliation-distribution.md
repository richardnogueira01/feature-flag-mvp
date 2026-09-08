# Task 017 — Distribuição push + reconciliação

status: ativo
owner: developer

## Objetivo

Fazer o NATS/JetStream ser o mecanismo primário de atualização dos Agents e manter polling somente como reconciliação.

## Fluxo

1. Control Plane grava alteração e outbox na mesma transação.
2. Outbox publica evento com `application`, `environment`, `configuration`, `revision` e `snapshot_hash`.
3. Agent inscrito no subject do escopo recebe o evento.
4. Agent baixa o snapshot bulk apenas se a revisão recebida for maior.
5. Evento duplicado ou antigo é ignorado.
6. Gap de revisão dispara full resync imediato.
7. Polling periódico consulta apenas `/version` e baixa `/snapshot` quando necessário.

## Subjects

```text
feature-flags.{application}.{environment}.{configuration}.updated
```

A implementação deve validar nomes e não permitir wildcard vindo do usuário.

## Aceite

- alteração chega ao Agent sem aguardar o intervalo de polling;
- evento duplicado não causa download adicional;
- gap detectado gera métrica e full resync;
- Agent reiniciado recupera estado por snapshot/version;
- polling sem alteração retorna 304 ou resposta de versão sem payload.