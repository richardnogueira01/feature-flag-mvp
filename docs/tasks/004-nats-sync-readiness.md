# Task: 004-nats-sync-readiness

- Projeto: feature-flag-mvp
- Tipo: feature
- Status: planned
- Branch: task/feature-flag-mvp-nats-sync-readiness
- Dependências: 003-postgres-outbox

## Contexto

Instâncias recebem revisões via NATS/JetStream. A próxima revisão esperada é local + 1; revisão
maior indica gap e exige full resync pela fonte de verdade. Snapshot válido continua servindo
durante falhas externas.

## Objetivo e implementação

- Definir evento versionado com id, revision, event_type e payload.
- Subscriber descarta eventos antigos, aplica o próximo e troca snapshot atomicamente.
- Gap marca sincronização degradada e solicita snapshot completo.
- Full resync obtém revisão atual, publica snapshot e limpa degradação.
- GET /internal/status retorna status, revision, snapshot_hash e synchronized.
- Readiness só tem sucesso após primeiro snapshot válido.
- Reconexão do NATS não interrompe avaliações com snapshot antigo.

## Testes e aceite

- [ ] Testar evento sequencial, atrasado e com gap.
- [ ] Testar que gap dispara resync sem concorrência duplicada.
- [ ] Testar instância sem snapshot como não-ready.
- [ ] Testar falha de NATS/PostgreSQL preservando avaliação local.
- [ ] Passar testes de integração e mover spec concluída para docs/history/.
