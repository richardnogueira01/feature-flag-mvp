# Task 018 — Last-known-good, disco e failure mode

status: ativo
owner: developer

after: 014-scoped-agent-polling

## Objetivo

Garantir que falhas do Control Plane não interrompam aplicações consumidoras.

## Modelo por flag

```json
{
  "enabled": false,
  "value": true,
  "default_value": false,
  "failure_mode": "LAST_KNOWN"
}
```

Valores permitidos: `LAST_KNOWN`, `FAIL_OPEN`, `FAIL_CLOSED`.

## Fallback

Ordem de inicialização do Agent:

1. snapshot de disco last-known-good;
2. defaults embarcados/configurados;
3. sincronização remota em background.

O snapshot deve ser escrito atomicamente via arquivo temporário + rename e conter revision/hash.

## Aceite

- snapshot corrompido é rejeitado sem apagar o anterior;
- API central indisponível mantém última configuração válida;
- flag sem last-known usa seu failure mode;
- `/readyz` e `/internal/status` informam origem do snapshot e idade;
- testes cobrem restart e indisponibilidade do upstream.